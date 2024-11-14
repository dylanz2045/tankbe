package jwtutils

import (
	"fmt"
	"net/http"
	"os"
	"time"
	"user_mgt/utils"

	"github.com/dgrijalva/jwt-go"
)

// 这里是用于生成不同的Token类型的   tokentype
func (*JWTutils) GenerateToken(userID string, tokentype string) (string, error) {
	err := utils.CheckLogger()
	if err != nil {
		return "", err
	}
	if userID == "" {
		utils.Logger.Errorf("传入的参数有误")
		return "", fmt.Errorf("传入的参数有误")
	}
	Token := jwt.New(jwt.SigningMethodHS256)
	//创建一个Map用于存储claims

	claims := Token.Claims.(jwt.MapClaims)
	claims["iss"] = defaultIssuer                                         // 签发者
	claims["sub"] = defaultSubject                                        // 主题
	claims["aud"] = tokentype                                             //受众
	claims["iat"] = time.Now().Unix()                                     // 签发时间
	claims["exp"] = time.Now().Add(time.Hour * defaultExpireHours).Unix() // 过期时间
	claims["id"] = userID

	// 签名并获取完整的编码后的字符串token
	tokenString, err := Token.SignedString([]byte("holryLBFQUIVXYfqrg-HC4Bde3cYiZQoTFeoW_3J-ug"))
	if err != nil {
		utils.Logger.Errorf("生成Token失败：%v", err)
		return "", err
	}
	return tokenString, nil
}

// TODO
func (*JWTutils) GetTokenByCookie(cookies []*http.Cookie) (string, error) {
	err := utils.CheckLogger()
	if err != nil {
		return "", fmt.Errorf("cannot find Logger")
	}
	//这个是没有cookie的字段
	if len(cookies) == 0 {
		return "", nil
	}

	for _, cookie := range cookies {
		if cookie.Name == "token" {
			return cookie.Value, nil
		}
	}
	return "", nil
}

// 根据不同受众从Token中获取存放的ID
func (*JWTutils) VerifyAndGetIdFromToken(tokenstring string, tokentype string) (bool, string, error) {
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println("日志实体消失了，请检查！")
		os.Exit(-1)
		return false, "", err
	}
	if tokenstring == "" {
		utils.Logger.Errorf("参数传入的形式不正确，请传入正确的参数！")
		return false, "", fmt.Errorf("参数传入的形式不正确")
	}
	Token, err := parseJWT(tokenstring)
	if err != nil {
		utils.Logger.Errorf("解析Token的令牌失败！：%v", err)
		return false, "", fmt.Errorf("解析Token的令牌失败！：%v", err.Error())
	}
	ok, err := VerifyToken(Token, tokentype)
	if err != nil {
		utils.Logger.Errorf("解析Token的令牌失效！：%v", err)
		return false, "", fmt.Errorf("解析Token的令牌失效！：%v", err.Error())
	}
	if ok {
		utils.Logger.Errorf("解析Token的令牌过时了！：%v", err)
		return true, "", fmt.Errorf("解析Token的令牌过时了！：%v", err.Error())
	}
	//下面是从Token里面取出前面分配的Userid
	userid, err := GetUserIDFromToken(Token)
	if err != nil {
		utils.Logger.Errorf("cannot GetUserIDFromToken :%v", err.Error())
		return false, "", fmt.Errorf("cannot get userid from token:%v", err.Error())
	}
	if userid == "" {
		utils.Logger.Errorf("userid is empty")
		return false, "", fmt.Errorf("userid is empty")
	}
	return false, userid, nil
}

// 解析并检验用户发来的Token
func (*JWTutils) ParseAndVerifyToken(tokenstring string, tokentype string) (bool, error) {
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println("日志实体消失了，请检查！")
		os.Exit(-1)
		return false, err
	}
	if tokenstring == "" {
		utils.Logger.Errorf("参数传入的形式不正确，请传入正确的参数！")
		return false, fmt.Errorf("参数传入的形式不正确")
	}
	Token, err := parseJWT(tokenstring)
	if err != nil {
		utils.Logger.Errorf("解析Token的令牌失败！：%v", err)
		return false, fmt.Errorf("解析Token的令牌失败！：%v", err.Error())
	}
	isValid, err := VerifyToken(Token, tokentype)
	if err != nil {
		utils.Logger.Errorf("解析Token的令牌失效！：%v", err)
		return false, fmt.Errorf("解析Token的令牌失效！：%v", err.Error())
	}
	if isValid {
		utils.Logger.Errorf("解析Token的令牌过时了！：%v", err)
		return true, fmt.Errorf("解析Token的令牌过时了！：%v", err.Error())
	}
	return isValid, nil
}

// 可以根据不同受众类型进行分配不同的cookie
func (jwtUser *JWTutils) SetNewCookie(w http.ResponseWriter, id string, tokentype string) error {
	// 检查utils.Logger是否为空
	if utils.Logger == nil {
		fmt.Println("utils.Logger is nil")
		return fmt.Errorf("utils.Logger is nil")
	}
	utils.Logger.Info("---->user_mgt/internal/auth.setUserCookie is run")

	//检测传入id是否为空
	if id == "" {
		utils.Logger.Error("id is empty")
		return fmt.Errorf("id is empty")
	}

	//转换为token格式
	token, err := jwtUser.GenerateToken(id, tokentype)
	if err != nil {
		utils.Logger.Errorf("GenerateToken failed, err:%v", err)
		return fmt.Errorf("GenerateToken failed, err:%v", err)
	}

	//设置用户cookie
	cookie := &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(2 * 24 * time.Hour),
		MaxAge:   172800, //设置2天后这个cookie就会失效
		Secure:   false,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
	utils.Logger.Info("成功设置cookie")
	return nil
}

func (jwtUser *JWTutils) IdentifyToken(tokenstring string) (string, error) {
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println("日志实体消失了，请检查！")
		os.Exit(-1)
		return "", err
	}
	// 检查token是否为空
	if tokenstring == "" {
		utils.Logger.Errorf("token is empty")
		return "", fmt.Errorf("token is empty")
	}
	token, err := parseJWT(tokenstring)
	if err != nil {
		utils.Logger.Errorf("parseJWT is failed :%v", err)
		return "", fmt.Errorf("parseJWT is failed :%v", err)
	}
	// 检查token是否有效
	if !token.Valid {
		utils.Logger.Errorf("token is invalid")
		return "", fmt.Errorf("token is invalid")
	}
	// 提取claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		utils.Logger.Errorf("failed to extract claims")
		return "", fmt.Errorf("failed to extract claims")
	}
	// 检查claims是否为空
	if claims == nil {
		utils.Logger.Errorf("claims is empty")
		return "", fmt.Errorf("claims is empty")
	}
	// 从claims中获取受众
	aud, ok := claims["aud"].(string)
	if !ok {
		utils.Logger.Errorf("failed to get audience")
		return "", fmt.Errorf("failed to get audience")
	}
	if aud == "" {
		utils.Logger.Errorf("audience is empty")
		return "", fmt.Errorf("audience is empty")
	}
	switch aud {
	case guestAudience:
		return TokenTypeGuest, nil
	case regAudience:
		return TokenTypeRegistered, nil
	case adminAudience:
		return TokenTypeAdmin, nil
	default:
		return "", fmt.Errorf("invalid audience")
	}
}
