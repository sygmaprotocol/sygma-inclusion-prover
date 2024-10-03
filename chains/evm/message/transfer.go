// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package message

import (
	"math/big"

	"github.com/sygmaprotocol/sygma-core/relayer/message"
	"github.com/sygmaprotocol/sygma-core/relayer/proposal"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/listener/events"
)

type TransferType string

const (
	EVMTransferMessage  message.MessageType   = "EVMTransferMessage"
	EVMTransferProposal proposal.ProposalType = "EVMTransferProposal"

	GenericTransfer  TransferType = "genericTransfer"
	FungibleTransfer TransferType = "fungibleTransfer"
)

type TransferData struct {
	Deposit      *events.Deposit
	Slot         *big.Int
	AccountProof []string
	StorageProof []string
	Type         TransferType
}

func NewEVMTransferMessage(source uint8, destination uint8, transfer TransferData, messageID string) *message.Message {
	return &message.Message{
		Source:      source,
		Destination: destination,
		Data:        transfer,
		ID:          messageID,
		Type:        EVMTransferMessage,
	}
}

type TransferHandler struct{}

func (h *TransferHandler) HandleMessage(m *message.Message) (*proposal.Proposal, error) {
	return &proposal.Proposal{
		Source:      m.Source,
		Destination: m.Destination,
		Type:        EVMTransferProposal,
		Data:        m.Data,
	}, nil
}
