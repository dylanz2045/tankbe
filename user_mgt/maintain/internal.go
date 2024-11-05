package maintain

import (
	"errors"
	"fmt"
	"os"
	"time"
	"user_mgt/utils"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/websocket"
)

// 添加一条在线用户记录在redis中，这个仅仅是把用户添加到redis，并且保存上活跃时间的
func (redismaintainer *GuestRedisMaintainer) AddGuestRedis(user User) error {
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println("Logger is nil")
		os.Exit(-1)
		return err
	}
	userid := user.userid
	userActiveTime := user.LastActive
	if userid == "" || userActiveTime.IsZero() {
		utils.Logger.Errorf("userid is invalid or activetime is invalid")
		return fmt.Errorf("parameter is invalid")
	}
	if ctx == nil {
		utils.Logger.Errorf("ctx is nil")
		return fmt.Errorf("ctx is nil")
	}
	// 对redis做连接测试，以检验这个redis是正常运行
	_, err = pingToRedis(rdb)
	if err != nil {
		utils.Logger.Fatalf("redis connection error: %v", err)
		os.Exit(-1)
		return fmt.Errorf("cannot get connection from redis")
	}
	//对时间转化成时间戳，并且再转成64位浮点型
	ActiveTime := float64(userActiveTime.Unix())

	//创建记录在redis
	err = rdb.ZAdd(ctx, activeusers, redis.Z{
		Member: userid,
		Score:  ActiveTime,
	}).Err()
	if err != nil {
		utils.Logger.Errorf("cannot create info into redis:%v", err)
		return fmt.Errorf("cannot create info into redis")
	}
	return nil

}

// 这个是websocket端亲自操作子的线程，自己的这个用户是否断开连接自己是知道的，所以自己可以维护
func (redismaintainer *GuestRedisMaintainer) DelGuestRedis(userid string) error {
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println("Logger is nil")
		os.Exit(-1)
		return fmt.Errorf("logger is nil")
	}
	if userid == "" {
		utils.Logger.Errorf("userid is nil")
		return fmt.Errorf("userid is nil")
	}
	// 检查ctx是否为空
	if ctx == nil {
		utils.Logger.Errorf("ctx is nil")
		return fmt.Errorf("ctx is nil")
	}
	_, err = pingToRedis(rdb)
	if err != nil {
		utils.Logger.Errorf("redis connection error: %s", err)
		return err
	}
	// 从redis中删除用户,这个语句，即使没有这个用户的信息，也可以执行，不会返回错误的
	err = rdb.ZRem(ctx, activeusers, userid).Err()
	if err != nil {
		utils.Logger.Errorf("failed to remove user %s from redis: %s", userid, err)
		return err
	}
	return nil

}

// 用于获取所有的User
func (redismaintainer *GuestRedisMaintainer) GetAllActiveUser(isstring bool) ([]redis.Z, []string, error) {
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println("Logger is nil")
		os.Exit(-1)
		return nil, nil, fmt.Errorf("logger is nil")
	}
	if !isstring {
		users, err := rdb.ZRangeWithScores(ctx, activeusers, 0, -1).Result()
		if err != nil {
			utils.Logger.Errorf("cannot get activeuser from redis:%v", err.Error())
			return nil, nil, fmt.Errorf("cannot get activeuser from redis")
		}
		return users, nil, nil
	} else {
		users, err := rdb.ZRange(ctx, activeusers, 0, -1).Result()
		if err != nil {
			utils.Logger.Errorf("cannot get activeuser from redis:%v", err.Error())
			return nil, nil, fmt.Errorf("cannot get activeuser from redis")
		}
		return nil, users, nil
	}

}

func (redismaintainer *GuestRedisMaintainer) PrintOnlineUser() error {
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println("Logger is nil")
		os.Exit(-1)
		return fmt.Errorf("logger is nil")
	}
	users, err := rdb.ZRange(ctx, activeusers, 0, -1).Result()
	if err != nil {
		utils.Logger.Errorf("cannot get activeuser from redis:%v", err.Error())
		return fmt.Errorf("cannot get activeuser from redis")
	}
	utils.Logger.Infof("online user : %s", users)
	return nil
}

