package maintain

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
	"user_mgt/user_mgt/jwtutils"
	"user_mgt/utils"

	"golang.org/x/net/websocket"
)

//就是操作redis跟操作Websocket连接，有时是需要同时进行的

// 这个函数的作用不仅仅是添加一个连接，并且需要进行轮询发送心跳的作用，并且现在需要判断这个发送过来的Token是否符合预期，也就是判断现在用户的一个网页状态
func (guestmaintainer *GuestMaintainer) KeepAlive(ws *websocket.Conn) {
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println("logger is nil")
		os.Exit(-1)
		return
	}
	utils.Logger.Info("guest KeepAlive is running")
	//每一个处理http请求，都是独立的一个协程，都可以开启一个Token是否过期的一个信息
	maintainer := NewOnlineUser()
	var (
		//在一个周期之内进行判断
		jwtServer jwtutils.JWTutils
		isExpired = false
	)

	//先能正常吧用户添加到redis中，并且能正常发送信息，检测前后端分别都能收到
	//这个是第一次收前端给后端发送的websocket信息
	tokenstring, err := maintainer.HandleReceiveFirstMessage(ws)
	if err != nil {
		utils.Logger.Errorf("handle receive first message failed: %s", err.Error())
		return
	}
	ok, userID, err := jwtServer.VerifyAndGetIdFromToken(tokenstring, GuestType)
	if err != nil {
		utils.Logger.Errorf("get adn verify token is something wrong：%v", err.Error())
		return
	}
	if ok {
		isExpired = true
		utils.Logger.Errorf("获取成功，但token超时了：%v", err.Error())
		return
	}
	user := User{
		userid:     userID,
		LastActive: time.Now(),
	}
	//首先需要分别添加到Redis跟Websocket里面，这里需要考虑到是否在线的情况，就是假设现在发送请求的是同一个人，也就是说呀是携带token重玩进入了一次这个游戏，但是当他掉线的时候，并没有及时从在线用户列表中删除，是重新建立连接
	err = maintainer.SetUserOnline(userID, ws, &user)
	if err != nil {
		utils.Logger.Errorf("SetUserOnline is failed:%v", err.Error())
		return
	}
	utils.Logger.Infof("user %s is online", userID)
	//添加完之后，应该开启一个轮询，去持续给这个用户发心跳，直到什么时候识别用户下线，就是心跳发送包失败5次之后
	//初始化一个ticker来设置心跳时间
	// 创建一个5秒间隔的Ticker,在协程中获取信号
	ticker := time.NewTicker(heartbreakTime)
	// deadTimer := time.NewTimer(disconnectTimeout)
	//定义一个管道，用于心跳协程与主线程进行响应数据的数据共享
	message := make(chan string)
	//定义一个管道，用于控制主线程退出的信号
	quit := make(chan bool)

	defer func() {
		ticker.Stop()
		// deadTimer.Stop()
		close(message)
		close(quit)
	}()

	//这个连接呢，是用于判断这个连接是否真实有效的改变着用户的活跃时间的，那要是检查出来着个时间
	lastActiveTime := time.Now()
	user.LastActive = lastActiveTime

	go maintainer.heartbreak(ws, userID, message, quit)
	//与其每一个协程都开一个计时器，不如一起开一个协程来管理这个在线用户列表

	//这个循环是会一直执行的，前端发来的Token也会过期的，所以需要一直进行判断,这里是维持这个主线程不结束的代码，通道也需要关闭，也就defer func(){}close所有的管道
	//TODO  修改这个接收心跳处理器的结构
	for {
		select {
		//这里定义好这个定时发送心跳的信号,这里也需要定一个时间，如果已经有新的信息进来，就重新重置超时时间
		case msg := <-message:
			//重置超时计时器，如果超过超时时间都没有收到心跳信息，心跳协程自己会停止，但是这个不会，因此要是这个timer有信息到，就表示要将循环停止
			// if !deadTimer.Stop() {
			// 	<-deadTimer.C
			// }
			// deadTimer.Reset(disconnectTimeout)
			// utils.Logger.Infof("重置超时时间")
			// //从Map里面检查这个连接是否存在
			ok, err := maintainer.CheckConn(userID)
			if err != nil {
				utils.Logger.Errorf("failed to get that new conn from Map")
				return
			}
			if !ok {
				utils.Logger.Infof("conn is close")
			}
			//这里Token 过期了的话，怎么办？就将玩家踢下线，不重新获取一个Token返回给前端么？并且这个是游客
			if isExpired == true {
				utils.Logger.Errorf("Token is expired")
				// 向前端发送"token expired"消息
				ws.SetWriteDeadline(time.Now().Add(receiveTimeout))
				_, err := ws.Write([]byte("token expired"))
				if err != nil {
					utils.Logger.Errorf("send 'token expired' message failed: %s", err.Error())
				}
				//同时释放redis跟websocketMap里面的资源
				err = maintainer.OSHandleSetUserOffline(userID, lastActiveTime)
				if err != nil {
					utils.Logger.Errorf("HandleSetUserOffline is failed")
					return
				}
				return
			}
			//下面这一段代码就是用于等待心跳的反馈信息的，现在不需要了，只要没收到信息，处理协程就会一直等待，直到有任何信息过来或者已经超时了
			// if msg == "waitedRec" || msg == "waitedSend" {
			// 	//这里是将整个协程久阻塞在这里，再次等到心跳的时候才次检测是否要等待，所以在这个等待的过程中，是不会更新用户的在线时间的,所以在这个期间，这个用户我还是会维持在线的状态
			// 	utils.Logger.Errorf("now cannot send or receive msg bewteen client to server")
			// 	continue
			// }
			//这里仅仅用于检验Token是否过期的，不需要返回一个userid
			inValid, err := jwtServer.ParseAndVerifyToken(msg, GuestType)
			if err != nil {
				//这里面出错的，都应该关掉这个连接
				utils.Logger.Errorf("verify token is something wrong：%v", err.Error())
				//同时释放redis跟websocketMap里面的资源
				err := maintainer.OSHandleSetUserOffline(userID, lastActiveTime)
				if err != nil {
					utils.Logger.Errorf("HandleSetUserOffline is failed")
					return
				}
			}
			//表示现在的token已经超时了，就直接返回，如果没有超时，就更新这个
			if inValid {
				isExpired = true
				utils.Logger.Errorf("获取成功，但token超时了：%v", err.Error())
				//同时释放redis跟websocketMap里面的资源
				err := maintainer.OSHandleSetUserOffline(userID, lastActiveTime)
				if err != nil {
					utils.Logger.Errorf("HandleSetUserOffline is failed")
					return
				}
				//这里的超时，有可能是盗取而来的websocket连接，所以只需要踢掉就好
			}
			now := time.Now()
			lastActiveTime = now
			user.LastActive = now
			err = maintainer.UpdateActiveTime(userID, now)
			if err != nil {
				utils.Logger.Errorf("update user %s activeTime failed :%v", userID, err)
				return
			}
			//如果能正常取值，则可以给她刷新Token时间，不是Token时间，而是在线活跃的最后时间
		//收到close的信号，这里分为两种情况：前端主动断；还有就是无法收到或者发送信号到前端
		case <-quit:
			//同时释放redis跟websocketMap里面的资源
			utils.Logger.Info("try set user:%s Offline!", userID)
			err := maintainer.OSHandleSetUserOffline(userID, lastActiveTime)
			if err != nil {
				utils.Logger.Errorf("HandleSetUserOffline is failed")
				return
			}
			return
			//心跳超时,需要从在线用户列表中移除
			// case <-deadTimer.C:
			// 	utils.Logger.Info("os toughly set user:%s Offline!", userID)
			// 	isValid, err := maintainer.checkActive(userID)
			// 	if err != nil {
			// 		utils.Logger.Errorf("Failed to validate active time: %v", err)
			// 		return
			// 	}
			// 	if !isValid {
			// 		err = maintainer.HandleSetUserOffline(userID)
			// 		if err != nil {
			// 			utils.Logger.Errorf("Failed to set user offline: %v", err)
			// 			return
			// 		}
			// 	}
		}
	}

}

