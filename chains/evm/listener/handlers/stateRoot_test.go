// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package handlers_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/suite"
	"github.com/sygmaprotocol/sygma-core/relayer/message"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/listener/events"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/listener/handlers"
	evmMessage "github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/message"
	"github.com/sygmaprotocol/sygma-inclusion-prover/mock"
	"go.uber.org/mock/gomock"
)

func readFromChannel(msgChan chan []*message.Message) ([]*message.Message, error) {
	select {
	case msgs := <-msgChan:
		return msgs, nil
	default:
		return make([]*message.Message, 0), fmt.Errorf("no message sent")
	}
}

func SliceTo32Bytes(in []byte) [32]byte {
	var res [32]byte
	copy(res[:], in)
	return res
}

type StateRootHandlerTestSuite struct {
	suite.Suite

	stateRootHandler *handlers.StateRootEventHandler

	msgChan          chan []*message.Message
	mockEventFetcher *mock.MockEventFetcher
	sourceDomain     uint8
	stateRootAddress common.Address
}

func TestRunConfigTestSuite(t *testing.T) {
	suite.Run(t, new(StateRootHandlerTestSuite))
}

func (s *StateRootHandlerTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.mockEventFetcher = mock.NewMockEventFetcher(ctrl)
	s.msgChan = make(chan []*message.Message, 10)
	s.sourceDomain = 1
	s.stateRootAddress = common.HexToAddress("0xa83114A443dA1CecEFC50368531cACE9F37fCCcb")
	s.stateRootHandler = handlers.NewStateRootEventHandler(
		s.msgChan,
		s.mockEventFetcher,
		s.stateRootAddress,
		s.sourceDomain)
}

func (s *StateRootHandlerTestSuite) Test_HandleEvents_FetchingArgsFails() {
	startBlock := big.NewInt(5)
	endBlock := big.NewInt(10)
	s.mockEventFetcher.EXPECT().FetchEventLogs(
		context.Background(), s.stateRootAddress, string(events.StateRootSubmittedSig), startBlock, endBlock,
	).Return([]types.Log{}, fmt.Errorf("error"))

	err := s.stateRootHandler.HandleEvents(startBlock, endBlock)
	s.NotNil(err)

	_, err = readFromChannel(s.msgChan)
	s.NotNil(err)
}

func (s *StateRootHandlerTestSuite) Test_HandleEvents_NoRootFound() {
	startBlock := big.NewInt(5)
	endBlock := big.NewInt(10)
	s.mockEventFetcher.EXPECT().FetchEventLogs(
		context.Background(), s.stateRootAddress, string(events.StateRootSubmittedSig), startBlock, endBlock,
	).Return([]types.Log{}, nil)

	err := s.stateRootHandler.HandleEvents(startBlock, endBlock)
	s.Nil(err)

	_, err = readFromChannel(s.msgChan)
	s.NotNil(err)
}

func (s *StateRootHandlerTestSuite) Test_HandleEvents_ValidRoots() {
	startBlock := big.NewInt(5)
	endBlock := big.NewInt(10)
	stateRootLogData, _ := hex.DecodeString("000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000f1060d3d6b9d842f0cd581c75d6d9d94e5da5fa1440708386c3ca4dba069cf5eb6abd")
	s.mockEventFetcher.EXPECT().FetchEventLogs(
		context.Background(), s.stateRootAddress, string(events.StateRootSubmittedSig), startBlock, endBlock,
	).Return([]types.Log{
		{
			Data: stateRootLogData,
		},
		{
			Data: stateRootLogData,
		},
	}, nil)

	err := s.stateRootHandler.HandleEvents(startBlock, endBlock)
	s.Nil(err)

	expectedStateRoot, _ := hex.DecodeString("D3D6B9D842F0CD581C75D6D9D94E5DA5FA1440708386C3CA4DBA069CF5EB6ABD")
	msg, err := readFromChannel(s.msgChan)
	s.Nil(err)
	s.Equal(msg, []*message.Message{evmMessage.NewEvmStateRootMessage(s.sourceDomain, 1, evmMessage.StateRootData{
		StateRoot: SliceTo32Bytes(expectedStateRoot),
		Slot:      big.NewInt(987232),
	})})
	msg, err = readFromChannel(s.msgChan)
	s.Nil(err)
	s.Equal(msg, []*message.Message{evmMessage.NewEvmStateRootMessage(s.sourceDomain, 1, evmMessage.StateRootData{
		StateRoot: SliceTo32Bytes(expectedStateRoot),
		Slot:      big.NewInt(987232),
	})})
	_, err = readFromChannel(s.msgChan)
	s.NotNil(err)
}
