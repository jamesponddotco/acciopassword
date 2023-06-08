// Package server is the main server for the application.
package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"git.sr.ht/~jamesponddotco/acciopassword/internal/cerrors"
	"git.sr.ht/~jamesponddotco/acciopassword/internal/config"
	"git.sr.ht/~jamesponddotco/acciopassword/internal/database"
	"git.sr.ht/~jamesponddotco/acciopassword/internal/endpoint"
	"git.sr.ht/~jamesponddotco/acciopassword/internal/server/handler"
	"git.sr.ht/~jamesponddotco/acciopassword/internal/server/middleware"
	"git.sr.ht/~jamesponddotco/xstd-go/xcrypto/xtls"
	"go.uber.org/zap"
)

type Server struct {
	httpServer *http.Server
	logger     *zap.Logger
}

func New(cfg *config.Config, db *database.DB, logger *zap.Logger) (*Server, error) {
	cert, err := tls.LoadX509KeyPair(cfg.Server.TLS.Certificate, cfg.Server.TLS.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
	}

	var tlsConfig *tls.Config

	if cfg.Server.TLS.Version == "1.3" {
		tlsConfig = xtls.ModernServerConfig()
	}

	if cfg.Server.TLS.Version == "1.2" {
		tlsConfig = xtls.IntermediateServerConfig()
	}

	tlsConfig.Certificates = []tls.Certificate{cert}

	middlewares := []func(http.Handler) http.Handler{
		func(h http.Handler) http.Handler { return middleware.PanicRecovery(logger, h) },
		func(h http.Handler) http.Handler { return middleware.UserAgent(logger, h) },
		func(h http.Handler) http.Handler { return middleware.AcceptRequests(logger, h) },
		func(h http.Handler) http.Handler { return middleware.PrivacyPolicy(cfg.PrivacyPolicy, h) },
		func(h http.Handler) http.Handler { return middleware.TermsOfService(cfg.TermsOfService, h) },
		middleware.CORS,
	}

	var (
		dicewareHandler = handler.NewDicewareHandler(db, logger)
		randomHandler   = handler.NewRandomHandler(db, logger)
		pinHandler      = handler.NewPINHandler(db, logger)
		metricsHandler  = handler.NewMetricsHandler(db, logger)
		healthHandler   = handler.NewHealthHandler(db, logger)
		pingHandler     = handler.NewPingHandler(logger)
	)

	mux := http.NewServeMux()
	mux.HandleFunc(endpoint.Root, func(w http.ResponseWriter, r *http.Request) {
		cerrors.JSON(w, logger, cerrors.ErrorResponse{
			Code:    http.StatusNotFound,
			Message: "Page not found. Check the URL and try again.",
		})
	})

	mux.Handle(endpoint.Diceware, middleware.Chain(dicewareHandler, middlewares...))
	mux.Handle(endpoint.Random, middleware.Chain(randomHandler, middlewares...))
	mux.Handle(endpoint.PIN, middleware.Chain(pinHandler, middlewares...))
	mux.Handle(endpoint.Metrics, middleware.Chain(metricsHandler, middlewares...))
	mux.Handle(endpoint.Health, middleware.Chain(healthHandler, middlewares...))
	mux.Handle(endpoint.Ping, middleware.Chain(pingHandler, middlewares...))

	httpServer := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      mux,
		TLSConfig:    tlsConfig,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpServer: httpServer,
		logger:     logger,
	}, nil
}

func (s *Server) Start() error {
	var (
		sigint            = make(chan os.Signal, 1)
		shutdownCompleted = make(chan struct{})
	)

	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigint

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.Error("HTTP server Shutdown:", zap.Error(err))
		}

		close(shutdownCompleted)
	}()

	if err := s.httpServer.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start server: %w", err)
	}

	<-shutdownCompleted

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	return nil
}
