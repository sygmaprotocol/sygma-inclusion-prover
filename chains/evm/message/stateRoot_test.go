// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package message_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/attestantio/go-eth2-client/api"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/deneb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/suite"
	evmMessage "github.com/sygmaprotocol/sygma-core/relayer/message"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/listener/events"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/message"
	"github.com/sygmaprotocol/sygma-inclusion-prover/mock"
	"go.uber.org/mock/gomock"
)

func readFromChannel(msgChan chan []*evmMessage.Message) ([]*evmMessage.Message, error) {
	select {
	case msgs := <-msgChan:
		return msgs, nil
	default:
		return make([]*evmMessage.Message, 0), fmt.Errorf("no message sent")
	}
}

type StateRootHandlerTestSuite struct {
	suite.Suite

	stateRootHandler *message.StateRootHandler

	msgChan          chan []*evmMessage.Message
	mockClient       *mock.MockClient
	mockBlockStorer  *mock.MockBlockStorer
	mockBlockFetcher *mock.MockBlockFetcher
	sourceDomain     uint8
	slotIndex        uint8
	routerAddress    common.Address
}

func TestRunConfigTestSuite(t *testing.T) {
	suite.Run(t, new(StateRootHandlerTestSuite))
}

func (s *StateRootHandlerTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.mockClient = mock.NewMockClient(ctrl)
	s.mockBlockFetcher = mock.NewMockBlockFetcher(ctrl)
	s.mockBlockStorer = mock.NewMockBlockStorer(ctrl)
	s.msgChan = make(chan []*evmMessage.Message, 10)
	s.sourceDomain = 1
	s.slotIndex = 2
	s.routerAddress = common.HexToAddress("0xa83114A443dA1CecEFC50368531cACE9F37fCCcb")
	s.stateRootHandler = message.NewStateRootHandler(
		s.mockBlockFetcher,
		s.mockBlockStorer,
		s.mockClient,
		s.routerAddress,
		s.msgChan,
		s.sourceDomain,
		s.slotIndex,
		[]string{"0x0000000000000000000000000000000000000000000000000000000000000500"},
	)
}

func (s *StateRootHandlerTestSuite) Test_HandleEvents_InvalidBlock() {
	s.mockBlockFetcher.EXPECT().SignedBeaconBlock(context.Background(), &api.SignedBeaconBlockOpts{
		Block: "10",
	}).Return(nil, fmt.Errorf("error"))

	_, err := s.stateRootHandler.HandleMessage(message.NewEvmStateRootMessage(2, s.sourceDomain, message.StateRootData{
		Slot: big.NewInt(10),
	}))

	s.NotNil(err)
}

func (s *StateRootHandlerTestSuite) Test_HandleEvents_Invalidstore() {
	s.mockBlockFetcher.EXPECT().SignedBeaconBlock(context.Background(), &api.SignedBeaconBlockOpts{
		Block: "10",
	}).Return(&api.Response[*spec.VersionedSignedBeaconBlock]{
		Data: &spec.VersionedSignedBeaconBlock{
			Deneb: &deneb.SignedBeaconBlock{
				Message: &deneb.BeaconBlock{
					Body: &deneb.BeaconBlockBody{
						ExecutionPayload: &deneb.ExecutionPayload{
							BlockNumber: 100,
						},
					},
				},
			},
		},
	}, nil)
	s.mockBlockStorer.EXPECT().LatestBlock(s.sourceDomain, uint8(2)).Return(nil, fmt.Errorf("error"))

	_, err := s.stateRootHandler.HandleMessage(message.NewEvmStateRootMessage(2, s.sourceDomain, message.StateRootData{
		Slot: big.NewInt(10),
	}))

	s.NotNil(err)
}

func (s *StateRootHandlerTestSuite) Test_HandleEvents_NoDeposits() {
	s.mockBlockFetcher.EXPECT().SignedBeaconBlock(context.Background(), &api.SignedBeaconBlockOpts{
		Block: "10",
	}).Return(&api.Response[*spec.VersionedSignedBeaconBlock]{
		Data: &spec.VersionedSignedBeaconBlock{
			Deneb: &deneb.SignedBeaconBlock{
				Message: &deneb.BeaconBlock{
					Body: &deneb.BeaconBlockBody{
						ExecutionPayload: &deneb.ExecutionPayload{
							BlockNumber: 100,
						},
					},
				},
			},
		},
	}, nil)
	s.mockBlockStorer.EXPECT().LatestBlock(s.sourceDomain, uint8(2)).Return(big.NewInt(80), nil)
	s.mockBlockStorer.EXPECT().StoreBlock(s.sourceDomain, uint8(2), big.NewInt(100)).Return(nil)
	s.mockClient.EXPECT().FetchEventLogs(context.Background(), s.routerAddress, string(events.DepositSig), big.NewInt(80), big.NewInt(100))

	_, err := s.stateRootHandler.HandleMessage(message.NewEvmStateRootMessage(2, s.sourceDomain, message.StateRootData{
		Slot: big.NewInt(10),
	}))

	s.Nil(err)
	_, err = readFromChannel(s.msgChan)
	s.NotNil(err)
}

func (s *StateRootHandlerTestSuite) Test_HandleEvents_ValidDeposits() {
	s.mockBlockFetcher.EXPECT().SignedBeaconBlock(context.Background(), &api.SignedBeaconBlockOpts{
		Block: "10",
	}).Return(&api.Response[*spec.VersionedSignedBeaconBlock]{
		Data: &spec.VersionedSignedBeaconBlock{
			Deneb: &deneb.SignedBeaconBlock{
				Message: &deneb.BeaconBlock{
					Body: &deneb.BeaconBlockBody{
						ExecutionPayload: &deneb.ExecutionPayload{
							BlockNumber: 100,
						},
					},
				},
			},
		},
	}, nil)
	s.mockBlockStorer.EXPECT().LatestBlock(s.sourceDomain, uint8(2)).Return(big.NewInt(80), nil)
	s.mockBlockStorer.EXPECT().StoreBlock(s.sourceDomain, uint8(2), big.NewInt(100)).Return(nil)
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
		func(ctx context.Context, target *message.AccountProof, rpcMethod string, args ...interface{}) error {
			*target = message.AccountProof{
				AccountProof: []string{"1"},
				StorageProof: []message.StorageProof{
					{
						Proof: []string{"2"},
					},
				},
			}
			return nil
		}).Times(2)

	prop, err := s.stateRootHandler.HandleMessage(message.NewEvmStateRootMessage(2, s.sourceDomain, message.StateRootData{
		Slot: big.NewInt(10),
	}))

	s.Nil(prop)
	s.Nil(err)
	msgs, err := readFromChannel(s.msgChan)
	s.Nil(err)
	s.Equal(len(msgs), 2)
	s.Equal(msgs[0].Destination, uint8(2))
	s.Equal(msgs[1].Destination, uint8(2))
	_, err = readFromChannel(s.msgChan)
	s.NotNil(err)
}
