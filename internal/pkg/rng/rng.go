package rng

import (
	"crypto/rand"
)

const (
	KuznyechikKeySize = 32
	NonceSize         = 16
	IVSize            = 16
)

type RNG struct{}

func NewRNG() *RNG {
	return &RNG{}
}

func (rng RNG) GenerateNonce() ([]byte, error) {
	key := make([]byte, NonceSize)
	if _, err := rand.Reader.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

func (rng RNG) GenerateIV() ([]byte, error) {
	key := make([]byte, IVSize)
	if _, err := rand.Reader.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

func (rng RNG) GenerateKey(bytes int) ([]byte, error) {
	key := make([]byte, bytes)
	if _, err := rand.Reader.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}
