// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package message

import (
	"context"
	"math/big"

	"github.com/attestantio/go-eth2-client/api"
	"github.com/attestantio/go-eth2-client/spec"
	ethereumABI "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog/log"
	"github.com/sygmaprotocol/sygma-core/relayer/message"
	"github.com/sygmaprotocol/sygma-core/relayer/proposal"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/listener/events"
)

const (
	EVMStateRootMessage message.MessageType = "EVMStateRootMessage"
)

type StateRootData struct {
	StateRoot [32]byte
	Slot      *big.Int
}

func NewEvmStateRootMessage(source uint8, destination uint8, stateRoot StateRootData) *message.Message {
	return &message.Message{
		Source:      source,
		Destination: destination,
		Data:        stateRoot,
		Type:        EVMStateRootMessage,
	}
}

type BlockFetcher interface {
	SignedBeaconBlock(ctx context.Context, opts *api.SignedBeaconBlockOpts) (*api.Response[*spec.VersionedSignedBeaconBlock], error)
}

type EventFetcher interface {
	FetchEventLogs(ctx context.Context, contractAddress common.Address, event string, startBlock *big.Int, endBlock *big.Int) ([]types.Log, error)
}

type StateRootHandler struct {
	blockFetcher  BlockFetcher
	eventFetcher  EventFetcher
	routerAddress common.Address
	routerABI     ethereumABI.ABI
	domainID      uint8

	msgChan chan []*message.Message

	// TODO
	// latest block number for the source-destination pair
	latestBlockNumber map[uint8]map[uint8]uint64
}

func (h *StateRootHandler) HandleMessage(m *message.Message) (*proposal.Proposal, error) {
	log.Debug().Uint8("domainID", m.Destination).Msgf("Received rotate message from domain %d", m.Source)

	stateRoot := m.Data.(StateRootData)
	block, err := h.blockFetcher.SignedBeaconBlock(context.Background(), &api.SignedBeaconBlockOpts{
		Block: stateRoot.Slot.String(),
	})
	if err != nil {
		return nil, err
	}
	blockNumber := big.NewInt(int64(block.Data.Deneb.Message.Body.ExecutionPayload.BlockNumber))
	deposits, err := h.fetchDeposits(blockNumber, blockNumber)
	if err != nil {
		return nil, err
	}
	msgs := []*message.Message{}
	for _, d := range deposits {
		msgs = append(msgs, NewEVMDepositMessage(h.domainID, d.DestinationDomainID, d))
	}

	h.msgChan <- msgs
	return nil, nil
}

func (h *StateRootHandler) fetchDeposits(startBlock *big.Int, endBlock *big.Int) ([]*events.Deposit, error) {
	logs, err := h.eventFetcher.FetchEventLogs(context.Background(), h.routerAddress, string(events.DepositSig), startBlock, endBlock)
	if err != nil {
		return nil, err
	}

	deposits := make([]*events.Deposit, 0)
	for _, dl := range logs {
		d, err := h.unpackDeposit(dl.Data)
		if err != nil {
			log.Error().Msgf("Failed unpacking deposit event log: %v", err)
			continue
		}

		log.Debug().Msgf("Found deposit log in block: %d, TxHash: %s, contractAddress: %s, sender: %s", dl.BlockNumber, dl.TxHash, dl.Address, d.SenderAddress)
		deposits = append(deposits, d)
	}

	return deposits, nil
}

func (h *StateRootHandler) unpackDeposit(data []byte) (*events.Deposit, error) {
	var d events.Deposit
	err := h.routerABI.UnpackIntoInterface(&d, "Deposit", data)
	if err != nil {
		return &events.Deposit{}, err
	}

	return &d, nil
}
