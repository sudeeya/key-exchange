package trent

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sudeeya/key-exchange/internal/pkg/pem"
	"github.com/sudeeya/key-exchange/internal/pkg/rng"
)

type Trent struct {
	cfg         *config
	clientsList clients
	mux         *chi.Mux
	rng         *rng.RNG
	privateKey  []byte
	publicKey   []byte
}

func NewTrent() *Trent {
	cfg, err := newConfig()
	if err != nil {
		log.Fatal(err)
	}

	privateKey, err := pem.ExtractRSAPrivateKey(cfg.PrivateKey)
	if err != nil {
		log.Fatal(err)
	}
	publicKey, err := pem.ExtractRSAPublicKey(cfg.PublicKey)
	if err != nil {
		log.Fatal(err)
	}

	clientsList, err := newClients(cfg.ClientIDs, cfg.ClientPublicKeys)
	if err != nil {
		log.Fatal(err)
	}

	mux := chi.NewRouter()

	rng := rng.NewRNG()

	return &Trent{
		cfg:         cfg,
		clientsList: clientsList,
		mux:         mux,
		rng:         rng,
		privateKey:  privateKey,
		publicKey:   publicKey,
	}
}

func (t Trent) Run() {
	t.addRoutes()

	if err := http.ListenAndServe(t.cfg.Addr, t.mux); err != nil {
		log.Fatal(err)
	}
}

func (t *Trent) addRoutes() {
	t.mux.Post("/initiate/", newInitiateHandler(t))
	t.mux.Post("/confirm/", newConfirmHandler(t))
}
