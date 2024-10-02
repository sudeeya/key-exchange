package trent

import (
	"encoding/json"
	"net/http"

	"github.com/sudeeya/key-exchange/internal/pkg/api"
	"github.com/sudeeya/key-exchange/internal/pkg/crypto"
	"github.com/sudeeya/key-exchange/internal/pkg/rng"
)

// Step 2
func step2Handler(t *Trent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req api.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		acceptorKey := t.agentList[req.Acceptor].PublicKey

		info := api.Info{
			AcceptorKey: acceptorKey,
		}
		infoJSON, err := json.Marshal(info)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		signature := crypto.SignRSA(infoJSON, t.privateKey)

		resp := api.Response{
			Certificate: api.Cert{
				Information: info,
				Signature:   signature,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// Step 5
func step5Handler(t *Trent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req api.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		initiatorKey := t.agentList[req.Initiator].PublicKey
		info := api.Info{
			InitiatorKey: initiatorKey,
		}
		infoJSON, err := json.Marshal(info)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		signature := crypto.SignRSA(infoJSON, t.privateKey)

		initiatorNonce := crypto.DecryptRSA(req.Ciphertext, t.privateKey)

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

		signatureToEncrypt := crypto.SignRSA(infoToEncryptJSON, t.privateKey)

		certToEncrypt := api.Cert{
			Information: infoToEncrypt,
			Signature:   signatureToEncrypt,
		}
		certToEncryptJSON, err := json.Marshal(certToEncrypt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		acceptorKey := t.agentList[req.Acceptor].PublicKey
		ciphertext := crypto.EncryptRSA(certToEncryptJSON, acceptorKey)

		resp := api.Response{
			Certificate: api.Cert{
				Information: info,
				Signature:   signature,
			},
			Ciphertext: ciphertext,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
