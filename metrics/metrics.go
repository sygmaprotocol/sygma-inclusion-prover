package metrics

import (
	"math/big"

	"github.com/rs/zerolog/log"
)

type RelayerMetrics struct{}

func (t *RelayerMetrics) TrackBlockDelta(domainID uint8, head *big.Int, current *big.Int) {
	log.Trace().Uint8("domainID", domainID).Msgf("Block delta is %d", new(big.Int).Sub(head, current))
}
