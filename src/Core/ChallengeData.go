package Core

// ChallengeData 用于存储挑战数据，包括两个部分：系数和索引
type ChallengeData struct {
	Coefficients []byte
	Index        []int
}

func initCD(index []int, coefficients []byte) *ChallengeData {
	return &ChallengeData{
		Coefficients: coefficients,
		Index:        index,
	}
}
