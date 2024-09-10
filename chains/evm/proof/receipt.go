// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package proof

import (
	"bytes"
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/trie/trienode"
)

type TransactionFetcher interface {
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error)
}

type ReceiptProver struct {
	txFetcher TransactionFetcher
}

func NewReceiptProver(txFetcher TransactionFetcher) *ReceiptProver {
	return &ReceiptProver{
		txFetcher: txFetcher,
	}
}

// ReceiptProof calculates the Ethereum receipt proof for the specified transaction hash.
// The root hash of the proof is the block `receiptRoof` field.
func (p *ReceiptProver) ReceiptProof(txHash common.Hash) ([][]byte, error) {
	receipt, err := p.txFetcher.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		return nil, err
	}

	siblings, err := p.siblings(receipt.BlockHash)
	if err != nil {
		return nil, err
	}

	trie, err := p.trie(siblings)
	if err != nil {
		return nil, err
	}

	key, err := rlp.EncodeToBytes(receipt.TransactionIndex)
	if err != nil {
		return nil, err
	}

	var proofList trienode.ProofList
	err = trie.Prove(key, &proofList)
	if err != nil {
		return nil, err
	}

	proof := make([][]byte, len(proofList))
	for i, e := range proofList {
		proof[i] = e
	}
	return proof, nil
}

func (p *ReceiptProver) trie(siblings []*types.Receipt) (*trie.Trie, error) {
	memDB := rawdb.NewMemoryDatabase()
	db := trie.NewDatabase(memDB, nil)
	trie, err := trie.New(&trie.ID{}, db)
	if err != nil {
		return nil, err
	}

	for _, sibling := range siblings {
		key, err := rlp.EncodeToBytes(sibling.TransactionIndex)
		if err != nil {
			return nil, err
		}

		var buffer bytes.Buffer
		err = sibling.EncodeRLP(&buffer)
		if err != nil {
			return nil, err
		}

		if sibling.Type == 0 {
			err = trie.Update(key, buffer.Bytes())
			if err != nil {
				return nil, err
			}
		} else {
			err = trie.Update(key, buffer.Bytes()[3:])
			if err != nil {
				return nil, err
			}
		}
	}

	return trie, nil
}

func (p *ReceiptProver) siblings(blockHash common.Hash) ([]*types.Receipt, error) {
	block, err := p.txFetcher.BlockByHash(context.Background(), blockHash)
	if err != nil {
		return nil, err
	}

	siblings := make([]*types.Receipt, len(block.Transactions()))
	for i, tx := range block.Transactions() {
		receipt, err := p.txFetcher.TransactionReceipt(context.Background(), tx.Hash())
		if err != nil {
			return nil, err
		}
		siblings[i] = receipt
	}

	return siblings, nil
}
