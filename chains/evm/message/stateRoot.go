// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package message

import (
	"context"
	"encoding/hex"
	"math/big"

	"github.com/attestantio/go-eth2-client/api"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/rs/zerolog/log"
	"github.com/sygmaprotocol/sygma-core/relayer/message"
	"github.com/sygmaprotocol/sygma-core/relayer/proposal"
)

const (
	EVMStateRootMessage message.MessageType = "EVMStateRootMessage"
	MAX_BLOCK_RANGE     int64               = 1000
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

type BlockStorer interface {
	StoreBlock(sourceDomainID uint8, destinationDomainID uint8, blockNumber *big.Int) error
	LatestBlock(sourceDomainID uint8, destinationDomainID uint8) (*big.Int, error)
}

type DepositHandler interface {
	HandleDeposits(destination uint8, startBlock *big.Int, endBlock *big.Int, slot *big.Int) error
}

type StateRootHandler struct {
	blockFetcher   BlockFetcher
	blockStorer    BlockStorer
	depositHandler DepositHandler
	startBlock     *big.Int
	domainID       uint8
}

func NewStateRootHandler(
	domainID uint8,
	depositHandler DepositHandler,
	blockFetcher BlockFetcher,
	blockStorer BlockStorer,
	startBlock *big.Int,
) *StateRootHandler {
	return &StateRootHandler{
		blockFetcher: blockFetcher,
		blockStorer:  blockStorer,
		domainID:     domainID,
		startBlock:   startBlock,
	}
}

// HandleMessage fetches deposits for the given state root and submits a transfer message
// with execution state proofs per transfer
func (h *StateRootHandler) HandleMessage(m *message.Message) (*proposal.Proposal, error) {
	stateRoot := m.Data.(StateRootData)
	log.Debug().Uint8(
		"domainID", m.Destination).Str(
		"stateRoot", hex.EncodeToString(stateRoot.StateRoot[:]),
	).Msgf("Received state root message from domain %d", m.Source)
	block, err := h.blockFetcher.SignedBeaconBlock(context.Background(), &api.SignedBeaconBlockOpts{
		Block: stateRoot.Slot.String(),
	})
	if err != nil {
		return nil, err
	}
	startBlock, err := h.blockStorer.LatestBlock(h.domainID, m.Source)
	if err != nil {
		return nil, err
	}
	if startBlock.Cmp(big.NewInt(0)) == 0 {
		startBlock = h.startBlock
	}
	endBlock := big.NewInt(int64(block.Data.Deneb.Message.Body.ExecutionPayload.BlockNumber))

	err = h.depositHandler.HandleDeposits(m.Source, startBlock, endBlock, stateRoot.Slot)
	if err != nil {
		return nil, err
	}

	err = h.blockStorer.StoreBlock(h.domainID, m.Source, endBlock)
	if err != nil {
		log.Err(err).Msgf("Failed saving latest block for %d-%d", h.domainID, m.Source)
	}
	return nil, nil
}
