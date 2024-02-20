// Copyright 2021 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package store

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	"github.com/sygmaprotocol/sygma-core/store"
	"github.com/syndtr/goleveldb/leveldb"
)

type BlockStore struct {
	db store.KeyValueReaderWriter
}

func NewBlockStore(db store.KeyValueReaderWriter) *BlockStore {
	return &BlockStore{
		db: db,
	}
}

// StoreBlock stores latest block number per route
func (ns *BlockStore) StoreBlock(sourceDomainID uint8, destinationDomainID uint8, blockNumber *big.Int) error {
	key := bytes.Buffer{}
	keyS := fmt.Sprintf("source:%d:destination:%d:blockNumber", sourceDomainID, destinationDomainID)
	key.WriteString(keyS)
	err := ns.db.SetByKey(key.Bytes(), blockNumber.Bytes())
	if err != nil {
		return err
	}

	return nil
}

// LatestBlock returns the latest block indexer per router
func (ns *BlockStore) LatestBlock(sourceDomainID uint8, destinationDomainID uint8) (*big.Int, error) {
	key := bytes.Buffer{}
	keyS := fmt.Sprintf("source:%d:destination:%d:blockNumber", sourceDomainID, destinationDomainID)
	key.WriteString(keyS)

	v, err := ns.db.GetByKey(key.Bytes())
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return big.NewInt(0), nil
		}
		return nil, err
	}

	block := big.NewInt(0).SetBytes(v)
	return block, nil
}
