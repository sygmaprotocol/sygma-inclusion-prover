// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package events

import "math/big"

const (
	StateRootSubmittedSig = "StateRootSubmitted(uint8,uint256,bytes32)"
)

type StateRootSubmitted struct {
	// ID of chain from which the state root is from
	SourceDomainID uint8
	// Finalized beacon slot belonging to the state root
	Slot *big.Int
	// Execution state root
	StateRoot [32]byte
}
