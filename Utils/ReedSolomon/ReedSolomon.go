package ReedSolomon

import "fmt"

type ReedSolomon struct {
	dataShardCount   int
	parityShardCount int
	totalShardCount  int
	matrix           *Matrix
	parityRows       [][]byte
}

func NewReedSolomon(dataShards, parityShards int) *ReedSolomon {
	total := dataShards + parityShards
	matrix := buildMatrix(dataShards, total)
	parityRows := make([][]byte, parityShards)
	for i := 0; i < parityShards; i++ {
		parityRows[i] = matrix.GetRow(dataShards + i)
	}
	return &ReedSolomon{
		dataShardCount:   dataShards,
		parityShardCount: parityShards,
		totalShardCount:  total,
		matrix:           matrix,
		parityRows:       parityRows,
	}
}

func (rs *ReedSolomon) GetDataShardCount() int   { return rs.dataShardCount }
func (rs *ReedSolomon) GetParityShardCount() int { return rs.parityShardCount }
func (rs *ReedSolomon) GetTotalShardCount() int  { return rs.totalShardCount }

//func (rs *ReedSolomon) CheckBuffersAndSizes(shards [][]byte, offset, byteCount int) {
//	if len(shards) != rs.totalShardCount {
//		panic("wrong number of shards")
//	}
//	shardLength := len(shards[0])
//	for _, s := range shards {
//		if len(s) != shardLength {
//			panic("shards different sizes")
//		}
//	}
//	if offset < 0 || byteCount < 0 || shardLength < offset+byteCount {
//		panic("invalid offset/byteCount")
//	}
//}

func (rs *ReedSolomon) CheckBuffersAndSize(shards []byte, offset, byteCount int) {
	// 检查分片数量是否与数据分片数量一致
	if len(shards) != rs.dataShardCount {
		panic(fmt.Sprintf("wrong number of datashards: %d", len(shards)))
	}

	// 检查偏移量和字节计数是否非负
	if offset < 0 {
		panic(fmt.Sprintf("offset is negative: %d", offset))
	}
	if byteCount < 0 {
		panic(fmt.Sprintf("byteCount is negative: %d", byteCount))
	}

	// 检查分片长度是否足够（确保与 Java 逻辑完全一致）
	if len(shards) < rs.dataShardCount {
		panic(fmt.Sprintf("buffers too small: %d", rs.dataShardCount))
	}
}

//func (rs *ReedSolomon) codeSomeShards(matrixRows, inputs, outputs [][]byte, outputCount, offset, byteCount int) [][]byte {
//	for iByte := offset; iByte < offset+byteCount; iByte++ {
//		for iRow := 0; iRow < outputCount; iRow++ {
//			matrixRow := matrixRows[iRow]
//			value := byte(0)
//			for c := 0; c < rs.dataShardCount; c++ {
//				value ^= Multiply(matrixRow[c], inputs[c][iByte])
//			}
//			outputs[iRow][iByte] = byte(value)
//		}
//	}
//	return outputs
//}

//func (rs *ReedSolomon) EncodeParities(shards [][]byte, offset, byteCount int) [][]byte {
//	rs.CheckBuffersAndSizes(shards, offset, byteCount)
//	outputs := make([][]byte, rs.parityShardCount)
//	for i := 0; i < rs.parityShardCount; i++ {
//		outputs[i] = shards[rs.dataShardCount+i]
//	}
//	return rs.codeSomeShards(rs.parityRows, shards, outputs, rs.parityShardCount, offset, byteCount)
//}

func (rs *ReedSolomon) EncodeParity(shards []byte, offset, byteCount int) []byte {
	// Check arguments.
	rs.CheckBuffersAndSize(shards, offset, byteCount)

	// Build the array of output buffers.
	outputs := make([]byte, rs.parityShardCount)

	// Do the coding.
	return rs.codeSingleShards(rs.parityRows, shards, outputs, rs.parityShardCount, offset, byteCount)
}

