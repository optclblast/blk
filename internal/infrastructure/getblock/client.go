package getblock

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/optclblast/blk/internal/entities"
	"github.com/ybbus/jsonrpc/v3"
)

// B ase getblock API url
const baseURL = "https://go.getblock.io/"

// JSON rpc client
type Client struct {
	log *slog.Logger
	cc  jsonrpc.RPCClient
}

// NewClient returns a new GetBlock JSON rpc client
func NewClient(
	log *slog.Logger,
	accessToken string,
) *Client {
	return &Client{
		log: log,
		cc:  jsonrpc.NewClient(baseURL + accessToken),
	}
}

// LastBlockNumber returns a last block number
func (c *Client) LastBlockNumber(ctx context.Context) (entities.BlockNumber, error) {
	const method = "eth_blockNumber"

	res, err := c.cc.Call(ctx, method)
	if err != nil {
		if _, ok := err.(*jsonrpc.HTTPError); ok {
			return "", ErrorRateLimitExceeded
		}

		return "", fmt.Errorf("error fetch last block number. %w", err)
	} else if res.Error != nil {
		return "", fmt.Errorf("error fetch last block number. %w", res.Error)
	}

	response, err := res.GetString()
	if err != nil {
		return "", fmt.Errorf("error parse response. %w", err)
	}

	c.log.Debug(
		"last block number",
		slog.String("method", method),
		slog.String("resp", response),
	)

	return entities.BlockNumber(response), nil
}

// BlockInfoByNumber returns an info about block by its number
func (c *Client) BlockInfoByNumber(ctx context.Context, num entities.BlockNumber) (*entities.Block, error) {
	const method = "eth_getBlockByNumber"

	res, err := c.cc.Call(ctx, method, num, true)
	if err != nil {
		if _, ok := err.(*jsonrpc.HTTPError); ok {
			return nil, ErrorRateLimitExceeded
		}

		return nil, fmt.Errorf("error fetch block info. %w", err)
	} else if res.Error != nil {
		return nil, fmt.Errorf("error fetch block info. %w", res.Error)
	}

	out := new(entities.Block)

	if err := res.GetObject(out); err != nil {
		return nil, fmt.Errorf("error marshal response body into block object. %w", err)
	}

	return out, nil
}
