package jwtutils

import (
	"net/http"
)

const (
	defaultIssuer      = "tank user mgt"
	defaultSubject     = "Auth-Token"
	defaultExpireHours = 48
	guestAudience      = "guest"
	regAudience        = "reg"
	adminAudience      = "admin"

	TokenTypeGuest      = "guest"
	TokenTypeRegistered = "reg"
	TokenTypeAdmin      = "admin"
)

type JWTutils struct {
}

// 这个接口是用于外部调用的，专门用于对
// 1、token转string，string转token的工具包
// 2、生成一个正常token令牌的函数
type JWTserver interface {
	//生成一个新的Token令牌
	GenerateToken(string, string) (string, error)
	//从前端的cookie中取出Token
	GetTokenByCookie(cookies []*http.Cookie) (string, error)
	//从获取到的Tokenstring中，验证他是否合法,并返回一个guestID
	VerifyAndGetIdFromToken(string, string) (bool, string, error)
	ParseAndVerifyToken(string, string) (bool, error)
	SetNewCookie(http.ResponseWriter, string, string) error
	IdentifyToken(string) (string, error)
}

func NewJWTserve() JWTserver {
	return &JWTutils{}
}
