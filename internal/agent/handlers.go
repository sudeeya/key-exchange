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
		// Step 3
		var req api.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		infoJSON3 := crypto.DecryptRSA(req.Ciphertext, a.keys.privateKey)

		info3 := api.Info{}
		err := json.Unmarshal(infoJSON3, &info3)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		initiator := info3.Initiator

		// Step 4
		ciphertext4 := crypto.EncryptRSA(info3.InitiatorNonce, a.keys.trentKey)
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

		infoJSON5, err := json.Marshal(resp5.Certificate.Information)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ok := crypto.VerifyRSA(infoJSON5, resp5.Certificate.Signature, a.keys.trentKey)
		if !ok {
			http.Error(w, "signature verification failed", http.StatusInternalServerError)
			return
		}

		a.keys.agentKey = resp5.Certificate.Information.InitiatorKey

		certJSON5 := crypto.DecryptRSA(resp5.Ciphertext, a.keys.privateKey)
		var cert5 api.Cert
		if err = json.Unmarshal(certJSON5, &cert5); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		certInfoJSON5, err := json.Marshal(cert5.Information)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ok = crypto.VerifyRSA(certInfoJSON5, cert5.Signature, a.keys.trentKey)
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
		respJSON6, err := json.Marshal(resp6)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ciphertext6 := crypto.EncryptRSA(respJSON6, a.keys.agentKey)

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

		a.msgCh <- SessionEstablishedMsg(a.cfg.AgentID)

		w.WriteHeader(http.StatusOK)
	}
}
