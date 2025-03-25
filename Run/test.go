package main

import (
	"serveless_Go/Core"
	// "serveless_Go/Utils"
)

func mai1n() {
	dd := [][]byte{{116, 46, 103}, {101, 116, 67}, {108, 97, 115}}
	dp := [][]byte{{188, 198}, {196, 235}, {44, 21}}
	index := []int{1, 2, 3}
	coefficients := []byte{0x01, 0x02, 0x03}
	challengeData := Core.InitCD(index, coefficients)
	it := Core.NewIntegrityAuditing(3, 2)
	p := it.Prove(challengeData, dd, dp)
	println(p.DataProof)
	println(p.ParityProof)
}

//
////func auditTask1() {
////
////
////}
//
//// package Core
//
//// import (
//// 	"bytes"
//// 	"crypto/aes"
//// 	"crypto/cipher"
//// 	"encoding/binary"
//// 	"errors"
//// 	"fmt"
//// 	"math"
//// 	"strconv"
//// )
//
//// // PseudoRandom 伪随机生成器
//// type PseudoRandom struct{}
//
//// /****************** AES ECB 加密工具集 (私有方法) ******************/
//
//// // ecbEncrypt AES-ECB模式加密
//// func ecbEncrypt(block cipher.Block, plaintext []byte) []byte {
//// 	bs := block.BlockSize()
//// 	plaintext = pkcs5Padding(plaintext, bs)
//// 	ciphertext := make([]byte, len(plaintext))
//// 	for i := 0; i < len(plaintext); i += bs {
//// 		block.Encrypt(ciphertext[i:i+bs], plaintext[i:i+bs])
//// 	}
//// 	return ciphertext
//// }
//
//// // ecbDecrypt AES-ECB模式解密
//// func ecbDecrypt(block cipher.Block, ciphertext []byte) ([]byte, error) {
//// 	bs := block.BlockSize()
//// 	if len(ciphertext)%bs != 0 {
//// 		return nil, errors.New("ciphertext length not multiple of block size")
//// 	}
//
//// 	plaintext := make([]byte, len(ciphertext))
//// 	for i := 0; i < len(ciphertext); i += bs {
//// 		block.Decrypt(plaintext[i:i+bs], ciphertext[i:i+bs])
//// 	}
//// 	return pkcs5Trimming(plaintext)
//// }
//
//// // pkcs5Padding PKCS5填充
//// func pkcs5Padding(src []byte, blockSize int) []byte {
//// 	padding := blockSize - len(src)%blockSize
//// 	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
//// 	return append(src, padtext...)
//// }
//
//// // pkcs5Trimming PKCS5去除填充
//// func pkcs5Trimming(src []byte) ([]byte, error) {
//// 	length := len(src)
//// 	if length == 0 {
//// 		return nil, errors.New("empty input")
//// 	}
//// 	padding := int(src[length-1])
//// 	if padding < 1 || padding > aes.BlockSize {
//// 		return nil, errors.New("invalid padding")
//// 	}
//// 	return src[:length-padding], nil
//// }
//
//// /****************** 公有方法实现 ******************/
//
//// // Encrypt AES加密 (数字转字符串后加密)
//// func (p *PseudoRandom) Encrypt(strKey string, content int) ([]byte, error) {
//// 	// 验证密钥长度
//// 	key := []byte(strKey)
//// 	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
//// 		return nil, fmt.Errorf("invalid AES key length: %d bytes", len(key))
//// 	}
//
//// 	// 初始化加密块
//// 	block, err := aes.NewCipher(key)
//// 	if err != nil {
//// 		return nil, err
//// 	}
//
//// 	// 处理明文（将数字转为字符串）
//// 	plaintext := []byte(strconv.Itoa(content))
//// 	return ecbEncrypt(block, plaintext), nil
//// }
//
//// // Decrypt AES解密(返回数字字符串的字节形式)
//// func (p *PseudoRandom) Decrypt(strKey string, ciphertext []byte) ([]byte, error) {
//// 	key := []byte(strKey)
//// 	block, err := aes.NewCipher(key)
//// 	if err != nil {
//// 		return nil, err
//// 	}
//// 	return ecbDecrypt(block, ciphertext)
//// }
//
//// // BytesToInt 字节转整数（大端序）
//// func (p *PseudoRandom) BytesToInt(bytes []byte) int {
//// 	if len(bytes) < 4 {
//// 		return 0
//// 	}
//// 	return int(binary.BigEndian.Uint32(bytes[:4]))
//// }
//
//// // GenerateRandom 生成随机字节流(兼容原Java逻辑)
//// func (p *PseudoRandom) GenerateRandom(index int, key string, paritySize int) ([]byte, error) {
//// 	// 计算需要生成的数据块数量
//// 	aesCount := int(math.Ceil(float64(paritySize) / 16))
//// 	offsetIndex := aesCount * index // 关键偏移逻辑
//
//// 	result := make([]byte, aesCount*16)
//// 	for i := 0; i < aesCount; i++ {
//// 		encrypted, err := p.Encrypt(key, offsetIndex)
//// 		if err != nil {
//// 			return nil, fmt.Errorf("generate random failed at index %d: %v", offsetIndex, err)
//// 		}
//
//// 		if len(encrypted) != 16 { // AES ECB加密后必须是16字节
//// 			return nil, errors.New("unexpected encrypted data length")
//// 		}
//// 		copy(result[i*16:], encrypted)
//// 		offsetIndex++
//// 	}
//// 	return result[:paritySize], nil // 截断到所需长度
//// }
//
//import (
//	"bytes"
//	"crypto/aes"
//	"errors"
//	"fmt"
//	"math"
//	"strconv"
//)
//
//// PseudoRandom 提供基于AES的伪随机数生成功能
//type PseudoRandom struct{}
//
//// Encrypt 使用AES ECB模式加密内容（注意：ECB模式不安全，建议在实际应用中使用更安全的模式）
//func Encrypt(strKey string, content int) ([]byte, error) {
//	key := []byte(strKey)
//	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
//		return nil, fmt.Errorf("invalid AES key length: %d - must be 16, 24 or 32 bytes", len(key))
//	}
//
//	// 将整数转换为字符串并进行PKCS5填充
//	plaintext := pkcs5Padding([]byte(strconv.Itoa(content)), aes.BlockSize)
//
//	block, err := aes.NewCipher(key)
//	if err != nil {
//		return nil, err
//	}
//
//	ciphertext := make([]byte, len(plaintext))
//	for i := 0; i < len(plaintext); i += aes.BlockSize {
//		block.Encrypt(ciphertext[i:i+aes.BlockSize], plaintext[i:i+aes.BlockSize])
//	}
//
//	return ciphertext, nil
//}
//
//// GenerateRandom 生成伪随机字节序列
//func GenerateRandom(index int, key string, paritySize int) ([]byte, error) {
//	aesCount := int(math.Ceil(float64(paritySize) / float64(aes.BlockSize)))
//	index = aesCount * index
//	result := make([]byte, aesCount*aes.BlockSize)
//
//	for i := 0; i < aesCount; i++ {
//		encrypted, err := Encrypt(key, index)
//		if err != nil {
//			return nil, fmt.Errorf("encryption failed: %v", err)
//		}
//
//		if len(encrypted) != aes.BlockSize {
//			return nil, errors.New("unexpected ciphertext length")
//		}
//
//		start := i * aes.BlockSize
//		end := start + aes.BlockSize
//		if end > len(result) {
//			return nil, errors.New("index out of bounds")
//		}
//		copy(result[start:end], encrypted)
//		index++
//	}
//
//	return result, nil
//}
//
//// pkcs5Padding 实现PKCS#5填充
//func pkcs5Padding(src []byte, blockSize int) []byte {
//	padding := blockSize - len(src)%blockSize
//	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
//	return append(src, padtext...)
//}
//
///* 测试示例*/
//func main() {
//	// 示例使用16字节密钥（AES-128）
//	key := "KQmmnw0FpbD8H826"
//	paritySize := 32
//
//	result, err := GenerateRandom(1, key, paritySize)
//	if err != nil {
//		fmt.Println("Error:", err)
//		return
//	}
//
//	fmt.Printf("Generated %d bytes:\n%x\n", len(result), result)
//}
