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
	"github.com/sygmaprotocol/sygma-core/relayer/proposal"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/abi"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/message"
)

type ExecutorProposal struct {
	OriginDomainID uint8
	SecurityModel  uint8
	DepositNonce   uint64
	ResourceID     [32]byte
	Data           []byte
	StorageProof   [][]byte
}

type Executor struct {
	coreContracts.Contract
}

func NewExecutorContract(
	address common.Address,
	client client.Client,
	transactor transactor.Transactor,
) *Executor {
	a, _ := ethereumABI.JSON(strings.NewReader(abi.ExecutorABI))
	return &Executor{
		Contract: coreContracts.NewContract(address, a, nil, client, transactor),
	}
}

func (c *Executor) ExecuteProposals(
	props []ExecutorProposal,
	accountProof [][]byte,
	slot *big.Int,
	opts transactor.TransactOptions,
) (*common.Hash, error) {
	return c.ExecuteTransaction(
		"executeProposals",
		opts,
		props, accountProof, slot,
	)
}

func (c *Executor) IsProposalExecuted(p *proposal.Proposal) (bool, error) {
	t := p.Data.(message.TransferData)
	res, err := c.CallContract("isProposalExecuted", p.Source, big.NewInt(int64(t.Deposit.DepositNonce)))
	if err != nil {
		return false, err
	}
	out := *ethereumABI.ConvertType(res[0], new(bool)).(*bool)
	return out, nil
}
