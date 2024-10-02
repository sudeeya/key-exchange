package trent

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sudeeya/key-exchange/internal/pkg/api"
	"github.com/sudeeya/key-exchange/internal/pkg/pem"
	"github.com/sudeeya/key-exchange/internal/pkg/rng"
	"go.uber.org/zap"
)

type Trent struct {
	cfg        *config
	logger     *zap.Logger
	agentList  agents
	mux        *chi.Mux
	rng        *rng.RNG
	privateKey []byte
	publicKey  []byte
}

func NewTrent() *Trent {
	cfg, err := newConfig()
	if err != nil {
		log.Fatal(err)
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}

	logger.Info("Extracting RSA private key")
	privateKey, err := pem.ExtractRSAPrivateKey(cfg.PrivateKey)
	if err != nil {
		logger.Fatal(err.Error())
	}
	logger.Info("Extracting RSA public key")
	publicKey, err := pem.ExtractRSAPublicKey(cfg.PublicKey)
	if err != nil {
		logger.Fatal(err.Error())
	}

	logger.Info("Forming agent list")
	agentList, err := newAgents(cfg.AgentIDs, cfg.AgentPublicKeys)
	if err != nil {
		logger.Fatal(err.Error())
	}

	logger.Info("Initializing router")
	mux := chi.NewRouter()
	logger.Info("Initializing middleware")
	mux.Use(middleware.Logger)

	logger.Info("Initializing RNG")
	rng := rng.NewRNG()

	return &Trent{
		cfg:        cfg,
		logger:     logger,
		agentList:  agentList,
		mux:        mux,
		rng:        rng,
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

func (t Trent) Run() {
	t.logger.Info("Initializing endpoints")
	t.addRoutes()

	t.logger.Info("Server is running")
	if err := http.ListenAndServe(t.cfg.Addr, t.mux); err != nil {
		log.Fatal(err)
	}
}

func (t *Trent) addRoutes() {
	t.mux.Post(api.Step2Endpoint, step2Handler(t))
	t.mux.Post(api.Step5Endpoint, step5Handler(t))
}
