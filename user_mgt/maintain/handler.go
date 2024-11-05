package maintain

import (
	"fmt"
	"os"
	"time"
	"user_mgt/utils"

	"golang.org/x/net/websocket"
)

// 现在这个是处理器，用于处理来自心跳的信号
// func (server *OnlineUserMaintainerServer) handleHeartbeat(ws *websocket.Conn, message chan string, quit chan bool) {
// 	err := utils.CheckLogger()
// 	if err != nil {
// 		fmt.Println("Logger is nil")
// 		os.Exit(-1)
// 		return
// 	}
// 	utils.Logger.Info("HandleHeartbeat is running")

//		//用于发送心跳携带的信息管道
//		if message == nil {
//			utils.Logger.Errorf("channel message  is null or empty")
//			return
//		}
//		//用于控制链接是否继续的信号管道
//		if quit == nil {
//			utils.Logger.Errorf("channel quit is null or empty")
//			return
//		}
//		maintainer := NewOnlineUser()
//		var (
//			//在一个周期之内进行判断
//			jwtServer jwtutils.JWTutils
//			isExpired = false
//		)
//		tokenstring, err := maintainer.HandleReceiveFirstMessage(ws)
//		if err != nil {
//			utils.Logger.Errorf("handle receive first message failed: %s", err.Error())
//			return
//		}
//		ok, userID, err := jwtServer.VerifyAndGetIdFromToken(tokenstring)
//		if err != nil {
//			utils.Logger.Errorf("get adn verify token is something wrong：%v", err.Error())
//			return
//		}
//		if ok {
//			isExpired = true
//			utils.Logger.Errorf("获取成功，但token超时了：%v", err.Error())
//			return
//		}
//		//获取最后的活跃的时间
//		user := User{
//			userid:     userID,
//			LastActive: time.Now(),
//		}
//		err = maintainer.SetUserOnline(userID, ws, &user)
//		if err != nil {
//			utils.Logger.Errorf("SetUserOnline is failed:%v", err.Error())
//			return
//		}
//		utils.Logger.Infof("user %s is online", userID)
//	}
func (server *OnlineUserMaintainerServer) OSCloseWebsocket(userid string) error {
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println("Logger is nil")
		os.Exit(-1)
		return fmt.Errorf("Logger is nil")
	}
	if userid == "" {
		utils.Logger.Errorf("id is null or empty")
		return fmt.Errorf("userid is null or empty")
	}
	//将游客从在线用户的
	err = server.DelGuestRedis(userid)
	if err != nil {
		utils.Logger.Errorf("delGuestRedis is failed :%v", err)
		return fmt.Errorf("delGuestRedis is failed :%v", err)
	}
	//发送这个删除的信息给到前端，并且真实的close这个websocket连接
	err = server.CloseConn(userid)
	if err != nil {
		utils.Logger.Errorf("closeConn is failed :%v", err)
		return fmt.Errorf("closeConn is failed :%v", err)
	}
	err = server.DelOneConn(userid)
	if err != nil {
		utils.Logger.Errorf("delOneConn is failed :%v", err)
		return fmt.Errorf("delOneConn is failed :%v", err)
	}
	err = server.PrintOnlineUser()
	if err != nil {
		utils.Logger.Errorf("Failed to print online users:%v", err)
		return err
	}
	return nil

}

