## 教案学到的GOlang知识点

### 全局变量的使用

在普通一个包里面进行定义类型

```go
package db

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	pool  *pgxpool.Pool
	query Query
)
```

在别的函数内部就不能使用类型断言，这样就会覆盖掉全局性，导致指针丢失的问题

```go
func Init() {
	if utils.Logger == nil {
		fmt.Println("utils logger is nil")
		return
	}
	pool = pg_conn.Getpgpool()
	if pool == nil {
		utils.Logger.Fatalf("获取数据库连接池失败")
	}
	utils.Logger.Infof("Init db pool success")
}
```

### 接口的使用

- 可以定义一个接口类型的结构体，仅仅是一个定义，到达具体的实现，还是需要绑定一个结构体

```go
//接口类型，用于看起来一眼就能看到这个结构体包含什么方法
type WsMaintainer interface {
	addWebSocketConn(userid string)
}

//专门定义的结构体，用于实现这个接口的所有方法
type GuestWsMaintainer struct {
}

//实现
func (wsmaintainer *GuestWsMaintainer) addWebSocketConn(userid string) {

}
```

- 并且，当一个结构体实体想要实现一体化使用两个接口的所有方法

```go
// 一个结构体，嵌入多个接口类型作为,将接口类型嵌入进来
type OnlineUserMaintainerServer struct {
	WsMaintainer
	RedisMaintainer
}
//并且自己结构体中，也包含进来，同时也可以定义自己的接口方法
type OnlineUserMaintainer interface {
	WsMaintainer
	RedisMaintainer
	SetwebSocketConn(userid string)
}
//随即需要定义好工厂函数，实现结构体与接口之间的类型隐式转换
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
//这样使用函数创建出来的实体，就能包含两个接口，并且自己定义出来的函数方法了
```

### 返回的错误，是可以进行fmt.Errorf()转化成字符类型的

- 在处理token的时候，这个*jwt.Token类型的实体，是有valid属性的
- 并且是可以在token中放置一个声明，用于后面检测收到从前端返回过来的Token

在使用websocket的包的时候，可以使用websocket.Message.Receive这个函数，用于固定某一个连接的收发信息





### 处理redis，是可以使用*redis.Client这个类型的变量在redis里面插入键值对的

```go
//添加一个键值对在redis中
err = rdb.ZAdd(ctx, "activeUsers", redis.Z{
		Score:  currentTime,
		Member: userId,
	}).Err()


//查找键值对key为一个值的分数
score, err := rdb.ZScore(ctx, "activeUsers", userId).Result()
if err == nil {
    fmt.Printf("User %s has score %f\n", userId, score)
} else {
    fmt.Printf("User %s does not exist in the activeUsers set\n", userId)
}

//删除一个key为一个值的有序键值对
err = rdb.ZRem(ctx, "activeUsers", userId).Err()
if err != nil {
    fmt.Printf("Error removing user %s from activeUsers set: %v\n", userId, err)
} else {
    fmt.Printf("User %s removed from activeUsers set\n", userId)
}
```

#### 所有声明类型的变量，不开辟创建实体不行，这样会变成空指针，执行不了任何函数

```go
var maintainer OnlineUserMaintainer
	
activeUsers, _, err := maintainer.getAllActiveUser(false)
```

在执行点函数之后，会显示空指针的错误

#### for-select结构，实现协程之间通信，高性能的处理高并发

```go
for {
		select {
		case msg := <-message:
            ......
        case <-quit:
            
        }
}
```

只要通道此时没有信息进入，协程会被挂起，释放CPU的资源。

#### 只有一个计时功能，需要分配一个case 超时管道，增强代码的健壮性，避免造成资源的浪费

- 假如此时计时通道被意外关掉了，那么此时的for-select循环并不会停止，会一直占用内存，一直阻塞在goroutine的调度任务列表中，造成资源的浪费
- 第二种，可以刷新此时管道的新的超时时间，实现随着新任务的执行过程定时开启，定时关闭，提高代码的实时操作性
  - 这里有个坑，就是Reset的函数有bug，在1.22版本以下的GO语言，需要先调用stop让计时器停下，才能重新让计时器重置时间

```go
for {
		select {
        //只要此时有新的信息到达管道，并且能够正确的读取到，就更新超时时间
		case msg := <-message:
            timer.Reset(5 * time.Second) 
            ......
        case <-quit:
            ......
        case <-timer.C
            
        }
}
```

#### websocket的连接发送与接收数据失败

这里面的关闭通道，是有一个等待时间，通道在关闭之后不会马上对通道资源进行释放，而是会等待最长超时时间之后，才会报错。因此需要添加ws.SetDeadReadline(time.Duration)的语句进行控制

- 在修改一个在线用户的状态，还需要检查他的在线状态，否则会误删，在网络状态中，一些硬件的反应慢了的话，就有可能在前面添加了用户的在线状态，但是在后面再收到close信号，就会让添加在线状态失败，最终将连接也删除了。

  - 例子：这个就是处于先添加成功，之后根据用户的在线状态中最终的活跃时间，去查看此时的

  - 王皓解决方案：在每一次更新时间的时候，记录一次时间，之后在删除的时候，将这个时间传进去函数中跟用户在redis中的时间进行比较，如果这两个值的间隔时间在1秒之外的话，就表示着此时用户

    ```go
    
    ```

    ![image-20241104192932497](C:\Users\zdlff\AppData\Roaming\Typora\typora-user-images\image-20241104192932497.png)