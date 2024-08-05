// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package message_test

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/attestantio/go-eth2-client/api"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/deneb"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/stretchr/testify/suite"
	evmMessage "github.com/sygmaprotocol/sygma-core/relayer/message"
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

	msgChan            chan []*evmMessage.Message
	mockBlockStorer    *mock.MockBlockStorer
	mockBlockFetcher   *mock.MockBlockFetcher
	mockDepositHandler *mock.MockDepositHandler
	mockHashiHandler   *mock.MockHashiHandler
	sourceDomain       uint8
}

func TestRunConfigTestSuite(t *testing.T) {
	suite.Run(t, new(StateRootHandlerTestSuite))
}

func (s *StateRootHandlerTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.mockBlockFetcher = mock.NewMockBlockFetcher(ctrl)
	s.mockBlockStorer = mock.NewMockBlockStorer(ctrl)
	s.mockDepositHandler = mock.NewMockDepositHandler(ctrl)
	s.mockHashiHandler = mock.NewMockHashiHandler(ctrl)
	s.msgChan = make(chan []*evmMessage.Message, 10)
	s.sourceDomain = 1
	s.stateRootHandler = message.NewStateRootHandler(
		s.sourceDomain,
		s.mockDepositHandler,
		s.mockHashiHandler,
		s.mockBlockFetcher,
		s.mockBlockStorer,
		big.NewInt(50),
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

func (s *StateRootHandlerTestSuite) Test_HandleEvents_MissingStartBlock() {
	s.mockBlockFetcher.EXPECT().SignedBeaconBlock(context.Background(), &api.SignedBeaconBlockOpts{
		Block: "1000",
	}).Return(&api.Response[*spec.VersionedSignedBeaconBlock]{
		Data: &spec.VersionedSignedBeaconBlock{
			Deneb: &deneb.SignedBeaconBlock{
				Message: &deneb.BeaconBlock{
					Slot: phase0.Slot(1000),
					Body: &deneb.BeaconBlockBody{
						ExecutionPayload: &deneb.ExecutionPayload{
							BlockNumber: 100,
						},
					},
				},
			},
		},
	}, nil)
	s.mockBlockStorer.EXPECT().LatestBlock(s.sourceDomain, uint8(2)).Return(big.NewInt(0), nil)
	s.mockBlockStorer.EXPECT().StoreBlock(s.sourceDomain, uint8(2), big.NewInt(100)).Return(nil)

	s.mockDepositHandler.EXPECT().HandleDeposits(uint8(2), big.NewInt(50), big.NewInt(100), big.NewInt(1000)).Return(nil)
	s.mockHashiHandler.EXPECT().HandleMessages(uint8(2), big.NewInt(50), big.NewInt(100), big.NewInt(1000)).Return(nil)

	_, err := s.stateRootHandler.HandleMessage(message.NewEvmStateRootMessage(2, s.sourceDomain, message.StateRootData{
		Slot: big.NewInt(1000),
	}))

	s.Nil(err)
}

func (s *StateRootHandlerTestSuite) Test_HandleEvents_ExistingStartBlock() {
	s.mockBlockFetcher.EXPECT().SignedBeaconBlock(context.Background(), &api.SignedBeaconBlockOpts{
		Block: "1000",
	}).Return(&api.Response[*spec.VersionedSignedBeaconBlock]{
		Data: &spec.VersionedSignedBeaconBlock{
			Deneb: &deneb.SignedBeaconBlock{
				Message: &deneb.BeaconBlock{
					Slot: phase0.Slot(1000),
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

	s.mockDepositHandler.EXPECT().HandleDeposits(uint8(2), big.NewInt(80), big.NewInt(100), big.NewInt(1000)).Return(nil)
	s.mockHashiHandler.EXPECT().HandleMessages(uint8(2), big.NewInt(80), big.NewInt(100), big.NewInt(1000)).Return(nil)

	_, err := s.stateRootHandler.HandleMessage(message.NewEvmStateRootMessage(2, s.sourceDomain, message.StateRootData{
		Slot: big.NewInt(1000),
	}))

	s.Nil(err)
}
