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
	"strings"
"encoding/base64"
	"strconv"
)

type Claims struct {
	Issuer         string                 `json:"Issuer"`
	OriginalIssuer string                 `json:"OriginalIssuer"`
	Properties     map[string]interface{} `json:"Properties"`
	Type           string                 `json:"Type"`
	Value          string                 `json:"Value"`
	ValueType      string                 `json:"ValueType"`
}

type Identity struct {
	AuthenticationType string   `json:"AuthenticationType"`
	IsAuthenticated    bool     `json:"IsAuthenticated"`
	Claims             []Claims `json:"Claims"`
}

type Req struct {
	Url     string              `json:"Url"`
	Method  string              `json:"Method"`
	Headers map[string][]string `json:"Headers"`
	Body    string              `json:"Body"`
}

type Data struct {
	Req Req `json:"req"`
}

type Metadata struct {
	Headers map[string]string `json:"Headers"`
	Sys     struct {
		MethodName string `json:"MethodName"`
		UtcNow     string `json:"UtcNow"`
		RandGuid   string `json:"RandGuid"`
	} `json:"sys"`
}

type JsonData struct {
	Data     Data     `json:"Data"`
	Metadata Metadata `json:"Metadata"`
}

type InvokeRequest struct {
	Data     map[string]json.RawMessage
	Metadata map[string]interface{}
}

type InvokeResponse struct {
	Outputs     map[string]interface{}
	Logs        []string
	ReturnValue interface{}
}

// export Handler
func HttpHandler(w http.ResponseWriter, r *http.Request) {
	logger := log.Default()

	// 1. 检查请求体
	if r.Body == nil {
		http.Error(w, "Request body is missing", http.StatusBadRequest)
		return
	}
	logger.Println(r.Body)
	logger.Println("Received data successfully")

	//var challenge Core.ChallengeData
	bodyBytes,
		_ := io.ReadAll(r.Body)
	bodyStr := string(bodyBytes)

	logger.Println(bodyStr)

	var data JsonData

	// 解码 JSON 数据
	err := json.Unmarshal(bodyBytes, &data)
	if err != nil {
		log.Fatal(err)
	}

	// 2. 输出解析的数据
	challenge := Core.ChallengeData{}

	parityShards, _ := strconv.Atoi(extractJsonValue(data.Data.Req.Body, "PARITY_SHARDS"))
	dataShards, _ := strconv.Atoi(extractJsonValue(data.Data.Req.Body, "DATA_SHARDS"))

	//Coefficients是byte数组，Go 的 json.Marshal([]byte) 会自动转换成 Base64 字符串，因为它认为 []byte 是二进制数据。
	//所以传数据的时候coefficient被转成字符串了，而这边解析出来的字符串，不能直接用byte转换成[]byte，需要先解码成原始的二进制数据。
	//解码的时候，就要使用base64.StdEncoding.DecodeString()函数来解析出原始数据！！！。
	s := extractJsonValue(data.Data.Req.Body, "Coefficients")
	decodedData, _ := base64.StdEncoding.DecodeString(s)
	challenge.Coefficients = decodedData

	challenge.Index = extractJsonArray(data.Data.Req.Body, "Index")

	// logger.Println("PARITY_SHARDS:", parityShards)
	// logger.Println("DATA_SHARDS:", dataShards)
	// logger.Println("coefficients:", string(challenge.Coefficients))
	// logger.Println("coefficients(string):", s)
	// logger.Println("coefficients(true):", decodedData)
	// logger.Println("index:", challenge.Index)

	if dataShards == 0 || parityShards == 0 {
		logger.Print("Invalid shard values")
		http.Error(w, "Invalid shard values", http.StatusBadRequest)
		return
	}

	// 4. 初始化服务
	cloudAPI := Utils.NewCloudAPI()
	auditor := Core.NewIntegrityAuditing(dataShards, parityShards)

	// 5. 下载数据块
	downloadData := make([][]byte, len(challenge.Index))
	downloadParity := make([][]byte, len(challenge.Index))
	for i := range challenge.Index {
		downloadData[i], _ = cloudAPI.DownloadPartFile(
			"source.txt",
			challenge.Index[i]*dataShards,
			dataShards,
		)
		// logger.Println("downloadData[i]:", downloadData[i])
		downloadParity[i], _ = cloudAPI.DownloadPartFile(
			"parities.txt",
			challenge.Index[i]*parityShards,
			parityShards,
		)
		// logger.Println("downloadParity[i]:", downloadParity[i])
	}
	logger.Println("Downloaded data successfully")

	// 6. 生成证明
	proof := auditor.Prove(&challenge, downloadData, downloadParity)
	logger.Println("Generated proof:", proof)
	// 7. 返回响应
	outputs := make(map[string]interface{})
	resData := make(map[string]interface{})
	resData["body"] = proof
	outputs["res"] = resData
	invokeResponse := InvokeResponse{outputs, nil, nil}
	responseJson, _ := json.Marshal(invokeResponse)
	// 使用 json.Marshal 将结构体转换为 JSON 格式并返回

	logger.Println("Content-Type:", w.Header().Get("Content-Type"))
	logger.Println("Generated jsonData:", responseJson)
	w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusOK)
	// 写入响应体
	w.Write(responseJson)
}

// extractJsonValue 从 JSON 字符串中提取指定键的值
func extractJsonValue(jsonStr, key string) string {
	start := `"` + key + `":"`
	startIndex := bytes.Index([]byte(jsonStr), []byte(start))
	if startIndex == -1 {
		return "-1"
	}
	startIndex += len(start)
	endIndex := bytes.IndexByte([]byte(jsonStr[startIndex:]), '"')
	if endIndex == -1 {
		return "-1"
	}
	return jsonStr[startIndex : startIndex+endIndex]
}

func extractJsonArray(jsonStr, key string) []int {
	start := `"` + key + `":[` // 例如 `"index":[`
	startIndex := bytes.Index([]byte(jsonStr), []byte(start))
	if startIndex == -1 {
		return nil
	}
	startIndex += len(start)

	// 提取数组字符串部分
	arrayStr := jsonStr[startIndex : len(jsonStr)-2]

	// 解析字符串成整数数组
	strValues := strings.Split(arrayStr, ",")
	var result []int
	for _, strVal := range strValues {
		num, err := strconv.Atoi(strings.TrimSpace(strVal)) // 去掉空格并转换
		if err != nil {
			continue // 忽略转换失败的值
		}
		result = append(result, num)
	}

	return result
}

func main() {
	customHandlerPort, exists := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT")
	if exists {
		fmt.Println("FUNCTIONS_CUSTOMHANDLER_PORT: " + customHandlerPort)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/Handler", HttpHandler) // Handler is the entry point for the Azure Function，必须同名

	fmt.Println("Go server Listening...on FUNCTIONS_CUSTOMHANDLER_PORT:", customHandlerPort)
	log.Fatal(http.ListenAndServe(":"+customHandlerPort, mux))
}