// 检验此时的用户最后的活跃时间是否符合距离现在不超过30秒
func (redismaintainer *GuestRedisMaintainer) CheckActive(userid string) (bool, error) {
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println("Logger is nil")
		os.Exit(-1)
		return false, fmt.Errorf("logger is nil")
	}
	// 检查ctx是否为空
	if ctx == nil {
		utils.Logger.Errorf("ctx is nil")
		return false, fmt.Errorf("ctx is nil")
	}
	// 检查rdb是否为空
	if rdb == nil {
		utils.Logger.Errorf("rdb is nil")
		return false, fmt.Errorf("rdb is nil")
	}
	// 检查userId是否为空
	if userid == "" {
		utils.Logger.Errorf("userid is nil")
		return false, fmt.Errorf("userid is nil")
	}
	// 对redis做连接测试
	_, err = pingToRedis(rdb)
	if err != nil {
		utils.Logger.Errorf("redis connection error: %s", err)
		return false, err
	}
	// 检查用户是否在线
	isOnline, err := redismaintainer.CheckOnline(userid)
	if err != nil {
		utils.Logger.Errorf("failed to check user online: %s", err)
		return false, err
	}
	if !isOnline {
		utils.Logger.Errorf("user %s is not online", userid)
		return false, fmt.Errorf("user %s is not online", userid)
	}
	//这里就应该是去提取当前的用户的活跃时间，看看是否真实的被改变着，如果是漏删的话，就需要额外的进行删除
	// 提取该用户的lastActiveTime
	lastActiveTime, err := rdb.ZScore(ctx, activeusers, userid).Result()
	if errors.Is(err, redis.Nil) {
		utils.Logger.Errorf("user %s is not online", userid)
		return false, fmt.Errorf("user %s is not online", userid)
	} else if err != nil {
		utils.Logger.Errorf("failed to get user %s from redis: %s", userid, err)
		return false, err
	}
	//现在获取到的这个时间是当时存放float类型的
	// 将lastActiveTime从float64转为time.Time格式
	lastActive := float64ToTime(lastActiveTime)
	// 计算最后活跃时间与当前时间的间隔
	timeSpan := calculateTimeSpan(lastActive)
	// timeSpan取绝对值
	if timeSpan < 0 {
		timeSpan = -timeSpan
	}
	if timeSpan >= disconnectTimeout {
		return false, nil // 用户已断开连接
	}
	return true, nil

}

// 这里是判断是否为重新刷新而删除的用户的，如果此时的活跃时间没有变化的话，这里需要将处理redis跟websocket通道应该是两种处理，我先行将redis的下线了，之后
func (redisMaintainer *GuestRedisMaintainer) checkActiveTimeChanged(userId string, lastActiveTime time.Time) (bool, error) {
	// 检查utils.Logger是否为空
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	// 检查ctx是否为空
	if ctx == nil {
		utils.Logger.Errorf("ctx is nil")
		return false, fmt.Errorf("ctx is nil")
	}
	// 检查rdb是否为空
	if rdb == nil {
		utils.Logger.Errorf("rdb is nil")
		return false, fmt.Errorf("rdb is nil")
	}
	// 检查userId是否为空
	if userId == "" {
		utils.Logger.Errorf("userId is nil")
		return false, fmt.Errorf("userId is nil")
	}

	// 对redis做连接测试
	_, err = pingToRedis(rdb)
	if err != nil {
		utils.Logger.Errorf("redis connection error: %s", err)
		return false, err
	}

	// 检查用户是否在线
	isOnline, err := redisMaintainer.CheckOnline(userId)
	if err != nil {
		utils.Logger.Errorf("failed to check user online: %s", err)
		return false, err
	}
	if !isOnline {
		utils.Logger.Errorf("user %s is not online", userId)
		return false, fmt.Errorf("user %s is not online", userId)
	}

	// 提取该用户的lastActiveTime
	currentActiveTime, err := rdb.ZScore(ctx, activeusers, userId).Result()
	if errors.Is(err, redis.Nil) {
		utils.Logger.Errorf("user %s is not online", userId)
		return false, fmt.Errorf("user %s is not online", userId)
	} else if err != nil {
		utils.Logger.Errorf("failed to get user %s from redis: %s", userId, err)
		return false, err
	}

	// 将 Redis 中的 lastActiveTime 转换为 time.Time 进行比较
	redisActiveTime := time.Unix(int64(currentActiveTime), 0)

	// 定义允许的误差范围为500毫秒
	allowedDrift := 1 * time.Second
	// 判断两时间是否在误差范围内相等
	if lastActiveTime.Sub(redisActiveTime).Abs() > allowedDrift {
		utils.Logger.Warnf("user %s last active time has changed", userId)
		return true, nil
	}
	utils.Logger.Info("这里用户没有及时更新自己的活跃时间，导致现在代码认为用户已经下线了")
	return false, nil // 用户的lastActiveTime未发生变化
}

