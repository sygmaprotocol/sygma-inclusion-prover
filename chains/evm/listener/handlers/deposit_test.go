// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package handlers_test

import (
	"context"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/suite"
	evmMessage "github.com/sygmaprotocol/sygma-core/relayer/message"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/listener/events"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/listener/handlers"
	"github.com/sygmaprotocol/sygma-inclusion-prover/mock"
	"go.uber.org/mock/gomock"
)

type DepositHandlerTestSuite struct {
	suite.Suite

	depositHandler *handlers.DepositEventHandler

	msgChan           chan []*evmMessage.Message
	mockClient        *mock.MockClient
	mockBlockStorer   *mock.MockBlockStorer
	mockBlockFetcher  *mock.MockBlockFetcher
	sourceDomain      uint8
	destinationDomain uint8
	slotIndex         uint8
	routerAddress     common.Address
}

func TestRunDepositHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(DepositHandlerTestSuite))
}

func (s *DepositHandlerTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.mockClient = mock.NewMockClient(ctrl)
	s.mockBlockFetcher = mock.NewMockBlockFetcher(ctrl)
	s.mockBlockStorer = mock.NewMockBlockStorer(ctrl)
	s.msgChan = make(chan []*evmMessage.Message, 10)
	s.sourceDomain = 1
	s.destinationDomain = 2
	s.slotIndex = 2
	s.routerAddress = common.HexToAddress("0xa83114A443dA1CecEFC50368531cACE9F37fCCcb")
	s.depositHandler = handlers.NewDepositEventHandler(
		s.sourceDomain,
		s.mockClient,
		s.routerAddress,
		s.slotIndex,
		[]string{"0x0000000000000000000000000000000000000000000000000000000000000500"},
		s.msgChan,
	)
}

func (s *DepositHandlerTestSuite) Test_HandleEvents_NoDeposits() {
	s.mockClient.EXPECT().FetchEventLogs(context.Background(), s.routerAddress, string(events.DepositSig), big.NewInt(80), big.NewInt(100))

	err := s.depositHandler.HandleDeposits(s.destinationDomain, big.NewInt(80), big.NewInt(100), big.NewInt(150))

	s.Nil(err)
	_, err = readFromChannel(s.msgChan)
	s.NotNil(err)
}

func (s *DepositHandlerTestSuite) Test_HandleEvents_ValidDeposits() {
	validDepositData, _ := hex.DecodeString("000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000")
	s.mockClient.EXPECT().FetchEventLogs(context.Background(), s.routerAddress, string(events.DepositSig), big.NewInt(80), big.NewInt(100)).Return(
		[]types.Log{
			{
				Data: validDepositData,
				Topics: []common.Hash{
					{},
					common.HexToHash("0xd68eb9b5E135b96c1Af165e1D8c4e2eB0E1CE4CD"),
				},
			},
			{
				Data: validDepositData,
				Topics: []common.Hash{
					{},
					common.HexToHash("0xd68eb9b5E135b96c1Af165e1D8c4e2eB0E1CE4CD"),
				},
			},
		},
		nil,
	)

	expectedSlotKey := "0x9fffbb9e89029b0baa965344cab51a6b05088fdd0a0df87ecf7dddfe9e4c7b74"
	s.mockClient.EXPECT().CallContext(context.Background(), gomock.Any(), "eth_getProof", s.routerAddress, []string{expectedSlotKey}, hexutil.EncodeBig(big.NewInt(100))).DoAndReturn(
		func(ctx context.Context, target *handlers.AccountProof, rpcMethod string, args ...interface{}) error {
			*target = handlers.AccountProof{
				AccountProof: []string{"1"},
				StorageProof: []handlers.StorageProof{
					{
						Proof: []string{"2"},
					},
				},
			}
			return nil
		}).Times(2)

	err := s.depositHandler.HandleDeposits(s.destinationDomain, big.NewInt(80), big.NewInt(100), big.NewInt(150))

	s.Nil(err)
	msgs, err := readFromChannel(s.msgChan)
	s.Nil(err)
	s.Equal(len(msgs), 2)
	s.Equal(msgs[0].Destination, uint8(2))
	s.Equal(msgs[1].Destination, uint8(2))
	_, err = readFromChannel(s.msgChan)
	s.NotNil(err)
}

func (s *DepositHandlerTestSuite) Test_HandleEvents_ValidDeposits_LargeBlockRange() {
	validDepositData, _ := hex.DecodeString("000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000")
	s.mockClient.EXPECT().FetchEventLogs(context.Background(), s.routerAddress, string(events.DepositSig), big.NewInt(80), big.NewInt(1080)).Return(
		[]types.Log{
			{
				Data: validDepositData,
				Topics: []common.Hash{
					{},
					common.HexToHash("0xd68eb9b5E135b96c1Af165e1D8c4e2eB0E1CE4CD"),
				},
			},
		},
		nil,
	)
	s.mockClient.EXPECT().FetchEventLogs(context.Background(), s.routerAddress, string(events.DepositSig), big.NewInt(1081), big.NewInt(2081)).Return(
		[]types.Log{},
		nil,
	)
	s.mockClient.EXPECT().FetchEventLogs(context.Background(), s.routerAddress, string(events.DepositSig), big.NewInt(2082), big.NewInt(2432)).Return(
		[]types.Log{
			{
				Data: validDepositData,
				Topics: []common.Hash{
					{},
					common.HexToHash("0xd68eb9b5E135b96c1Af165e1D8c4e2eB0E1CE4CD"),
				},
			},
			{
				Data: validDepositData,
				Topics: []common.Hash{
					{},
					common.HexToHash("0xd68eb9b5E135b96c1Af165e1D8c4e2eB0E1CE4CD"),
				},
			},
		},
		nil,
	)

	expectedSlotKey := "0x9fffbb9e89029b0baa965344cab51a6b05088fdd0a0df87ecf7dddfe9e4c7b74"
	s.mockClient.EXPECT().CallContext(context.Background(), gomock.Any(), "eth_getProof", s.routerAddress, []string{expectedSlotKey}, hexutil.EncodeBig(big.NewInt(2432))).DoAndReturn(
		func(ctx context.Context, target *handlers.AccountProof, rpcMethod string, args ...interface{}) error {
			*target = handlers.AccountProof{
				AccountProof: []string{"1"},
				StorageProof: []handlers.StorageProof{
					{
						Proof: []string{"2"},
					},
				},
			}
			return nil
		}).Times(3)

	err := s.depositHandler.HandleDeposits(s.destinationDomain, big.NewInt(80), big.NewInt(2432), big.NewInt(150))

	s.Nil(err)
	msgs, err := readFromChannel(s.msgChan)
	s.Nil(err)
	s.Equal(len(msgs), 3)
	s.Equal(msgs[0].Destination, uint8(2))
	s.Equal(msgs[1].Destination, uint8(2))
	s.Equal(msgs[2].Destination, uint8(2))
	_, err = readFromChannel(s.msgChan)
	s.NotNil(err)
}
