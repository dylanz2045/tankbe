package maintain

import (
	"errors"
	"fmt"
	"net/http"
	"user_mgt/user_mgt/jwtutils"
	"user_mgt/utils"
)

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查utils.Logger是否为空
		err := utils.ValidateLogger()
		if err != nil {
			fmt.Println(err)
			return
		}
		// 从Cookie中获取token
		tokenCookie, err := r.Cookie("token")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				utils.Logger.Errorf("No token in cookie")
				http.Error(w, "No token in cookie", http.StatusBadRequest)
			}
			utils.Logger.Errorf("Failed to get token from cookie: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if tokenCookie == nil {
			utils.Logger.Errorf("tokenCookie is nil")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		utils.Logger.Infof("receive tokenCookie: %v", tokenCookie)

		token := tokenCookie.Value

		// 验证token
		jwtService := jwtutils.NewJWTserve()
		isToken, userId, err := jwtService.VerifyAndGetIdFromToken(token, jwtutils.TokenTypeRegistered)
		if err != nil {
			utils.Logger.Errorf("check token failed: %s", err.Error())
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if !isToken {
			utils.Logger.Errorf("message is not a token")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if userId == "" {
			utils.Logger.Errorf("userId is empty")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// 将userId存入HTTP请求context
		ctx, err := setToContext("regId", userId, r.Context())

		// 调用下一个处理程序
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
