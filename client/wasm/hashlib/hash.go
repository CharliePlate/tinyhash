package hashlib

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"math/bits"
	"sync"
)

var leadingZerosLookup [256]byte

func init() {
	for i := 0; i < 256; i++ {
		leadingZerosLookup[i] = byte(bits.LeadingZeros8(uint8(i)))
	}
}

type SHA256Hash struct {
	InputBytes   []byte
	HashBytes    [32]byte
	LeadingZeros int
}

var sha256Pool = sync.Pool{
	New: func() interface{} {
		return sha256.New()
	},
}

func NewSHA256HashFromString(input string) SHA256Hash {
	return NewSHA256HashFromInput([]byte(input))
}

func NewSHA256HashFromInput(input []byte) SHA256Hash {
	h := sha256Pool.Get().(hash.Hash)
	defer sha256Pool.Put(h)
	h.Reset()
	h.Write(input)
	var hash [32]byte
	h.Sum(hash[:0])
	l := CountLeadingZeros(hash[:])
	return SHA256Hash{
		InputBytes:   input,
		HashBytes:    hash,
		LeadingZeros: l,
	}
}

func NewSHA256Hash(input []byte, hash [32]byte) SHA256Hash {
	l := CountLeadingZeros(hash[:])
	return SHA256Hash{
		InputBytes:   input,
		HashBytes:    hash,
		LeadingZeros: l,
	}
}

func (h *SHA256Hash) Input() string {
	return string(h.InputBytes)
}

func (h *SHA256Hash) Hash() string {
	return hex.EncodeToString(h.HashBytes[:])
}

func (h *SHA256Hash) IsLessThan(comp SHA256Hash) bool {
	if h.LeadingZeros < comp.LeadingZeros {
		return false
	}
	if h.LeadingZeros > comp.LeadingZeros {
		return true
	}
	return bytes.Compare(h.HashBytes[:], comp.HashBytes[:]) < 0
}

func CountLeadingZeros(hash []byte) int {
	totalZeros := 0
	for _, b := range hash {
		if b == 0 {
			totalZeros += 8
		} else {
			return totalZeros + int(leadingZerosLookup[b])
		}
	}
	return totalZeros
}
