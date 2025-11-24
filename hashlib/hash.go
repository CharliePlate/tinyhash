package hashlib

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"math/bits"
	"sync"

	byteslice "github.com/charlieplate/TinyHash/client/wasm/slice"
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
	return NewSHA256HashFromInput(byteslice.String2ByteSlice(input, 64))
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

func (h *SHA256Hash) Input() string {
	return byteslice.ByteSlice2String(h.InputBytes, 64)
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
