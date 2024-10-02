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

func EncryptAES(plaintext, key, iv []byte) []byte {
	cipher := dongle.NewCipher()
	cipher.SetMode(dongle.OFB)
	cipher.SetPadding(dongle.PKCS7)
	cipher.SetKey(key)
	cipher.SetIV(iv)

	ciphertext := dongle.Encrypt.
		FromBytes(plaintext).
		ByAes(cipher).
		ToRawBytes()

	return ciphertext
}

func DecryptAES(ciphertext, key, iv []byte) []byte {
	cipher := dongle.NewCipher()
	cipher.SetMode(dongle.OFB)
	cipher.SetPadding(dongle.PKCS7)
	cipher.SetKey(key)
	cipher.SetIV(iv)

	plaintext := dongle.Decrypt.
		FromRawBytes(ciphertext).
		ByAes(cipher).
		ToBytes()

	return plaintext
}