func (regmaintainer *RegMaintainer) KeepAlive(ws *websocket.Conn) {
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println("logger is nil")
		os.Exit(-1)
		return
	}
	utils.Logger.Info("Reg KeepAlive is running")
	//每一个处理http请求，都是独立的一个协程，都可以开启一个Token是否过期的一个信息
	maintainer := NewOnlineUser()
	var (
		//在一个周期之内进行判断
		jwtServer jwtutils.JWTutils
		isExpired = false
	)

	//还是要处理第一次信息处理，要提取第一次的Token么？
	//这个连接一直都是游客或提取的Token，之后就是需要
	//先能正常吧用户添加到redis中，并且能正常发送信息，检测前后端分别都能收到
	//这个是第一次收前端给后端发送的websocket信息
	tokenstring, err := maintainer.HandleReceiveFirstMessage(ws)
	if err != nil {
		utils.Logger.Errorf("handle receive first message failed: %s", err.Error())
		return
	}
	ok, userID, err := jwtServer.VerifyAndGetIdFromToken(tokenstring, RegType)
	if err != nil {
		utils.Logger.Errorf("get adn verify token is something wrong：%v", err.Error())
		return
	}
	if ok {
		isExpired = true
		utils.Logger.Errorf("获取成功，但token超时了：%v", err.Error())
		return
	}
	user := User{
		userid:     userID,
		LastActive: time.Now(),
	}
	//首先需要分别添加到Redis跟Websocket里面，这里需要考虑到是否在线的情况，就是假设现在发送请求的是同一个人，也就是说呀是携带token重玩进入了一次这个游戏，但是当他掉线的时候，并没有及时从在线用户列表中删除，是重新建立连接
	//这里是需要进行判断的，需要计算此时用户是否在刷新页面，若是的话，就直接把这个用户放置到在线列表中
	err = maintainer.SetUserOnline(userID, ws, &user)
	if err != nil {
		utils.Logger.Errorf("SetUserOnline is failed:%v", err.Error())
		return
	}
	utils.Logger.Infof("user %s is online", userID)
	//添加完之后，应该开启一个轮询，去持续给这个用户发心跳，直到什么时候识别用户下线，就是心跳发送包失败5次之后
	//初始化一个ticker来设置心跳时间
	// 创建一个5秒间隔的Ticker,在协程中获取信号
	ticker := time.NewTicker(heartbreakTime)
	// deadTimer := time.NewTimer(disconnectTimeout)
	//定义一个管道，用于心跳协程与主线程进行响应数据的数据共享
	message := make(chan string)
	//定义一个管道，用于控制主线程退出的信号
	quit := make(chan bool)

	defer func() {
		ticker.Stop()
		// deadTimer.Stop()
		close(message)
		close(quit)
	}()

	//这个连接呢，是用于判断这个连接是否真实有效的改变着用户的活跃时间的，那要是检查出来着个时间
	lastActiveTime := time.Now()
	user.LastActive = lastActiveTime

	go maintainer.heartbreak(ws, userID, message, quit)
	//与其每一个协程都开一个计时器，不如一起开一个协程来管理这个在线用户列表

	//这个循环是会一直执行的，前端发来的Token也会过期的，所以需要一直进行判断,这里是维持这个主线程不结束的代码，通道也需要关闭，也就defer func(){}close所有的管道
	//TODO  修改这个接收心跳处理器的结构
	for {
		select {
		//这里定义好这个定时发送心跳的信号,这里也需要定一个时间，如果已经有新的信息进来，就重新重置超时时间
		case msg := <-message:
			//重置超时计时器，如果超过超时时间都没有收到心跳信息，心跳协程自己会停止，但是这个不会，因此要是这个timer有信息到，就表示要将循环停止
			// if !deadTimer.Stop() {
			// 	<-deadTimer.C
			// }
			// deadTimer.Reset(disconnectTimeout)
			// utils.Logger.Infof("重置超时时间")
			// //从Map里面检查这个连接是否存在
			ok, err := maintainer.CheckConn(userID)
			if err != nil {
				utils.Logger.Errorf("failed to get that new conn from Map")
				return
			}
			if !ok {
				utils.Logger.Infof("conn is close")
			}
			//这里Token 过期了的话，怎么办？就将玩家踢下线，不重新获取一个Token返回给前端么？并且这个是游客
			if isExpired == true {
				utils.Logger.Errorf("Token is expired")
				// 向前端发送"token expired"消息
				ws.SetWriteDeadline(time.Now().Add(receiveTimeout))
				_, err := ws.Write([]byte("token expired"))
				if err != nil {
					utils.Logger.Errorf("send 'token expired' message failed: %s", err.Error())
				}
				//同时释放redis跟websocketMap里面的资源
				err = maintainer.OSHandleSetUserOffline(userID, lastActiveTime)
				if err != nil {
					utils.Logger.Errorf("HandleSetUserOffline is failed")
					return
				}
				return
			}
			//下面这一段代码就是用于等待心跳的反馈信息的，现在不需要了，只要没收到信息，处理协程就会一直等待，直到有任何信息过来或者已经超时了
			// if msg == "waitedRec" || msg == "waitedSend" {
			// 	//这里是将整个协程久阻塞在这里，再次等到心跳的时候才次检测是否要等待，所以在这个等待的过程中，是不会更新用户的在线时间的,所以在这个期间，这个用户我还是会维持在线的状态
			// 	utils.Logger.Errorf("now cannot send or receive msg bewteen client to server")
			// 	continue
			// }
			//这里仅仅用于检验Token是否过期的，不需要返回一个userid
			inValid, err := jwtServer.ParseAndVerifyToken(msg, RegType)
			if err != nil {
				//这里面出错的，都应该关掉这个连接
				utils.Logger.Errorf("verify token is something wrong：%v", err.Error())
				//同时释放redis跟websocketMap里面的资源
				err := maintainer.OSHandleSetUserOffline(userID, lastActiveTime)
				if err != nil {
					utils.Logger.Errorf("HandleSetUserOffline is failed")
					return
				}
			}
			//表示现在的token已经超时了，就直接返回，如果没有超时，就更新这个
			if inValid {
				isExpired = true
				utils.Logger.Errorf("获取成功，但token超时了：%v", err.Error())
				//同时释放redis跟websocketMap里面的资源
				err := maintainer.OSHandleSetUserOffline(userID, lastActiveTime)
				if err != nil {
					utils.Logger.Errorf("HandleSetUserOffline is failed")
					return
				}
				//这里的超时，有可能是盗取而来的websocket连接，所以只需要踢掉就好
			}
			now := time.Now()
			lastActiveTime = now
			user.LastActive = now
			err = maintainer.UpdateActiveTime(userID, now)
			if err != nil {
				utils.Logger.Errorf("update user %s activeTime failed :%v", userID, err)
				return
			}
			//如果能正常取值，则可以给她刷新Token时间，不是Token时间，而是在线活跃的最后时间
		//收到close的信号，这里分为两种情况：前端主动断；还有就是无法收到或者发送信号到前端
		case <-quit:
			//同时释放redis跟websocketMap里面的资源
			utils.Logger.Info("set user:%s Offline!", userID)
			err := maintainer.OSHandleSetUserOffline(userID, lastActiveTime)
			if err != nil {
				utils.Logger.Errorf("HandleSetUserOffline is failed")
				return
			}
			return
			//心跳超时,需要从在线用户列表中移除
			// case <-deadTimer.C:
			// 	utils.Logger.Info("os toughly set user:%s Offline!", userID)
			// 	isValid, err := maintainer.checkActive(userID)
			// 	if err != nil {
			// 		utils.Logger.Errorf("Failed to validate active time: %v", err)
			// 		return
			// 	}
			// 	if !isValid {
			// 		err = maintainer.HandleSetUserOffline(userID)
			// 		if err != nil {
			// 			utils.Logger.Errorf("Failed to set user offline: %v", err)
			// 			return
			// 		}
			// 	}
		}
	}
}

