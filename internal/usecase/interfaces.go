package usecase

import (
	"context"

	"github.com/optclblast/blk/internal/entities"
)

type (
	NodeClient interface {
		LastBlockNumber(ctx context.Context) (entities.BlockNumber, error)
		BlockInfoByNumber(ctx context.Context, num entities.BlockNumber) (*entities.Block, error)
	}
)
