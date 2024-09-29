package pem

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

const (
	PrivateKeyBlockType = "RSA PRIVATE KEY"
	PublicKeyBlockType  = "RSA PUBLIC KEY"
)

func SaveRSAPrivateKey(key *rsa.PrivateKey, file string) error {
	var privateKeyPEM bytes.Buffer
	pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  PrivateKeyBlockType,
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	if err := os.WriteFile(file, privateKeyPEM.Bytes(), 0666); err != nil {
		return err
	}

	return nil
}

func ExtractRSAPrivateKey(file string) (*rsa.PrivateKey, error) {
	pemData, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != PrivateKeyBlockType {
		return nil, fmt.Errorf("PEM file does not contain %s", PrivateKeyBlockType)
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func SaveRSAPublicKey(key *rsa.PublicKey, file string) error {
	var publicKeyPEM bytes.Buffer
	pem.Encode(&publicKeyPEM, &pem.Block{
		Type:  PublicKeyBlockType,
		Bytes: x509.MarshalPKCS1PublicKey(key),
	})

	if err := os.WriteFile(file, publicKeyPEM.Bytes(), 0666); err != nil {
		return err
	}

	return nil
}

func ExtractRSAPublicKey(file string) (*rsa.PublicKey, error) {
	pemData, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != PublicKeyBlockType {
		return nil, fmt.Errorf("PEM file does not contain %s", PublicKeyBlockType)
	}

	key, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return key, nil
}
