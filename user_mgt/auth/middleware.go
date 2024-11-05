package auth

import (
	"fmt"
	"net/http"
)

// 给返回的标头设置允许的标头
func AddCoresHeader(w http.ResponseWriter) error {
	if w == nil {
		return fmt.Errorf("ResponseWriter is nil")
	}
	w.Header().Set("Access-Control-Allow-Origin", "*") // 允许所有域名访问，可以根据需要进行限制
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	return nil
}




