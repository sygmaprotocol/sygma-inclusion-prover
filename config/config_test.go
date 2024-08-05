// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/sygmaprotocol/sygma-inclusion-prover/config"
)

type ConfigTestSuite struct {
	suite.Suite
}

func TestRunConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (c *ConfigTestSuite) TearDownTest() {
	os.Clearenv()
}

func (s *ConfigTestSuite) Test_LoadConfig_MissingField() {
	_, err := config.LoadConfig()

	s.NotNil(err)
}

func (s *ConfigTestSuite) Test_LoadConfig_DefaultValues() {
	os.Setenv("INCLUSION_PROVER_DOMAINS", "1:evm,2:evm")
	os.Setenv("INCLUSION_PROVER_CHAINIDS", "1:3,2:6")

	c, err := config.LoadConfig()

	chainIDS := make(map[uint8]uint64)
	chainIDS[1] = 3
	chainIDS[2] = 6
	domains := make(map[uint8]string)
	domains[1] = "evm"
	domains[2] = "evm"
	s.Nil(err)
	s.Equal(c, &config.Config{
		Observability: &config.Observability{
			LogLevel:   "debug",
			LogFile:    "out.log",
			HealthPort: 9001,
		},
		Store: &config.Store{
			Path: "./lvldbdata",
		},
		Domains:  domains,
		ChainIDS: chainIDS,
	})
}

func (s *ConfigTestSuite) Test_LoadEVMConfig_SuccessfulLoad() {
	os.Setenv("INCLUSION_PROVER_OBSERVABILITY_LOG_LEVEL", "info")
	os.Setenv("INCLUSION_PROVER_OBSERVABILITY_LOG_FILE", "out2.log")
	os.Setenv("INCLUSION_PROVER_OBSERVABILITY_HEALTH_PORT", "9003")
	os.Setenv("INCLUSION_PROVER_STORE_PATH", "./custom_path")
	os.Setenv("INCLUSION_PROVER_DOMAINS", "1:evm,2:evm")
	os.Setenv("INCLUSION_PROVER_CHAINIDS", "1:3,2:6")

	c, err := config.LoadConfig()

	chainIDS := make(map[uint8]uint64)
	chainIDS[1] = 3
	chainIDS[2] = 6
	domains := make(map[uint8]string)
	domains[1] = "evm"
	domains[2] = "evm"
	s.Nil(err)
	s.Equal(c, &config.Config{
		Observability: &config.Observability{
			LogLevel:   "info",
			LogFile:    "out2.log",
			HealthPort: 9003,
		},
		Store: &config.Store{
			Path: "./custom_path",
		},
		Domains:  domains,
		ChainIDS: chainIDS,
	})
}
