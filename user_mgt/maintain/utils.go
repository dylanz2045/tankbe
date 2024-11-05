package maintain

import (
	"fmt"
	"net/http"
	"time"
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
