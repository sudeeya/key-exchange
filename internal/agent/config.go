package agent

import "github.com/caarlos0/env"

type config struct {
	AgentID      string `env:"AGENT_ID,required"`
	NeighborID   string `env:"NEIGHBOR_ID,required"`
	TrentAddr    string `env:"TRENT_ADDR,required"`
	NeighborAddr string `env:"NEIGHBOR_ADDR,required"`
	PublicKey    []byte `env:"PUBLIC_KEY,required"`
	PrivateKey   []byte `env:"PRIVATE_KEY,required"`
}

func NewConfig() (*config, error) {
	var cfg config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
