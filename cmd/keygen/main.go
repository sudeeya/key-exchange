package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"log"
	"os"
)

const keySize = 2048

func main() {
	privatePath := flag.String("private", "private.pem", "Path to the file that will store the private key")
	publicPath := flag.String("public", "public.pem", "Path to the file that will store the public key")

	flag.Parse()

	privateKey, publicKey := generateKeyPair()

	savePrivateKey(privateKey, *privatePath)
	savePublicKey(publicKey, *publicPath)
}

func generateKeyPair() (*rsa.PrivateKey, *rsa.PublicKey) {
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		log.Fatal(err)
	}

	return privateKey, &privateKey.PublicKey
}

func savePrivateKey(key *rsa.PrivateKey, path string) {
	var privateKeyPEM bytes.Buffer
	pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	if err := os.WriteFile(path, privateKeyPEM.Bytes(), 0666); err != nil {
		log.Fatal(err)
	}
}

func savePublicKey(key *rsa.PublicKey, path string) {
	var publicKeyPEM bytes.Buffer
	pem.Encode(&publicKeyPEM, &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(key),
	})

	if err := os.WriteFile(path, publicKeyPEM.Bytes(), 0666); err != nil {
		log.Fatal(err)
	}
}
