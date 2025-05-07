package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"serveless_Go/Core"
	"serveless_Go/Utils"
	"time"
)

func main() {
	//the path of document need to be store
	filePath := ""

	var blockShards, dataShards int
	fmt.Println("Please input the BLOCK_SHARDS:")
	fmt.Scan(&blockShards)
	fmt.Println("Please input the DATA_SHARDS:")
	fmt.Scan(&dataShards)

	fmt.Printf("The ReedSolomon parameters: (BLOCK_SHARDS, DATA_SHARDS)=(%d, %d)\n", blockShards, dataShards)
	auditTask(filePath, blockShards, dataShards, 1)
}

// auditTask 对应 Java 中的 auditTask 方法
func auditTask(filePath string, blockShards, dataShards, taskCount int) {
	integrityAuditing := Core.NewIntegrityAuditingFromFile(filePath, blockShards, dataShards)

	timeCosts := make([]int64, 5)

	// KeyGen 阶段
	fmt.Println("---KeyGen phase start---")
	startTime := time.Now()
	integrityAuditing.GenKey()
	timeCosts[0] = time.Since(startTime).Nanoseconds()
	fmt.Println("---KeyGen phase finished---")

	// OutSource 阶段
	fmt.Println("---OutSource phase start---")
	startTime = time.Now()
	timeCosts[1] = integrityAuditing.OutSource()
	//the path that storing source file
	uploadSourceFilePath := ""
	//the path that storing parity file
	uploadParitiesPath := ""

	// 将源文件存储到本地
	os.WriteFile(uploadSourceFilePath, bytes.Join(integrityAuditing.OriginalData, nil), 0644)
	fmt.Println("store file in local")

	// 计算源文件大小
	sourceFileInfo, _ := os.Stat(uploadSourceFilePath)
	sourceFileSize := sourceFileInfo.Size()

	// 将奇偶校验数据存储到本地
	os.WriteFile(uploadParitiesPath, bytes.Join(integrityAuditing.Parity, nil), 0644)
	fmt.Println("store tags in local")

	// 计算额外存储开销
	paritiesInfo, _ := os.Stat(uploadParitiesPath)
	extraStorageSize := paritiesInfo.Size()
	fmt.Printf("extraStorageSize is %d Bytes\n", extraStorageSize)

	// 上传文件和奇偶校验数据到云存储
	startTime = time.Now()
	cloudAPI := Utils.NewCloudAPI()
	print(cloudAPI.BlobName + "\n")
	cloudAPI.UploadFile(uploadSourceFilePath, "source.txt")
	cloudAPI.UploadFile(uploadParitiesPath, "parities.txt")
	fmt.Println("upload File and tags to COS")
	timeCosts[2] = time.Since(startTime).Nanoseconds()
	fmt.Println("---OutSource phase finished---")

	////自己给parity赋值,控制变量
	//integrityAuditing.Key = "cmbSGrmHl4ozuiuf"
	//integrityAuditing.Setskey("h2")
	//combinedBase64 := "java输出的parity"
	//parts := strings.Split(combinedBase64, ",")
	//var result [][]byte
	//for _, part := range parts {
	//	decoded, err := base64.StdEncoding.DecodeString(part)
	//	if err != nil {
	//		fmt.Println("Error decoding:", err)
	//		return
	//	}
	//	result = append(result, decoded)
	//}
	//integrityAuditing.Parity = result

	// 触发 SCF 处理
	reqPath := "https://goscf.azurewebsites.net/api/Handler?code=lQAf3tzpe-mrPQXl5lZU1FrCtY2R2zR9O2MuSjZZ-o8jAzFumY5Rzg=="

	// 准备挑战数据
	fmt.Println("---Audit phase start---")
	startTime = time.Now()
	challengeData := integrityAuditing.Audit(460)

	// 这里先将index和coefficients写死！！
	//index := []int{1, 2, 3}
	//coefficients := []byte{1, 2, 3}
	//challengeData := Core.InitCD(index, coefficients)

	timeCosts[3] = time.Since(startTime).Nanoseconds()

	// // 发送挑战数据
	// requestBody, _ := json.Marshal(map[string]interface{}{
	// 	"DATA_SHARDS":   dataShards,
	// 	"PARITY_SHARDS": blockShards - dataShards,
	// 	"challengeData": challengeData,
	// })
	// fmt.Println("challengeData str:", string(requestBody))
	// resp, err := http.Post(reqPath, "application/json", bytes.NewBuffer(requestBody))
	// if err != nil {
	// 	log.Fatalf("Error sending request: %v", err)
	// }
	// defer resp.Body.Close()

	params := map[string]interface{}{
		"DATA_SHARDS":   fmt.Sprint(dataShards),
		"PARITY_SHARDS": fmt.Sprint(blockShards - dataShards),
	}

	paramsJSON, _ := json.Marshal(params)
	challengeJSON, _ := json.Marshal(challengeData)

	var requestBody bytes.Buffer
	requestBody.Write(paramsJSON)
	requestBody.Write(challengeJSON)
	fmt.Printf("paramsJSON str: %s\n", paramsJSON)
	fmt.Printf("challengeData str: %s\n", challengeJSON)

	resp, err := http.Post(reqPath, "application/json", &requestBody)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	responseDataStr, _ := io.ReadAll(resp.Body)
	fmt.Printf("Get respond content: %s\n", responseDataStr)

	// 解析响应中的证明数据
	var proofData Core.ProofData
	if err := json.Unmarshal(responseDataStr, &proofData); err != nil {
		log.Fatalf("Error parsing proof data: %v", err)
	}

	// 计算通信开销
	proofDataSize := int64(len(proofData.DataProof) + len(proofData.ParityProof))
	fmt.Printf("proofDataSize is %d Bytes\n", proofDataSize)

	fmt.Println(proofData)
	// 验证阶段
	fmt.Println("---Verify phase start---")
	startTime = time.Now()

	//a := []byte{10, 101, 116}
	//b := []byte{87, 178}
	//proofData.DataProof = a
	//proofData.ParityProof = b

	if integrityAuditing.Verify(&challengeData, &proofData) {
		fmt.Println("---Verify phase finished---")
		fmt.Println("The data is intact in the cloud. The auditing process is success!")
	} else {
		fmt.Println("The data is not intact in the cloud. The auditing process is failed!")
	}
	timeCosts[4] = time.Since(startTime).Nanoseconds()

	// 存储性能结果
	//the path that storing result file
	performanceFilePath := ""
	performanceFile, err := os.OpenFile(performanceFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error opening performance file: %v", err)
	}
	defer performanceFile.Close()

	title := fmt.Sprintf("Audit data size is %d. No. %d audit process. \r\n", sourceFileSize, taskCount)
	performanceFile.WriteString(title)

	performanceFile.WriteString(fmt.Sprintf("StorageCost %d  CommunicationCost %d\r\n", extraStorageSize, proofDataSize))
	for i, t := range timeCosts {
		performanceFile.WriteString(fmt.Sprintf("time[%d] = %d  ", i, t))
	}
	performanceFile.WriteString("\r\n")
}
