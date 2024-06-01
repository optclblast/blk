package usecase

import (
	"context"
	"log/slog"
	"math/big"
	"math/rand"
	"slices"
	"testing"

	"github.com/optclblast/blk/internal/entities"
)

func TestAddressWithBiggestDelta(t *testing.T) {
	ethInteractor := &ethInteractor{log: slog.Default()}

	for _, tc := range tests {
		t.Run(tc.Title, func(t *testing.T) {
			txCh := make(chan *entities.Transaction, len(tc.Block.Transactions))

			for _, txs := range tc.Block.Transactions {
				txCh <- txs
			}

			close(txCh)

			walletAddr, err := ethInteractor.addressWithBiggestDelta(context.TODO(), txCh)
			if err != nil {
				t.Fatalf("error: %s\n", err.Error())
			}

			if !slices.Contains(tc.ExpectedResult, walletAddr) {
				t.Fatalf("invalid result: %s | Expected: %v", walletAddr, tc.ExpectedResult)
			}
		})
	}
}

func BenchmarkAddressWithBiggestDelta(b *testing.B) {
	ethInteractor := &ethInteractor{log: slog.Default()}

	txs := make([]*entities.Transaction, 20000)

	for i := 0; i < 20000; i++ {
		txs[i] = &entities.Transaction{
			Value: big.NewInt(rand.Int63()),
			From:  RandStringRunes(rand.Intn(1000)),
			To:    RandStringRunes(rand.Intn(1000)),
		}
	}

	txCh := make(chan *entities.Transaction, len(txs))

	for _, txs := range txs {
		txCh <- txs
	}

	close(txCh)

	_, err := ethInteractor.addressWithBiggestDelta(context.TODO(), txCh)
	if err != nil {
		b.Fatalf("error: %s\n", err.Error())
	}
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)

	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(b)
}

type TestCase struct {
	Title          string
	Block          *entities.Block
	ExpectedResult []string
}

var tests []TestCase = []TestCase{
	{
		Title: "1 block. 3 wallets, 3 txs. Negative is the highest",
		Block: &entities.Block{
			Transactions: []*entities.Transaction{
				{
					From:  "A",
					To:    "B", // + 100
					Value: big.NewInt(100),
				},
				{
					From:  "B", // -90
					To:    "A", // -10
					Value: big.NewInt(90),
				},
				{
					From:  "A", // -5
					To:    "F",
					Value: big.NewInt(5),
				},
			},
		},
		ExpectedResult: []string{"A"},
	},
	{
		Title: "1 block, 3 wallets, 4 txs. Positive is the highest",
		Block: &entities.Block{
			Transactions: []*entities.Transaction{
				{
					From:  "A",
					To:    "B", // + 1000
					Value: big.NewInt(1000),
				},
				{
					From:  "F", // -90
					To:    "A", // -10
					Value: big.NewInt(1000),
				},
				{
					From:  "A",
					To:    "B", // +10
					Value: big.NewInt(10),
				},
				{
					From:  "A",
					To:    "B", // +15
					Value: big.NewInt(15),
				},
			},
		},
		ExpectedResult: []string{"B"},
	},
	{
		Title: "1 block, 0 wallets, 0 txs",
		Block: &entities.Block{
			Transactions: []*entities.Transaction{},
		},
		ExpectedResult: []string{""},
	},
	{
		Title: "1 block, 3 wallets, 3 txs. Txs are negative",
		Block: &entities.Block{
			Transactions: []*entities.Transaction{
				{
					From:  "A",
					To:    "B",
					Value: big.NewInt(-10),
				},
				{
					From:  "A",
					To:    "B",
					Value: big.NewInt(-1000),
				},
				{
					From:  "F",
					To:    "A",
					Value: big.NewInt(-1000),
				},
				{
					From:  "A",
					To:    "B",
					Value: big.NewInt(-15),
				},
			},
		},
		ExpectedResult: []string{"B"},
	},
	{
		Title: "1 block, 2 wallets, 3 txs. Both are equal",
		Block: &entities.Block{
			Transactions: []*entities.Transaction{
				{
					From:  "A",
					To:    "B",
					Value: big.NewInt(-10),
				},
				{
					From:  "A",
					To:    "B",
					Value: big.NewInt(50),
				},
				{
					From:  "B",
					To:    "A",
					Value: big.NewInt(10),
				},
				{
					From:  "A",
					To:    "B",
					Value: big.NewInt(-15),
				},
			},
		},
		ExpectedResult: []string{"B", "A"},
	},
}
