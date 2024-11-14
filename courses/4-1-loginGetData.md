# 第四章 继续编写用户管理模块和个人信息修改部分

## 1. 首先我们需要获取到用户输入的信息

把以下代码复制到 <repl-file path="/user_mgt/auth/serve.go">serve.go</repl-file>文件里去：
```go 
package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"user_mgt/utils"
)

func (server *RegHTTPServer) Login(w http.ResponseWriter, r *http.Request) {
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println("Logger is nil")
		http.Error(w, "utils.Logger is nil", http.StatusInternalServerError)
		os.Exit(-1)
		return
	}
	utils.Logger.Info("---->user_mgt/internal/auth.Login is run")
	if r.Method != http.MethodPost {
		utils.Logger.Error("request method isn't POST")
		http.Error(w, "the request isn't POST", http.StatusMethodNotAllowed)
		return
	}

	var user User
	err = json.NewDecoder(r.Body).Decode(&user)
	if user.Email == "" {
		utils.Logger.Errorf("email is empty")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	utils.Logger.Infof("receive email:%s  , password:%s", user.Email, user.Password)
} 
```

把以下代码复制到 <repl-file path="/user_mgt/auth/init.go">init.go</repl-file>文件里去：
```go
package auth

import (
	"net/http"

	"fmt"
)

// Init 初始化auth模块
func Init() {
	InitAuth()

	http.Handle("/api/GuestLogin", http.HandlerFunc(GuestServer.Login))
	http.Handle("/api/RandName", http.HandlerFunc(GuestServer.RegisterRandName))
	http.Handle("/api/VerifyRegData", http.HandlerFunc(GuestServer.VerifyRegData))
	http.Handle("/api/ValidateRegCode", GuestServer.AuthMiddleware(http.HandlerFunc(GuestServer.ValidateRegEmail)))
	http.Handle("/api/ValidateEmailForReset", http.HandlerFunc(GuestServer.ValidateEmailForReset))
	http.Handle("/api/SendCodeWithoutLogin", http.HandlerFunc(GuestServer.SendCodeWithoutLogin))
	http.Handle("/api/ValidateCodeForReset", http.HandlerFunc(GuestServer.ValidateCodeForReset))
	http.Handle("/api/ResetPasswordWithoutLogin", http.HandlerFunc(GuestServer.ResetPasswordWithoutLogin))
	http.Handle("/api/UserLogin", GuestServer.AuthMiddleware(http.HandlerFunc(UserServer.Login)))
	}
// InitAuth auth模块初始化
func InitAuth() {
	// 检查utils.Logger是否为空
	if utils.Logger == nil {
		fmt.Println("utils.Logger is nil")
		return
	}
	utils.Logger.Info("---->user_mgt/internal/auth.InitAuth is run")

	//建表
	err := createTable()
	if err != nil {
		utils.Logger.Errorf("createTable failed, err:%v", err)
		return
	}
}
```

<t->***代码说明***

1. <nt->`"encoding/json"`包可用于操作 `io.Reader` 数据流中携带的 `json` 的数据。

2. <nt->`NewDecoder`方法是创建一个新的 `json` 解码器 ， 并且将 `io.Reader` 数据流作为参数 ，这样解码器就知道从什么地方读取数据啦~

3. <nt->解码器是 `*json.Decoder` 类型 ， 并且使用 `Decode` 对数据流中携带的 `json` 数据解析到GO的结构体中。

    * <t-->此处我们使用了 `User` 这个结构体去接收 ， 用实体 `user` 来存储接收到的信息。

    * <t-->`Decode`这个方法有一个 `error` 类型的返回值 ， 用于检测是否正确读取了 `json` 数据。

    * >数据流中的 `json` 数据跟结构体中定义 `json:""` 是一一对应的，也就是说 `json` 数据流中拥有名称"userid",那么当解码器对数据流进行解码时，会映射到`json:"userid"`的字段中。若没有字段匹配，那么结构体默认是其类型对应的零值

4. <nt->最后就能使用user的实体属性来使用 `json` 数据流中的数据啦！

::: tip 总结
***恭喜你完成了这段代码的学习，你能正常获取到前端发来的信息啦！在前后端通信中，已经迈出了很大的一步，想要验证这个信息是否符合您的要求？快来学习下一节吧~***
:::
