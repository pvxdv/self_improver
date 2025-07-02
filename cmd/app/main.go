// Package main implements a Telegram bot task tracker application.
//
// Example usage:
//
//	go run cmd/app/main.go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/pvxdv/self_improver/internal/config"
	"github.com/pvxdv/self_improver/internal/config/loader/env"
	"github.com/pvxdv/self_improver/internal/http-server/handlers/health"
	challengeAdd "github.com/pvxdv/self_improver/internal/http-server/handlers/url/challenge/add"
	trendAdd "github.com/pvxdv/self_improver/internal/http-server/handlers/url/trend/add"
	trendDelete "github.com/pvxdv/self_improver/internal/http-server/handlers/url/trend/delete"
	trendGet "github.com/pvxdv/self_improver/internal/http-server/handlers/url/trend/get"
	"github.com/pvxdv/self_improver/internal/storage/postgres"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfgLoader := env.NewLoader()
	cfg, err := config.LoadAndValidate(cfgLoader)
	if err != nil {
		log.Fatal(err)
	}

	consoleLogger, err := setUpLogger(cfg)
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer consoleLogger.Sync()

	consoleLogger.Debugf("config: %+v", cfg)
	fmt.Print()

	strg, err := postgres.New(ctx, cfg.Storage, consoleLogger)
	if err != nil {
		consoleLogger.Fatalf("failed to initialize storage: %v", err)
	}
	defer func(strg *postgres.Storage, ctx context.Context) {
		err := strg.Close(ctx)
		if err != nil {
			consoleLogger.Errorf("failed to close storage: %v", err)
		}
	}(strg, ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("api/v1/challenge/add", challengeAdd.New(ctx, strg, consoleLogger))
	mux.HandleFunc("api/v1/trend/add", trendAdd.New(ctx, strg, consoleLogger))
	mux.HandleFunc("api/v1/trend/get", trendGet.New(ctx, strg, consoleLogger))
	mux.HandleFunc("api/v1/trend/delete", trendDelete.New(ctx, strg, consoleLogger))

	mux.HandleFunc("/health", health.New(consoleLogger))

	server := &http.Server{
		Addr:    ":" + cfg.HTTP.Port,
		Handler: mux,
	}

	go func() {
		consoleLogger.Infof("server started at %s port", cfg.HTTP.Port)
		if err := server.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				consoleLogger.Errorf("failed to start http server: %v", err)
			}
		}
	}()

	<-ctx.Done()
	consoleLogger.Info("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		consoleLogger.Errorf("server shutdown error: %v", err)
	}
}

func setUpLogger(cfg *config.Config) (*zap.SugaredLogger, error) {
	logConfig := zap.NewProductionConfig()

	logConfig.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(time.RFC1123))
	}
	logConfig.Encoding = "json"

	switch cfg.App.Debug {
	case true:
		logConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case false:
		logConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	logConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	logger, err := logConfig.Build()
	if err != nil {
		return nil, err
	}

	return logger.Sugar(), nil
}
