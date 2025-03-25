package Core

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"serveless_Go/Utils/ReedSolomon"

	// "strconv"
	"time"
)

// IntegrityAuditing represents the integrity auditing process.
type IntegrityAuditing struct {
	DATA_SHARDS   int      // Number of data shards
	PARITY_SHARDS int      // Number of parity shards
	SHARD_NUMBER  int      // Total number of shards
	fileSize      int64    // Size of the input file
	storeSize     int64    // Size of the stored data
	Key           string   // Secret key
	sKey          string   // Second secret key
	filePath      string   // Path to the input file
	OriginalData  [][]byte // Original data stored as blocks
	Parity        [][]byte // Calculated parity data
	len           int      // Length of the key (16 characters)
}

// NewIntegrityAuditing creates a new IntegrityAuditing instance (used in SCF).
func NewIntegrityAuditing(DATA_SHARDS, PARITY_SHARDS int) *IntegrityAuditing {
	return &IntegrityAuditing{
		DATA_SHARDS:   DATA_SHARDS,
		PARITY_SHARDS: PARITY_SHARDS,
	}
}

func NewIntegrityAuditingFromFile1() *IntegrityAuditing {
	return &IntegrityAuditing{
		DATA_SHARDS:   5,
		PARITY_SHARDS: 2,
		len:           16,
		SHARD_NUMBER:  11420,
		fileSize:      34255,
		storeSize:     34259,
		sKey:          "Wb",
		Key:           "KQmmnw0FpbD8H826",
	}
}

// NewIntegrityAuditingFromFile creates a new IntegrityAuditing instance (used in client).
func NewIntegrityAuditingFromFile(filePath string, BLOCK_SHARDS, DATA_SHARDS int) *IntegrityAuditing {
	ia := &IntegrityAuditing{
		filePath:      filePath,
		DATA_SHARDS:   DATA_SHARDS,
		PARITY_SHARDS: BLOCK_SHARDS - DATA_SHARDS,
		len:           16,
	}

	// Calculate SHARD_NUMBER
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil
	}
	ia.fileSize = fileInfo.Size()
	ia.storeSize = ia.fileSize + 4 // BYTES_IN_INT = 4
	ia.SHARD_NUMBER = (int(ia.storeSize) + DATA_SHARDS - 1) / DATA_SHARDS

	// Read original data
	ia.OriginalData = make([][]byte, ia.SHARD_NUMBER)
	for i := range ia.OriginalData {
		ia.OriginalData[i] = make([]byte, DATA_SHARDS)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer file.Close()

	for i := 0; i < ia.SHARD_NUMBER; i++ {
		_, err := file.Read(ia.OriginalData[i])
		if err != nil && err != io.EOF {
			return nil
		}
	}

	return ia
}

func (ia *IntegrityAuditing) Setskey(key string) {
	ia.sKey = key
}

// GenKey generates two secret keys.
func (ia *IntegrityAuditing) GenKey() {
	chars := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	r := rand.New(rand.NewSource(1234))
	sBuffer1 := make([]byte, ia.len)
	for i := range sBuffer1 {
		sBuffer1[i] = chars[r.Intn(len(chars))]
	}

	sBuffer2 := make([]byte, ia.PARITY_SHARDS)
	for i := range sBuffer2 {
		sBuffer2[i] = chars[r.Intn(len(chars))]
	}

	ia.Key = string(sBuffer1)
	ia.sKey = string(sBuffer2)
	println("Key: ", ia.Key)
	println("sKey: ", ia.sKey)
}

