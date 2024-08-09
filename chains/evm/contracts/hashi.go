// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package contracts

import (
	"math/big"
	"strings"

	ethereumABI "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/sygmaprotocol/sygma-core/chains/evm/client"
	coreContracts "github.com/sygmaprotocol/sygma-core/chains/evm/contracts"
	"github.com/sygmaprotocol/sygma-core/chains/evm/transactor"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/abi"
)

type HashiAdapterContract struct {
	coreContracts.Contract
}

func NewHashiAdapterContract(
	address common.Address,
	client client.Client,
	transactor transactor.Transactor,
) *HashiAdapterContract {
	a, _ := ethereumABI.JSON(strings.NewReader(abi.HashiAdapterABI))
	return &HashiAdapterContract{
		Contract: coreContracts.NewContract(address, a, nil, client, transactor),
	}
}

func (c *HashiAdapterContract) VerifyAndStoreDispatchedMessage(
	srcSlot uint64,
	txSlot uint64,
	receiptsRootProof [][]byte,
	receiptsRoot [32]byte,
	receiptProof [][]byte,
	txIndexRLPEncoded []byte,
	logIndex *big.Int,
	opts transactor.TransactOptions,
) (*common.Hash, error) {
	return c.ExecuteTransaction(
		"verifyAndStoreDispatchedMessage",
		opts,
		srcSlot, txSlot, Bytes32Array(receiptsRootProof), SliceTo32Bytes(receiptsRoot[:]), receiptProof, txIndexRLPEncoded, logIndex,
	)
}

func SliceTo32Bytes(in []byte) [32]byte {
	var res [32]byte
	copy(res[:], in)
	return res
}

func Bytes32Array(array [][]byte) [][32]byte {
	bytes32Array := make([][32]byte, len(array))
	for i, e := range array {
		bytes32Array[i] = SliceTo32Bytes(e[:])
	}
	return bytes32Array
}
