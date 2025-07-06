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
	telegram "github.com/pvxdv/self_improver/internal/telegram/bot"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/pvxdv/self_improver/internal/config"
	"github.com/pvxdv/self_improver/internal/config/app"
	"github.com/pvxdv/self_improver/internal/config/loader/env"
	"github.com/pvxdv/self_improver/internal/http-server/handlers/health"
	challengeAdd "github.com/pvxdv/self_improver/internal/http-server/handlers/url/challenge/add"
	debtAdd "github.com/pvxdv/self_improver/internal/http-server/handlers/url/finance/debt/add"
	debtDelete "github.com/pvxdv/self_improver/internal/http-server/handlers/url/finance/debt/delete"
	debtGet "github.com/pvxdv/self_improver/internal/http-server/handlers/url/finance/debt/get"
	debtPay "github.com/pvxdv/self_improver/internal/http-server/handlers/url/finance/debt/pay"
	debtUpdate "github.com/pvxdv/self_improver/internal/http-server/handlers/url/finance/debt/update"
	trendAdd "github.com/pvxdv/self_improver/internal/http-server/handlers/url/trend/add"
	trendDelete "github.com/pvxdv/self_improver/internal/http-server/handlers/url/trend/delete"
	trendGet "github.com/pvxdv/self_improver/internal/http-server/handlers/url/trend/get"
	debtService "github.com/pvxdv/self_improver/internal/service/finance/debt"
	"github.com/pvxdv/self_improver/internal/storage/postgres"
)

// @title Self Improver API
// @version 0.0.0

// @description Available in local/dev environments only.
// @description This API allows users to track personal improvement challenges and trends.
// @description It provides endpoints for:
// @description - Creating and deleting trends
// @description - Adding challenges to trends
// @description - Fetching trend data with associated challenges

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @BasePath /
// @schemes http https
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfgLoader := env.NewLoader()
	cfg, err := config.LoadAndValidate(cfgLoader)
	if err != nil {
		log.Fatal(err)
	}

	consoleLogger, err := setUpLogger(cfg.App)
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer consoleLogger.Sync()

	consoleLogger.Debugf("config: %+v", cfg)
	fmt.Print()

	storage, err := postgres.New(ctx, cfg.Storage, consoleLogger)
	if err != nil {
		consoleLogger.Fatalf("failed to initialize storage: %v", err)
	}
	defer func(strg *postgres.Storage, ctx context.Context) {
		err := strg.Close(ctx)
		if err != nil {
			consoleLogger.Errorf("failed to close storage: %v", err)
		}
	}(storage, ctx)

	debtSrvc := debtService.NewDebtService(storage, consoleLogger)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health.New(consoleLogger))

	mux.HandleFunc("POST /api/v1/challenge/add", challengeAdd.New(ctx, storage, consoleLogger))
	mux.HandleFunc("POST /api/v1/trend/add", trendAdd.New(ctx, storage, consoleLogger))
	mux.HandleFunc("GET /api/v1/trend/get", trendGet.New(ctx, storage, consoleLogger))
	mux.HandleFunc("DELETE /api/v1/trend/delete", trendDelete.New(ctx, storage, consoleLogger))

	mux.HandleFunc("POST /api/v1/finance/debt/add", debtAdd.New(ctx, debtSrvc, consoleLogger))
	mux.HandleFunc("DELETE /api/v1/finance/debt/delete", debtDelete.New(ctx, debtSrvc, consoleLogger))
	mux.HandleFunc("GET /api/v1/finance/debt/get", debtGet.New(ctx, debtSrvc, consoleLogger))
	mux.HandleFunc("POST /api/v1/finance/debt/pay", debtPay.New(ctx, debtSrvc, consoleLogger))
	mux.HandleFunc("PUT /api/v1/finance/debt/update", debtUpdate.New(ctx, debtSrvc, consoleLogger))

	if cfg.App.Env == "local" || cfg.App.Env == "dev" {
		mux.Handle("/swagger/", httpSwagger.Handler(
			httpSwagger.URL("doc.json"),
		))

		mux.HandleFunc("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "./docs/swagger.json")
		})
	}

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

	bot, err := telegram.New(ctx, cfg.Telegram, consoleLogger, debtSrvc)
	if err != nil {
		consoleLogger.Fatalf("failed to create bot: %v", err)
	}

	go func() {
		consoleLogger.Infof("starting telegram bot")
		err = bot.Start()
		if err != nil {
			consoleLogger.Infof("stoping telergam bot: %v", err)
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

func setUpLogger(cfg *app.Config) (*zap.SugaredLogger, error) {
	logConfig := zap.NewProductionConfig()

	logConfig.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(time.RFC1123))
	}

	switch cfg.Env {
	case "local":
		logConfig.Encoding = "console"
	case "dev":
		logConfig.Encoding = "console"
	case "prod":
		logConfig.Encoding = "json"
	}

	switch cfg.Debug {
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