// 心跳协程，可以再定义一个关闭信号通道，用于避免协程占用cpu资源，进入“忙等”状态
func (server *OnlineUserMaintainerServer) heartbreak(ws *websocket.Conn, userid string, messages chan string, quit chan bool) {
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println("Logger is nil")
		os.Exit(-1)
		return
	}
	if userid == "" {
		utils.Logger.Errorf("id is null or empty")
		return
	}
	//用于发送心跳携带的信息管道
	if messages == nil {
		utils.Logger.Errorf("channel message  is null or empty")
		return
	}
	//用于控制链接是否继续的信号管道
	if quit == nil {
		utils.Logger.Errorf("channel quit is null or empty")
		return
	}
	//定义变量，用于接受来自前端发过来的Token心跳信息
	var tokenstring string
	maintainer := NewOnlineUser()
	failedRecMsg := 3 //3次失败收到信息的次数
	failedSedMsg := 3 //3次失败发送到信息的次数

	// 创建一个5秒间隔的Ticker
	ticker := time.NewTicker(heartbreakTime)
	defer ticker.Stop()
	//从连接池中获取websocket的连接
	conn, ok := WebsocketConns.Load(userid)
	if !ok {
		utils.Logger.Errorf("cannot get conn from Map")
		quit <- true
		return
	}
	_, isget := conn.(*websocket.Conn)
	if !isget {
		utils.Logger.Errorf("failed to convert conn to websocket.Conn")
		quit <- true
		return
	}

	if conn == nil {
		utils.Logger.Errorf("conn is nil")
		quit <- true
		return
	}
	for {
		select {
		case <-ticker.C:
			//从Map里面检查这个连接是否存在
			ok, err := maintainer.CheckConn(userid)
			if err != nil {
				utils.Logger.Errorf("failed to get that new conn from Map")
				return
			}
			if !ok {
				utils.Logger.Infof("conn has been closed")
				return
			}
			//若此时websocket能正常运作，就开始发送心跳包轮询，并且承受失败次数最大5次
			ws.SetWriteDeadline(time.Now().Add(receiveTimeout))
			err = websocket.Message.Send(ws, "heartbeat")
			if err != nil {
				utils.Logger.Errorf("user %s send heartbreak failed", userid)
				//表示还在接收的范围之内
				if failedSedMsg < 0 {
					utils.Logger.Errorf("user %s failed send message more than 5 times :%v", userid, err.Error())
					quit <- true
					return
				}
				failedSedMsg--
				utils.Logger.Errorf("failedSenMsg failed %d times", 5-failedSedMsg)
				continue
			}
			failedSedMsg = 3
			ws.SetReadDeadline(time.Now().Add(receiveTimeout))
			//就是阻塞，为什么他的time.Sleep就不会阻塞
			err = websocket.Message.Receive(ws, &tokenstring)
			if err != nil {
				utils.Logger.Errorf("cannot receive info from user")
				if failedRecMsg < 0 {
					utils.Logger.Errorf("user %s failed receive message more than 5 times :%v", userid, err.Error())
					quit <- true
					return
				}
				failedRecMsg--
				continue
			}
			utils.Logger.Infof("此时前端发送的信息为：%s", tokenstring)
			if tokenstring == "close" {
				quit <- true
				return
			}
			//如果能成功运行
			failedRecMsg = 3
			//这里是最后将信息转发给主处理线程的,若前面的能正常发，正常收，并且不是收到close信号的话
			messages <- tokenstring
		}
	}
}

func (server *OnlineUserMaintainerServer) HandleReceiveFirstMessage(ws *websocket.Conn) (string, error) {
	//获取第一次的信息，肯定就是需要获取信息，就是获取从cookie中解析而来的cookie
	var message string
	ws.SetReadDeadline(time.Now().Add(receiveTimeout))
	err := websocket.Message.Receive(ws, &message)
	if err != nil {
		utils.Logger.Errorf("failed read message : %v", err.Error())
		return "", nil
	}
	if len(message) == 0 {
		utils.Logger.Errorf("message is empty")
		return "", fmt.Errorf("message is empty")
	}
	utils.Logger.Infof("receive message : %s", message)
	ws.SetWriteDeadline(time.Now().Add(receiveTimeout))
	err = websocket.Message.Send(ws, message)
	if err != nil {
		utils.Logger.Errorf("write message failed: %s", err.Error())
	}
	return string(message), nil
}

