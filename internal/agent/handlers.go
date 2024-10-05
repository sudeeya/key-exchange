package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sudeeya/key-exchange/internal/pkg/api"
	"github.com/sudeeya/key-exchange/internal/pkg/crypto"
)

func step4Handler(a *Agent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Step 4
		var req api.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		info4JSON := crypto.DecryptRSA(req.Ciphertext, a.keys.privateKey)

		info4 := api.Info{}
		err := json.Unmarshal(info4JSON, &info4)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		initiator := info4.Initiator

		ciphertext4 := crypto.EncryptRSA(info4.InitiatorNonce, a.keys.trentKey)
		req4 := api.Request{
			Initiator:  initiator,
			Acceptor:   a.cfg.ID,
			Ciphertext: ciphertext4,
		}
		var resp5 api.Response
		rawResp5, err := a.client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(req4).
			SetResult(&resp5).
			Post(httpPrefix + a.cfg.TrentAddr + api.Step5Endpoint)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if rawResp5.StatusCode() != http.StatusOK {
			http.Error(w, fmt.Sprintf("step 5 status code is %d", rawResp5.StatusCode()), http.StatusInternalServerError)
			return
		}

		info5JSON, err := json.Marshal(resp5.Certificate.Information)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ok := crypto.VerifyRSA(info5JSON, resp5.Certificate.Signature, a.keys.trentKey)
		if !ok {
			http.Error(w, "signature verification failed", http.StatusInternalServerError)
			return
		}

		a.keys.agentKey = resp5.Certificate.Information.InitiatorKey

		cert5JSON := crypto.DecryptRSA(resp5.Ciphertext, a.keys.privateKey)
		var cert5 api.Cert
		if err = json.Unmarshal(cert5JSON, &cert5); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		certInfo5JSON, err := json.Marshal(cert5.Information)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ok = crypto.VerifyRSA(certInfo5JSON, cert5.Signature, a.keys.trentKey)
		if !ok {
			http.Error(w, "signature verification failed", http.StatusInternalServerError)
			return
		}

		a.keys.sessionKey = cert5.Information.SessionKey

		// Step 6
		acceptorNonce, err := a.rng.GenerateNonce()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		a.keys.nonce = acceptorNonce

		resp6 := api.Response{
			Certificate:   cert5,
			AcceptorNonce: acceptorNonce,
		}
		resp6JSON, err := json.Marshal(resp6)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ciphertext6 := crypto.EncryptRSA(resp6JSON, a.keys.agentKey)

		resp7 := api.Response{
			Ciphertext: ciphertext6,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp7); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// Step 7
func step7Handler(a *Agent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var msg api.Message
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		acceptorNonce := crypto.DecryptAES(msg.Ciphertext, a.keys.sessionKey, msg.IV)

		if !bytes.Equal(acceptorNonce, a.keys.nonce) {
			http.Error(w, "nonce verification failed", http.StatusBadRequest)
			return
		}

		a.tui.session = a.cfg.AgentID
		a.tui.active[writeMessageItem] = struct{}{}

		w.WriteHeader(http.StatusOK)
	}
}

func messageHandler(a *Agent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var msg api.Message
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		message := crypto.DecryptAES(msg.Ciphertext, a.keys.sessionKey, msg.IV)
		a.tui.messages = append(a.tui.messages, string(message))
		a.tui.unread = true

		w.WriteHeader(http.StatusOK)
	}
}
