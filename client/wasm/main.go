package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"math/bits"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

var (
	globalCurrMin SHA256Hash
	globalMutex   sync.Mutex
)

type Stats struct {
	StartTime   time.Time
	TotalHashes uint64
}

func (s *Stats) Report() {
	total := atomic.LoadUint64(&s.TotalHashes)
	elapsed := time.Since(s.StartTime).Seconds()
	if elapsed < 0.1 {
		return
	}
	hr := float64(total) / elapsed
	fmt.Printf(
		"Hashrate: %.2f MH/s | Total Hashes: %d | Current Min: %s\n",
		hr/1000000,
		total,
		globalCurrMin.Hash(),
	)
}

type HashInstance struct {
	cancelChan chan struct{}
}

func NewHashInstance() *HashInstance {
	return &HashInstance{
		cancelChan: make(chan struct{}),
	}
}

func (hi *HashInstance) HashLoop(stats *Stats) {
	hi.cancelChan = make(chan struct{})

	src := rand.New(rand.NewSource(time.Now().UnixNano()))

	go func() {
		for {
			select {
			case <-hi.cancelChan:
				return
			default:
				atomic.AddUint64(&stats.TotalHashes, 1)

				h := NewSHA256HashFromString(
					RandStringBytesMaskImprSrcUnsafe(src, 32),
				)

				if h.IsLessThan(globalCurrMin) {
					globalMutex.Lock()
					if h.IsLessThan(globalCurrMin) {
						globalCurrMin = h
					}
					globalMutex.Unlock()
				}
			}
		}
	}()
}

func (hi *HashInstance) CancelLoop() {
	if hi.cancelChan != nil {
		close(hi.cancelChan)
		hi.cancelChan = nil
	}
}

func main() {
	maxHashBytes := bytes.Repeat([]byte{0xFF}, 32)
	var hashBytes [32]byte
	copy(hashBytes[:], maxHashBytes)
	globalCurrMin = NewSHA256Hash(nil, hashBytes)

	stopper := make(chan struct{})
	hi := NewHashInstance()
	stats := Stats{StartTime: time.Now()}

	numWorkers := runtime.NumCPU()
	fmt.Printf("Starting %d hashing workers...\n", numWorkers)
	for range numWorkers {
		hi.HashLoop(&stats)
	}

	go func() {
		for {
			select {
			case <-stopper:
				return
			default:
				stats.Report()
				time.Sleep(time.Second)
			}
		}
	}()

	<-stopper
}

// SHA256Hash struct remains an efficient way to hold hash data.
type SHA256Hash struct {
	InputBytes   []byte
	HashBytes    [32]byte
	LeadingZeros int
}

// sync.Pool is an excellent optimization for reducing GC pressure.
var sha256Pool = sync.Pool{
	New: func() interface{} {
		return sha256.New()
	},
}

func NewSHA256HashFromString(input string) SHA256Hash {
	return NewSHA256HashFromInput(unsafeStringToBytes(input))
}

func NewSHA256HashFromInput(input []byte) SHA256Hash {
	h := sha256Pool.Get().(hash.Hash)
	h.Reset()
	h.Write(input)
	var hash [32]byte
	h.Sum(hash[:0])
	sha256Pool.Put(h)
	return NewSHA256Hash(input, hash)
}

func NewSHA256Hash(input []byte, hash [32]byte) SHA256Hash {
	l := CountLeadingZeros(hash[:])
	return SHA256Hash{
		InputBytes:   input,
		HashBytes:    hash,
		LeadingZeros: l,
	}
}

func (h *SHA256Hash) Hash() string {
	return hex.EncodeToString(h.HashBytes[:])
}

func (h *SHA256Hash) IsLessThan(comp SHA256Hash) bool {
	if h.LeadingZeros > comp.LeadingZeros {
		return true
	}
	if h.LeadingZeros < comp.LeadingZeros {
		return false
	}
	return bytes.Compare(h.HashBytes[:], comp.HashBytes[:]) < 0
}

// The lookup table is the fastest way to do this. Great implementation.
var leadingZerosLookup [256]byte

func init() {
	for i := 0; i < 256; i++ {
		leadingZerosLookup[i] = byte(bits.LeadingZeros8(uint8(i)))
	}
}

func CountLeadingZeros(hash []byte) int {
	totalZeros := 0
	for _, b := range hash {
		zeros := int(leadingZerosLookup[b])
		totalZeros += zeros
		if zeros < 8 {
			break // No need to check further bytes.
		}
	}
	return totalZeros
}

// --- Unsafe String/Byte Generation (Unchanged) ---
// This is a known fast method for generating random strings.
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

const (
	letterIdxBits = 6
	letterIdxMask = 1<<letterIdxBits - 1
	letterIdxMax  = 63 / letterIdxBits
)

func RandStringBytesMaskImprSrcUnsafe(src *rand.Rand, n int) string {
	b := make([]byte, n)
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

func unsafeStringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}