// SetUserOnline 这里是已经成功的能够操作redis跟存放的connMap的情况下的
func (server *OnlineUserMaintainerServer) SetUserOnline(id string, conn *websocket.Conn, user *User) error {
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println("Logger is nil")
		os.Exit(-1)
		return err
	}
	if id == "" {
		utils.Logger.Errorf("id is null or empty")
		return fmt.Errorf("id is null or empty")
	}
	if conn == nil {
		utils.Logger.Errorf("conn is null or empty")
		return fmt.Errorf("conn is null or empty")
	}
	//多加一个判断：就是判断用户是否在线
	isOnline, err := server.CheckOnline(id)
	if err != nil {
		utils.Logger.Errorf("CheckOnline is failed：%v", err)
		return fmt.Errorf("CheckOnline is failed：%v", err)
	}
	utils.Logger.Info("执行判断是否在线")
	//这里的在线，并不是同用户登录一般的在线，而是在判断redis列表中的在线
	//如果是出于在线状态的话，还是需要添加这个
	if isOnline {
		onlineuser, err := server.getUserInstance(id)
		if err != nil {
			utils.Logger.Errorf("failed to get user %s from users map: %s", id, err)
			return err
		}
		if onlineuser == nil {
			utils.Logger.Errorf("user %s is nil", id)
			return fmt.Errorf("user %s is nil", id)
		}
		timespan := calculateTimeSpan(user.LastActive)
		// timeSpan取绝对值
		if timespan < 0 {
			timespan = -timespan
		}
		if timespan < refreshTimeSpan {
			//接下来是操作redis
			err = server.AddGuestRedis(*onlineuser)
			if err != nil {
				utils.Logger.Errorf("addGuestRedis is failed")
				return fmt.Errorf("addGuestRedis is failed:%v", nil)
			}
			//操作websocket
			err = server.AddNewConn(id, conn)
			if err != nil {
				utils.Logger.Errorf("addNewConn is failed!:%v", err.Error())
				return fmt.Errorf("addNewConn is failed!!%v", nil)
			}
			return nil
		}

	}
	utils.Logger.Info("该用户不在线")
	//接下来是操作redis
	err = server.AddGuestRedis(*user)
	if err != nil {
		utils.Logger.Errorf("addGuestRedis is failed")
		return fmt.Errorf("addGuestRedis is failed:%v", nil)
	}
	//操作websocket
	err = server.AddNewConn(id, conn)
	if err != nil {
		utils.Logger.Errorf("addNewConn is failed!:%v", err.Error())
		return fmt.Errorf("addNewConn is failed!!%v", nil)
	}
	//接下来是列出现在所有在线人数，看看是否真的存在与redis中
	err = server.PrintOnlineUser()
	if err != nil {
		utils.Logger.Errorf("cannot print online user :%v", err)
	}
	return nil
}

// TODO
func (server *OnlineUserMaintainerServer) HandleTokenExpired(ws *websocket.Conn) error {
	return nil
}

// 这个踢下线是后端直接不通知前端，直接关掉的操作，在切换websocket通道的时候，可能会问题，会报错
// HandleSetUserOffline 这里是将用户踢下线，websocket自行调用的函数，这里也需要判断这个用户正确在线时间
// 这里进行判断是否将用户真正的删除，需要判断是否真的要将这个用户从在线队列中删掉，如果不是的话，就保留这个ID在在线用户列表那里，之后的连接也不需要删，那个心跳的协程也会删掉的，会进行return的。这个逻辑运用的可以，还是可以的
func (server *OnlineUserMaintainerServer) OSHandleSetUserOffline(userid string, lastActiveTime time.Time) error {
	//这里应该将用户踢下线，踢下线这里他有个前提，就是去获取一下游客的最近活跃时间，去检测一下这个最近活跃时间，是否真实的
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println("Logger is nil")
		os.Exit(-1)
		return err
	}
	if userid == "" {
		utils.Logger.Errorf("userid is nil")
		return fmt.Errorf("userid is nil")
	}
	//必须先判断此时用户最后的一次活跃时间，是否正在在线的状态，否则不能将该用户的连接与在线用户列表中移除
	//需要检验这个在线状态，如果最后一次的活跃时间没有超过超时时间，是不能将他踢掉的
	//先将在线列表中移除user
	istimechanged, err := server.checkActiveTimeChanged(userid, lastActiveTime)
	if err != nil {
		utils.Logger.Errorf("checkActiveTimeChanged is failed ：%v", err)
		return fmt.Errorf("checkActiveTimeChanged is failed ：%v", err)
	}
	if !istimechanged {
		err = server.DelGuestRedis(userid)
		if err != nil {
			utils.Logger.Errorf("cannot remove active user from redis:%v", err.Error())
			return fmt.Errorf("cannot remove active user from redis:%v", err.Error())
		}
		utils.Logger.Info("successful remove user from redis")
		//随后关闭websocket的连接
		err = server.CloseConn(userid)
		if err != nil {
			utils.Logger.Errorf("cannot release that conn")
			return fmt.Errorf("cannot release that conn")
		}
		//将连接从Map中移除
		err = server.DelOneConn(userid)
		if err != nil {
			utils.Logger.Errorf("cannot remove Conn from Websocket Map")
			return fmt.Errorf("cannot remove Conn from Websocket Map")
		}
		utils.Logger.Info("successful remove Conn from Map")
		err = server.PrintOnlineUser()
		if err != nil {
			utils.Logger.Errorf("cannot print online user :%v", err)
		}
	} else {
		utils.Logger.Errorf("user %s is still active, can not be remove", userid)
		return fmt.Errorf("user %s is still active, can not be remove", userid)
	}
	return nil
}
