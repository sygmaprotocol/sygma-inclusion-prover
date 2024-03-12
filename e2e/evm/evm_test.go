// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package evm_test

import (
	"context"
	"math/big"
	"strings"
	"testing"

	ethereumABI "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"
	"github.com/sygmaprotocol/sygma-core/chains/evm/client"
	"github.com/sygmaprotocol/sygma-core/crypto/secp256k1"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/abi"
)

const EXECUTION_SIG = "ProposalExecution(uint8 originDomainID, uint64 depositNonce, bytes handlerResponse)"

type ProposalExecution struct {
	OriginDomainID  uint8
	DepositNonce    uint64
	HandlerResponse []byte
}

type EVME2ETestSuite struct {
	suite.Suite
	client        *client.EVMClient
	routerAddress common.Address
	routerABI     ethereumABI.ABI
}

func TestRunEVME2ETestSuite(t *testing.T) {
	suite.Run(t, new(EVME2ETestSuite))
}

func (s *EVME2ETestSuite) SetupSuite() {
	kp, _ := secp256k1.NewKeypairFromString("4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d")
	s.client, _ = client.NewEVMClient("http://localhost:8645", kp)
	s.routerAddress = common.HexToAddress("0xC89Ce4735882C9F0f0FE26686c53074E09B0D550")
	s.routerABI, _ = ethereumABI.JSON(strings.NewReader(abi.RouterABI))
}

func (s *EVME2ETestSuite) Test_SuccessfulExecutions() {
	executions, err := s.client.FetchEventLogs(
		context.Background(),
		s.routerAddress,
		EXECUTION_SIG,
		big.NewInt(0),
		big.NewInt(30))
	s.Nil(err)

	s.Equal(len(executions), 2)

	for index, execution := range executions {
		var e ProposalExecution
		err := s.routerABI.UnpackIntoInterface(&e, "Deposit", execution.Data)
		s.Nil(err)
		s.Equal(e.DepositNonce, index+1)
		s.Equal(e.OriginDomainID, 1)
	}
}
