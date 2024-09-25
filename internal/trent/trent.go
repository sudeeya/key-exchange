package trent

import (
	"log"
)

type Trent struct {
	cfg     *config
	clients map[string]client
}

func NewTrent() *Trent {
	cfg, err := newConfig()
	if err != nil {
		log.Fatal(err)
	}

	clients := newClients(cfg.ClientIDs, cfg.ClientPublicKeys)

	return &Trent{
		cfg:     cfg,
		clients: clients,
	}
}

func (a Trent) Run() {
}
