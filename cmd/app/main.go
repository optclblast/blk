package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/optclblast/blk/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Init(ctx); err != nil {
		panic(err)
	}
}
