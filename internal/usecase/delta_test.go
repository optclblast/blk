package usecase

import (
	"context"
	"log/slog"
	"math/big"
	"testing"

	"github.com/optclblast/blk/internal/entities"
)

type TestCase struct {
	Title          string
	Block          *entities.Block
	ExpectedResult string
}

var testBlocks []TestCase = []TestCase{
	{
		Title: "1 block. 3 wallets, 3 txs. Negative is the higest",
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
		ExpectedResult: "A",
	},
	{
		Title: "2 block, 3 wallets, 4 txs. Positive is the higest",
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
					To:    "B", // +10
					Value: big.NewInt(15),
				},
			},
		},
		ExpectedResult: "B",
	},
}

func TestAddressWithBiggestDelta(t *testing.T) {
	ethInteractor := &ethInteractor{log: slog.Default()}

	for _, tc := range testBlocks {
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

			if walletAddr != tc.ExpectedResult {
				t.Fatalf("invalid result: %s | Expected: %s", walletAddr, tc.ExpectedResult)
			}
		})
	}
}
