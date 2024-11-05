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

1. <nt->我们通过使用 `err = json.NewDecoder(r.Body).Decode(&user)` ,这个语句来对携带json数据流的数据进行解码并且获取。`json` 是一个GO语言的 `"encoding/json"` 中的一个包,能使用这个包的方法来操作 `json` 结构的数据

   1. <nt->其中的 `NewDecoder` 方法是用于将随着请求体携带的JSON数据进行解码，接受一个`io.Reader`接口作为参数，并且返回一个 `*json.Decoder` 类型的实例

   2. <nt-> `io.Reader` 的参数通常是一个文件、网络连接或者其他类型的数据流 `*json.Decoder` 的实例能使用一个方法: `Decode` ，将从字节流中读到的json解码到结构体中

   3. <nt->我们使用 `User` 这个结构体去接收来自请求中的数据流，请求中有的字段，都会解析到这个字段对应 `json` 码的字段上。结构体中定义的字段数量与 `json` 数据中的字段数量之间的关系会影响解码过程,以下是两种情况

      * <t-->结构体字段少于 `json` 数据字段：解码器会忽略 `json` 中那些没有对应字段的数据，过程中不会报错
      
        例如：
          ```go
          type User struct {
             Name string `json:"name"`
             Age int    `json:"age"`
		  } 
          // JSON 数据
          jsonData := `{"name": "John", "age": 30, "email": "john@example.com"}`
          ```
        
        在这个例子中，User 结构体只有 `Name` 和 `Age` 两个字段，而 `json` 数据中有三个字段。解码后，`User` 结构体的 `Name` 和 `Age` 字段会被填充，但 email 字段会被忽略。
      * <t-->结构体字段多于 `json` 数据字段：解码器会将结构体中未在 `json` 数据中找到对应字段的属性设置为其类型的零值。例如如果一个字段是 `int` 类型，它的零值是 `0`；如果是 `string` 类型，零值是空字符串 `""`；如果是指针类型，零值是 `nil`

         例如：
           ```go
           type User struct {
              Name string `json:"name"`
              Age int    `json:"age"`
		   } 
          // JSON 数据
          jsonData := `{"name": "John", "age": 30, "email": "john@example.com"}`
          ```
        在这个例子中，`User` 结构体有三个字段，但 `json` 数据中只有两个字段。解码后，`User` 结构体的 `Name` 和 `Age` 字段会被填充，而 `Email` 字段会被设置为空字符串 `""`
      
2. <nt->我们将 `*http.request` 的实体r中的body，作为 `io.Reader` 接口的参数，请求体携带的数据放在body中，将这个 `body` 中的 `json` 数据流读取到 `User` 的结构体中，最后就能通过 `user` 的属性来引用请求的数据啦~

::: tip 总结
***恭喜你完成了这段代码的学习，你能正常获取到前端发来的信息啦！在前后端通信中，已经迈出了很大的一步，想要验证这个信息是否符合您的要求？快来学习下一节吧~***
:::
