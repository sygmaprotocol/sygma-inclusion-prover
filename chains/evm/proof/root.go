// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package proof

import (
	"context"
	"math/big"
	"strconv"
	"time"

	"github.com/attestantio/go-eth2-client/api"
	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/deneb"
	ssz "github.com/ferranbt/fastssz"
	gnosisDeneb "github.com/mpetrun5/go-eth2-client/spec/deneb"
	cache "github.com/patrickmn/go-cache"
	"github.com/rs/zerolog/log"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/config"
)

const (
	BEACON_STATE_GINDEX              = 11
	RECEIPTS_ROOT_GINDEX             = 6435
	BLOCK_ROOTS_GINDEX         int64 = 37
	SLOTS_PER_HISTORICAL_LIMIT       = 8192
)

type BeaconClient interface {
	BeaconBlockHeader(
		ctx context.Context,
		opts *api.BeaconBlockHeaderOpts,
	) (
		*api.Response[*apiv1.BeaconBlockHeader],
		error,
	)
	BeaconState(
		ctx context.Context,
		opts *api.BeaconStateOpts,
	) (
		*api.Response[*spec.VersionedBeaconState],
		error,
	)
	SignedBeaconBlock(ctx context.Context,
		opts *api.SignedBeaconBlockOpts,
	) (
		*api.Response[*spec.VersionedSignedBeaconBlock],
		error,
	)
}

type ReceiptRootProver struct {
	beaconClient BeaconClient
	spec         config.Spec
	stateCache   *cache.Cache
}

func NewReceiptRootProver(beaconClient BeaconClient, spec config.Spec) *ReceiptRootProver {
	return &ReceiptRootProver{
		beaconClient: beaconClient,
		spec:         spec,
		stateCache:   cache.New(time.Minute*5, time.Minute*5),
	}
}

// ReceiptRootProof returns the prove from the beacon block root to the receipt roof of the given slot.
// The path for the proof is beacon block -> beacon state -> block roots -> execution payload header -> receipt root.
func (p *ReceiptRootProver) ReceiptsRootProof(ctx context.Context, currentSlot *big.Int, targetSlot *big.Int) ([][]byte, error) {
	receiptsRootProof, err := p.receiptsRootProof(ctx, targetSlot)
	if err != nil {
		return nil, err
	}

	if currentSlot.Cmp(targetSlot) != 0 {
		beaconStateProof, err := p.historicalRootProof(ctx, currentSlot, targetSlot)
		if err != nil {
			return nil, err
		}

		return append(receiptsRootProof, beaconStateProof...), nil
	} else {
		return receiptsRootProof, nil
	}
}

func (p *ReceiptRootProver) historicalRootProof(ctx context.Context, currentSlot *big.Int, targetSlot *big.Int) ([][]byte, error) {
	beaconBlock, err := p.beaconClient.BeaconBlockHeader(ctx, &api.BeaconBlockHeaderOpts{
		Block: currentSlot.String(),
	})
	if err != nil {
		return nil, err
	}
	headerTree, err := beaconBlock.Data.Header.Message.GetTree()
	if err != nil {
		return nil, err
	}
	stateProof, err := headerTree.Prove(BEACON_STATE_GINDEX)
	if err != nil {
		return nil, err
	}

	state, err := p.beaconState(ctx, currentSlot)
	if err != nil {
		return nil, err
	}

	stateTree, err := p.stateTree(state.Data.Deneb)
	if err != nil {
		return nil, err
	}
	rootGindex := calculateArrayGindex(new(big.Int).Mod(targetSlot, big.NewInt(SLOTS_PER_HISTORICAL_LIMIT)))
	historicalRootProof, err := stateTree.Prove(int(concatGindices([]int64{BLOCK_ROOTS_GINDEX, rootGindex})))
	if err != nil {
		return nil, err
	}

	return append(historicalRootProof.Hashes, stateProof.Hashes...), nil
}

func (p *ReceiptRootProver) beaconState(ctx context.Context, slot *big.Int) (*api.Response[*spec.VersionedBeaconState], error) {
	cachedState, ok := p.stateCache.Get(slot.String())
	if ok {
		return cachedState.(*api.Response[*spec.VersionedBeaconState]), nil
	}

	state, err := p.beaconClient.BeaconState(ctx, &api.BeaconStateOpts{
		State: strconv.FormatUint(slot.Uint64(), 10),
	})
	if err != nil {
		return nil, err
	}

	err = p.stateCache.Add(slot.String(), state, cache.DefaultExpiration)
	if err != nil {
		log.Err(err).Msgf("Failed saving state to cache")
	}

	return state, nil
}

func (p *ReceiptRootProver) stateTree(state *deneb.BeaconState) (*ssz.Node, error) {
	if p.spec == config.GnosisSpec {
		stateData, err := state.MarshalJSON()
		if err != nil {
			return nil, err
		}
		gnosisState := &gnosisDeneb.BeaconState{}
		err = gnosisState.UnmarshalJSON(stateData)
		if err != nil {
			return nil, err
		}

		return gnosisState.GetTree()
	}

	return state.GetTree()
}

func (p *ReceiptRootProver) blockTree(block *deneb.BeaconBlock) (*ssz.Node, error) {
	if p.spec == config.GnosisSpec {
		blockData, err := block.MarshalJSON()
		if err != nil {
			return nil, err
		}
		gnosisBlock := &gnosisDeneb.BeaconBlock{}
		err = gnosisBlock.UnmarshalJSON(blockData)
		if err != nil {
			return nil, err
		}

		return gnosisBlock.GetTree()
	}

	return block.GetTree()
}

func (p *ReceiptRootProver) receiptsRootProof(ctx context.Context, slot *big.Int) ([][]byte, error) {
	beaconBlock, err := p.beaconClient.SignedBeaconBlock(ctx, &api.SignedBeaconBlockOpts{
		Block: slot.String(),
	})
	if err != nil {
		return nil, err
	}
	blockTree, err := p.blockTree(beaconBlock.Data.Deneb.Message)
	if err != nil {
		return nil, err
	}
	receiptsRootProof, err := blockTree.Prove(RECEIPTS_ROOT_GINDEX)
	if err != nil {
		return nil, err
	}
	return receiptsRootProof.Hashes, nil
}

func calculateArrayGindex(elementIndex *big.Int) int64 {
	gindex := int64(1)
	index := elementIndex.Int64()

	depth := 0
	for (1 << depth) < SLOTS_PER_HISTORICAL_LIMIT {
		depth++
	}
	for d := 0; d < depth; d++ {
		gindex = (gindex << 1) | ((index >> (depth - d - 1)) & 1)
	}
	return gindex
}

func concatGindices(gindices []int64) int64 {
	binaryStr := "1"
	for _, gindex := range gindices {
		binary := big.NewInt(int64(gindex)).Text(2)
		binaryStr += binary[1:] // Skip the leading "1" for concatenation
	}

	result := new(big.Int)
	result.SetString(binaryStr, 2)
	return result.Int64()
}
