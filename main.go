// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package main

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/attestantio/go-eth2-client/http"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sygmaprotocol/sygma-core/chains/evm"
	"github.com/sygmaprotocol/sygma-core/chains/evm/client"
	"github.com/sygmaprotocol/sygma-core/chains/evm/listener"
	"github.com/sygmaprotocol/sygma-core/chains/evm/transactor/gas"
	"github.com/sygmaprotocol/sygma-core/chains/evm/transactor/monitored"
	"github.com/sygmaprotocol/sygma-core/chains/evm/transactor/transaction"
	"github.com/sygmaprotocol/sygma-core/crypto/secp256k1"
	"github.com/sygmaprotocol/sygma-core/observability"
	"github.com/sygmaprotocol/sygma-core/relayer"
	"github.com/sygmaprotocol/sygma-core/relayer/message"
	coreStore "github.com/sygmaprotocol/sygma-core/store"
	"github.com/sygmaprotocol/sygma-core/store/lvldb"
	evmConfig "github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/config"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/contracts"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/executor"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/listener/handlers"
	evmMessage "github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/message"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/proof"
	"github.com/sygmaprotocol/sygma-inclusion-prover/config"
	"github.com/sygmaprotocol/sygma-inclusion-prover/health"
	"github.com/sygmaprotocol/sygma-inclusion-prover/metrics"
	"github.com/sygmaprotocol/sygma-inclusion-prover/store"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	logLevel, err := zerolog.ParseLevel(cfg.Observability.LogLevel)
	if err != nil {
		panic(err)
	}
	observability.ConfigureLogger(logLevel, os.Stdout)

	log.Info().Msg("Loaded configuration")

	go health.StartHealthEndpoint(cfg.Observability.HealthPort)
	var db *lvldb.LVLDB
	for {
		db, err = lvldb.NewLvlDB(cfg.Store.Path)
		if err != nil {
			log.Error().Err(err).Msg("Unable to connect to blockstore file, retry in 10 seconds")
			time.Sleep(10 * time.Second)
		} else {
			log.Info().Msg("Successfully connected to blockstore file")
			break
		}
	}
	latestBlockStore := store.NewBlockStore(db)
	blockStore := coreStore.NewBlockStore(db)

	msgChan := make(chan []*message.Message)
	chains := make(map[uint8]relayer.RelayedChain)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for id, nType := range cfg.Domains {
		switch nType {
		case "evm":
			{
				config, err := evmConfig.LoadEVMConfig(id)
				if err != nil {
					panic(err)
				}

				kp, err := secp256k1.NewKeypairFromString(config.Key)
				if err != nil {
					panic(err)
				}

				client, err := client.NewEVMClient(config.Endpoint, kp)
				if err != nil {
					panic(err)
				}

				startBlock, err := blockStore.GetStartBlock(
					id,
					new(big.Int).SetUint64(config.StartBlock),
					config.Latest,
					config.FreshStart,
				)
				if err != nil {
					panic(err)
				}
				if startBlock == nil {
					latestBlock, err := client.LatestBlock()
					if err != nil {
						panic(err)
					}
					startBlock = latestBlock
				}

				var evmListener *listener.EVMListener
				if len(config.StateRootAddresses) > 0 {
					eventHandlers := []listener.EventHandler{}
					for _, stateRootAddress := range config.StateRootAddresses {
						eventHandlers = append(eventHandlers, handlers.NewStateRootEventHandler(msgChan, client, common.HexToAddress(stateRootAddress), id))
					}
					evmListener = listener.NewEVMListener(
						client,
						eventHandlers,
						blockStore,
						&metrics.RelayerMetrics{},
						id,
						time.Duration(config.BlockRetryInterval)*time.Second,
						big.NewInt(config.BlockConfirmations),
						big.NewInt(config.BlockInterval))
				}

				messageHandler := message.NewMessageHandler()
				messageHandler.RegisterMessageHandler(evmMessage.EVMTransferMessage, &evmMessage.TransferHandler{})
				messageHandler.RegisterMessageHandler(evmMessage.HashiMessage, &evmMessage.HashiMessageHandler{})
				if config.Yaho != "" || config.Router != "" {
					beaconClient, err := http.New(ctx,
						http.WithAddress(config.BeaconEndpoint),
						http.WithLogLevel(logLevel),
						http.WithTimeout(time.Minute*15),
						http.WithEnforceJSON(false),
					)
					if err != nil {
						panic(err)
					}
					beaconProvider := beaconClient.(*http.Service)
					receiptProver := proof.NewReceiptProver(client)
					rootProver := proof.NewReceiptRootProver(beaconProvider, config.Spec)

					stateRootEventHandlers := make([]evmMessage.EventHandler, 0)
					if config.Yaho != "" {
						yahoAddress := common.HexToAddress(config.Yaho)
						stateRootEventHandlers = append(
							stateRootEventHandlers,
							handlers.NewHashiEventHandler(
								id, client, beaconProvider, receiptProver, rootProver, yahoAddress, cfg.ChainIDS, msgChan),
						)

					}
					if config.Router != "" {
						routerAddress := common.HexToAddress(config.Router)
						stateRootEventHandlers = append(
							stateRootEventHandlers,
							handlers.NewDepositEventHandler(
								id, client, routerAddress, config.SlotIndex, config.GenericResources, msgChan),
						)
					}
					messageHandler.RegisterMessageHandler(
						evmMessage.EVMStateRootMessage,
						evmMessage.NewStateRootHandler(id, stateRootEventHandlers, beaconProvider, latestBlockStore, new(big.Int).Set(startBlock)))
				}

				gasPricer := gas.NewLondonGasPriceClient(client, &gas.GasPricerOpts{
					UpperLimitFeePerGas: big.NewInt(config.MaxGasPrice),
					GasPriceFactor:      big.NewFloat(config.GasMultiplier),
				})
				t := monitored.NewMonitoredTransactor(transaction.NewTransaction, gasPricer, client, big.NewInt(config.MaxGasPrice), big.NewInt(config.GasIncreasePercentage))
				go t.Monitor(ctx, time.Minute*3, time.Minute*10, time.Minute)
				evmExecutor := executor.NewEVMExecutor(
					id,
					contracts.NewExecutorContract(common.HexToAddress(config.Executor), client, t),
					contracts.NewHashiAdapterContract(common.HexToAddress(config.Hashi), client, t),
				)
				chain := evm.NewEVMChain(evmListener, messageHandler, evmExecutor, id, startBlock)
				chains[id] = chain
			}
		default:
			{
				panic(fmt.Sprintf("invalid network type %s for id %d", nType, id))
			}
		}
	}

	r := relayer.NewRelayer(chains)
	go r.Start(ctx, msgChan)

	sysErr := make(chan os.Signal, 1)
	signal.Notify(sysErr,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGHUP,
		syscall.SIGQUIT)
	log.Info().Msgf("Started Sygma inclusion prover")

	se := <-sysErr
	log.Info().Msgf("terminating got ` [%v] signal", se)
}
