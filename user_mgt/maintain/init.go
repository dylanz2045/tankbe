package maintain

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"
	"user_mgt/utils"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/websocket"
)

var (
	//专门处理游客发来的请求函数
	guestMaintainer      GuestMaintainer
	guestRedisMaintainer GuestRedisMaintainer
	regmaintainer        RegMaintainer
	onlineMaintain       OnlineUserMaintainerServer

	//专门处理已经注册的用户发来的请求函数
	//TODO
)

func NewGuestRedis() RedisMaintainer {
	return &GuestRedisMaintainer{}
}

func NewGuestWS() WsMaintainer {
	return &GuestWsMaintainer{}
}

func NewOnlineUser() OnlineUserMaintainer {
	return &OnlineUserMaintainerServer{
		WsMaintainer:    NewGuestWS(),
		RedisMaintainer: NewGuestRedis(),
	}
}

func Init() {
	regmaintainerPtr := &regmaintainer
	http.Handle("/msg/GuestKeepAlive", websocket.Handler(guestMaintainer.KeepAlive))
	http.Handle("/api/GetActiveUsers", http.HandlerFunc(guestRedisMaintainer.GetActiveUsers))
	http.Handle("/msg/RegKeepAlive", websocket.Handler(regmaintainer.KeepAlive))
	http.Handle("/api/CloseRegWebSocket", authMiddleware(http.HandlerFunc(regmaintainerPtr.CloseWebsocketHTTP)))

	// 初始化context
	ctx = context.Background()
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = InitRdeisConn()
	if err != nil {
		utils.Logger.Fatalf("cannot connect redis")
		os.Exit(-1)
		return
	}

	go maintainRedis()

}

// 初始化redis连接
func InitRdeisConn() error {
	err := utils.ValidateLogger()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	utils.Logger.Infof("trying to connect to redis:%s", "0.0.0.0:6910")
	rdb = redis.NewClient(&redis.Options{
		Addr:     "0.0.0.0:6910",
		Password: "QDFAS23SDVasdfQWFAS5EA42A29282628295E_7B7Ddasdfasdflc2Cvsdffffdfdazzxcvnm3Fasdf21235E262A2829_2B7C7D7BPOIUYTREWQ3ALKJHGFDSA3F3E3CMNBVCXZqwertyuiopasdfghjklzxcvbnm2C2F5B5D5C1234567890_3DadfdsdfWDFSwfdFSFA4023235E2628262A2829JHKljkL3ANM3C3EVBNMaEFesvrgwRTHDFGNBasdfWSVSDFSDVXCVXGDFGDFGSFWVF",
		DB:       0,
		PoolSize: 100,
	})
	// 对redis做连接测试
	msg, err := pingToRedis(rdb)
	if err != nil {
		utils.Logger.Errorf("redis connection error: %v", err)
		return err
	}
	utils.Logger.Infof("redis 返回的信息是：%v", msg)
	//这里需要清理之前在redis保存的信息g
	maintainer := NewOnlineUser()
	err = maintainer.PrintOnlineUser()
	if err != nil {
		utils.Logger.Errorf("cannot paintOnlineUser :%v", err)
		return err
	}
	activeUsers, _, err := maintainer.GetAllActiveUser(false)
	if err != nil {
		utils.Logger.Errorf("cannot get Activeuser from redis:%v", err)
		return err
	}
	if activeUsers != nil {
		err = maintainer.ClearRedisUser()
		if err != nil {
			utils.Logger.Errorf("failed to delete all user from redis :%v", err)
			return err
		}
	}
	utils.Logger.Infof("init redis connection success")
	err = maintainer.PrintOnlineUser()
	if err != nil {
		utils.Logger.Errorf("cannot paintOnlineUser :%v", err)
		return err
	}
	return nil
}

// pingToRedis 测试Redis连接
func pingToRedis(rdb *redis.Client) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 对redis做连接测试
	msg, err := rdb.Ping(ctx).Result()
	if err != nil {
		utils.Logger.Errorf("redis connection error: %v", err)
		return "", err
	}

	return msg, nil
}

// 这个是一个协程，用于手动的删除并未删除资源
func maintainRedis() {
	err := utils.ValidateLogger()
	if err != nil {
		fmt.Println(err)
		return
	}
	Maintainer := NewOnlineUser()
	ticker := time.NewTicker(disconnectTimeout)
	defer ticker.Stop()
	//这里是重复在做的，一到时间就去维护，查看最后的谁的时间不到位的
	for {
		<-ticker.C
		// 对redis做连接测试
		_, err := pingToRedis(rdb)
		if err != nil {
			utils.Logger.Errorf("redis connection error: %v", err)
			return
		}
		//获取里面的活跃用户，随后在里面
		ActiveUsers, _, err := Maintainer.GetAllActiveUser(false)
		if err != nil {
			utils.Logger.Errorf("cannot get active user from redis:%v", err)
			return
		}
		if ActiveUsers == nil {
			//应该等待此时的用户处于在线状态
			continue
		}

		for index := range ActiveUsers {
			// 类型断言
			id, ok := ActiveUsers[index].Member.(string)
			if !ok {
				utils.Logger.Errorf("Failed to assert id: %v", ActiveUsers[index].Member)
				return
			}
			isValid, err := Maintainer.CheckActive(id)
			if err != nil {
				utils.Logger.Errorf("Failed to validate active time: %v", err)
				return
			}

			if !isValid {
				err = Maintainer.DelGuestRedis(id)
				if err != nil {
					utils.Logger.Errorf("Failed to set user offline: %v", err)
					return
				}
				utils.Logger.Infof("user %s is offline in redis because of not active", id)
			}
		}
	}
}
