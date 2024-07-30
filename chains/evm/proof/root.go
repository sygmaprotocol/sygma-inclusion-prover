package proof

import (
	"context"
	"math/big"
	"strconv"

	"github.com/attestantio/go-eth2-client/api"
	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
)

const (
	BEACON_STATE_GINDEX        = 11
	RECEIPTS_ROOT_GINDEX       = 6435
	BLOCK_ROOTS_GINDEX         = 37
	SLOTS_PER_HISTORICAL_LIMIT = 8192
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
}

func NewReceiptRootProver(beaconClient BeaconClient) *ReceiptRootProver {
	return &ReceiptRootProver{
		beaconClient: beaconClient,
	}
}

// ReceiptRootProof returns the prove from the beacon block root to the receipt roof of the given slot.
// The path for the proof is beacon block -> beacon state -> block roots -> execution payload header -> receipt root.
func (p *ReceiptRootProver) ReceiptsRootProof(ctx context.Context, currentSlot *big.Int, targetSlot *big.Int) ([][]byte, error) {
	beaconStateProof, err := p.historicalRootProof(ctx, currentSlot, targetSlot)
	if err != nil {
		return nil, err
	}
	receiptsRootProof, err := p.receiptsRootProof(ctx, targetSlot)
	if err != nil {
		return nil, err
	}

	if currentSlot.Cmp(targetSlot) != 0 {
		return append(receiptsRootProof, beaconStateProof...), nil
	} else {
		return receiptsRootProof, nil
	}
}

func (p *ReceiptRootProver) historicalRootProof(ctx context.Context, currentSlot *big.Int, targetSlot *big.Int) ([][]byte, error) {
	beaconBlockHeader, err := p.beaconClient.BeaconBlockHeader(ctx, &api.BeaconBlockHeaderOpts{
		Block: currentSlot.String(),
	})
	if err != nil {
		return nil, err
	}
	headerTree, err := beaconBlockHeader.Data.Header.Message.GetTree()
	if err != nil {
		return nil, err
	}
	stateProof, err := headerTree.Prove(BEACON_STATE_GINDEX)
	if err != nil {
		return nil, err
	}

	state, err := p.beaconClient.BeaconState(ctx, &api.BeaconStateOpts{
		State: beaconBlockHeader.Data.Header.Message.StateRoot.String(),
	})
	if err != nil {
		return nil, err
	}
	stateTree, err := state.Data.Deneb.GetTree()
	if err != nil {
		return nil, err
	}

	rootGindex, err := calculateGindex(new(big.Int).Mod(targetSlot, big.NewInt(SLOTS_PER_HISTORICAL_LIMIT)))
	if err != nil {
		return nil, err
	}
	historicalRootProof, err := stateTree.Prove(int(concatGindices([]*big.Int{big.NewInt(BLOCK_ROOTS_GINDEX), big.NewInt(int64(rootGindex))}).Int64()))
	if err != nil {
		return nil, err
	}

	return append(historicalRootProof.Hashes, stateProof.Hashes...), nil
}

func (p *ReceiptRootProver) receiptsRootProof(ctx context.Context, slot *big.Int) ([][]byte, error) {
	beaconBlockHeader, err := p.beaconClient.SignedBeaconBlock(ctx, &api.SignedBeaconBlockOpts{
		Block: slot.String(),
	})
	if err != nil {
		return nil, err
	}
	headerTree, err := beaconBlockHeader.Data.Deneb.Message.GetTree()
	if err != nil {
		return nil, err
	}
	receiptsRootProof, err := headerTree.Prove(RECEIPTS_ROOT_GINDEX)
	if err != nil {
		return nil, err
	}
	return receiptsRootProof.Hashes, nil
}

func calculateGindex(index *big.Int) (uint64, error) {
	binaryIndex := strconv.FormatUint(index.Uint64(), 2)
	gindex := "1" + binaryIndex
	return strconv.ParseUint(gindex, 2, 64)
}

func concatGindices(gindices []*big.Int) *big.Int {
	binaryStr := "1"
	for _, gindex := range gindices {
		binary := gindex.Text(2)
		binaryStr += binary[1:] // Skip the leading "1" for concatenation
	}

	result := new(big.Int)
	result.SetString(binaryStr, 2)
	return result
}
