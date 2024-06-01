package app

import (
	"context"
	"log/slog"
	"os"

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

// Init is a main function in our application lifecycle.
// Init is responsible for bringing all the system's components together.
func Init(ctx context.Context) error {
	// Fetch env vars
	getblockAccessToken := os.Getenv(getblockAccessTokenEnv)
	logLevel := os.Getenv(logLevelEnv)
	httpAddr := os.Getenv(httpAddrEnv)

	// Build logger
	log := logger.NewBuilder().
		WithLevel(logger.MapLevel(logLevel)).
		Build()

	log.Info(
		"starting blk server 0w0",
		slog.String("address", httpAddr),
		slog.String("log level", logLevel),
	)

	// Initialize node provider client
	getblockClient := getblock.NewClient(
		log.WithGroup("getblock-client"),
		getblockAccessToken,
	)

	// Initialize application layer
	ethInteractor := usecase.NewEthInteractor(
		log.WithGroup("eth-interactor"),
		getblockClient,
	)

	// Initialize controller layer
	walletsController := http.NewWalletsController(
		log.WithGroup("wallets-controller"),
		ethInteractor,
	)

	// Build a router
	router := http.NewRouter(
		log.WithGroup("router"),
		walletsController,
	)

	// And run server with it
	server := server.New(router, httpAddr)

	select {
	case <-ctx.Done():
		log.Info("shutting down blk server. bye bye! =w=")
	case err := <-server.Notify():
		log.Error("error listen to net ;_;", logger.Err(err))
	}

	if err := server.Shutdown(); err != nil {
		return err
	}

	return nil
}
