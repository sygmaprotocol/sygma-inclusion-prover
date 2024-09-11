// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package proof_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/attestantio/go-eth2-client/http"
	"github.com/stretchr/testify/suite"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/config"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/proof"
	"github.com/sygmaprotocol/sygma-inclusion-prover/mock"
)

type ReceiptRootProofTestSuite struct {
	suite.Suite

	prover           *proof.ReceiptRootProver
	mockBeaconClient *mock.MockBeaconClient
}

func TestRunReceiptRootProofTestSuite(t *testing.T) {
	suite.Run(t, new(ReceiptRootProofTestSuite))
}

func (s *ReceiptRootProofTestSuite) SetupTest() {
	/*
		ctrl := gomock.NewController(s.T())
		s.mockBeaconClient = mock.NewMockBeaconClient(ctrl)

		beaconState := &spec.VersionedBeaconState{
			Deneb: &deneb.BeaconState{},
		}
		beaconBytes, err := os.ReadFile("./stubs/state.json")
		if err != nil {
			panic(err)
		}
		_ = beaconState.Deneb.UnmarshalJSON(beaconBytes)
		s.mockBeaconClient.EXPECT().BeaconState(gomock.Any(), gomock.Any()).Return(&api.Response[*spec.VersionedBeaconState]{
			Data: beaconState,
		}, nil).AnyTimes()

		block := &spec.VersionedSignedBeaconBlock{
			Deneb: &deneb.SignedBeaconBlock{},
		}
		blockBytes, err := os.ReadFile("./stubs/block.json")
		if err != nil {
			panic(err)
		}
		_ = block.Deneb.UnmarshalJSON(blockBytes)
		s.mockBeaconClient.EXPECT().SignedBeaconBlock(gomock.Any(), gomock.Any()).Return(&api.Response[*spec.VersionedSignedBeaconBlock]{
			Data: block,
		}, nil).AnyTimes()

		beaconBlockHeader := &apiv1.BeaconBlockHeader{}
		headerBytes, err := os.ReadFile("./stubs/header.json")
		if err != nil {
			panic(err)
		}
		_ = beaconBlockHeader.UnmarshalJSON(headerBytes)
		s.mockBeaconClient.EXPECT().BeaconBlockHeader(gomock.Any(), gomock.Any()).Return(&api.Response[*apiv1.BeaconBlockHeader]{
			Data: beaconBlockHeader,
		}, nil).AnyTimes()
	*/

	ctx, _ := context.WithCancel(context.Background())
	beaconClient, err := http.New(ctx,
		http.WithAddress("http://143.198.134.107:5053"),
		http.WithTimeout(time.Minute*15),
		http.WithEnforceJSON(false),
	)
	if err != nil {
		panic(err)
	}
	beaconProvider := beaconClient.(*http.Service)

	s.prover = proof.NewReceiptRootProver(beaconProvider, config.MainnetSpec)
}

func (s *ReceiptRootProofTestSuite) Test_ReceiptRootProof_SlotDifferent() {
	p, err := s.prover.ReceiptsRootProof(context.Background(), big.NewInt(9936787), big.NewInt(9936784))
	fmt.Println(p)
	s.Nil(err)
	p, err = s.prover.ReceiptsRootProof(context.Background(), big.NewInt(9936787), big.NewInt(9936784))
	s.Nil(err)
	fmt.Println(p)
}

func (s *ReceiptRootProofTestSuite) Test_ReceiptRootProof_SameSlot() {
	expectedProof := []string{
		"f84e9fc70392b8a60d7655b1924c9aa98fb8ad0a6cbb8a3f945a1e5ddf83b9b3",
		"97d7ab2b13aca2be860cd70d88b718f3ebe2ed251a4a032368812364c8298c9c",
		"8b0ec84446342d9bc11852eb2bd7f2fa64d29389e3a6e28bffd2d2edc4e037a5",
		"043cfab8f5f81f18af8f16eb656ccfbefebf8665aa9afd3a8639dd47168b7727",
		"91b2a710625c221d8c5b33bdfaadd77ec67c07a6950ef1e24f9d5e7822ecde33",
		"d5f4a630033673a515b904c5a74a439a9eaba6f18f207795a98974abb1335fbb",
		"b46f0c01805fe212e15907981b757e6c496b0cb06664224655613dcec82505bb",
		"db56114e00fdd4c1f85c892bf35ac9a89289aaecb1ebd0a96cde606a748b5d71",
		"44f37a37ebc7bf7ddcfb20d3b23e76e558e4f54ce6d902fa0e5de50f9ac36296",
		"0000000000000000000000000000000000000000000000000000000000000000",
		"f5a5fd42d16a20302798ef6ed309979b43003d2320d9f0e8ea9831a92759fb4b",
		"7dbd839dc4fcf9a1b024067aab7d6e3fe2a30e18ae08a19db4c9243f49a28c67",
	}

	p, err := s.prover.ReceiptsRootProof(context.Background(), big.NewInt(1), big.NewInt(1))

	s.Nil(err)
	for i, hash := range expectedProof {
		s.Equal(hash, hex.EncodeToString(p[i]))
	}

}
