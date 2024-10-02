package api

const (
	Step2Endpoint = "/step2/"
	Step4Endpoint = "/step4/"
	Step5Endpoint = "/step5/"
	Step7Endpoint = "/step7/"
)

type Request struct {
	Initiator  string `json:"initiator"`
	Acceptor   string `json:"acceptor"`
	Ciphertext []byte `json:"ciphertext"`
}

type Response struct {
	Certificate   Cert   `json:"certificate"`
	Ciphertext    []byte `json:"ciphertext"`
	AcceptorNonce []byte `json:"acceptor_nonce"`
}

type Cert struct {
	Information Info   `json:"info"`
	Signature   []byte `json:"signature"`
}

type Info struct {
	Initiator      string `json:"initiator"`
	Acceptor       string `json:"acceptor"`
	InitiatorNonce []byte `json:"initiator_nonce"`
	InitiatorKey   []byte `json:"initiator_key"`
	AcceptorKey    []byte `json:"acceptor_key"`
	SessionKey     []byte `json:"session_key"`
}

type Message struct {
	IV         []byte `json:"iv"`
	Ciphertext []byte `json:"ciphertext"`
}
