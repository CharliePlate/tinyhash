package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/bits"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var (
	globalBestHash  [32]byte
	globalBestZeros uint32 // Atomic access
	globalMutex     sync.Mutex
)

type Stats struct {
	StartTime   time.Time
	TotalHashes uint64 // Atomic
}

func (s *Stats) Report() {
	total := atomic.LoadUint64(&s.TotalHashes)
	elapsed := time.Since(s.StartTime).Seconds()
	if elapsed < 0.1 {
		return
	}

	globalMutex.Lock()
	currentHash := globalBestHash
	globalMutex.Unlock()

	hr := float64(total) / elapsed
	fmt.Printf(
		"\rHashrate: %.2f MH/s | Total: %d | Best: %s",
		hr/1_000_000,
		total,
		hex.EncodeToString(currentHash[:]),
	)
}

func main() {
	initBadHash := [32]byte{}
	for i := range initBadHash {
		initBadHash[i] = 0xFF
	}
	globalBestHash = initBadHash
	atomic.StoreUint32(&globalBestZeros, 0)

	numWorkers := runtime.NumCPU()
	stats := Stats{StartTime: time.Now()}

	stopChan := make(chan struct{})

	fmt.Printf("Starting %d workers optimized for zero-allocation mining...\n", numWorkers)

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func(id int) {
			defer wg.Done()
			workerLoop(&stats, stopChan)
		}(i)
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-stopChan:
				return
			case <-ticker.C:
				stats.Report()
			}
		}
	}()

	select {}
}

func workerLoop(stats *Stats, stopChan chan struct{}) {
	seed := make([]byte, 32)
	_, _ = rand.Read(seed)

	var input [40]byte
	copy(input[:32], seed)

	var nonce uint64 = 0

	localBestZeros := atomic.LoadUint32(&globalBestZeros)
	const batchSize = 1000
	counter := 0

	for {
		binary.LittleEndian.PutUint64(input[32:], nonce)
		nonce++

		hash := sha256.Sum256(input[:])

		if hash[0] != 0 {
			counter++
			if counter >= batchSize {
				atomic.AddUint64(&stats.TotalHashes, uint64(counter))
				localBestZeros = atomic.LoadUint32(&globalBestZeros)
				counter = 0

				select {
				case <-stopChan:
					return
				default:
				}
			}
			continue
		}

		zeros := countLeadingZeros(hash)

		if uint32(zeros) > localBestZeros || (uint32(zeros) == localBestZeros && compareHash(hash, globalBestHash)) {
			globalMutex.Lock()
			currentGlobalZeros := atomic.LoadUint32(&globalBestZeros)

			isBetter := false
			if uint32(zeros) > currentGlobalZeros {
				isBetter = true
			} else if uint32(zeros) == currentGlobalZeros {
				if bytes.Compare(hash[:], globalBestHash[:]) < 0 {
					isBetter = true
				}
			}

			if isBetter {
				globalBestHash = hash
				atomic.StoreUint32(&globalBestZeros, uint32(zeros))
				localBestZeros = uint32(zeros)
			}
			globalMutex.Unlock()
		}

		counter++
	}
}

func countLeadingZeros(hash [32]byte) int {
	zeros := 0
	for i := 0; i < 32; i++ {
		if hash[i] == 0 {
			zeros += 8
		} else {
			zeros += bits.LeadingZeros8(hash[i])
			return zeros
		}
	}
	return zeros
}

func compareHash(h1, h2 [32]byte) bool {
	return bytes.Compare(h1[:], h2[:]) < 0
}
