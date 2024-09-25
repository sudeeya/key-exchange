package trent

type client struct {
	PublicKey []byte
}

func newClients(ids, keys []string) map[string]client {
	clients := make(map[string]client, len(ids))
	for i, id := range ids {
		clients[id] = client{PublicKey: []byte(keys[i])}
	}
	return clients
}
