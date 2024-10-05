package trent

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/sudeeya/key-exchange/internal/pkg/api"
	"github.com/sudeeya/key-exchange/internal/pkg/middleware"
	"github.com/sudeeya/key-exchange/internal/pkg/pem"
	"github.com/sudeeya/key-exchange/internal/pkg/rng"
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

	loggerCfg := zap.NewDevelopmentConfig()
	loggerCfg.OutputPaths = []string{
		cfg.LogFile,
	}
	logger, err := loggerCfg.Build()
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
	mux.Use(middleware.WithLogging(logger))

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

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	t.logger.Info("Server is running")
	go func() {
		<-sigCh
		t.logger.Info("Trent is shutting down")
		t.Shutdown()
	}()

	if err := http.ListenAndServe(t.cfg.Addr, t.mux); err != nil {
		t.logger.Fatal(err.Error())
	}
}

func (t Trent) Shutdown() {
	if err := t.logger.Sync(); err != nil {
		t.logger.Sugar().Fatalf("failed to sync logger: %v", err)
	}

	os.Exit(0)
}

func (t *Trent) addRoutes() {
	t.mux.Post(api.Step2Endpoint, step2Handler(t))
	t.mux.Post(api.Step5Endpoint, step5Handler(t))
}
