// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package message

import (
	"context"
	"encoding/hex"
	"math/big"

	"github.com/attestantio/go-eth2-client/api"
	"github.com/attestantio/go-eth2-client/spec"
	ethereumABI "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
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
	CallContext(ctx context.Context, target interface{}, rpcMethod string, args ...interface{}) error
	FetchEventLogs(ctx context.Context, contractAddress common.Address, event string, startBlock *big.Int, endBlock *big.Int) ([]types.Log, error)
}

type StateRootHandler struct {
	blockFetcher  BlockFetcher
	eventFetcher  EventFetcher
	routerAddress common.Address
	routerABI     ethereumABI.ABI
	domainID      uint8
	slotIndex     uint8

	msgChan chan []*message.Message

	// TODO
	// latest block number for the source-destination pair
	latestBlockNumber map[uint8]map[uint8]uint64
}

func (h *StateRootHandler) HandleMessage(m *message.Message) (*proposal.Proposal, error) {
	stateRoot := m.Data.(StateRootData)
	log.Debug().Uint8(
		"domainID", m.Destination).Str(
		"stateRoot", hex.EncodeToString(stateRoot.StateRoot[:])
	).Msgf("Received state root message from domain %d", m.Source)

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
		accountProof, storageProof, err := h.proof(blockNumber, d)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, NewEVMTransferMessage(h.domainID, d.DestinationDomainID, TransferData{
			Deposit:      d,
			Slot:         stateRoot.Slot,
			AccountProof: accountProof,
			StorageProof: storageProof,
		}))
	}

	h.msgChan <- msgs
	return nil, nil
}

func (h *StateRootHandler) proof(
	blockNumber *big.Int,
	deposit *events.Deposit,
) ([]string, []string, error) {
	type storageProof struct {
		Proof []string `json:"proof"`
	}
	type accountProof struct {
		AccountProof []string     `json:"accountProof"`
		StorageProof storageProof `json:"storageProof"`
	}
	type response struct {
		Result accountProof `json:"result"`
	}
	var resp response
	err := h.eventFetcher.CallContext(context.Background(), &resp, "eth_getProof", h.routerAddress, []string{h.slotKey(deposit)}, hexutil.EncodeBig(blockNumber))
	if err != nil {
		return nil, nil, err
	}

	return resp.Result.AccountProof, resp.Result.StorageProof.Proof, nil
}

// slotKey mimics slot key calculation from solidity
// https://github.com/sygmaprotocol/sygma-x-solidity/blob/bd43d1138b38328267f2bfdb65a37817f24e3286/src/contracts/Executor.sol#L235
func (h *StateRootHandler) slotKey(d *events.Deposit) string {
	// TODO arguments pack
	outerMap, _ := h.routerABI.Pack("", h.domainID, h.slotIndex)
	outerMapHash := crypto.Keccak256(outerMap)
	innerMap, _ := h.routerABI.Pack("", d.DepositNonce, outerMapHash)
	slotKey := crypto.Keccak256(innerMap)
	return hex.EncodeToString(slotKey)
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
