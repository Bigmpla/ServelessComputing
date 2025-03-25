package Core

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
