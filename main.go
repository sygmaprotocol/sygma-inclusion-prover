package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/sygmaprotocol/sygma-inclusion-prover/health"
)

func main() {
	go health.StartHealthEndpoint(3000)

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
