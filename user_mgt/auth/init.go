package auth

import (
	"net/http"
)

// 作为一个处理用户的功能，这个是一个游客的，之后还需要添加到转换到用户的活跃表
func Init() {
	http.Handle("/api/GuestLogin", http.HandlerFunc(GuestServer.Login))
	http.Handle("/api/UserLogin", GuestServer.AuthMiddleWare(http.HandlerFunc(RegServer.Login)))
	http.Handle("/api/VerifyAndChangeAvatar", RegServer.AuthMiddelWare(http.HandlerFunc(RegServer.VerifyAndChangeAvatar)))
	http.Handle("/api/getavater", RegServer.AuthMiddelWare(http.HandlerFunc(RegServer.GetAvatar)))

}
