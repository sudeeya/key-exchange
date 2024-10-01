package trent

import (
	"encoding/json"
	"net/http"

	"github.com/golang-module/dongle"
	"github.com/sudeeya/key-exchange/internal/pkg/api"
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

		info := api.Info{
			AcceptorKey: acceptorKey,
		}
		infoJSON, err := json.Marshal(info)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		signature := dongle.Sign.
			FromBytes(infoJSON).
			ByRsa(t.privateKey, dongle.SHA256).
			ToRawBytes()

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
		info := api.Info{
			InitiatorKey: initiatorKey,
		}
		infoJSON, err := json.Marshal(info)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		signature := dongle.Sign.
			FromBytes(infoJSON).
			ByRsa(t.privateKey, dongle.SHA256).
			ToRawBytes()

		initiatorNonce := dongle.Decrypt.
			FromRawBytes(req.Ciphertext).
			ByRsa(t.privateKey).
			ToBytes()

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

		signatureToEncrypt := dongle.Sign.
			FromBytes(infoToEncryptJSON).
			ByRsa(t.privateKey, dongle.SHA256).
			ToRawBytes()

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
		ciphertext := dongle.Encrypt.
			FromBytes(certToEncryptJSON).
			ByRsa(acceptorKey).
			ToRawBytes()

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