func (rs *ReedSolomon) codeSingleShards(matrixRows [][]byte, inputs []byte, outputs []byte, outputCount int, offset int, byteCount int) []byte {
	for iRow := 0; iRow < outputCount; iRow++ {
		matrixRow := matrixRows[iRow]
		value := byte(0)
		for c := 0; c < rs.dataShardCount; c++ {
			value ^= Multiply(matrixRow[c], inputs[c])
		}
		outputs[iRow] = value
	}
	return outputs
}

//func (rs *ReedSolomon) isParityCorrect(shards [][]byte, firstByte, byteCount int) bool {
//	rs.CheckBuffersAndSizes(shards, firstByte, byteCount)
//	toCheck := make([][]byte, rs.parityShardCount)
//	for i := 0; i < rs.parityShardCount; i++ {
//		toCheck[i] = shards[rs.dataShardCount+i]
//	}
//	return rs.checkSomeShards(rs.parityRows, shards, toCheck, rs.parityShardCount, firstByte, byteCount)
//}

//func (rs *ReedSolomon) checkSomeShards(matrixRows [][]byte, inputs [][]byte, toCheck [][]byte, checkCount int, offset int, byteCount int) bool {
//	for iByte := offset; iByte < offset+byteCount; iByte++ {
//		for iRow := 0; iRow < checkCount; iRow++ {
//			matrixRow := matrixRows[iRow]
//			value := byte(0)
//			for c := 0; c < rs.dataShardCount; c++ {
//				value ^= Multiply(matrixRow[c], inputs[c][iByte])
//			}
//			if toCheck[iRow][iByte] != byte(value) {
//				return false
//			}
//		}
//	}
//	return true
//}

func buildMatrix(dataShards, totalShards int) *Matrix {
	vandermonde := vandermonde(totalShards, dataShards)
	sub := vandermonde.Submatrix(0, 0, dataShards, dataShards)
	return vandermonde.Times(sub.Invert())
}

func vandermonde(rows, cols int) *Matrix {
	m := NewMatrix(rows, cols)
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			m.Set(r, c, Exp(byte(r), c))
		}
	}
	return m
}

//func (rs *ReedSolomon) decodeMissing(shards [][]byte, shardPresent []bool, offset, byteCount int) {
//	rs.CheckBuffersAndSizes(shards, offset, byteCount)
//
//	numberPresent := 0
//	for i := 0; i < rs.totalShardCount; i++ {
//		if shardPresent[i] {
//			numberPresent++
//		}
//	}
//	if numberPresent == rs.totalShardCount {
//		// Cool.  All of the shards data data.  We don't
//		// need to do anything.
//		return
//	}
//	if numberPresent < rs.dataShardCount {
//		panic("Not enough shards present")
//	}
//
//	subMatrix := NewMatrix(rs.dataShardCount, rs.dataShardCount)
//	subShards := make([][]byte, rs.dataShardCount)
//	subRow := 0
//	for matrixRow := 0; matrixRow < rs.totalShardCount && subRow < rs.dataShardCount; matrixRow++ {
//		if shardPresent[matrixRow] {
//			for c := 0; c < rs.dataShardCount; c++ {
//				subMatrix.Set(subRow, c, rs.matrix.Get(matrixRow, c))
//			}
//			subShards[subRow] = shards[matrixRow]
//			subRow++
//		}
//	}
//
//	dataDecodeMatrix := subMatrix.Invert()
//	outputs := make([][]byte, rs.parityShardCount)
//	matrixRows := make([][]byte, rs.parityShardCount)
//	outputCount := 0
//
//	for iShard := 0; iShard < rs.dataShardCount; iShard++ {
//		if !shardPresent[iShard] {
//			outputs[outputCount] = shards[iShard]
//			matrixRows[outputCount] = dataDecodeMatrix.GetRow(iShard)
//			outputCount++
//		}
//	}
//	rs.codeSomeShards(matrixRows, subShards, outputs, outputCount, offset, byteCount)
//
//	outputCount = 0
//	for iShard := rs.dataShardCount; iShard < rs.totalShardCount; iShard++ {
//		if !shardPresent[iShard] {
//			outputs[outputCount] = shards[iShard]
//			matrixRows[outputCount] = rs.parityRows[iShard-rs.dataShardCount]
//			outputCount++
//		}
//	}
//	rs.codeSomeShards(matrixRows, shards, outputs, outputCount, offset, byteCount)
//}
