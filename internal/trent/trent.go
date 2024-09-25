package trent

import (
	"log"
)

type Trent struct {
	cfg *config
}

func NewTrent() *Trent {
	cfg, err := NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	return &Trent{
		cfg: cfg,
	}
}

func (a Trent) Run() {
}
