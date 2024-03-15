// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package util

import (
	"encoding/hex"
	"strings"
)

func SliceTo32Bytes(in []byte) [32]byte {
	var res [32]byte
	copy(res[:], in)
	return res
}

func ToByteArray(slice []string) ([][]byte, error) {
	result := make([][]byte, len(slice))
	for i, elem := range slice {
		elem = strings.TrimPrefix(elem, "0x")
		elemBytes, err := hex.DecodeString(elem)
		if err != nil {
			return nil, err
		}
		result[i] = elemBytes
	}
	return result, nil
}
