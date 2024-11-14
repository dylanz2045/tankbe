package jwtutils

import (
	"fmt"
	"os"
	"time"
	"user_mgt/utils"

	"github.com/dgrijalva/jwt-go"
)

// 从前端发来的Token中，解析成出真正的Token
func parseJWT(tokenstring string) (*jwt.Token, error) {
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println("日志实体消失了，请检查！")
		os.Exit(-1)
		return nil, err
	}
	if tokenstring == "" {
		utils.Logger.Errorf("tokenstring 是空，请检查参数输入")
		return nil, fmt.Errorf("tokenString is empty")
	}
	token, err := jwt.Parse(tokenstring, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			utils.Logger.Errorf("Token的签名不合法！:%v", token.Header["alg"])
			return nil, fmt.Errorf("不合法的Token签名：%v", token.Header["alg"])
		}
		return []byte("holryLBFQUIVXYfqrg-HC4Bde3cYiZQoTFeoW_3J-ug"), nil
	})
	//若此时的Token是过期的话，上面的连接应该发不过来
	if err != nil {
		utils.Logger.Errorf("解析Token时出错了！:%v", err.Error())
		if err.Error() == "Token is expired" {
			return nil, fmt.Errorf("Token is expired")
		}
		return nil, fmt.Errorf("解析Token时出错了！:%v", err.Error())
	}
	if token == nil {
		utils.Logger.Error("token is empty")
		return nil, fmt.Errorf("token is empty")
	}
	if !token.Valid {
		utils.Logger.Error("token is inValid")
		return nil, fmt.Errorf("token is inValid")
	}
	return token, nil

}

// 检验发送过来的Token,若首个返回值是真，则表示Token过期
func VerifyToken(token *jwt.Token, tokentype string) (bool, error) {
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println("日志实体消失了，请检查！")
		os.Exit(-1)
		return false, err
	}
	if token == nil {
		utils.Logger.Errorf("token is empty")
		return false, fmt.Errorf("token is empty")
	}
	if !token.Valid {
		utils.Logger.Errorf("token is inValid")
		return false, fmt.Errorf("token is inValid")
	}
	//提取这里面的声明，并且一一对应
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		utils.Logger.Errorf("failed to extract claims")
		return false, fmt.Errorf("failed to extract claims")
	}
	if claims == nil {
		utils.Logger.Errorf(" claims is nil")
		return false, fmt.Errorf("claims is nil")
	}
	//检验前面的所有的自己的声明是否符合预期
	if ok = claims.VerifyAudience(tokentype, true); !ok {
		utils.Logger.Errorf("invalid andience")
		return false, fmt.Errorf("invalid andience")
	}
	if ok = claims.VerifyIssuer(defaultIssuer, true); !ok {
		utils.Logger.Errorf("invalid issuer")
		return false, fmt.Errorf("invalid issuer")
	}
	if ok = claims.VerifyExpiresAt(time.Now().Unix(), true); !ok {
		utils.Logger.Errorf("Token is expired")
		return true, fmt.Errorf("Token is expired")
	}
	return false, nil
}

func GetUserIDFromToken(token *jwt.Token) (string, error) {
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println("日志实体消失了，请检查！")
		os.Exit(-1)
		return "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		utils.Logger.Errorf("failed to extract claims")
		return "", fmt.Errorf("failed to extract claims")
	}
	if claims == nil {
		utils.Logger.Errorf(" claims is nil")
		return "", fmt.Errorf("claims is nil")
	}
	userid, ok := claims["id"].(string)
	if !ok || userid == "" {
		utils.Logger.Errorf("解析userid失败")
		return "", fmt.Errorf("解析userid失败")
	}
	return userid, nil
}
