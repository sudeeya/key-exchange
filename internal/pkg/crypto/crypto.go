package crypto

import (
	"github.com/golang-module/dongle"
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

func VerifyRSA(message, signature, publicKey []byte) bool {
	ok := dongle.Verify.
		FromRawBytes(signature, message).
		ByRsa(publicKey, dongle.SHA256).
		ToBool()
	return ok
}
