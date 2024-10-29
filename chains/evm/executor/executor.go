// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package executor

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog/log"
	"github.com/sygmaprotocol/sygma-core/chains/evm/transactor"
	"github.com/sygmaprotocol/sygma-core/relayer/proposal"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/contracts"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/message"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/util"
)

const TRANSFER_GAS_COST = 600000
const HASHI_GAS_COST = 3000000

type Batch struct {
	proposals []contracts.ExecutorProposal
	gasLimit  uint64
}

type ExecutorContract interface {
	IsProposalExecuted(p *proposal.Proposal) (bool, error)
	ExecuteProposals(proposals []contracts.ExecutorProposal, accountProof [][]byte, slot *big.Int, opts transactor.TransactOptions) (*common.Hash, error)
}

type HashiContract interface {
	VerifyAndStoreDispatchedMessage(
		srcSlot uint64,
		txSlot uint64,
		receiptsRootProof [][]byte,
		receiptsRoot [32]byte,
		receiptProof [][]byte,
		txIndexRLPEncoded []byte,
		logIndex *big.Int,
		opts transactor.TransactOptions,
	) (*common.Hash, error)
}

type EVMExecutor struct {
	domainID          uint8
	executor          ExecutorContract
	hashiAdapter      HashiContract
	transactionMaxGas uint64
}

func NewEVMExecutor(domainID uint8, executor ExecutorContract, hashiAdapter HashiContract) *EVMExecutor {
	return &EVMExecutor{
		domainID:          domainID,
		executor:          executor,
		hashiAdapter:      hashiAdapter,
		transactionMaxGas: 10000000,
	}
}

func (e *EVMExecutor) Execute(props []*proposal.Proposal) error {
	switch prop := props[0]; prop.Type {
	case message.EVMTransferProposal:
		return e.transfer(props)
	case message.HashiProposal:
		return e.storeMessage(props)
	default:
		return fmt.Errorf("no executor configured for prop type %s", prop.Type)
	}
}

func (e *EVMExecutor) storeMessage(props []*proposal.Proposal) error {
	data := props[0].Data.(message.HashiData)
	hash, err := e.hashiAdapter.VerifyAndStoreDispatchedMessage(
		data.SrcSlot.Uint64(),
		data.TxSlot.Uint64(),
		data.ReceiptRootProof,
		data.ReceiptRoot,
		data.ReceiptProof,
		data.TxIndexRLPEncoded,
		data.LogIndex,
		transactor.TransactOptions{
			GasLimit: HASHI_GAS_COST,
		},
	)
	if err != nil {
		return err
	}

	log.Info().Str("messageID", props[0].MessageID).Uint8("domainID", e.domainID).Msgf("Sent hashi message execution with hash: %s", hash)
	return nil
}

func (e *EVMExecutor) transfer(props []*proposal.Proposal) error {
	batches, err := e.proposalBatches(props)
	if err != nil {
		return err
	}

	batchData := props[0].Data.(message.TransferData)
	proofBytes, _ := util.ToByteArray(batchData.AccountProof)
	for _, batch := range batches {
		if len(batch.proposals) == 0 {
			continue
		}

		hash, err := e.executor.ExecuteProposals(batch.proposals, proofBytes, batchData.Slot, transactor.TransactOptions{
			GasLimit: batch.gasLimit,
		})
		if err != nil {
			log.Err(err).Msgf("Failed executing proposals")
			continue
		}

		log.Info().Str("messageID", props[0].MessageID).Uint8("domainID", e.domainID).Msgf("Sent proposals execution with hash: %s", hash)
	}
	return nil
}

func (e *EVMExecutor) proposalBatches(props []*proposal.Proposal) ([]*Batch, error) {
	batches := make([]*Batch, 1)
	currentBatch := &Batch{
		proposals: make([]contracts.ExecutorProposal, 0),
		gasLimit:  0,
	}
	batches[0] = currentBatch

	for _, prop := range props {
		isExecuted, err := e.executor.IsProposalExecuted(prop)
		if err != nil {
			return nil, err
		}
		if isExecuted {
			log.Info().Msgf("Proposal %+v already executed", prop)
			continue
		}

		propGasLimit := e.proposalGas(prop)
		currentBatch.gasLimit += propGasLimit
		if currentBatch.gasLimit >= e.transactionMaxGas {
			currentBatch = &Batch{
				proposals: make([]contracts.ExecutorProposal, 0),
				gasLimit:  0,
			}
			batches = append(batches, currentBatch)
		}

		d := prop.Data.(message.TransferData)
		proofBytes, _ := util.ToByteArray(d.StorageProof)
		currentBatch.proposals = append(currentBatch.proposals, contracts.ExecutorProposal{
			OriginDomainID: prop.Source,
			SecurityModel:  d.Deposit.SecurityModel,
			DepositNonce:   d.Deposit.DepositNonce,
			ResourceID:     d.Deposit.ResourceID,
			Data:           d.Deposit.Data,
			StorageProof:   proofBytes,
		})
	}

	return batches, nil
}

func (e *EVMExecutor) proposalGas(prop *proposal.Proposal) uint64 {
	transferData := prop.Data.(message.TransferData)
	if transferData.Type != message.GenericTransfer {
		return TRANSFER_GAS_COST
	}

	genericFee := new(big.Int).SetBytes(transferData.Deposit.Data[:32])
	return uint64(TRANSFER_GAS_COST) + genericFee.Uint64()
}