// 获取userid为id的user实体，用于获取这个最后的活跃时间
func (redisMaintainer *GuestRedisMaintainer) getUserInstance(id string) (*User, error) {
	// 检查utils.Logger是否为空
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	// 检查ctx是否为空
	if ctx == nil {
		utils.Logger.Errorf("ctx is nil")
		return nil, fmt.Errorf("ctx is nil")
	}
	// 检查rdb是否为空
	if rdb == nil {
		utils.Logger.Errorf("rdb is nil")
		return nil, fmt.Errorf("rdb is nil")
	}
	// 检查id是否为空
	if id == "" {
		utils.Logger.Errorf("id is nil")
		return nil, fmt.Errorf("id is nil")
	}

	// 对redis做连接测试
	_, err = pingToRedis(rdb)
	if err != nil {
		utils.Logger.Errorf("redis connection error: %s", err)
		return nil, err
	}

	// 检查用户是否在线
	isOnline, err := redisMaintainer.CheckOnline(id)
	if err != nil {
		utils.Logger.Errorf("failed to check user online: %s", err)
		return nil, err
	}
	if !isOnline {
		utils.Logger.Errorf("user %s is not online", id)
		return nil, fmt.Errorf("user %s is not online", id)
	}

	// 获取用户的lastActiveTime
	lastActiveTime, err := rdb.ZScore(ctx, activeusers, id).Result()
	if errors.Is(err, redis.Nil) {
		utils.Logger.Errorf("user %s is not online", id)
		return nil, fmt.Errorf("user %s is not online", id)
	} else if err != nil {
		utils.Logger.Errorf("failed to get user %s from redis: %s", id, err)
		return nil, err
	}

	// 将lastActiveTime从float64转为time.Time格式
	lastActive := float64ToTime(lastActiveTime)

	return &User{
		userid:     id,
		LastActive: lastActive,
	}, nil
}

// 添加一条新的连接在sync.Map中
func (wsmaintainer *GuestWsMaintainer) AddNewConn(userid string, conn *websocket.Conn) error {
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
	if conn == nil {
		utils.Logger.Errorf("conn is nil")
		return fmt.Errorf("conn is nil")
	}
	WebsocketConns.Store(userid, conn)
	return nil
}

func (wsmaintainer *GuestWsMaintainer) CheckConn(userid string) (bool, error) {
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println("Logger is nil")
		os.Exit(-1)
		return false, err
	}
	if userid == "" {
		utils.Logger.Errorf("userid is nil")
		return false, fmt.Errorf("userid is nil")
	}

	_, ok := WebsocketConns.Load(userid)
	return ok, nil

}

// 从Map里面删除连接
func (wsmaintainer *GuestWsMaintainer) DelOneConn(userid string) error {
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
	WebsocketConns.Delete(userid)
	return nil

}

// 关闭这个websocket实例的连接,这里是告诉前端，让他关闭，也就是让执行游客心跳的协程不再执行
func (wsmaintainer *GuestWsMaintainer) CloseConn(id string) error {
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println("Logger is nil")
		os.Exit(-1)
		return err
	}
	// 检查id是否为空
	if id == "" {
		utils.Logger.Errorf("id is empty")
		return fmt.Errorf("id is empty")
	}
	//需要从连接池中获取这个连接，不能使用这个ws进行传输
	conn, ok := WebsocketConns.Load(id)
	if !ok {
		utils.Logger.Errorf("conn not found")
		return fmt.Errorf("conn not found")
	}
	// 检查conn是否为空
	if conn == nil {
		utils.Logger.Errorf("conn is nil")
		return fmt.Errorf("conn is nil")
	}
	// 类型断言
	wsConn, ok := conn.(*websocket.Conn)
	if !ok {
		utils.Logger.Errorf("failed to convert conn to websocket.Conn")
		return fmt.Errorf("failed to convert conn to websocket.Conn")
	}
	// 向客户端发送关闭消息,必须是有一个超时时间的，否则会一直都堵塞，到真正关闭的时候，会将日志的输出的IO都占用了
	wsConn.SetWriteDeadline(time.Now().Add(receiveTimeout))
	err = websocket.Message.Send(wsConn, "close")
	if err != nil {
		utils.Logger.Errorf("failed to send close message: %s", err)
	}
	// 关闭连接
	err = wsConn.Close()
	if err != nil {
		utils.Logger.Errorf("failed to close conn: %s", err)
		return err
	}
	utils.Logger.Infof("close conn %s successfully", id)

	return nil
}

