//go:build js

package wasm

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"syscall/js"
	"time"
	"unsafe"

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
	hi.cancelChan = make(chan struct{})

	src := rand.New(rand.NewSource(time.Now().UnixNano()))

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
				h := hashlib.NewSHA256HashFromString(RandStringBytesMaskImprSrcUnsafe(src, 32))

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
	js.Global().Call("updateStats", s.TotalHashes, fmt.Sprintf("%.2f", hr))
}

// fast way to calcualte random bytes: https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go/31832326#31832326
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

const (
	letterIdxBits = 6
	letterIdxMask = 1<<letterIdxBits - 1
	letterIdxMax  = 63 / letterIdxBits
)

func RandStringBytesMaskImprSrcUnsafe(src *rand.Rand, n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}
