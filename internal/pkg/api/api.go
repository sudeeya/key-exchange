package api

const (
	Step2Endpoint   = "/step2/"
	Step4Endpoint   = "/step4/"
	Step5Endpoint   = "/step5/"
	Step7Endpoint   = "/step7/"
	MessageEndpoint = "/msg/"
)

type Request struct {
	Initiator  string `json:"initiator,omitempty"`
	Acceptor   string `json:"acceptor,omitempty"`
	Ciphertext []byte `json:"ciphertext,omitempty"`
}

type Response struct {
	Certificate   Cert   `json:"certificate,omitempty"`
	Ciphertext    []byte `json:"ciphertext,omitempty"`
	AcceptorNonce []byte `json:"acceptor_nonce,omitempty"`
}

type Cert struct {
	Information Info   `json:"info"`
	Signature   []byte `json:"signature"`
}

type Info struct {
	Initiator      string `json:"initiator,omitempty"`
	Acceptor       string `json:"acceptor,omitempty"`
	InitiatorNonce []byte `json:"initiator_nonce,omitempty"`
	InitiatorKey   []byte `json:"initiator_key,omitempty"`
	AcceptorKey    []byte `json:"acceptor_key,omitempty"`
	SessionKey     []byte `json:"session_key,omitempty"`
}

type Message struct {
	IV         []byte `json:"iv"`
	Ciphertext []byte `json:"ciphertext"`
}
