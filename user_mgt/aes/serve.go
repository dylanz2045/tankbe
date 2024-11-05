package aes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"os"
	"user_mgt/utils"
)

// 对密码进行加密
func (*UserInfo) Encrypt(data string) (string, error) {
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println("Logger is nil")
		os.Exit(-1)
		return "", err
	}
	utils.Logger.Info("---->user_mgt/aes.Encrypt is run")
	//先取出这个string，之后转成字节数组
	byteKey := []byte(utils.AllConfig.Aeskey)
	byteIv := []byte(utils.AllConfig.Aesiv)

	plaintext := []byte(data)
	padding := aes.BlockSize - len(data)%aes.BlockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	plaintext = append(plaintext, padText...)

	block, err := aes.NewCipher(byteKey)
	if err != nil {
		utils.Logger.Errorf("failed to create new cipher block: %v", err.Error())
		return "", fmt.Errorf("failed to create new cipher block: %v", err.Error())
	}
	// 检查iv长度是否为aes.BlockSize
	if len(byteIv) != aes.BlockSize {
		utils.Logger.Errorf("iv length is not equal to aes.BlockSize, iv length: %d", len(byteIv))
		return "", fmt.Errorf("iv length is not equal to aes.BlockSize, iv length: %d", len(byteIv))
	}
	// 创建CBC模式加密器
	mode := cipher.NewCBCEncrypter(block, byteIv)

	// 加密
	mode.CryptBlocks(plaintext, plaintext)

	// 将加密后的数据从byte数组转换为Hex编码字符串
	encryptedStr := hex.EncodeToString(plaintext)

	// 检查加密后的数据是否为空
	if encryptedStr == "" {
		utils.Logger.Errorf("encrypt failed")
		return "", fmt.Errorf("encrypt failed")
	}
	return encryptedStr, nil

}

func (*UserInfo) Decrypt(data string) (string, error) {
	return "", nil
}
