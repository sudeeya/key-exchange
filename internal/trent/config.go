package trent

import "github.com/caarlos0/env"

type config struct {
	PublicKey  string `env:"PUBLIC_KEY,required"`
	PrivateKey string `env:"PRIVATE_KEY,required"`

	ClientIDs        []string `env:"CLIENT_IDS,required"`
	ClientPublicKeys []string `env:"CLIENT_PUBLIC_KEYS,required"`
}

func NewConfig() (*config, error) {
	var cfg config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
