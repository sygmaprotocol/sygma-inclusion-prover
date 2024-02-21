// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package handlers

import (
	"context"
	"encoding/hex"
	"math/big"
	"strings"

	ethereumABI "github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog/log"
	"github.com/sygmaprotocol/sygma-core/relayer/message"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/abi"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/listener/events"
	evmMessage "github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/message"
)

type EventFetcher interface {
	FetchEventLogs(ctx context.Context, contractAddress common.Address, event string, startBlock *big.Int, endBlock *big.Int) ([]types.Log, error)
}

type StateRootEventHandler struct {
	msgChan chan []*message.Message

	eventFetcher        EventFetcher
	stateRootAddress    common.Address
	stateRootStorageABI ethereumABI.ABI

	domainID uint8
}

func NewStateRootEventHandler(
	msgChan chan []*message.Message,
	eventFetcher EventFetcher,
	stateRootAddress common.Address,
	domainID uint8,
) *StateRootEventHandler {
	stateRootStorageABI, _ := ethereumABI.JSON(strings.NewReader(abi.StateRootStorageABI))
	return &StateRootEventHandler{
		eventFetcher:        eventFetcher,
		msgChan:             msgChan,
		stateRootAddress:    stateRootAddress,
		domainID:            domainID,
		stateRootStorageABI: stateRootStorageABI,
	}
}

// HandleEvents fetches state root submitted events and sends message to origin domain
func (h *StateRootEventHandler) HandleEvents(startBlock *big.Int, endBlock *big.Int) error {
	stateRoots, err := h.fetchStateRoots(startBlock, endBlock)
	if err != nil {
		return err
	}

	for _, sr := range stateRoots {
		log.Debug().Uint8("domainID", h.domainID).Msgf("Sending state root message to domain %d", sr.SourceDomainID)
		h.msgChan <- []*message.Message{evmMessage.NewEvmStateRootMessage(h.domainID, sr.SourceDomainID, evmMessage.StateRootData{
			StateRoot: sr.StateRoot,
			Slot:      sr.Slot,
		})}
	}
	return nil
}

func (h *StateRootEventHandler) fetchStateRoots(startBlock *big.Int, endBlock *big.Int) ([]*events.StateRootSubmitted, error) {
	logs, err := h.eventFetcher.FetchEventLogs(context.Background(), h.stateRootAddress, string(events.StateRootSubmittedSig), startBlock, endBlock)
	if err != nil {
		return nil, err
	}

	stateRoots := make([]*events.StateRootSubmitted, 0)
	for _, l := range logs {
		sr, err := h.unpackStateRoot(l.Data)
		if err != nil {
			log.Error().Msgf("Failed unpacking state root event log: %v", err)
			continue
		}
		log.Debug().Uint8("domainID", h.domainID).Uint8("sourceDomainID", sr.SourceDomainID).Msgf(
			"Found state root %s in block: %d", hex.EncodeToString(sr.StateRoot[:]), l.BlockNumber)
		stateRoots = append(stateRoots, sr)
	}

	return stateRoots, nil
}

func (h *StateRootEventHandler) unpackStateRoot(data []byte) (*events.StateRootSubmitted, error) {
	var sr events.StateRootSubmitted
	err := h.stateRootStorageABI.UnpackIntoInterface(&sr, "StateRootSubmitted", data)
	if err != nil {
		return &events.StateRootSubmitted{}, err
	}

	return &sr, nil
}
