// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package message

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

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
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/abi"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/listener/events"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/util"
	"golang.org/x/exp/slices"
)

const (
	EVMStateRootMessage message.MessageType = "EVMStateRootMessage"
)

type StateRootData struct {
	StateRoot [32]byte
	Slot      *big.Int
}

type StorageProof struct {
	Proof []string `json:"proof"`
}

type AccountProof struct {
	AccountProof []string       `json:"accountProof"`
	StorageProof []StorageProof `json:"storageProof"`
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

type Client interface {
	CallContext(ctx context.Context, target interface{}, rpcMethod string, args ...interface{}) error
	FetchEventLogs(ctx context.Context, contractAddress common.Address, event string, startBlock *big.Int, endBlock *big.Int) ([]types.Log, error)
}

type StateRootHandler struct {
	blockFetcher     BlockFetcher
	blockStorer      BlockStorer
	client           Client
	routerAddress    common.Address
	routerABI        ethereumABI.ABI
	domainID         uint8
	slotIndex        uint8
	genericResources []string

	msgChan chan []*message.Message
}

func NewStateRootHandler(
	blockFetcher BlockFetcher,
	blockStorer BlockStorer,
	client Client,
	routerAddress common.Address,
	msgChan chan []*message.Message,
	domainID uint8,
	slotIndex uint8,
	genericResources []string,
) *StateRootHandler {
	routerABI, _ := ethereumABI.JSON(strings.NewReader(abi.RouterABI))
	return &StateRootHandler{
		blockFetcher:     blockFetcher,
		blockStorer:      blockStorer,
		client:           client,
		routerAddress:    routerAddress,
		routerABI:        routerABI,
		domainID:         domainID,
		slotIndex:        slotIndex,
		msgChan:          msgChan,
		genericResources: genericResources,
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
	endBlock := big.NewInt(int64(block.Data.Deneb.Message.Body.ExecutionPayload.BlockNumber))
	deposits, err := h.fetchDeposits(m.Source, startBlock, endBlock)
	if err != nil {
		return nil, err
	}
	msgs := []*message.Message{}
	for _, d := range deposits {
		accountProof, storageProof, err := h.proof(endBlock, d)
		if err != nil {
			return nil, err
		}
		log.Debug().Uint8("domainID", h.domainID).Uint8("destination", d.DestinationDomainID).Msg("Sending transfer message")

		msgs = append(msgs, NewEVMTransferMessage(h.domainID, d.DestinationDomainID, TransferData{
			Deposit:      d,
			Slot:         stateRoot.Slot,
			AccountProof: accountProof,
			StorageProof: storageProof,
			Type:         h.transferType(d),
		}))
	}

	err = h.blockStorer.StoreBlock(h.domainID, m.Source, endBlock)
	if err != nil {
		log.Err(err).Msgf("Failed saving latest block for %d-%d", h.domainID, m.Source)
	}

	if len(msgs) == 0 {
		log.Warn().Uint8("domainID", h.domainID).Msgf("No deposits found for block range %s-%s", startBlock, endBlock)
		return nil, nil
	}
	h.msgChan <- msgs

	return nil, nil
}

func (h *StateRootHandler) fetchDeposits(destinationDomain uint8, startBlock *big.Int, endBlock *big.Int) ([]*events.Deposit, error) {
	logs, err := h.client.FetchEventLogs(context.Background(), h.routerAddress, string(events.DepositSig), startBlock, endBlock)
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
		if d.DestinationDomainID != destinationDomain {
			continue
		}

		log.Debug().Msgf("Found deposit log in block: %d, TxHash: %s, contractAddress: %s, sender: %s", dl.BlockNumber, dl.TxHash, dl.Address, d.SenderAddress)
		deposits = append(deposits, d)
	}

	return deposits, nil
}

func (h *StateRootHandler) proof(
	blockNumber *big.Int,
	deposit *events.Deposit,
) ([]string, []string, error) {
	var resp AccountProof
	err := h.client.CallContext(
		context.Background(),
		&resp,
		"eth_getProof",
		h.routerAddress,
		[]string{fmt.Sprintf("0x%s", h.slotKey(deposit))},
		hexutil.EncodeBig(blockNumber))
	if err != nil {
		return nil, nil, err
	}

	return resp.AccountProof, resp.StorageProof[0].Proof, nil
}

func (h *StateRootHandler) transferType(d *events.Deposit) TransferType {
	if slices.Contains(h.genericResources, hex.EncodeToString(d.ResourceID[:])) {
		return GenericTransfer
	} else {
		return FungibleTransfer
	}
}

// slotKey mimics slot key calculation from solidity
// https://github.com/sygmaprotocol/sygma-x-solidity/blob/bd43d1138b38328267f2bfdb65a37817f24e3286/src/contracts/Executor.sol#L235
func (h *StateRootHandler) slotKey(d *events.Deposit) string {
	u8, _ := ethereumABI.NewType("uint8", "uint8", []ethereumABI.ArgumentMarshaling{})
	u64, _ := ethereumABI.NewType("uint64", "uint64", []ethereumABI.ArgumentMarshaling{})
	b32, _ := ethereumABI.NewType("bytes32", "bytes32", []ethereumABI.ArgumentMarshaling{})
	outerArguments := ethereumABI.Arguments{
		ethereumABI.Argument{Name: "", Type: u8},
		ethereumABI.Argument{Name: "", Type: u8},
	}
	outerMap, _ := outerArguments.Pack(d.DestinationDomainID, h.slotIndex)
	outerMapHash := crypto.Keccak256(outerMap)
	innerArguments := ethereumABI.Arguments{
		ethereumABI.Argument{Name: "", Type: u64},
		ethereumABI.Argument{Name: "", Type: b32},
	}
	innerMap, _ := innerArguments.Pack(d.DepositNonce, util.SliceTo32Bytes(outerMapHash))
	slotKey := crypto.Keccak256(innerMap)
	return hex.EncodeToString(slotKey)
}

func (h *StateRootHandler) unpackDeposit(data []byte) (*events.Deposit, error) {
	var d events.Deposit
	err := h.routerABI.UnpackIntoInterface(&d, "Deposit", data)
	if err != nil {
		return &events.Deposit{}, err
	}

	return &d, nil
}
