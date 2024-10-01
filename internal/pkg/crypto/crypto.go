package crypto

import (
	"encoding/json"

	"github.com/golang-module/dongle"
	"github.com/sudeeya/key-exchange/internal/pkg/api"
)

func EncryptRSA(plaintext, publicKey []byte) []byte {
	ciphertext := dongle.Encrypt.
		FromBytes(plaintext).
		ByRsa(publicKey).
		ToRawBytes()
	return ciphertext
}

func DecryptRSA(ciphertext, privateKey []byte) []byte {
	plaintext := dongle.Decrypt.
		FromRawBytes(ciphertext).
		ByRsa(privateKey).
		ToBytes()
	return plaintext
}

func SignRSA(message, privateKey []byte) []byte {
	signature := dongle.Sign.
		FromBytes(message).
		ByRsa(privateKey, dongle.SHA256).
		ToRawBytes()
	return signature
}

func VerifyCertRSA(cert api.Cert, publicKey []byte) (bool, error) {
	infoJSON, err := json.Marshal(cert.Information)
	if err != nil {
		return false, nil
	}
	ok := dongle.Verify.
		FromRawBytes(cert.Signature, infoJSON).
		ByRsa(publicKey, dongle.SHA256).
		ToBool()
	return ok, nil
}
