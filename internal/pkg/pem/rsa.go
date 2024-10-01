package pem

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
)

const (
	PrivateKeyBlockType = "RSA PRIVATE KEY"
	PublicKeyBlockType  = "RSA PUBLIC KEY"
)

func EncodeRSAPrivateKey(key *rsa.PrivateKey) []byte {
	var privateKeyPEM bytes.Buffer
	pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  PrivateKeyBlockType,
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	return privateKeyPEM.Bytes()
}

func SaveRSAPrivateKey(key *rsa.PrivateKey, file string) error {
	if err := os.WriteFile(file, EncodeRSAPrivateKey(key), 0666); err != nil {
		return err
	}

	return nil
}

func ExtractRSAPrivateKey(file string) ([]byte, error) {
	key, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func EncodeRSAPublicKey(key *rsa.PublicKey) []byte {
	var publicKeyPEM bytes.Buffer
	pem.Encode(&publicKeyPEM, &pem.Block{
		Type:  PublicKeyBlockType,
		Bytes: x509.MarshalPKCS1PublicKey(key),
	})

	return publicKeyPEM.Bytes()
}

func SaveRSAPublicKey(key *rsa.PublicKey, file string) error {
	if err := os.WriteFile(file, EncodeRSAPublicKey(key), 0666); err != nil {
		return err
	}

	return nil
}

func ExtractRSAPublicKey(file string) ([]byte, error) {
	key, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return key, nil
}
