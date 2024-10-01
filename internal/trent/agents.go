package trent

import (
	"github.com/sudeeya/key-exchange/internal/pkg/pem"
)

type agent struct {
	PublicKey []byte
}

type agents map[string]agent

func newAgents(ids, keys []string) (agents, error) {
	clientsList := make(agents, len(ids))
	for i, id := range ids {
		publicKey, err := pem.ExtractRSAPublicKey(keys[i])
		if err != nil {
			return nil, err
		}
		clientsList[id] = agent{PublicKey: publicKey}
	}

	return clientsList, nil
}
