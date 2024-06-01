package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/optclblast/blk/internal/controller/http"
	"github.com/optclblast/blk/internal/infrastructure/getblock"
	"github.com/optclblast/blk/internal/logger"
	"github.com/optclblast/blk/internal/server"
	"github.com/optclblast/blk/internal/usecase"
)

const (
	// API access token
	getblockAccessTokenEnv = "BLK_GETBLOCK_ACCESS_TOKEN"
	// Log level
	logLevelEnv = "BLK_LOG_LEVEL"
	// Listen address
	httpAddrEnv = "BLK_HTTP_ADDR"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	getblockAccessToken := os.Getenv(getblockAccessTokenEnv)
	logLevel := os.Getenv(logLevelEnv)
	httpAddr := os.Getenv(httpAddrEnv)

	log := logger.NewBuilder().
		WithLevel(logger.MapLevel(logLevel)).
		Build()

	log.Info(
		"starting blk server 0w0",
		slog.String("address", httpAddr),
		slog.String("log level", logLevel),
	)

	getblockClient := getblock.NewClient(
		log.WithGroup("getblock-client"),
		getblockAccessToken,
	)

	ethInteractor := usecase.NewEthInteractor(
		log.WithGroup("eth-interactor"),
		getblockClient,
	)

	walletsController := http.NewWalletsController(
		log.WithGroup("wallets-controller"),
		ethInteractor,
	)

	router := http.NewRouter(
		log.WithGroup("router"),
		walletsController,
	)

	server := server.New(router, httpAddr)

	select {
	case <-ctx.Done():
		log.Info("shutting down blk server. bye bye! =w=")
	case err := <-server.Notify():
		log.Error("error listen to net ;_;", logger.Err(err))
	}

	if err := server.Shutdown(); err != nil {
		panic(err)
	}
}
