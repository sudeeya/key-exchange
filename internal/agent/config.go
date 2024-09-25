package agent

import "github.com/caarlos0/env"

type config struct {
	ID         string `env:"ID,required"`
	Addr       string `env:"ADDR,required"`
	PublicKey  string `env:"PUBLIC_KEY,required"`
	PrivateKey string `env:"PRIVATE_KEY,required"`

	TrentAddr      string `env:"TRENT_ADDR,required"`
	TrentPublicKey string `env:"TRENT_PUBLIC_KEY,required"`

	AgentIDs   []string `env:"AGENT_IDS,required"`
	AgentAddrs []string `env:"AGENT_ADDRS,required"`
}

func newConfig() (*config, error) {
	var cfg config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
