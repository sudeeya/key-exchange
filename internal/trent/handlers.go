package trent

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"net/http"

	"github.com/sudeeya/key-exchange/internal/pkg/api"
	"github.com/sudeeya/key-exchange/internal/pkg/pem"
	"github.com/sudeeya/key-exchange/internal/pkg/rng"
)

func newInitiateHandler(t *Trent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req api.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		acceptorKey := t.clientsList[req.Acceptor].PublicKey
		acceptorKeyPEM := pem.EncodeRSAPublicKey(acceptorKey)

		info := api.Info{
			AcceptorKey: acceptorKeyPEM,
		}
		infoJSON, err := json.Marshal(info)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		hash := sha256.New()
		if _, err := hash.Write(infoJSON); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		infoJSONHash := hash.Sum(nil)

		signature, err := rsa.SignPKCS1v15(rand.Reader, t.privateKey, crypto.SHA256, infoJSONHash)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := api.Response{
			Certificate: api.Certificate{
				Info:      info,
				Signature: signature,
			},
		}

		w.Header().Set("Content-Tyoe", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func newConfirmHandler(t *Trent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req api.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		initiatorKey := t.clientsList[req.Initiator].PublicKey
		initiatorKeyPEM := pem.EncodeRSAPublicKey(initiatorKey)
		info := api.Info{
			InitiatorKey: initiatorKeyPEM,
		}
		infoJSON, err := json.Marshal(info)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		hash := sha256.New()
		if _, err := hash.Write(infoJSON); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		infoJSONHash := hash.Sum(nil)

		signature, err := rsa.SignPKCS1v15(rand.Reader, t.privateKey, crypto.SHA256, infoJSONHash)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		initiatorNonce, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, t.privateKey, req.Ciphertext, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		sessionKey, err := t.rng.GenerateKey(rng.KuznyechikKeySize)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		infoToEncrypt := api.Info{
			InitiatorNonce: initiatorNonce,
			SessionKey:     sessionKey,
			Initiator:      req.Initiator,
			Acceptor:       req.Acceptor,
		}
		infoToEncryptJSON, err := json.Marshal(infoToEncrypt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		hash.Reset()
		if _, err := hash.Write(infoToEncryptJSON); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		infoToEncryptJSONHash := hash.Sum(nil)

		signatureToEncrypt, err := rsa.SignPKCS1v15(rand.Reader, t.privateKey, crypto.SHA256, infoToEncryptJSONHash)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		certToEncrypt := api.Certificate{
			Info:      infoToEncrypt,
			Signature: signatureToEncrypt,
		}
		certToEncryptJSON, err := json.Marshal(certToEncrypt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		acceptorKey := t.clientsList[req.Acceptor].PublicKey
		ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, acceptorKey, certToEncryptJSON, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := api.Response{
			Certificate: api.Certificate{
				Info:      info,
				Signature: signature,
			},
			Ciphertext: ciphertext,
		}

		w.Header().Set("Content-Tyoe", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
