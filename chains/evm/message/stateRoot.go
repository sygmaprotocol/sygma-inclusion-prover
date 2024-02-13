// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package message

import "github.com/sygmaprotocol/sygma-core/relayer/message"

const (
	EVMStateRootMessage message.MessageType = "EVMStateRootMessage"
)

func NewEvmStateRootMessage(source uint8, destination uint8, stateRoot [32]byte) *message.Message {
	return &message.Message{
		Source:      source,
		Destination: destination,
		Data:        stateRoot,
		Type:        EVMStateRootMessage,
	}
}
