package rand

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"user_mgt/utils"
)

// 生成一个随机的游客ID，用于唯一标识游客
func GenerateGuestID() (string, error) {
	if utils.Logger == nil {
		fmt.Println("utils.Logger is nil")
		return "", fmt.Errorf("utils.Logger is nil")
	}
	utils.Logger.Info("---->user_mgt/internal/rand.GenerateGuestId is run")

	buffer := make([]byte, 9)
	_, err := rand.Read(buffer)
	if err != nil {
		utils.Logger.Errorf("rand.Read failed, err:%v", err)
		return "", fmt.Errorf("rand.Read failed, err:%v", err)
	}

	//转换为base64编码
	id := base64.URLEncoding.EncodeToString(buffer)
	//加上前缀
	id = "guest" + id[5:12]
	return id, nil
}
