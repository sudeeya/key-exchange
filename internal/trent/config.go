package trent

import (
	"github.com/caarlos0/env"
)

type config struct {
	Addr       string `env:"ADDR,required"`
	PublicKey  string `env:"PUBLIC_KEY,required"`
	PrivateKey string `env:"PRIVATE_KEY,required"`

	AgentIDs        []string `env:"AGENT_IDS,required"`
	AgentPublicKeys []string `env:"AGENT_PUBLIC_KEYS,required"`
}

func newConfig() (*config, error) {
	var cfg config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
