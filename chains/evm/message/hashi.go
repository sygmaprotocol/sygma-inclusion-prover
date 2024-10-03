// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package message

import (
	"math/big"

	"github.com/sygmaprotocol/sygma-core/relayer/message"
	"github.com/sygmaprotocol/sygma-core/relayer/proposal"
)

const (
	HashiMessage  message.MessageType   = "HashiMessage"
	HashiProposal proposal.ProposalType = "HashiProposal"
)

type HashiData struct {
	SrcSlot           *big.Int
	TxSlot            *big.Int
	ReceiptRootProof  [][]byte
	ReceiptRoot       [32]byte
	ReceiptProof      [][]byte
	TxIndexRLPEncoded []byte
	LogIndex          *big.Int
}

func NewHashiMessage(source uint8, destination uint8, data HashiData, messageID string) *message.Message {
	return &message.Message{
		Source:      source,
		Destination: destination,
		Data:        data,
		Type:        HashiMessage,
		ID:          messageID,
	}
}

type HashiMessageHandler struct{}

func (h *HashiMessageHandler) HandleMessage(m *message.Message) (*proposal.Proposal, error) {
	return &proposal.Proposal{
		Source:      m.Source,
		Destination: m.Destination,
		Type:        HashiProposal,
		Data:        m.Data,
	}, nil
}