// TODO

// 根据键值对key来对固定的redis在线用户表进行查找是否在线
func (redismaintainer *GuestRedisMaintainer) CheckOnline(userid string) (bool, error) {
	// 检查utils.Logger是否为空
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	// 检查ctx是否为空
	if ctx == nil {
		utils.Logger.Errorf("ctx is nil")
		return false, fmt.Errorf("ctx is nil")
	}
	// 检查rdb是否为空
	if rdb == nil {
		utils.Logger.Errorf("rdb is nil")
		return false, fmt.Errorf("rdb is nil")
	}
	// 检查userId是否为空
	if userid == "" {
		utils.Logger.Errorf("userId is nil")
		return false, fmt.Errorf("userId is nil")
	}

	// 对redis做连接测试
	_, err = pingToRedis(rdb)
	if err != nil {
		utils.Logger.Errorf("redis connection error: %s", err)
		return false, err
	}

	// 检查ctx是否为空
	if ctx == nil {
		utils.Logger.Errorf("ctx is nil")
		return false, fmt.Errorf("ctx is nil")
	}

	// 检查用户是否在线
	_, err = rdb.ZScore(ctx, activeusers, userid).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil // 用户不在线
	} else if err != nil {
		return false, err
	}

	return true, nil
}

// 更新玩家最后的活跃时间
func (redismaintainer *GuestRedisMaintainer) UpdateActiveTime(userid string, LastActive time.Time) error {
	// 检查utils.Logger是否为空
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println(err)
		return err
	}
	// 检查ctx是否为空
	if ctx == nil {
		utils.Logger.Errorf("ctx is nil")
		return fmt.Errorf("ctx is nil")
	}
	// 检查rdb是否为空
	if rdb == nil {
		utils.Logger.Errorf("rdb is nil")
		return fmt.Errorf("rdb is nil")
	}
	// 检查user.Id是否为空
	if userid == "" {
		utils.Logger.Errorf("user.Id is nil")
		return fmt.Errorf("user.Id is nil")
	}
	// 检查user.LastActive是否为空
	if LastActive.IsZero() {
		utils.Logger.Errorf("user.Id is zero")
		return fmt.Errorf("user.Id is zero")
	}

	currentTime := float64(LastActive.Unix())
	// 对redis做连接测试
	_, err = pingToRedis(rdb)
	if err != nil {
		utils.Logger.Errorf("redis connection error: %s", err)
		return err
	}
	isOnline, err := redismaintainer.CheckOnline(userid)
	if err != nil {
		utils.Logger.Errorf("cannot check online user:%s for:%v", userid, err)
		return err
	}
	if !isOnline {
		utils.Logger.Errorf("user %s is not online", userid)
		return fmt.Errorf("user %s is not online", userid)
	}
	//这个指令是更新分数的
	err = rdb.ZAdd(ctx, activeusers, redis.Z{
		Score:  currentTime,
		Member: userid,
	}).Err()
	if err != nil {
		utils.Logger.Errorf("failed to update user %s active time :%v", userid, err)
		return err
	}
	return nil
}

func (redismaintainer *GuestRedisMaintainer) ClearRedisUser() error {
	// 检查utils.Logger是否为空
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println(err)
		return err
	}
	// 检查ctx是否为空
	if ctx == nil {
		utils.Logger.Errorf("ctx is nil")
		return fmt.Errorf("ctx is nil")
	}
	// 检查rdb是否为空
	if rdb == nil {
		utils.Logger.Errorf("rdb is nil")
		return fmt.Errorf("rdb is nil")
	}
	// 获取所有活跃用户
	activeUsers, err := rdb.ZRangeWithScores(ctx, activeusers, 0, -1).Result()
	if err != nil {
		utils.Logger.Errorf("failed to get all active activeUsers: %s", err)
		return err
	}
	if activeUsers == nil {
		utils.Logger.Warnf("no active activeUsers in redis")
		return fmt.Errorf("no active activeUsers in redis")
	}

	// 删除所有活跃用户
	err = rdb.ZRemRangeByScore(ctx, activeusers, "-inf", "+inf").Err()
	if err != nil {
		utils.Logger.Errorf("failed to remove all active activeUsers: %s", err)
		return err
	}

	return nil
}

// TODO
func (redismaintainer *GuestRedisMaintainer) SetOffline(userid string) error {
	return nil
}

func (wsmaintainer *GuestWsMaintainer) CloseWebsocket(userid string) error {
	return nil
}
