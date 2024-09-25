package trent

import (
	"log"
)

type Trent struct {
	cfg     *config
	clients map[string]client
}

func NewTrent() *Trent {
	cfg, err := NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	clients := NewClients(cfg.ClientIDs, cfg.ClientPublicKeys)

	return &Trent{
		cfg:     cfg,
		clients: clients,
	}
}

func (a Trent) Run() {
}
