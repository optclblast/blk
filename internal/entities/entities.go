package entities

import (
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"time"
)

type Wallet struct {
	Address string
	Delta   *big.Int
}

type walletJSON struct {
	Address string `json:"address"`
	Delta   string `json:"delta"`
}

func (w *Wallet) MarshalJSON() ([]byte, error) {
	deltaHex := w.Delta.Text(16)

	if w.Delta.Sign() == -1 {
		deltaHex = "-0x" + deltaHex[1:]
	}

	return json.Marshal(&walletJSON{
		Address: w.Address,
		Delta:   deltaHex,
	})
}

type Wallets []*Wallet

func (w Wallets) Sort() {
	sort.Slice(w, func(r, l int) bool {
		if r := w[r].Delta.Cmp(w[l].Delta); r < 0 {
			return true
		}

		return false
	})
}

type BlockNumber string

func (n BlockNumber) ToInt() (*big.Int, error) {
	return hexToInt((string)(n))
}

type Block struct {
	Difficulty       *big.Int       `json:"difficulty"`
	BaseFeePerGas    *big.Int       `json:"baseFeePerGas"`
	ExtraData        string         `json:"extraData"`
	GasLimit         *big.Int       `json:"gasLimit"`
	GasUsed          *big.Int       `json:"gasUsed"`
	Hash             string         `json:"hash"`
	LogsBloom        string         `json:"logsBloom"`
	Miner            string         `json:"miner"`
	MixHash          string         `json:"mixHash"`
	Nonce            string         `json:"nonce"`
	Number           *big.Int       `json:"number"`
	ParentHash       string         `json:"parentHash"`
	ReceiptsRoot     string         `json:"receiptsRoot"`
	Sha3Uncles       string         `json:"sha3Uncles"`
	Size             *big.Int       `json:"size"`
	StateRoot        string         `json:"stateRoot"`
	Timestamp        time.Time      `json:"timestamp"`
	TotalDifficulty  *big.Int       `json:"totalDifficulty"`
	Transactions     []*Transaction `json:"transactions"`
	TransactionsRoot string         `json:"transactionsRoot"`
}

// helper alias for proper unmarshal of a block object
type blockAlias Block

type blockRaw struct {
	*blockAlias
	BaseFeePerGas   string `json:"baseFeePerGas"`
	Difficulty      string `json:"difficulty"`
	GasLimit        string `json:"gasLimit"`
	GasUsed         string `json:"gasUsed"`
	Number          string `json:"number"`
	Size            string `json:"size"`
	Timestamp       string `json:"timestamp"`
	TotalDifficulty string `json:"totalDifficulty"`
}

func (b *Block) UnmarshalJSON(data []byte) error {
	raw := &blockRaw{
		blockAlias: (*blockAlias)(b),
	}

	if err := json.Unmarshal(data, raw); err != nil {
		return fmt.Errorf("error unmarshal base block data. %w", err)
	}

	b.Difficulty = hexToIntMust(raw.Difficulty)
	b.TotalDifficulty = hexToIntMust(raw.TotalDifficulty)
	b.BaseFeePerGas = hexToIntMust(raw.BaseFeePerGas)
	b.GasLimit = hexToIntMust(raw.GasLimit)
	b.GasUsed = hexToIntMust(raw.GasUsed)
	b.Number = hexToIntMust(raw.Number)
	b.Size = hexToIntMust(raw.Size)
	b.Timestamp = time.Unix(hexToIntMust(raw.Timestamp).Int64(), 0)

	return nil
}

type Transaction struct {
	BlockHash            string        `json:"blockHash"`
	BlockNumber          *big.Int      `json:"blockNumber"`
	From                 string        `json:"from"`
	Gas                  *big.Int      `json:"gas"`
	GasPrice             *big.Int      `json:"gasPrice"`
	Hash                 string        `json:"hash"`
	Input                string        `json:"input"`
	Nonce                *big.Int      `json:"nonce"`
	To                   string        `json:"to"`
	TransactionIndex     *big.Int      `json:"transactionIndex"`
	Value                *big.Int      `json:"value"`
	Type                 *big.Int      `json:"type"`
	V                    *big.Int      `json:"v"`
	R                    string        `json:"r"`
	S                    string        `json:"s"`
	MaxFeePerGas         *big.Int      `json:"maxFeePerGas"`
	MaxPriorityFeePerGas *big.Int      `json:"maxPriorityFeePerGas"`
	AccessList           []interface{} `json:"accessList"`
	ChainID              *big.Int      `json:"chainId"`
}

// helper alias for proper unmarshal of a transaction object
type transactionAlias Transaction

type transactionRaw struct {
	*transactionAlias
	BlockNumber          string `json:"blockNumber"`
	Gas                  string `json:"gas"`
	GasPrice             string `json:"gasPrice"`
	Nonce                string `json:"nonce"`
	TransactionIndex     string `json:"transactionIndex"`
	Value                string `json:"value"`
	Type                 string `json:"type"`
	V                    string `json:"v"`
	MaxFeePerGas         string `json:"maxFeePerGas"`
	MaxPriorityFeePerGas string `json:"maxPriorityFeePerGas"`
	ChainID              string `json:"chainId"`
}

func (t *Transaction) UnmarshalJSON(data []byte) error {
	txRaw := &transactionRaw{
		transactionAlias: (*transactionAlias)(t),
	}

	if err := json.Unmarshal(data, txRaw); err != nil {
		return fmt.Errorf("error unmarshal base tx data. %w", err)
	}

	t.BlockNumber = hexToIntMust(txRaw.BlockNumber)
	t.Gas = hexToIntMust(txRaw.Gas)
	t.GasPrice = hexToIntMust(txRaw.GasPrice)
	t.Nonce = hexToIntMust(txRaw.Nonce)
	t.TransactionIndex = hexToIntMust(txRaw.TransactionIndex)
	t.Value = hexToIntMust(txRaw.Value)
	t.Type = hexToIntMust(txRaw.Type)
	t.V = hexToIntMust(txRaw.V)
	t.MaxFeePerGas = hexToIntMust(txRaw.MaxFeePerGas)
	t.MaxPriorityFeePerGas = hexToIntMust(txRaw.MaxPriorityFeePerGas)
	t.ChainID = hexToIntMust(txRaw.ChainID)

	return nil
}

func hexToIntMust(s string) *big.Int {
	i, err := hexToInt(s)
	if err != nil {
		panic(err)
	}

	return i
}

func hexToInt(s string) (*big.Int, error) {
	bi := new(big.Int)

	if len(s) < 2 {
		return bi, nil
	}

	if _, ok := bi.SetString(s[2:], 16); !ok {
		return nil, fmt.Errorf("error invalid hex number")
	}

	return bi, nil
}
