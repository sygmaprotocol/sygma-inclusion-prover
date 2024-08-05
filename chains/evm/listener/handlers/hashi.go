// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package handlers

import (
	"context"
	"fmt"
	"math/big"

	"github.com/attestantio/go-eth2-client/api"
	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	ethereumABI "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sygmaprotocol/sygma-core/relayer/message"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/listener/events"
	evmMessage "github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/message"
)

type ReceiptProver interface {
	ReceiptProof(txHash common.Hash) ([][]byte, error)
}

type RootProver interface {
	ReceiptsRootProof(ctx context.Context, currentSlot *big.Int, targetSlot *big.Int) ([][]byte, error)
}

type BeaconClient interface {
	BeaconBlockHeader(
		ctx context.Context,
		opts *api.BeaconBlockHeaderOpts,
	) (
		*api.Response[*apiv1.BeaconBlockHeader],
		error,
	)
}

type HashiEventHandler struct {
	log           zerolog.Logger
	domainID      uint8
	yahoAddress   common.Address
	yahoABI       ethereumABI.ABI
	msgChan       chan []*message.Message
	receiptProver ReceiptProver
	rootProver    RootProver
	beaconClient  BeaconClient
	client        Client
	chainIDS      map[uint8]*big.Int
}

func NewHashiEventHandler(
	domainID uint8,
	client Client,
	yahoAddress common.Address,
	msgChan chan []*message.Message) *HashiEventHandler {
	return &HashiEventHandler{
		log:    log.With().Uint8("domainID", domainID).Logger(),
		client: client,
	}
}

func (h *HashiEventHandler) HandleMessages(destination uint8, startBlock *big.Int, endBlock *big.Int, slot *big.Int) error {
	logs, err := h.fetchMessages(startBlock, endBlock)
	if err != nil {
		return err
	}

	msgs := make([]*message.Message, 0)
	for _, l := range logs {
		msg, err := h.handleMessage(l, destination, slot)
		if err != nil {
			return err
		}
		if msg == nil {
			continue
		}

		msgs = append(msgs, msg)
	}

	for _, msg := range msgs {
		h.msgChan <- []*message.Message{msg}
	}
	return nil
}

func (h *HashiEventHandler) handleMessage(l types.Log, destination uint8, slot *big.Int) (*message.Message, error) {
	msg, err := h.unpackMessage(l.Data)
	if err != nil {
		return nil, err
	}
	chainID, ok := h.chainIDS[destination]
	if !ok {
		return nil, fmt.Errorf("no chain ID for destination %d", destination)
	}
	if chainID.Cmp(msg.Message.TargetChainID) != 0 {
		return nil, nil
	}

	block, err := h.client.BlockByHash(context.Background(), l.TxHash)
	if err != nil {
		return nil, err
	}
	beaconBlock, err := h.beaconClient.BeaconBlockHeader(context.Background(), &api.BeaconBlockHeaderOpts{
		Block: block.BeaconRoot().Hex(),
	})
	if err != nil {
		return nil, err
	}
	txSlot := new(big.Int).SetUint64(uint64(beaconBlock.Data.Header.Message.Slot))
	rootProof, err := h.rootProver.ReceiptsRootProof(context.Background(), slot, txSlot)
	if err != nil {
		return nil, err
	}

	receipt, err := h.client.TransactionReceipt(context.Background(), l.TxHash)
	if err != nil {
		return nil, err
	}
	txIndexRLP, err := rlp.EncodeToBytes(receipt.TransactionIndex)
	if err != nil {
		return nil, err
	}

	receiptProof, err := h.receiptProver.ReceiptProof(l.TxHash)
	if err != nil {
		return nil, err
	}

	return evmMessage.NewHashiMessage(h.domainID, destination, evmMessage.HashiData{
		SrcSlot:           slot,
		TxSlot:            txSlot,
		ReceiptProof:      receiptProof,
		ReceiptRootProof:  rootProof,
		ReceiptRoot:       block.ReceiptHash(),
		TxIndexRLPEncoded: txIndexRLP,
		LogIndex:          h.logIndex(receipt, l),
	}), nil
}

func (h *HashiEventHandler) logIndex(receipt *types.Receipt, log types.Log) *big.Int {
	for i, l := range receipt.Logs {
		if l.Index == log.Index {
			return big.NewInt(int64(i))
		}
	}

	return big.NewInt(0)
}

func (h *HashiEventHandler) fetchMessages(startBlock *big.Int, endBlock *big.Int) ([]types.Log, error) {
	return fetchLogs(h.client, startBlock, endBlock, h.yahoAddress, string(events.MessageDispatchedSig))
}

func (h *HashiEventHandler) unpackMessage(data []byte) (*events.MessageDispatched, error) {
	var m events.MessageDispatched
	err := h.yahoABI.UnpackIntoInterface(&m, "MessageDispatched", data)
	if err != nil {
		return &events.MessageDispatched{}, err
	}

	return &m, nil
}
