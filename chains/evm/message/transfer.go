// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package message

import (
	"github.com/sygmaprotocol/sygma-core/relayer/message"
	"github.com/sygmaprotocol/sygma-core/relayer/proposal"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/listener/events"
)

const (
	EVMDepositMessage  message.MessageType   = "EVMDepositMessage"
	EVMDepositProposal proposal.ProposalType = "EVMDepositProposal"
)

func NewEVMDepositMessage(source uint8, destination uint8, deposit *events.Deposit) *message.Message {
	return &message.Message{
		Source:      source,
		Destination: destination,
		Data:        deposit,
		Type:        EVMDepositMessage,
	}
}

type DepositHandler struct{}

func (h *DepositHandler) HandleMessage(m *message.Message) (*proposal.Proposal, error) {
	return &proposal.Proposal{
		Source:      m.Source,
		Destination: m.Destination,
		Type:        EVMDepositProposal,
		Data:        m.Data,
	}, nil
}
