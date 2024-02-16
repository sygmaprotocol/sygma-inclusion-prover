// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/config"
	baseConfig "github.com/sygmaprotocol/sygma-inclusion-prover/config"
)

type EVMConfigTestSuite struct {
	suite.Suite
}

func TestRunEVMConfigTestSuite(t *testing.T) {
	suite.Run(t, new(EVMConfigTestSuite))
}

func (c *EVMConfigTestSuite) TearDownTest() {
	os.Clearenv()
}

func (s *EVMConfigTestSuite) Test_LoadEVMConfig_MissingField() {
	os.Setenv("INCLUSION_PROVER_DOMAINS_1_ENDPOINT", "http://endpoint.com")
	os.Setenv("INCLUSION_PROVER_DOMAINS_1_KEY", "key")
	os.Setenv("INCLUSION_PROVER_DOMAINS_1_SPECTRE", "spectre")
	os.Setenv("INCLUSION_PROVER_DOMAINS_2_ROUTER", "invalid")

	_, err := config.LoadEVMConfig(1)

	s.NotNil(err)
}

func (s *EVMConfigTestSuite) Test_LoadEVMConfig_SuccessfulLoad_DefaultValues() {
	os.Setenv("INCLUSION_PROVER_DOMAINS_1_ENDPOINT", "http://endpoint.com")
	os.Setenv("INCLUSION_PROVER_DOMAINS_1_KEY", "key")
	os.Setenv("INCLUSION_PROVER_DOMAINS_1_SPECTRE", "spectre")
	os.Setenv("INCLUSION_PROVER_DOMAINS_1_ROUTER", "router")
	os.Setenv("INCLUSION_PROVER_DOMAINS_1_EXECUTOR", "executor")
	os.Setenv("INCLUSION_PROVER_DOMAINS_1_STATE_ROOT_ADDRESSES", "0x1,0x2")
	os.Setenv("INCLUSION_PROVER_DOMAINS_1_BEACON_ENDPOINT", "endpoint")
	os.Setenv("INCLUSION_PROVER_DOMAINS_2_ROUTER", "invalid")
	os.Setenv("INCLUSION_PROVER_DOMAINS_1_STATE_ROOT_ADDRESSES", "0x1,0x2")

	c, err := config.LoadEVMConfig(1)

	s.Nil(err)
	s.Equal(c, &config.EVMConfig{
		BaseNetworkConfig: baseConfig.BaseNetworkConfig{
			Key:      "key",
			Endpoint: "http://endpoint.com",
		},
		Router:                "router",
		Executor:              "executor",
		GasMultiplier:         1,
		GasIncreasePercentage: 15,
		MaxGasPrice:           500000000000,
		BeaconEndpoint:        "endpoint",
		StateRootAddresses:    []string{"0x1", "0x2"},
	})
}

func (s *EVMConfigTestSuite) Test_LoadEVMConfig_SuccessfulLoad() {
	os.Setenv("INCLUSION_PROVER_DOMAINS_1_ENDPOINT", "http://endpoint.com")
	os.Setenv("INCLUSION_PROVER_DOMAINS_1_KEY", "key")
	os.Setenv("INCLUSION_PROVER_DOMAINS_1_ROUTER", "router")
	os.Setenv("INCLUSION_PROVER_DOMAINS_1_EXECUTOR", "executor")
	os.Setenv("INCLUSION_PROVER_DOMAINS_1_BEACON_ENDPOINT", "endpoint")
	os.Setenv("INCLUSION_PROVER_DOMAINS_1_MAX_GAS_PRICE", "1000")
	os.Setenv("INCLUSION_PROVER_DOMAINS_1_BLOCK_INTERVAL", "10")
	os.Setenv("INCLUSION_PROVER_DOMAINS_1_GAS_MULTIPLIER", "1")
	os.Setenv("INCLUSION_PROVER_DOMAINS_1_GAS_INCREASE_PERCENTAGE", "20")
	os.Setenv("INCLUSION_PROVER_DOMAINS_2_ROUTER", "invalid")
	os.Setenv("INCLUSION_PROVER_DOMAINS_1_STATE_ROOT_ADDRESSES", "0x1,0x2")

	c, err := config.LoadEVMConfig(1)

	s.Nil(err)
	s.Equal(c, &config.EVMConfig{
		BaseNetworkConfig: baseConfig.BaseNetworkConfig{
			Key:      "key",
			Endpoint: "http://endpoint.com",
		},
		Router:                "router",
		Executor:              "executor",
		GasMultiplier:         1,
		GasIncreasePercentage: 20,
		MaxGasPrice:           1000,
		BeaconEndpoint:        "endpoint",
		StateRootAddresses:    []string{"0x1", "0x2"},
	})
}
