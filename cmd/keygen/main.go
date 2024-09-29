package main

import (
	"crypto/rand"
	"crypto/rsa"
	"flag"
	"log"

	"github.com/sudeeya/key-exchange/internal/pkg/pem"
)

const keySize = 2048

func main() {
	privatePath := flag.String("private", "private.pem", "Path to the file that will store the private key")
	publicPath := flag.String("public", "public.pem", "Path to the file that will store the public key")

	flag.Parse()

	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		log.Fatal(err)
	}

	if err := pem.SaveRSAPrivateKey(privateKey, *privatePath); err != nil {
		log.Fatal(err)
	}
	if err := pem.SaveRSAPublicKey(&privateKey.PublicKey, *publicPath); err != nil {
		log.Fatal(err)
	}
}
