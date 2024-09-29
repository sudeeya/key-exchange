package api

type Request struct {
	Initiator  string `json:"initiator"`
	Acceptor   string `json:"acceptor"`
	Ciphertext []byte `json:"ciphertext"`
}

type Response struct {
	Certificate
	Ciphertext []byte `json:"ciphertext"`
}

type Certificate struct {
	Info
	Signature []byte `json:"signature"`
}

type Info struct {
	Initiator      string `json:"initiator"`
	Acceptor       string `json:"acceptor"`
	InitiatorNonce []byte `json:"initiator_nonce"`
	InitiatorKey   []byte `json:"initiator_key"`
	AcceptorKey    []byte `json:"acceptor_key"`
	SessionKey     []byte `json:"session_key"`
}