func (redismaintainer *GuestRedisMaintainer) GetActiveUsers(w http.ResponseWriter, q *http.Request) {
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println("logger is nil")
		os.Exit(-1)
		return
	}
	utils.Logger.Info("GetActiveUsers is running")
	err = AddCoresHeader(w)
	if err != nil {
		utils.Logger.Errorf("ResponseWriter is nil")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	maintain := NewOnlineUser()
	_, activers, err := maintain.GetAllActiveUser(true)
	if err != nil {
		utils.Logger.Errorf("cannot get acitve user from redis:%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonBytes, err := json.Marshal(activers)
	if err != nil {
		utils.Logger.Errorf("cannot marshal data %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// 写入JSON响应体
	w.WriteHeader(254)
	w.Write(jsonBytes)
	utils.Logger.Info("GetActiveUsers successfully")

}

// 用于后端主动断前端的websocket连接，切换用户的websocket连接心跳
func (guestMaintainer *GuestMaintainer) CloseWebSocket(userid string) error {
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println("logger is nil")
		os.Exit(-1)
		return fmt.Errorf("logger is nil")
	}
	utils.Logger.Info("CloseWebSocket is running")
	if userid == "" {
		utils.Logger.Errorf("userid is empty")
		return fmt.Errorf("userid is empty")
	}
	//需要创建一个在线管理器
	maintainer := NewOnlineUser()
	//先关闭websocket的连接，再从redis的在线列表中移除
	err = maintainer.CloseConn(userid)
	if err != nil {
		utils.Logger.Errorf("CloseConn is failed ：%v", err)
		return fmt.Errorf("CloseConn is failed ：%v", err)
	}
	err = maintainer.DelOneConn(userid)
	if err != nil {
		utils.Logger.Errorf("DelOneConn is failed ：%v", err)
		return fmt.Errorf("DelOneConn is failed ：%v", err)
	}
	err = maintainer.DelGuestRedis(userid)
	if err != nil {
		utils.Logger.Errorf("DelGuestRedis is failed ：%v", err)
		return fmt.Errorf("DelGuestRedis is failed ：%v", err)
	}
	isonline, err := maintainer.CheckOnline(userid)
	if err != nil {
		utils.Logger.Errorf("CheckOnline is failed :%v", err)
		return fmt.Errorf("CheckOnline is failed :%v", err)
	}
	if !isonline {
		utils.Logger.Info("successfully set guest :%s offline", userid)
		return nil
	}
	return fmt.Errorf("something wrong in CloseWebSocket")
}
func (regmaintainer *RegMaintainer) CloseWebsocketHTTP(w http.ResponseWriter, r *http.Request) {
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println("logger is nil")
		os.Exit(-1)
		return
	}
	utils.Logger.Info("CloseWebsocketHTTP is running")
	regid, err := getFromContext("regid", r.Context())
	if err != nil {
		utils.Logger.Errorf("Failed to get regId from context:%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var reconn GuestMaintainer
	err = reconn.CloseWebSocket(regid)
	if err != nil {
		utils.Logger.Errorf("Failed to remove user:%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	utils.Logger.Infof("user %s is offline", regid)
	// 返回成功
	w.WriteHeader(http.StatusOK)
}
