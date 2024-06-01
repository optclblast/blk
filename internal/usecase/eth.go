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
	cmap "github.com/orcaman/concurrent-map/v2"
)

type EthInteractor interface {
	MostChangedAddress(ctx context.Context, numBlocks int) (string, error)
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

func (t *ethInteractor) MostChangedAddress(ctx context.Context, numBlocks int) (string, error) {
	// We need to fetch current head block
	head, err := t.client.LastBlockNumber(ctx)
	if err != nil {
		return "", fmt.Errorf("error fetch last block number. %w", err)
	}

	t.log.Debug(
		"most_changed_address",
		slog.String("head block number", (string)(head)),
		slog.Int("num blocks parameter", numBlocks),
	)

	headBlockNumber, err := head.ToInt()
	if err != nil {
		return "", fmt.Errorf("error map last block number to numeric. %w", err)
	}

	txChan := make(chan *entities.Transaction, numBlocks)

	t.fetchBlocksTransactions(ctx, headBlockNumber, numBlocks, txChan)

	walletAddress, err := t.addressWithBiggestDelta(ctx, txChan)
	if err != nil {
		return "", fmt.Errorf("error fetch wallets. %w", err)
	}

	return walletAddress, nil
}

var defaultWorkersNum = runtime.GOMAXPROCS(0) * 2

func (t *ethInteractor) addressWithBiggestDelta(
	ctx context.Context,
	txChan chan *entities.Transaction,
) (string, error) {
	// map [Wallet address => Delta]
	addresses := cmap.New[*big.Int]()

	var wg sync.WaitGroup

	// Fill the map with address / delta pairs
	for i := 0; i < defaultWorkersNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for t := range txChan {
				deltaFrom, ok := addresses.Get(t.From)
				if !ok {
					deltaFrom = new(big.Int)
				}

				addresses.Set(t.From, deltaFrom.Sub(deltaFrom, t.Value))

				deltaTo, ok := addresses.Get(t.To)
				if !ok {
					deltaTo = new(big.Int)
				}

				addresses.Set(t.To, deltaTo.Add(deltaTo, t.Value))
			}
		}()
	}

	wg.Wait()

	t.log.Debug(
		"processing done",
		slog.Int("fetched wallets", len(addresses.Items())),
	)

	return biggestDeltaAddres(addresses.Items()), nil
}

const fetchWorkersPoolSize = 4

func (t *ethInteractor) fetchBlocksTransactions(
	ctx context.Context,
	headBlock *big.Int,
	numBlocks int,
	txChan chan<- *entities.Transaction,
) {
	blockToFetch := new(big.Int).Set(headBlock)
	endBlock := big.NewInt(0)

	blocksChan := make(chan *entities.Block, numBlocks)

	var fetchWg sync.WaitGroup
	fetchPool := pond.New(fetchWorkersPoolSize, numBlocks)

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

	processPool := pond.New(defaultWorkersNum, numBlocks)

	var processWg sync.WaitGroup

	go func() {
		dispatchBlockTransactions(&processWg, processPool, blocksChan, txChan)

		processWg.Wait()
		close(txChan)
	}()
}

// Dispatches blocks from blocksChan into txsChan
func dispatchBlockTransactions(
	wg *sync.WaitGroup,
	pool *pond.WorkerPool,
	blocksChan <-chan *entities.Block,
	txsChan chan<- *entities.Transaction,
) {
	for b := range blocksChan {
		wg.Add(1)

		pool.Submit(func() {
			defer wg.Done()

			for _, tx := range b.Transactions {
				txsChan <- tx
			}
		})
	}
}

// Returns an address of a wallet with balances mod|delta| is the highest
func biggestDeltaAddres(set map[string]*big.Int) string {
	// Build wallets array
	wallets := make(entities.Wallets, 0, len(set))
	for addr, dlt := range set {
		wallets = append(wallets, &entities.Wallet{
			Address: addr,
			Delta:   dlt.Abs(dlt),
		})
	}

	// Sort
	wallets.Sort()

	return wallets[len(wallets)-1].Address
}
