package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"runtime"
	"sync"

	"github.com/alitto/pond"
	"github.com/optclblast/blk/internal/entities"
	"github.com/optclblast/blk/internal/logger"
)

type EthInteractor interface {
	MostChangedAddress(ctx context.Context, numBlocks int) (*entities.Wallet, error)
}

type ethInteractor struct {
	log    *slog.Logger
	client NodeClient
}

func NewEthInteractor(
	log *slog.Logger,
	client NodeClient,
) EthInteractor {
	return &ethInteractor{
		log:    log,
		client: client,
	}
}

type MostChangedAddressResult struct {
	Address string
	Delta   *big.Int
}

func (t *ethInteractor) MostChangedAddress(ctx context.Context, numBlocks int) (*entities.Wallet, error) {
	// We need to fetch current head block
	head, err := t.client.LastBlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("error fetch last block number. %w", err)
	}

	t.log.Debug(
		"MostChangedAddress",
		slog.String("head block number", (string)(head)),
		slog.Int("num blocks parameter", numBlocks),
	)

	lastBlockNumber, err := head.ToInt()
	if err != nil {
		return nil, fmt.Errorf("error map last block number to numeric. %w", err)
	}

	wallets, err := t.fetchWallets(ctx, lastBlockNumber, numBlocks)
	if err != nil {
		return nil, fmt.Errorf("error fetch wallets. %w", err)
	}

	t.log.Debug(
		"MostChangedAddress",
		slog.Int("num blocks parameter", numBlocks),
		slog.Int("wallets fetched", len(wallets)),
	)

	if len(wallets) == 0 {
		return nil, fmt.Errorf("error empty wallets set")
	}

	return wallets[0], nil
}

func (t *ethInteractor) fetchWallets(
	ctx context.Context,
	headBlock *big.Int,
	numBlocks int,
) ([]*entities.Wallet, error) {
	txChan := make(chan *entities.Transaction, numBlocks)

	t.fetchBlocksTransactions(ctx, headBlock, numBlocks, txChan)

	// map [Wallet address => Delta]
	addresses := make(map[string]*big.Int, numBlocks)

	// Fill the map with deltas
	for t := range txChan {
		delta, ok := addresses[t.From]
		if !ok {
			delta = new(big.Int)
		}

		addresses[t.From] = delta.Sub(delta, t.Value)

		delta, ok = addresses[t.To]
		if !ok {
			delta = new(big.Int)
		}

		addresses[t.To] = delta.Add(delta, t.Value)
	}

	// Build wallets array
	wallets := make(entities.Wallets, len(addresses))
	{
		i := 0
		for addr, dlt := range addresses {
			wallets[i] = &entities.Wallet{
				Address: addr,
				Delta:   dlt,
			}

			i++
		}
	}

	// Sort
	wallets.Sort()

	return wallets, nil
}

func (t *ethInteractor) fetchBlocksTransactions(
	ctx context.Context,
	headBlock *big.Int,
	numBlocks int,
	txChan chan<- *entities.Transaction,
) {
	blockToFetch := new(big.Int).Set(headBlock)
	endBlock := big.NewInt(0)

	processPool := pond.New(runtime.GOMAXPROCS(0)*2, numBlocks)
	fetchPool := pond.New(2, numBlocks)

	blocksChan := make(chan *entities.Block, numBlocks/2)

	var fetchWg sync.WaitGroup

	for i := 0; i < numBlocks; i++ {
		blockNumber := entities.BlockNumber("0x" + blockToFetch.Text(16))

		fetchWg.Add(1)
		fetchPool.Submit(func() {
			defer fetchWg.Done()

			block, err := t.client.BlockInfoByNumber(
				ctx,
				blockNumber,
			)
			if err != nil {
				t.log.Error(
					"error fetch block info",
					logger.Err(err),
					slog.Any("block number", blockNumber),
				)

				return
			}

			blocksChan <- block
		})

		blockToFetch.Sub(blockToFetch, big.NewInt(1))

		if blockToFetch.Cmp(endBlock) == 0 {
			break
		}
	}

	go func() {
		fetchWg.Wait()
		close(blocksChan)
	}()

	var processWg sync.WaitGroup

	for b := range blocksChan {
		processWg.Add(1)

		processPool.Submit(func() {
			defer processWg.Done()

			for _, tx := range b.Transactions {
				txChan <- tx
			}
		})
	}

	go func() {
		processWg.Wait()
		close(txChan)
	}()
}
