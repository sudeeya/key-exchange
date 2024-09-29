package trent

import (
	"crypto/rsa"

	"github.com/sudeeya/key-exchange/internal/pkg/pem"
)

type client struct {
	PublicKey *rsa.PublicKey
}

type clients map[string]client

func newClients(ids, keys []string) (clients, error) {
	clientsList := make(clients, len(ids))
	for i, id := range ids {
		publicKey, err := pem.ExtractRSAPublicKey(keys[i])
		if err != nil {
			return nil, err
		}
		clientsList[id] = client{PublicKey: publicKey}
	}

	return clientsList, nil
}
