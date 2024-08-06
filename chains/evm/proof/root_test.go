// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package proof_test

import (
	"context"
	"encoding/hex"
	"math/big"
	"os"
	"testing"

	"github.com/attestantio/go-eth2-client/api"
	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/deneb"
	"github.com/stretchr/testify/suite"
	"github.com/sygmaprotocol/sygma-inclusion-prover/chains/evm/proof"
	"github.com/sygmaprotocol/sygma-inclusion-prover/mock"
	"go.uber.org/mock/gomock"
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
	ctrl := gomock.NewController(s.T())
	s.mockBeaconClient = mock.NewMockBeaconClient(ctrl)

	beaconBlockHeader := &apiv1.BeaconBlockHeader{}
	headerBytes, err := os.ReadFile("./stubs/header.json")
	if err != nil {
		panic(err)
	}
	_ = beaconBlockHeader.UnmarshalJSON(headerBytes)
	s.mockBeaconClient.EXPECT().BeaconBlockHeader(gomock.Any(), gomock.Any()).Return(&api.Response[*apiv1.BeaconBlockHeader]{
		Data: beaconBlockHeader,
	}, nil).AnyTimes()

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

	s.prover = proof.NewReceiptRootProver(s.mockBeaconClient)
}

func (s *ReceiptRootProofTestSuite) Test_ReceiptRootProof_SlotDifferent() {
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
		"324de181c3f04cb1a16f1bb16da93281c6d0201f220d28a849ce8ab2472f5baf",
		"154c88980564c7b6112ffc6c5f231f64755cc61d8dd3630ea62449605d4ee1ab",
		"4ee7055436bf367dc95da1447e06bcb3a01f666b6b0a3849e41531f9fe1a027b",
		"0f7f0a89fec31344d9951eef426f6e66c2b3046110dc89dd3cf8bdd67a1629ae",
		"c55a2b1579d3b1070ebd7c632dc015c2534f7d571dfefb51bf6fe84b61d08a44",
		"08ff7acf2cafee5dc6830950f71a11cfbce05fa3b75df7b99f5525df86f63792",
		"1ef8cdf86082b48a9f28e6a9db31cee9b6808cd5b3ed8c8ab9eac64c01cdfba6",
		"a12c7734c8170b73272ec633ac1b3c5441ca3f1f0508d024c43c1a3162cf5af4",
		"c6671a889699ae057f4a9b4f836d1913c2059a2c78f1dfeed8891f4ce27ec00d",
		"f40656ff428eed796debe9a19f1caa71c64750f70641ad325298cb9c22f5f7bc",
		"955510b9f25c5799bdb2641599dbe5acc2c30527db35b0f29750fc556aea565d",
		"5dbab0cfd3d8cb1ecf549645837c9e13f40e224147462d8d602f6604aa91effe",
		"df45c7d3b2b5bf7827f2fb9d827c87076a05a8b314d200b8d4832e1c3413eef2",
		"521b6c5edd664ac6e622e668bfb2fb1c13590d31318b24179106c04a35c9f5bb",
		"c54ea91aff432927f329abac170cf288c086159e640b6c24fcc03a416b5fc0b1",
		"549bdebf58b3048217f3853462337b18fb410d36768f35fa227b6e8bbedd5e82",
		"eab572026cded9cddada269b7054f84d839d2d424ee93065f9d07c026c24b7eb",
		"51df3ff10a34eafbecc355d793e4a12d43bc0df8787d2ec345e5b685e7f71da1",
		"7f81497ad30b4d3215e6c18222b277bf0dbc6ba315b1df3dd49a45c09f8aa569",
		"59f0fb8994c3f33d8aefaf635e57bd4732cc6f718e09ed4e5d78ca1e4876a802",
		"01a16c157ec2dca5bd7eef2436d104d5c6fd970b476bcec30f41baabbeddc815",
	}

	p, err := s.prover.ReceiptsRootProof(context.Background(), big.NewInt(5544654), big.NewInt(5544653))

	s.Nil(err)
	for i, hash := range expectedProof {
		s.Equal(hash, hex.EncodeToString(p[i]))
	}
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
