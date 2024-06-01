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

var defaultWorkersNum = runtime.GOMAXPROCS(0) * 2

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

	txChan := make(chan *entities.Transaction, defaultWorkersNum)

	// Begin a thansactions data stream
	t.streamTransactions(ctx, headBlockNumber, numBlocks, txChan)

	// Handle transactions stream and calculate th result
	walletAddress, err := t.addressWithBiggestDelta(ctx, txChan)
	if err != nil {
		return "", fmt.Errorf("error fetch wallets. %w", err)
	}

	return walletAddress, nil
}

func (t *ethInteractor) addressWithBiggestDelta(
	ctx context.Context,
	txChan chan *entities.Transaction,
) (string, error) {
	outChan := make(chan string, 1)

	go func() {
		defer func() {
			if panic := recover(); panic != nil {
				t.log.Error("addressWithBiggestDelta", slog.Any("panic", panic))
				return
			}
		}()

		// map [Wallet address => Delta]
		addresses := cmap.New[*big.Int]()

		var wg sync.WaitGroup

		// Fill the map with address / delta pairs
		for i := 0; i < defaultWorkersNum; i++ {
			// Run a writer worker
			wg.Add(1)

			go t.appendAddressDeltaWorker(&wg, &addresses, txChan)
		}

		wg.Wait()

		outChan <- biggestDeltaAddres(addresses.Items())
	}()

	select {
	case out := <-outChan:
		return out, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func (t *ethInteractor) appendAddressDeltaWorker(
	wg *sync.WaitGroup,
	cmp *cmap.ConcurrentMap[string, *big.Int],
	txsChan <-chan *entities.Transaction,
) {
	defer func() {
		if panic := recover(); panic != nil {
			t.log.Error("appendAddressDeltaWorker", slog.Any("panic", panic))
			return
		}
	}()

	defer wg.Done()

	for t := range txsChan {
		deltaFrom, ok := cmp.Get(t.From)
		if !ok {
			deltaFrom = new(big.Int)
		}

		cmp.Set(t.From, deltaFrom.Sub(deltaFrom, t.Value))

		deltaTo, ok := cmp.Get(t.To)
		if !ok {
			deltaTo = new(big.Int)
		}

		cmp.Set(t.To, deltaTo.Add(deltaTo, t.Value))
	}
}

// Returns an address of a wallet with balances mod|delta| is the highest
func biggestDeltaAddres(set map[string]*big.Int) string {
	if len(set) == 0 {
		return ""
	}

	i := 0

	// Build wallets array
	wallets := make(entities.Wallets, len(set))
	for addr, dlt := range set {
		wallets[i] = entities.Wallet{
			Address: addr,
			Delta:   dlt.Abs(dlt),
		}

		i++
	}

	// Sort
	wallets.Sort()

	return wallets[len(wallets)-1].Address
}

const fetchWorkersPoolSize = 4

// streamTransactions fetches blocks from getblock node API and
// dispatches related transaction into a dedicated channel for
// other workers to process.
// The channels used by streamTransactions will be closed
// internally.
func (t *ethInteractor) streamTransactions(
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