// OutSource calculates the tags of the source data.
func (ia *IntegrityAuditing) OutSource() int64 {
	startTime := time.Now().UnixNano()

	ia.Parity = make([][]byte, ia.SHARD_NUMBER)
	reedSolomon := ReedSolomon.NewReedSolomon(ia.DATA_SHARDS, ia.PARITY_SHARDS)
	for i := 0; i < ia.SHARD_NUMBER; i++ {
		ia.Parity[i] = reedSolomon.EncodeParity(ia.OriginalData[i], 0, 1)
	}

	// Multiply by sKey
	for i := range ia.Parity {
		sKeyBytes := []byte(ia.sKey)
		if len(sKeyBytes) != len(ia.Parity[i]) {
			fmt.Println("Error: sKeyBytes.length != parity.length")
		} else {
			for j := range ia.Parity[i] {
				ia.Parity[i][j] = ReedSolomon.Multiply(ia.Parity[i][j], sKeyBytes[j])
			}
		}
	}

	// Add a pseudo-random number
	for i := range ia.Parity {
		// pr := &PseudoRandom{}
		randoms, _ := GenerateRandom(i, ia.Key, ia.PARITY_SHARDS)
		for j := 0; j < ia.PARITY_SHARDS; j++ {
			ia.Parity[i][j] = ReedSolomon.Add(ia.Parity[i][j], randoms[j])
		}
	}

	endTime := time.Now().UnixNano()
	fmt.Println("Process phase finished")
	return endTime - startTime
}

// Audit generates the challenge data for auditing.
func (ia *IntegrityAuditing) Audit(challengeLen int) ChallengeData {
	coefficients := make([]byte, challengeLen)
	index := make([]int, challengeLen)

	r := rand.New(rand.NewSource(1234))

	// rand.Seed(time.Now().UnixNano())
	for i := range index {
		index[i] = r.Intn(ia.SHARD_NUMBER)
	}
	r.Read(coefficients)
	return ChallengeData{
		Index:        index,
		Coefficients: coefficients,
	}
}


// Prove calculates the proof data after receiving the challenge data.
func (ia *IntegrityAuditing) Prove(challengeData *ChallengeData, downloadData, downloadParity [][]byte) ProofData {
	dataProof := make([]byte, ia.DATA_SHARDS)
	parityProof := make([]byte, ia.PARITY_SHARDS)

	for i := range challengeData.Index {
		tempData := make([]byte, ia.DATA_SHARDS)
		tempParity := make([]byte, ia.PARITY_SHARDS)

		for j := range tempParity {
			tempParity[j] = ReedSolomon.Multiply(challengeData.Coefficients[i], downloadParity[i][j])
		}
		for j := range tempData {
			tempData[j] = ReedSolomon.Multiply(challengeData.Coefficients[i], downloadData[i][j])
		}

		for j := range parityProof {
			parityProof[j] = ReedSolomon.Add(parityProof[j], tempParity[j])
		}
		for j := range dataProof {
			dataProof[j] = ReedSolomon.Add(dataProof[j], tempData[j])
		}
	}

	return ProofData{
		DataProof:   dataProof,
		ParityProof: parityProof,
	}
}

// Verify calculates the integrity audit result.
func (ia *IntegrityAuditing) Verify(challengeData *ChallengeData, proofData *ProofData) bool {
	verifyParity := make([]byte, ia.PARITY_SHARDS)
	sumTemp := make([]byte, ia.PARITY_SHARDS)

	for i := range challengeData.Index {
		// pr := &PseudoRandom{}
		AESRandomByte, _ := GenerateRandom(challengeData.Index[i], ia.Key, ia.PARITY_SHARDS)
		temp := make([]byte, ia.PARITY_SHARDS)

		for j := range temp {
			temp[j] = ReedSolomon.Multiply(challengeData.Coefficients[i], AESRandomByte[j])
		}
		for k := range sumTemp {
			sumTemp[k] = ReedSolomon.Add(sumTemp[k], temp[k])
		}
	}

	for i := range verifyParity {
		verifyParity[i] = ReedSolomon.Sub(proofData.ParityProof[i], sumTemp[i])
	}

	for j := range verifyParity {
		sKeyBytes := []byte(ia.sKey)
		verifyParity[j] = ReedSolomon.Divide(verifyParity[j], sKeyBytes[j])
	}

	reedSolomon := ReedSolomon.NewReedSolomon(ia.DATA_SHARDS, ia.PARITY_SHARDS)
	//reCalParity和proofData.DataProof有关
	reCalParity := reedSolomon.EncodeParity(proofData.DataProof, 0, 1)

	return bytes.Equal(verifyParity, reCalParity)
}

// // CompareByteArray compares two byte arrays.
// func compareByteArray(a, b []byte) bool {
// 	if a == nil || b == nil {
// 		return false
// 	}
// 	if len(a) != len(b) {
// 		return false
// 	}
// 	return bytes.Equal(a, b)
// }
