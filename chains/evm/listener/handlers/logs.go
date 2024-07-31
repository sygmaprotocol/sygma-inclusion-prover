package handlers

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const (
	MAX_BLOCK_RANGE int64 = 1000
)

// fetchLogs calls fetch event logs multiple times with a predefined block range to prevent
// rpc errors when the block range is too large
func fetchLogs(client Client, startBlock, endBlock *big.Int, contract common.Address, eventSignature string) ([]types.Log, error) {
	allLogs := make([]types.Log, 0)
	for startBlock.Cmp(endBlock) < 0 {
		rangeEnd := new(big.Int).Add(startBlock, big.NewInt(MAX_BLOCK_RANGE))
		if rangeEnd.Cmp(endBlock) > 0 {
			rangeEnd = endBlock
		}

		logs, err := client.FetchEventLogs(context.Background(), contract, eventSignature, startBlock, rangeEnd)
		if err != nil {
			return nil, err
		}
		allLogs = append(allLogs, logs...)
		startBlock = new(big.Int).Add(rangeEnd, big.NewInt(1))
	}

	return allLogs, nil
}
