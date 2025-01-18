package Core

import (
	"math/bits"
	"slices"
)

type ProofData struct {
	DataProof   []byte
	ParityProof []byte
}

func NewProofData(dataProof, parityProof []byte) *ProofData {
	return &ProofData{
		DataProof:   dataProof,
		ParityProof: parityProof,
	}
}

func largestCombination(candidates []int) (ans int) {
	n := bits.Len(uint(slices.Max(candidates)))

	for i := 0; i < n; i++ {
		cnt := 0
		for _, num := range candidates {
			cnt += (num >> i) & 1
		}
		ans = max(ans, cnt)

	}
	return
}




