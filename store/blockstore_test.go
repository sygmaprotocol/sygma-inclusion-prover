// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package store_test

import (
	"errors"
	"math/big"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/sygmaprotocol/sygma-inclusion-prover/mock"
	"github.com/sygmaprotocol/sygma-inclusion-prover/store"
	"github.com/syndtr/goleveldb/leveldb"
	"go.uber.org/mock/gomock"
)

type BlockStoreTestSuite struct {
	suite.Suite
	BlockStore           *store.BlockStore
	keyValueReaderWriter *mock.MockKeyValueReaderWriter
}

func TestRunBlockStoreTestSuite(t *testing.T) {
	suite.Run(t, new(BlockStoreTestSuite))
}

func (s *BlockStoreTestSuite) SetupTest() {
	gomockController := gomock.NewController(s.T())
	s.keyValueReaderWriter = mock.NewMockKeyValueReaderWriter(gomockController)
	s.BlockStore = store.NewBlockStore(s.keyValueReaderWriter)
}

func (s *BlockStoreTestSuite) Test_StoreBlock_FailedStore() {
	key := "source:1:destination:2:blockNumber"
	s.keyValueReaderWriter.EXPECT().SetByKey([]byte(key), []byte{5}).Return(errors.New("error"))

	err := s.BlockStore.StoreBlock(1, 2, big.NewInt(5))

	s.NotNil(err)
}

func (s *BlockStoreTestSuite) Test_StoreBlock_SuccessfulStore() {
	key := "source:1:destination:2:blockNumber"
	s.keyValueReaderWriter.EXPECT().SetByKey([]byte(key), []byte{5}).Return(nil)

	err := s.BlockStore.StoreBlock(1, 2, big.NewInt(5))

	s.Nil(err)
}

func (s *BlockStoreTestSuite) Test_LatestBlock_FailedFetch() {
	key := "source:1:destination:2:blockNumber"
	s.keyValueReaderWriter.EXPECT().GetByKey([]byte(key)).Return(nil, errors.New("error"))

	_, err := s.BlockStore.LatestBlock(1, 2)

	s.NotNil(err)
}

func (s *BlockStoreTestSuite) Test_LatestBlock_NotFound() {
	key := "source:1:destination:2:blockNumber"
	s.keyValueReaderWriter.EXPECT().GetByKey([]byte(key)).Return(nil, leveldb.ErrNotFound)

	block, err := s.BlockStore.LatestBlock(1, 2)

	s.Equal(block, big.NewInt(0))
	s.Nil(err)
}

func (s *BlockStoreTestSuite) Test_LatestBlock_Successful() {
	key := "source:1:destination:2:blockNumber"
	s.keyValueReaderWriter.EXPECT().GetByKey([]byte(key)).Return([]byte{5}, nil)

	block, err := s.BlockStore.LatestBlock(1, 2)

	s.Equal(block, big.NewInt(5))
	s.Nil(err)
}
