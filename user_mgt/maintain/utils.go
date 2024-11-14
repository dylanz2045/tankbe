package maintain

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"user_mgt/utils"
)

// float64ToTime 将float64格式的时间转换为time.Time格式
func float64ToTime(timestamp float64) time.Time {
	// 将float64格式的时间转换为int64格式的Unix时间戳
	unixTime := int64(timestamp)
	// 使用time.Unix函数将Unix时间戳转换为time.Time格式
	return time.Unix(unixTime, 0)
}

// calculateTimeSpan 计算时间间隔
func calculateTimeSpan(activeTime time.Time) time.Duration {
	// 检查time是否为空
	if activeTime.IsZero() {
		return 0
	}

	// 计算该时间与当前时间的间隔
	return activeTime.Sub(time.Now())
}

// 给返回的标头设置允许的标头
func AddCoresHeader(w http.ResponseWriter) error {
	if w == nil {
		return fmt.Errorf("ResponseWriter is nil")
	}
	w.Header().Set("Access-Control-Allow-Origin", "*") // 允许所有域名访问，可以根据需要进行限制
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "application/json")

	return nil
}

// setToContext 将key-value存入context
func setToContext(key, value string, ctx context.Context) (context.Context, error) {
	// 检查key是否为空
	if key == "" {
		utils.Logger.Errorf("key is empty")
		return ctx, fmt.Errorf("key is empty")
	}

	// 检查value是否为空
	if value == "" {
		utils.Logger.Errorf("value is empty")
		return ctx, fmt.Errorf("value is empty")
	}

	// 将key-value存入context
	ctx = context.WithValue(ctx, key, value)

	return ctx, nil
}

// getFromContext 从context中获取key对应的value
func getFromContext(key string, ctx context.Context) (string, error) {
	// 检查key是否为空
	if key == "" {
		utils.Logger.Errorf("key is empty")
		return "", fmt.Errorf("key is empty")
	}

	// 从context中获取key对应的value
	value, ok := ctx.Value(key).(string)
	if !ok {
		utils.Logger.Errorf("failed to assert value to string")
		return "", fmt.Errorf("failed to assert value to string")
	}

	return value, nil
}
