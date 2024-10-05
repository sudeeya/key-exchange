package agent

import (
	"github.com/caarlos0/env"
)

type config struct {
	ID         string `env:"ID,required"`
	Addr       string `env:"ADDR,required"`
	PublicKey  string `env:"PUBLIC_KEY,required"`
	PrivateKey string `env:"PRIVATE_KEY,required"`

	TrentAddr      string `env:"TRENT_ADDR,required"`
	TrentPublicKey string `env:"TRENT_PUBLIC_KEY,required"`

	AgentID   string `env:"AGENT_ID,required"`
	AgentAddr string `env:"AGENT_ADDR,required"`
}

func newConfig() (*config, error) {
	var cfg config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
