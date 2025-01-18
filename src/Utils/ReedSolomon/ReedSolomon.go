package ReedSolomon

import (
	"fmt"
)

type ReedSolomon struct {
	dataShardCount   int
	parityShardCount int
	totalShardCount  int
	matrix           *Matrix
	parityRows       [][]byte
}

func (rs *ReedSolomon) GetDataShardCount() int {
	return rs.dataShardCount
}

func (rs *ReedSolomon) GetParityShardCount() int {
	return rs.parityShardCount
}

func (rs *ReedSolomon) GetTotalShardCount() int {
	return rs.totalShardCount
}

func (rs *ReedSolomon) EncodeParity(shards []byte, offset, byteCount int) []byte {
	rs.checkBuffersAndSize(shards, offset, byteCount)
	outputs := make([]byte, rs.parityShardCount)
	return rs.codesingleShards(rs.parityRows, shards, outputs, rs.parityShardCount, offset, byteCount)
}

func (rs *ReedSolomon) IsParityCorrect(shards [][]byte, firstByte, byteCount int) bool {
	rs.checkBuffersAndSizes(shards, firstByte, byteCount)
	toCheck := make([][]byte, rs.parityShardCount)

	for i := 0; i < rs.parityShardCount; i++ {
		toCheck[i] = shards[rs.dataShardCount+i]
	}

	return rs.checkSomeShards(rs.parityRows, shards, toCheck, rs.parityShardCount, firstByte, byteCount)
}

func (rs *ReedSolomon) DecodeMissing(shards [][]byte, shardPresent []bool, offset, byteCount int) {
	rs.checkBuffersAndSizes(shards, offset, byteCount)
	numberPresent := 0

	for i := 0; i < rs.totalShardCount; i++ {
		if shardPresent[i] {
			numberPresent++
		}
	}

	if numberPresent != rs.totalShardCount {
		if numberPresent < rs.dataShardCount {
			panic("Not enough shards present")
		} else {
			subMatrix := NewMatrix(rs.dataShardCount, rs.dataShardCount)
			subShards := make([][]byte, rs.dataShardCount)
			subMatrixRow := 0

			for matrixRow := 0; matrixRow < rs.totalShardCount && subMatrixRow < rs.dataShardCount; matrixRow++ {
				if shardPresent[matrixRow] {
					for c := 0; c < rs.dataShardCount; c++ {
						subMatrix.Set(subMatrixRow, c, rs.matrix.Get(matrixRow, c))
					}

					subShards[subMatrixRow] = shards[matrixRow]
					subMatrixRow++
				}
			}

			dataDecodeMatrix := subMatrix.Invert()
			outputs := make([][]byte, rs.parityShardCount)
			matrixRows := make([][]byte, rs.parityShardCount)
			outputCount := 0

			for iShard := 0; iShard < rs.dataShardCount; iShard++ {
				if !shardPresent[iShard] {
					outputs[outputCount] = shards[iShard]
					matrixRows[outputCount] = dataDecodeMatrix.GetRow(iShard)
					outputCount++
				}
			}

			rs.codeSomeShards(matrixRows, subShards, outputs, outputCount, offset, byteCount)
			outputCount = 0

			for iShard := rs.dataShardCount; iShard < rs.totalShardCount; iShard++ {
				if !shardPresent[iShard] {
					outputs[outputCount] = shards[iShard]
					matrixRows[outputCount] = rs.parityRows[iShard-rs.dataShardCount]
					outputCount++
				}
			}

			rs.codeSomeShards(matrixRows, shards, outputs, outputCount, offset, byteCount)
		}
	}
}

func (rs *ReedSolomon) checkBuffersAndSizes(shards [][]byte, offset, byteCount int) {
	if len(shards) != rs.totalShardCount {
		panic(fmt.Sprintf("wrong number of shards: %d", len(shards)))
	}

	shardLength := len(shards[0])

	for i := 1; i < len(shards); i++ {
		if len(shards[i]) != shardLength {
			panic("Shards are different sizes")
		}
	}

	if offset < 0 {
		panic(fmt.Sprintf("offset is negative: %d", offset))
	}
	if byteCount < 0 {
		panic(fmt.Sprintf("byteCount is negative: %d", byteCount))
	}
	if shardLength < offset+byteCount {
		panic(fmt.Sprintf("buffers to small: %d + %d", offset, byteCount))
	}
}

func (rs *ReedSolomon) checkBuffersAndSize(shards []byte, offset, byteCount int) {
	if len(shards) != rs.dataShardCount {
		panic(fmt.Sprintf("wrong number of datashards: %d", len(shards)))
	}
	if offset < 0 {
		panic(fmt.Sprintf("offset is negative: %d", offset))
	}
	if byteCount < 0 {
		panic(fmt.Sprintf("byteCount is negative: %d", byteCount))
	}
	if len(shards) < rs.dataShardCount {
		panic(fmt.Sprintf("buffers to small: %d", rs.dataShardCount))
	}
}

func (rs *ReedSolomon) codesingleShards(matrixRows [][]byte, inputs []byte, outputs []byte, outputCount, offset, byteCount int) []byte {
	for iRow := 0; iRow < outputCount; iRow++ {
		matrixRow := matrixRows[iRow]
		var value byte = 0

		for c := 0; c < rs.dataShardCount; c++ {
			value ^= Multiply(matrixRow[c], inputs[c])
		}

		outputs[iRow] = byte(value)
	}

	return outputs
}

func (rs *ReedSolomon) codeSomeShards(matrixRows [][]byte, inputs [][]byte, outputs [][]byte, outputCount, offset, byteCount int) [][]byte {
	for iByte := offset; iByte < offset+byteCount; iByte++ {
		for iRow := 0; iRow < outputCount; iRow++ {
			matrixRow := matrixRows[iRow]
			var value byte = 0

			for c := 0; c < rs.dataShardCount; c++ {
				value ^= Multiply(matrixRow[c], inputs[c][iByte])
			}

			outputs[iRow][iByte] = value
		}
	}

	return outputs
}

func (rs *ReedSolomon) checkSomeShards(matrixRows [][]byte, inputs [][]byte, toCheck [][]byte, checkCount, offset, byteCount int) bool {
	for iByte := offset; iByte < offset+byteCount; iByte++ {
		for iRow := 0; iRow < checkCount; iRow++ {
			matrixRow := matrixRows[iRow]
			var value byte = 0

			for c := 0; c < rs.dataShardCount; c++ {
				value ^= Multiply(matrixRow[c], inputs[c][iByte])
			}

			if toCheck[iRow][iByte] != value {
				return false
			}
		}
	}

	return true
}

func buildMatrix(dataShardCount, totalShardCount int) *Matrix {
	vandermonde := Vandermonde(totalShardCount, dataShardCount)
	top := vandermonde.Submatrix(0, 0, dataShardCount, dataShardCount)
	return vandermonde.Times(top.Invert())
}

func Vandermonde(rows, cols int) *Matrix {
	result := NewMatrix(rows, cols)
	// Populate Vandermonde matrix using Galois field operations
	return result
}
