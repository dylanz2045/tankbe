package maintain

import (
	"context"
	"net/http"
	"sync"
	"time"
	"user_mgt/utils"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/websocket"
)

// 需要的上下文环境，不需要谁给，就是会存放在background里面
var (
	rdb            *redis.Client // Redis客户端
	ctx            context.Context
	WebsocketConns sync.Map //用于在服务器内存中保存websocket变量
	Config         = utils.SetConfig()
)

const (
	activeusers       = "activeusers"
	heartbreakTime    = 3 * time.Second
	refreshTimeSpan   = 5 * time.Second
	disconnectTimeout = 30 * time.Second
	receiveTimeout    = 100 * time.Millisecond
	GuestType         = "guest"
	RegType           = "reg"
	AdminType         = "admin"
)

// 用于保存每一个用户的信息，需要包括当前的标识身份的ID，最后存货的时间
type User struct {
	userid     string
	LastActive time.Time
}

// 这个是处理两个容器的处理接口（也就是工厂函数），里面的接口需要包含处理Redis跟websocket的接口,这个专门就是用于处理用户上下线的处理器，
// 这里的处理器，我是想着直接用online就一并处理了，并不需要分redis还是conn
type OnlineUserMaintainer interface {
	WsMaintainer
	RedisMaintainer
	SetUserOnline(string, *websocket.Conn, *User) error
	HandleTokenExpired(*websocket.Conn) error
	//处理收到第一次信息
	HandleReceiveFirstMessage(*websocket.Conn) (string, error)
	OSHandleSetUserOffline(string, time.Time) error
	heartbreak(*websocket.Conn, string, chan string, chan bool)
	OSCloseWebsocket(string) error
	GetOnlineUserAmount() (int, error)
}

// 一个结构体，嵌入多个接口类型作为
type OnlineUserMaintainerServer struct {
	WsMaintainer
	RedisMaintainer
}

// 还需要一个对外暴露的一个处理函数接口
type GuestMaintainer struct {
}

type RegMaintainer struct {
}

type connManager interface {
	KeepAlive(ws *websocket.Conn)
	CloseWebsocketHTTP(w http.ResponseWriter, r *http.Request)
}

// 这个是用于处理redis的一个全局变量，工厂函数
type GuestRedisMaintainer struct {
}

// 这个是处理ws连接的一个全局变量，工厂函数
type GuestWsMaintainer struct {
}

type RedisMaintainer interface {
	CheckOnline(string) (bool, error)
	AddGuestRedis(User) error
	DelGuestRedis(string) error
	GetAllActiveUser(bool) ([]redis.Z, []string, error)
	PrintOnlineUser() error
	UpdateActiveTime(string, time.Time) error
	ClearRedisUser() error
	CheckActive(string) (bool, error)
	GetActiveUsers(http.ResponseWriter, *http.Request)
	SetOffline(string) error
	getUserInstance(string) (*User, error)
	checkActiveTimeChanged(string, time.Time) (bool, error)
}

type WsMaintainer interface {
	CloseWebsocket(string) error
	AddNewConn(string, *websocket.Conn) error
	DelOneConn(string) error
	CheckConn(string) (bool, error)
	CloseConn(string) error
}
