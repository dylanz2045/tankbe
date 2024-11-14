package utils

import (
	"fmt"

	"go.uber.org/zap"
)

var (
	Logger    *zap.SugaredLogger
	AllConfig Config
)

func loginit() {
	Logger = zap.L().Sugar()
	// AllConfig = SetConfig()
}
func ValidateLogger() error {
	// 检查utils.Logger是否为空
	if Logger == nil {
		fmt.Println("utils.Logger is nil")
		return fmt.Errorf("utils.Logger is nil")
	}
	return nil
}
