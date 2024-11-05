package auth

import (
	"net/http"
	"time"
	"user_mgt/user_mgt/aes"
	"user_mgt/user_mgt/db"
	"user_mgt/user_mgt/jwtutils"
	"user_mgt/user_mgt/maintain"
)

const (
	key       = "guestID"
	GuestType = "guest"
	RegType   = "reg"
	AdminType = "admin"
)

// 定义一个类方法，可以作为类的属性进行调用
type GuestHTTPServer struct {
}

// 首先定义了几个不需要检测Token，就完成的工作,就是做获取Token，然后
type GuestHTTPHandle interface {
	Login(rep http.ResponseWriter, req *http.Request)
	AuthMiddleWare(http.Handler) http.Handler
}

type RegHTTPServer struct {
}
type RegHTTPHandle interface {
	Login(rep http.ResponseWriter, req *http.Request)
}

// 错误信息结构体
type ErrorMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

var (
	//用于操作游客的结构体接口主体
	GuestServer = &GuestHTTPServer{}
	RegServer   = &RegHTTPServer{}
	//用于操作游客数据库接口主体
	GuestDBServer = db.NewGuestDBServer()
	//关于JWT的处理包函数
	JWThandle                = jwtutils.NewJWTserve()
	OnlineMaintainer         = maintain.NewOnlineUser()
	Aes                      = aes.NewAes()
	errorCode        [14]int = [14]int{451, 452, 453, 454, 455, 456, 457, 458, 459, 460, 461, 462, 463, 464} //现确定的错误码
)

// 这个地方是可以进行一次判断的，就是使用sql.NullString的结构，用于更新结构体中特定字段
type User struct {
	UserId      string    `json:"userId"`        //用户id,由系统随机给出
	UserName    string    `json:"userName"`      //玩家昵称
	Password    string    `json:"password"`      //玩家密码
	Email       string    `json:"email"`         //玩家邮箱
	Avatar      string    `json:"avatar"`        //玩家头像，用于发送到前端
	CreateAt    time.Time `json:"createTime"`    //玩家创建时间
	LastLoginAt time.Time `json:"lastLoginTime"` //玩家最后登录时间
	IsBanned    bool      `json:"isBanned"`      //玩家是否被封禁
	BanTime     time.Time `json:"banTime"`       //玩家封禁剩余时间
	IsDeleted   bool      `json:"isDeleted"`     //玩家是否被删除
}

type RequestUser struct {
}

type ResponseUser struct {
}

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
