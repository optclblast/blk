package usecase

import (
	"context"

	"github.com/optclblast/blk/internal/entities"
)

// NodeClient is an node provider client presentation interface
type NodeClient interface {
	// LastBlockNumber return last block number.
	LastBlockNumber(ctx context.Context) (entities.BlockNumber, error)

	// BlockInfoByNumber accepts block number and returns all information, including
	// transactions, related to that block.
	// BlockInfoByNumber may return ErrorRateLimitExceeded and you may want to wrap it into
	// backoff
	BlockInfoByNumber(ctx context.Context, num entities.BlockNumber) (*entities.Block, error)
}
