package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"user_mgt/utils"
)

func setToContext(ctx context.Context, key string, value string) (context.Context, error) {
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println("Logger is nil")
		os.Exit(-1)
		return nil, fmt.Errorf("Logger is nil")
	}
	if key == "" {
		utils.Logger.Errorf("key is nil")
		return nil, fmt.Errorf("key is nil")
	}
	if value == "" {
		utils.Logger.Errorf("value is nil")
		return nil, fmt.Errorf("value is nil")
	}
	ctx = context.WithValue(ctx, key, value)
	utils.Logger.Info("successfully set Value to ctx")
	return ctx, nil
}

func getFromContest(ctx context.Context, key string) (string, error) {
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println("Logger is nil")
		os.Exit(-1)
		return "", fmt.Errorf("Logger is nil")
	}
	if key == "" {
		utils.Logger.Errorf("key is nil")
		return "", fmt.Errorf("key is nil")
	}
	// 从context中获取key对应的value
	value, ok := ctx.Value(key).(string)
	if !ok {
		utils.Logger.Errorf("failed to assert value to string")
		return "", fmt.Errorf("failed to assert value to string")
	}
	return value, nil
}

// sendErrorMessageToFe发送错误消息到前端的函数
func sendErrorMessageToFe(w http.ResponseWriter, code int, message string) error {
	// 检查utils.Logger是否为空
	if utils.Logger == nil {
		fmt.Println("utils.Logger is nil")
		return fmt.Errorf("utils.Logger is nil")
	}
	utils.Logger.Info("---->user_mgt/internal/auth.sendErrorMessageToFe is run")

	//检查传入的code是否在错误码列表中
	isInErrorCode := false
	for i := 0; i < len(errorCode); i++ {
		if code == errorCode[i] {
			isInErrorCode = true
			break
		}
	}
	if !isInErrorCode {
		utils.Logger.Errorf("传入的code不在错误码列表中")
		return fmt.Errorf("传入的code不在错误码列表中")
	}
	if message == "" {
		utils.Logger.Errorf("传入的message为空")
		return fmt.Errorf("传入的message为空")
	}

	//发送错误信息
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	errorResponse := ErrorMessage{
		Code:    code,
		Message: message,
	}
	err := json.NewEncoder(w).Encode(errorResponse)
	if err != nil {
		utils.Logger.Errorf("json.NewEncoder(w).Encode failed, err:%v", err)
		return fmt.Errorf("json.NewEncoder(w).Encode failed, err:%v", err)
	}

	return nil
}

// IsAvatarPathExists 检测头像路径是否存在，不存在设置为默认路径
func IsAvatarPathExists(path *string) (bool, error) {
	//检测传入的path是否为空
	if path == nil {
		utils.Logger.Errorf("传入头像路径为空")
		return false, fmt.Errorf("传入头像路径为空")
	}
	//检测路径下是否有图片
	_, err := os.Stat(*path)
	if os.IsNotExist(err) {
		//路径不存在,设置为默认头像路径
		utils.Logger.Infof("头像路径:%v", *path)
		return false, nil
	}
	return true, nil
}
