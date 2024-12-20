package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"user_mgt/user_mgt/jwtutils"
	"user_mgt/user_mgt/rand"
	"user_mgt/utils"
)

//这里需要区分关于游客登录还是已经注册的用户登录、如果是游客登录，就不需要对他的状态进行维护，就是一个

// Login 这是个游客，会自动登录的，那应该是发送什么方法呢？这个就是会自动获取一个cookie的函数
func (srg *GuestHTTPServer) Login(rep http.ResponseWriter, req *http.Request) {
	//登录进来之后，每一个函数都需要检测
	if utils.Logger == nil {
		fmt.Println("utils.Logger is nil")
		http.Error(rep, "utils.Logger is nil", http.StatusInternalServerError)
		return
	}
	//这里成功进入函数逻辑内
	utils.Logger.Info("---->user_mgt/internal/auth.Login is run")
	//在这里就应该获取里面的标头
	err := AddCoresHeader(rep)
	if err != nil {
		utils.Logger.Errorf("ResponseWriter is nil")
		http.Error(rep, err.Error(), http.StatusInternalServerError)
		return
	}
	if req.Method != http.MethodPost {
		utils.Logger.Error("request method isn't POST")
		http.Error(rep, "方法调用不是post", http.StatusMethodNotAllowed)
		return
	}
	//下面就可以对cookie进行检查，此时的游客
	//第一种：没有任何携带的token，打开这个网站
	//第二种，曾经登录过进来的游客
	cookies := req.Cookies()

	//获取里面携带的Token
	Token, err := JWThandle.GetTokenByCookie(cookies)
	if err != nil {
		http.Error(rep, err.Error(), http.StatusInternalServerError)
		return
	}
	//此时没有携带任何的信息，需要分配一个GuestID,跟正常的需要的秘钥不一样
	if Token == "" {
		utils.Logger.Info("此时没有携带任何身份信息，需要分配")
		GuestID, err := rand.GenerateGuestID()
		if err != nil || GuestID == "" {
			utils.Logger.Errorf("createGuestID failed, err:%v", err)
			http.Error(rep, err.Error(), http.StatusInternalServerError)
			return
		}
		//随后需要在数据库保存这个游客的数据，作为新增用户,这个我就先不做
		err = GuestDBServer.SavePlayer(GuestID)
		if err != nil {
			utils.Logger.Errorf("playerDb.SavePlayer failed, err:%v", err)
			http.Error(rep, err.Error(), http.StatusInternalServerError)
			return
		}
		//生成对应游客受众的Token令牌
		token, err := JWThandle.GenerateToken(GuestID, guestAudience)
		if err != nil || token == "" {
			utils.Logger.Errorf("获取的token失效：%v", err)
			http.Error(rep, err.Error(), http.StatusInternalServerError)
			return
		}
		//之后就是设置cookie，将Token放到cookie里面
		cookie := http.Cookie{
			Name:     "token",
			Value:    token,
			Path:     "/",
			Expires:  time.Now().Add(2 * 24 * time.Hour),
			MaxAge:   172800,
			Secure:   false,
			HttpOnly: false,
			SameSite: http.SameSiteLaxMode,
		}
		http.SetCookie(rep, &cookie)
		utils.Logger.Infof("成功给游客设置cookie")
		rep.WriteHeader(251)
		return
		//如果此时的游客是已经登录的过的，能够检测出来这个携带的Token
	} else {
		//现在就只处理游客的方式,就不需要进行处理么？，因为里面就已经有着需要的身份的东西
		//如果是已经有的Token，这个时候就不一样是Guest的，就需要去verify一下，产看这个区分
		//TODO
		//下一步需要添加这个验证Token,可以进行分流,进行分流
		tokentype, err := JWThandle.IdentifyToken(Token)
		if err != nil {
			utils.Logger.Errorf("IdentifyToken is failed：%v", err)
			http.Error(rep, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Println("---------------------tokenType:", tokentype)
		switch tokentype {
		case TokenTypeGuest:
			rep.WriteHeader(254)
		case TokenTypeRegistered:
			isValid, userid, err := JWThandle.VerifyAndGetIdFromToken(Token, RegType)
			if err != nil {
				utils.Logger.Errorf("VerifyAndGetIdFromToken is failed :%v", err)
				http.Error(rep, err.Error(), http.StatusInternalServerError)
				return
			}
			if isValid {
				utils.Logger.Errorf("token is expired")
				rep.WriteHeader(256)
				http.Error(rep, err.Error(), http.StatusBadRequest)
				return
			}
			err = JWThandle.SetNewCookie(rep, userid, RegType)
			if err != nil {
				utils.Logger.Errorf("SetNewCookie is failed :%v", err)
				http.Error(rep, err.Error(), http.StatusInternalServerError)
				return
			}
			rep.WriteHeader(255)
		case TokenTypeAdmin:
			rep.WriteHeader(257)
		default:
			utils.Logger.Errorf("IdentifyToken is failed：%v", err)
			http.Error(rep, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// 这个是以用户的视角来登录，登录成功之后就是需要断开原本游客身份的在线身份验证
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
		http.Error(w, "方法调用不是post", http.StatusMethodNotAllowed)
		return
	}
	//用于关闭游客的心跳连接的取出userid
	guestID, err := getFromContest(r.Context(), key)

	//TODO ：修改这个更新的语句。这里只能从user里面获取到邮箱跟密码
	var user User
	err = json.NewDecoder(r.Body).Decode(&user)
	//再去判断这个邮箱是否正常存在
	if user.Email == "" {
		utils.Logger.Errorf("email is empty")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	utils.Logger.Info(user.Email, user.Password)
	//查看这个邮箱是否存在
	utils.Logger.Info("---->CheckEmailExisst  run")
	exist, err := GuestDBServer.CheckEmailExist(user.Email)
	if err != nil {
		utils.Logger.Errorf("CheckEmailExist is failed :%v ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !exist {
		utils.Logger.Errorf("email is gone")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//这个地方是获取到邮箱，之后再根据这个邮箱获取ID，这个的ID是已经注册了的用户ID
	//这里我只是想模拟这个登录进行切换websocket的连接
	utils.Logger.Info("---->GetUserIdByEmail  run")
	userid, err := GuestDBServer.GetUserIdByEmail(user.Email)
	if err != nil {
		utils.Logger.Errorf("GetUserIdByEmail is failed ：%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	utils.Logger.Info("---->CheckOnline  run")
	isOnline, err := OnlineMaintainer.CheckOnline(userid)
	if err != nil {
		utils.Logger.Errorf("CheckOnline is failed : %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//如果是在线的话，就通知，告诉前端，无法进行登录，已有用户在线了
	if isOnline {
		utils.Logger.Errorf("user:%s is online", userid)
		sendErrorMessageToFe(w, 461, "user is already online")
		return
	}
	//之后需要在数据库查询密码是否正确
	utils.Logger.Infof("---->password:%s   email:%s", user.Password, user.Email)
	Valid, err := GuestDBServer.VerifyPassword(userid, user.Password)
	if err != nil {
		utils.Logger.Errorf("VerifyPassword is failed :%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !Valid {
		utils.Logger.Errorf("password is invalid ")
		sendErrorMessageToFe(w, 462, "password is invalid")
		return
	}
	//将游客处于下线的状态
	utils.Logger.Info("---->OSCloseWebsocket  run")
	err = OnlineMaintainer.OSCloseWebsocket(guestID)
	if err != nil {
		utils.Logger.Errorf("OSCloseWebsocket is failed ：%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//之后就是更新最后的登录时间，跟设置用户新的cookie时间，也就是更新Token
	utils.Logger.Info("---->UpdateLastLoginAt  run")
	err = GuestDBServer.UpdateLastLoginAt(userid)
	if err != nil {
		utils.Logger.Errorf("UpdateLastLoginAt is failed :%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//之后就是获取一个新的Token，随后分配一个新的Token时间给到响应头，会自动将Token存起来，放到浏览器的
	utils.Logger.Info("---->SetNewCookie  run")
	err = JWThandle.SetNewCookie(w, userid, regAudience)
	if err != nil {
		utils.Logger.Errorf("SetNewCookie is failed :%v ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(253)
	utils.Logger.Info("Geg login successfully")

}

// 中间件，用于检验登录用户的游客ID是否过期，也就是防止API的攻击，这里还有一个作用就是将获取到的用户ID存到上下文，通过上下文进行获取用户的ID
func (srg *GuestHTTPServer) AuthMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := utils.CheckLogger()
		if err != nil {
			fmt.Println("Logger is nil")
			os.Exit(-1)
			http.Error(w, "utils.Logger is nil", http.StatusInternalServerError)
			return
		}
		cookies := r.Cookies()
		//获取里面携带的Token
		Token, err := JWThandle.GetTokenByCookie(cookies)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if Token == "" {
			utils.Logger.Errorf("token is empty")
			http.Error(w, "token is empty", http.StatusInternalServerError)
			return
		}
		isValid, guestid, err := JWThandle.VerifyAndGetIdFromToken(Token, GuestType)
		if err != nil {
			utils.Logger.Errorf("VerifyAndGetIdFromToken is failed")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//这个是过期的情况
		if isValid {
			if err != nil {
				utils.Logger.Errorf("set a new Cookie failed :%v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		ctx, err := setToContext(r.Context(), key, guestid)
		if err != nil {
			utils.Logger.Errorf("setToContext is failed :%v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// 验证此时已经登录的用户的token
func (server *RegHTTPServer) AuthMiddelWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查utils.Logger是否为空
		err := utils.ValidateLogger()
		if err != nil {
			fmt.Println(err)
			return
		}
		// 从Cookie中获取token
		tokenCookie, err := r.Cookie("token")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				utils.Logger.Errorf("No token in cookie")
				http.Error(w, "No token in cookie", http.StatusBadRequest)
			}
			utils.Logger.Errorf("Failed to get token from cookie: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if tokenCookie == nil {
			utils.Logger.Errorf("tokenCookie is nil")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		utils.Logger.Infof("receive tokenCookie: %v", tokenCookie)

		token := tokenCookie.Value

		// 验证token
		jwtService := jwtutils.NewJWTserve()
		isToken, userId, err := jwtService.VerifyAndGetIdFromToken(token, jwtutils.TokenTypeRegistered)
		if err != nil {
			utils.Logger.Errorf("check token failed: %s", err.Error())
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if isToken {
			utils.Logger.Errorf("message is not a token")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if userId == "" {
			utils.Logger.Errorf("userId is empty")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// 将userId存入HTTP请求context
		ctx, err := setToContext(r.Context(), "userId", userId)
		if err != nil {
			utils.Logger.Errorf("cannot set ID to Context")
			http.Error(w, "Unauthorized", http.StatusInternalServerError)
			return
		}

		// 调用下一个处理程序
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// 假设我现在把头像就放在当前的目录之下，是需要添加上user_mgt之下的
func (server *RegHTTPServer) GetAvatar(w http.ResponseWriter, r *http.Request) {
	// 检查utils.Logger是否为空
	err := utils.ValidateLogger()
	if err != nil {
		fmt.Println(err)
		return
	}
	//需要从上下文中获取用户的ID
	userID, err := getFromContest(r.Context(), "userId")
	if err != nil {
		utils.Logger.Errorf("cannot get userid from context!:%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if userID == "" {
		utils.Logger.Errorf("got userid is empty!")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//这里是已经获取到存在数据库中的头像地址
	avatarpath, err := RegDBServer.GetAvatar(userID)
	if err != nil {
		utils.Logger.Errorf("GetAvatar is failed! :%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if avatarpath == "" {
		utils.Logger.Errorf("filepath is empty! :%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	truepath := filepath.Join("user_mgt", avatarpath)
	exist, err := IsAvatarPathExists(&truepath)
	if err != nil {
		utils.Logger.Errorf("cannot find avatar from that path!:%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !exist {
		utils.Logger.Errorf("cannot  find that avatar")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//打开这个文件，之后就可以对这个文件进行操作了
	file, err := os.Open(truepath)
	if err != nil {
		utils.Logger.Errorf("ReadFile failed, err:%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()
	//以二进制的形式发送到前端
	w.Header().Set("Content-Type", "image/jpeg")
	_, err = io.Copy(w, file)
	if err != nil {
		utils.Logger.Errorf("w.Write failed, err:%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// 用于验证数据库里是否存在这个属性，先获取头像，再修改，直接进行写文件的形式。
func (server *RegHTTPServer) VerifyAndChangeAvatar(w http.ResponseWriter, r *http.Request) {
	// 检查utils.Logger是否为空
	err := utils.ValidateLogger()
	if err != nil {
		fmt.Println(err)
		return
	}
	//检查完之后，就需要从数据库里面进行修改，之后执行一次获取头像的操作
}
