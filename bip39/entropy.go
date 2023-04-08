package bip39

import (
	"crypto/rand"
	"sync"

	"github.com/holiman/uint256"
	"github.com/pkg/errors"
)

var _32 = uint256.NewInt(32)

// NewEntropy will create random entropy bytes
// so long as the requested size bitSize is an appropriate size.
//
// bitSize has to be a multiple 32 and be within the inclusive range of {128, 256}.
func NewEntropy(bitSize int) ([]byte, error) {
	if err := validateEntropyBitSize(bitSize); err != nil {
		return nil, errors.WithStack(err)
	}

	entropy := make([]byte, bitSize/8)
	if _, err := rand.Read(entropy); err != nil {
		return nil, errors.WithStack(err)
	}

	return entropy, nil
}

// FastEntropy is a fast entropy generator. It use cumulative entropy instead of generating
// new random entropy. It will generate new random entropy every 2048 wallets.
type FastEntropy struct {
	bitSize       int
	mu            sync.Mutex
	count         uint64
	maxCumulative uint64
	entropyInt    *uint256.Int
}

// NewFastEntropy will create a new FastEntropy.
func NewFastEntropy(bitSize int, maxCumulative ...uint64) (*FastEntropy, error) {
	entropy, err := NewEntropy(bitSize)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// 2^11 | 2048
	max := shift11BitsMask.Uint64()
	if len(maxCumulative) > 0 {
		max = maxCumulative[0]
	}

	return &FastEntropy{
		bitSize:       bitSize,
		maxCumulative: max,
		entropyInt:    new(uint256.Int).SetBytes(entropy),
	}, nil
}

// Next will return the next entropy.
func (f *FastEntropy) Next() ([]byte, error) {
	f.mu.Lock()
	defer func() {
		f.count++
		f.entropyInt.Add(f.entropyInt, one)
		f.mu.Unlock()
	}()

	entropy := f.entropyInt.Bytes()
	entropyBitLength := len(entropy) * 8
	if err := validateEntropyBitSize(entropyBitLength); err != nil || f.count >= f.maxCumulative {
		newEntropy, err := NewEntropy(f.bitSize)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		f.count = 0
		entropy = newEntropy
		f.entropyInt.SetBytes(newEntropy)
	}

	return entropy, nil
}
