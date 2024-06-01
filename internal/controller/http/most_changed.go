package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/optclblast/blk/internal/usecase"
)

type WalletsController interface {
	MostChangedWalletAddress(w http.ResponseWriter, r *http.Request) (any, error)
}

const (
	defaultNumBlocks = 100
	maxNumBlocks     = 150
)

func (c *walletsController) MostChangedWalletAddress(w http.ResponseWriter, r *http.Request) (any, error) {
	defer r.Body.Close()

	var (
		query     = r.URL.Query()
		numBlocks = defaultNumBlocks
		err       error
	)

	if v, ok := query["blocks"]; ok && len(v) > 0 {
		numBlocks, err = strconv.Atoi(v[0])
		if err != nil {
			return nil, fmt.Errorf(
				"error invalid block param value. %w",
				errors.Join(err, ErrorBadQueryParams),
			)
		}

		if numBlocks > maxNumBlocks {
			numBlocks = maxNumBlocks
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	walletAddress, err := c.usecase.MostChangedAddress(ctx, numBlocks)
	if err != nil {
		return nil, fmt.Errorf("error fetch the most changed wallet. %w", err)
	}

	return MostChangedWalletAddressResponse{
		Address: walletAddress,
	}, nil
}

type walletsController struct {
	log     *slog.Logger
	usecase usecase.EthInteractor
}

func NewWalletsController(
	log *slog.Logger,
	usecase usecase.EthInteractor,
) WalletsController {
	return &walletsController{
		log:     log,
		usecase: usecase,
	}
}
