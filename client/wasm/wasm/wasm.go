//go:build js

package wasm

import (
	"encoding/base64"
	"math/rand"
	"runtime"
	"sync"
	"syscall/js"
	"time"

	"github.com/charlieplate/TinyHash/client/wasm/hashlib"
)

func HashSHA256(this js.Value, args []js.Value) interface{} {
	input := args[0].String()
	h := hashlib.NewSHA256HashFromString(input)
	return h.Hash()
}

func CountLeadingZeros(this js.Value, args []js.Value) interface{} {
	h := args[0].String()
	z := hashlib.CountLeadingZeros([]byte(h))
	return z
}

var (
	globalMutex   sync.Mutex
	globalCurrMin hashlib.SHA256Hash
)

type HashInstance struct {
	cancelChan     chan struct{}
	hashCountMutex sync.Mutex
}

func NewHashInstance() *HashInstance {
	return &HashInstance{
		cancelChan: make(chan struct{}),
	}
}

func (hi *HashInstance) HashLoop(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return "Error: Expected maxHPS argument"
	}
	maxMHPS := args[0].Int() * 1_000_000
	if maxMHPS <= 0 {
		return "Error: maxHPS must be positive"
	}
	hi.cancelChan = make(chan struct{})
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))

	go func() {
		stats := Stats{}
		stats.StartTime = time.Now()
		secondTicker := time.NewTicker(time.Second)
		defer secondTicker.Stop()

		for {
			select {
			case <-hi.cancelChan:
				return
			case <-secondTicker.C:
				stats.Report()
			default:
				stats.TotalHashes++
				rand := make([]byte, 32)
				seed.Read(rand)
				h := hashlib.NewSHA256HashFromString(base64.StdEncoding.EncodeToString(rand))

				isNewMin := h.IsLessThan(globalCurrMin)

				if isNewMin {
					globalMutex.Lock()
					if h.IsLessThan(globalCurrMin) {
						globalCurrMin = h
						js.Global().Call("updateHash", globalCurrMin.Hash(), globalCurrMin.Input())
					}
					globalMutex.Unlock()
				}
				if (stats.TotalHashes % 10000) == 0 {
					runtime.Gosched()
				}
			}
		}
	}()
	return nil
}

func (hi *HashInstance) CancelLoop(this js.Value, args []js.Value) interface{} {
	if hi.cancelChan != nil {
		close(hi.cancelChan)
		hi.cancelChan = nil
	}
	return nil
}

func SetCurrentMin(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return "Error: Expected hash argument"
	}
	h := args[0].String()
	newMin := hashlib.NewSHA256HashFromInput([]byte(h))

	globalMutex.Lock()
	defer globalMutex.Unlock()

	if newMin.IsLessThan(globalCurrMin) || globalCurrMin.Hash() == "" {
		globalCurrMin = newMin
	}
	return nil
}

type Stats struct {
	StartTime   time.Time
	TotalHashes uint64
}

func (s Stats) Report() {
	elapsed := time.Since(s.StartTime).Seconds()
	hr := float64(s.TotalHashes) / elapsed
	js.Global().Call("updateStats", s.TotalHashes, int(hr))
}
