// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package handlers

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"slices"
	"strings"

	ethereumABI "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sygmaprotocol/sygma-core/relayer/message"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/abi"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/listener/events"
	evmMessage "github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/message"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/util"
)

const (
	EVMStateRootMessage message.MessageType = "EVMStateRootMessage"
)

type Client interface {
	CallContext(ctx context.Context, target interface{}, rpcMethod string, args ...interface{}) error
	FetchEventLogs(ctx context.Context, contractAddress common.Address, event string, startBlock *big.Int, endBlock *big.Int) ([]types.Log, error)
	BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
}

type StorageProof struct {
	Proof []string `json:"proof"`
}

type AccountProof struct {
	AccountProof []string       `json:"accountProof"`
	StorageProof []StorageProof `json:"storageProof"`
}

type DepositEventHandler struct {
	log              zerolog.Logger
	client           Client
	domainID         uint8
	routerAddress    common.Address
	routerABI        ethereumABI.ABI
	slotIndex        uint8
	genericResources []string
	msgChan          chan []*message.Message
}

func NewDepositEventHandler(
	domainID uint8,
	client Client,
	routerAddres common.Address,
	slotIndex uint8,
	genericResources []string,
	msgChan chan []*message.Message) *DepositEventHandler {
	routerABI, _ := ethereumABI.JSON(strings.NewReader(abi.RouterABI))
	return &DepositEventHandler{
		log:              log.With().Uint8("domainID", domainID).Logger(),
		client:           client,
		routerAddress:    routerAddres,
		routerABI:        routerABI,
		slotIndex:        slotIndex,
		genericResources: genericResources,
		msgChan:          msgChan,
		domainID:         domainID,
	}
}

func (h *DepositEventHandler) HandleEvents(destination uint8, startBlock *big.Int, endBlock *big.Int, slot *big.Int) error {
	deposits, err := h.fetchDeposits(destination, startBlock, endBlock)
	if err != nil {
		return err
	}
	msgs := make(map[uint8][]*message.Message)
	for _, d := range deposits {
		accountProof, storageProof, err := h.proof(endBlock, d)
		if err != nil {
			return err
		}

		h.log.Debug().Uint8("destination", d.DestinationDomainID).Msg("Sending transfer message")

		msgs[d.DestinationDomainID] = append(msgs[d.DestinationDomainID], evmMessage.NewEVMTransferMessage(h.domainID, d.DestinationDomainID, evmMessage.TransferData{
			Deposit:      d,
			Slot:         slot,
			AccountProof: accountProof,
			StorageProof: storageProof,
			Type:         h.transferType(d),
		}))
	}
	if len(msgs) == 0 {
		log.Debug().Msgf("No deposits found for block range %s-%s", startBlock, endBlock)
		return nil
	}
	for _, msg := range msgs {
		h.msgChan <- msg
	}
	return nil
}

func (h *DepositEventHandler) fetchDeposits(destinationDomain uint8, startBlock *big.Int, endBlock *big.Int) ([]*events.Deposit, error) {
	logs, err := fetchLogs(h.client, startBlock, endBlock, h.routerAddress, string(events.DepositSig))
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

func (h *DepositEventHandler) proof(
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

func (h *DepositEventHandler) transferType(d *events.Deposit) evmMessage.TransferType {
	if slices.Contains(h.genericResources, hex.EncodeToString(d.ResourceID[:])) {
		return evmMessage.GenericTransfer
	} else {
		return evmMessage.FungibleTransfer
	}
}

// slotKey mimics slot key calculation from solidity
// https://github.com/sygmaprotocol/sygma-x-solidity/blob/bd43d1138b38328267f2bfdb65a37817f24e3286/src/contracts/Executor.sol#L235
func (h *DepositEventHandler) slotKey(d *events.Deposit) string {
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

func (h *DepositEventHandler) unpackDeposit(data []byte) (*events.Deposit, error) {
	var d events.Deposit
	err := h.routerABI.UnpackIntoInterface(&d, "Deposit", data)
	if err != nil {
		return &events.Deposit{}, err
	}

	return &d, nil
}
