// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package handlers_test

import (
	"context"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/attestantio/go-eth2-client/api"
	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/suite"
	evmMessage "github.com/sygmaprotocol/sygma-core/relayer/message"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/listener/events"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/listener/handlers"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/message"
	"github.com/sygmaprotocol/sygma-inclusion-prover/mock"
	"go.uber.org/mock/gomock"
)

type HashiHandlerTestSuite struct {
	suite.Suite

	hashiHandler *handlers.HashiEventHandler

	msgChan           chan []*evmMessage.Message
	mockClient        *mock.MockClient
	mockBeaconClient  *mock.MockBeaconClient
	mockReceiptProver *mock.MockReceiptProver
	mockRootProver    *mock.MockRootProver
	sourceDomain      uint8
	destinationDomain uint8
	yahoAddress       common.Address
}

func TestRunHashiHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HashiHandlerTestSuite))
}

func (s *HashiHandlerTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.mockClient = mock.NewMockClient(ctrl)
	s.mockBeaconClient = mock.NewMockBeaconClient(ctrl)
	s.mockReceiptProver = mock.NewMockReceiptProver(ctrl)
	s.mockRootProver = mock.NewMockRootProver(ctrl)
	s.msgChan = make(chan []*evmMessage.Message, 2)
	s.sourceDomain = 1
	s.destinationDomain = 2
	s.yahoAddress = common.HexToAddress("0xa83114A443dA1CecEFC50368531cACE9F37fCCcb")
	chainIDS := make(map[uint8]uint64)
	chainIDS[2] = 10200
	s.hashiHandler = handlers.NewHashiEventHandler(
		s.sourceDomain,
		s.mockClient,
		s.mockBeaconClient,
		s.mockReceiptProver,
		s.mockRootProver,
		s.yahoAddress,
		chainIDS,
		s.msgChan,
	)
}

func (s *HashiHandlerTestSuite) Test_HandleMessage_ValidMessage() {
	txHash := common.HexToHash("0x12345")
	messageData, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000027d800000000000000000000000000000000000000000000000000000000000000010000000000000000000000001c3a03d04c026b1f4b4208d2ce053c5686e6fb8d000000000000000000000000ba9165973963a6e5608f03b9648c34a737e48f68000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000001400000000000000000000000000000000000000000000000000000000000000180000000000000000000000000000000000000000000000000000000000000000b48656c6c6f20776f726c64000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000ba9165973963a6e5608f03b9648c34a737e48f68")
	s.mockClient.EXPECT().FetchEventLogs(context.Background(), s.yahoAddress, string(events.MessageDispatchedSig), big.NewInt(80), big.NewInt(100)).Return(
		[]types.Log{
			{
				Data:   messageData,
				TxHash: txHash,
			},
		},
		nil,
	)
	s.mockClient.EXPECT().BlockByHash(gomock.Any(), txHash).Return(types.NewBlock(&types.Header{
		ParentBeaconRoot: &common.Hash{},
	}, nil, nil, nil, nil), nil)
	s.mockClient.EXPECT().TransactionReceipt(gomock.Any(), txHash).Return(&types.Receipt{}, nil)
	s.mockBeaconClient.EXPECT().BeaconBlockHeader(gomock.Any(), gomock.Any()).Return(&api.Response[*apiv1.BeaconBlockHeader]{
		Data: &apiv1.BeaconBlockHeader{
			Header: &phase0.SignedBeaconBlockHeader{
				Message: &phase0.BeaconBlockHeader{
					Slot: phase0.Slot(121),
				},
			},
		},
	}, nil)
	s.mockReceiptProver.EXPECT().ReceiptProof(gomock.Any()).Return([][]byte{{1}}, nil)
	s.mockRootProver.EXPECT().ReceiptsRootProof(gomock.Any(), gomock.Any(), gomock.Any()).Return([][]byte{{2}}, nil)

	err := s.hashiHandler.HandleMessages(s.destinationDomain, big.NewInt(80), big.NewInt(100), big.NewInt(150))

	s.Nil(err)
	msgs, err := readFromChannel(s.msgChan)
	s.Nil(err)
	s.Equal(len(msgs), 1)
	s.Equal(msgs[0].Type, message.HashiMessage)
}
