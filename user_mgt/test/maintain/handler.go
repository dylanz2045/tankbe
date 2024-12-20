package maintain

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
	"user_mgt/user_mgt/maintain"
	"user_mgt/user_mgt/test/model"

	"golang.org/x/net/websocket"
)

type HandlerTest interface {
	TestGuestKeepAlive(t *testing.T)
	TestRegKeepAlive(t *testing.T)
	TestGuestKeepAliveWithToken(t *testing.T, guestToken string)
	TestRegKeepAliveWithToken(t *testing.T, regToken string)
	TestCloseRegWebSocketHTTPWithToken(t *testing.T, Token string)
}

type HandlerTestImpl struct {
}

func NewHandlerTest() HandlerTest {
	return &HandlerTestImpl{}
}

type ProcessHandlerTest interface {
	TestGuestKeepAlive(t *testing.T, guestToken string, timeoutCtx context.Context)
	TestUser0KeepAlive(t *testing.T, infoCtx context.Context, timeoutCtx context.Context)
	TestUser1KeepAlive(t *testing.T, infoCtx context.Context, timeoutCtx context.Context)
}

type ProcessHandlerTestImpl struct {
}

func NewProcessHandlerTest() ProcessHandlerTest {
	return &ProcessHandlerTestImpl{}
}

func (*HandlerTestImpl) TestGuestKeepAlive(t *testing.T) {
	// 创建测试服务器
	guest := maintain.RegMaintainer{}
	server := httptest.NewServer(websocket.Handler(guest.KeepAlive))
	defer server.Close()

	// 解析WebSocket URL
	wsURL := "ws" + server.URL[len("http"):]

	// 创建WebSocket连接
	ws, err := websocket.Dial(wsURL, "", "http://localhost/")
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer ws.Close()

	// 创建一个通道来阻塞主协程
	done := make(chan struct{})
	var once sync.Once // 用来确保通道只被关闭一次

	// 启动一个协程，定时发送消息
	go func() {
		ticker := time.NewTicker(2 * time.Second) // 每2秒发送一次消息
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				err := websocket.Message.Send(ws, guestToken)
				if err != nil {
					t.Errorf("Failed to send message: %v", err)
					once.Do(func() { close(done) }) // 确保通道只关闭一次
					return
				}
				t.Logf("Sent message: %s", guestToken)
			}
		}
	}()

	// 启动一个新的协程来保持连接
	go func() {
		for {
			var receivedMessage string
			err = websocket.Message.Receive(ws, &receivedMessage)
			if err != nil {
				t.Errorf("Failed to receive message: %v", err)
				once.Do(func() { close(done) }) // 确保通道只关闭一次
				return
			}

			// 处理收到的消息（在这里你可以进行断言或者其他操作）
			t.Logf("Received message: %s", receivedMessage)

			// 若收到close消息，则视为连接已关闭
			if receivedMessage == "close" {
				once.Do(func() { close(done) }) // 确保通道只关闭一次
				return
			}
		}
	}()

	// 发送消息
	message := guestToken
	err = websocket.Message.Send(ws, message)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// 等待一定时间，确保连接保持打开状态
	select {
	case <-done:
		t.Log("Connection closed")
		return
	case <-time.After(100 * time.Minute): // 等待10分钟
		t.Log("Test completed, connection remains open")
		return
	}
}

func (*HandlerTestImpl) TestRegKeepAlive(t *testing.T) {
	// 创建测试服务器s
	Reg := maintain.RegMaintainer{}
	server := httptest.NewServer(websocket.Handler(Reg.KeepAlive))
	defer server.Close()

	// 解析WebSocket URL
	wsURL := "ws" + server.URL[len("http"):] + "/echo"

	// 创建WebSocket连接
	ws, err := websocket.Dial(wsURL, "", "http://localhost/")
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer ws.Close()

	// 创建一个通道来阻塞主协程
	done := make(chan struct{})
	var once sync.Once // 用来确保通道只被关闭一次

	// 发送token进行身份验证
	message := regToken
	err = websocket.Message.Send(ws, message)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// 启动一个新的协程来保持连接
	go func() {
		for {
			var receivedMessage string
			err = websocket.Message.Receive(ws, &receivedMessage)
			if err != nil {
				t.Errorf("Failed to receive message: %v", err)
				once.Do(func() { close(done) }) // 确保通道只关闭一次
				return
			}

			// 处理收到的消息（在这里你可以进行断言或者其他操作）
			t.Logf("Received message: %s", receivedMessage)

			// 若收到close消息，则视为连接已关闭
			if receivedMessage == "close" {
				once.Do(func() { close(done) }) // 确保通道只关闭一次
				return
			}
		}
	}()

	// 启动一个协程，定时发送消息
	go func() {
		ticker := time.NewTicker(2 * time.Second) // 每2秒发送一次消息
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				err := websocket.Message.Send(ws, regToken)
				if err != nil {
					t.Errorf("Failed to send message: %v", err)
					once.Do(func() { close(done) }) // 确保通道只关闭一次
					return
				}
				t.Logf("Sent message: %s", regToken)
			}
		}
	}()

	// 等待一定时间，确保连接保持打开状态
	select {
	case <-done:
		t.Log("Connection closed")
		return
	case <-time.After(100 * time.Minute): // 等待10分钟
		t.Log("Test completed, connection remains open")
		return
	}
}

func (*HandlerTestImpl) TestGuestKeepAliveWithToken(t *testing.T, guestToken string) {
	// 创建测试服务器
	guest := maintain.GuestMaintainer{}
	server := httptest.NewServer(websocket.Handler(guest.KeepAlive))
	defer server.Close()

	// 解析WebSocket URL
	wsURL := "ws" + server.URL[len("http"):] + "/echo"

	// 创建WebSocket连接
	ws, err := websocket.Dial(wsURL, "", "http://localhost/")
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer ws.Close()

	// 创建一个通道来阻塞主协程
	done := make(chan struct{})
	var once sync.Once // 用来确保通道只被关闭一次

	// 启动一个协程，定时发送消息
	go func() {
		ticker := time.NewTicker(2 * time.Second) // 每2秒发送一次消息
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				err := websocket.Message.Send(ws, guestToken)
				if err != nil {
					t.Errorf("Failed to send message: %v", err)
					once.Do(func() { close(done) }) // 确保通道只关闭一次
					return
				}
				t.Logf("Sent message: %s", guestToken)
			}
		}
	}()

	// 启动一个新的协程来保持连接
	go func() {
		for {
			var receivedMessage string
			err = websocket.Message.Receive(ws, &receivedMessage)
			if err != nil {
				t.Errorf("Failed to receive message: %v", err)
				once.Do(func() { close(done) }) // 确保通道只关闭一次
				return
			}

			// 处理收到的消息（在这里你可以进行断言或者其他操作）
			t.Logf("Received message: %s", receivedMessage)

			// 若收到close消息，则视为连接已关闭
			if receivedMessage == "close" {
				once.Do(func() { close(done) }) // 确保通道只关闭一次
				return
			}
		}
	}()

	// 发送消息
	message := guestToken
	err = websocket.Message.Send(ws, message)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// 等待一定时间，确保连接保持打开状态
	select {
	case <-done:
		t.Log("Connection closed")
		return
	case <-time.After(100 * time.Minute): // 等待10分钟
		t.Log("Test completed, connection remains open")
		return
	}
}

func (*HandlerTestImpl) TestRegKeepAliveWithToken(t *testing.T, regToken string) {
	// 创建测试服务器
	Reg := maintain.RegMaintainer{}
	server := httptest.NewServer(websocket.Handler(Reg.KeepAlive))
	defer server.Close()

	// 解析WebSocket URL
	wsURL := "ws" + server.URL[len("http"):] + "/echo"

	// 创建WebSocket连接
	ws, err := websocket.Dial(wsURL, "", "http://localhost/")
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer ws.Close()

	// 创建一个通道来阻塞主协程
	done := make(chan struct{})
	var once sync.Once // 用来确保通道只被关闭一次

	// 发送token进行身份验证
	message := regToken
	err = websocket.Message.Send(ws, message)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// 启动一个新的协程来保持连接
	go func() {
		for {
			var receivedMessage string
			err = websocket.Message.Receive(ws, &receivedMessage)
			if err != nil {
				t.Errorf("Failed to receive message: %v", err)
				once.Do(func() { close(done) }) // 确保通道只关闭一次
				return
			}

			// 处理收到的消息（在这里你可以进行断言或者其他操作）
			t.Logf("Received message: %s", receivedMessage)

			// 若收到close消息，则视为连接已关闭
			if receivedMessage == "close" {
				once.Do(func() { close(done) }) // 确保通道只关闭一次
				return
			}
		}
	}()

	// 启动一个协程，定时发送消息
	go func() {
		ticker := time.NewTicker(2 * time.Second) // 每2秒发送一次消息
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				err := websocket.Message.Send(ws, regToken)
				if err != nil {
					t.Errorf("Failed to send message: %v", err)
					once.Do(func() { close(done) }) // 确保通道只关闭一次
					return
				}
				t.Logf("Sent message: %s", regToken)
			}
		}
	}()

	// 等待一定时间，确保连接保持打开状态
	select {
	case <-done:
		t.Log("Connection closed")
		return
	case <-time.After(100 * time.Minute): // 等待10分钟
		t.Log("Test completed, connection remains open")
		return
	}
}

func (*ProcessHandlerTestImpl) TestGuestKeepAlive(t *testing.T, guestToken string, timeoutCtx context.Context) {
	// 创建测试服务器
	guest := maintain.GuestMaintainer{}
	server := httptest.NewServer(websocket.Handler(guest.KeepAlive))
	defer server.Close()

	// 解析WebSocket URL
	wsURL := "ws" + server.URL[len("http"):] + "/echo"

	// 创建WebSocket连接
	ws, err := websocket.Dial(wsURL, "", "http://localhost/")
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer ws.Close()

	// 创建一个通道来阻塞主协程
	done := make(chan struct{})
	timeoutDone := make(chan struct{})

	// 启动一个新的协程来保持连接
	go func() {
		for {
			println("TestGuestKeepAlive running...")

			select { // 使用select来处理超时
			case <-timeoutCtx.Done():
				// 发送close消息
				err = websocket.Message.Send(ws, "close")
				close(timeoutDone)
				return

			default:
				// 设置ws的超时时间
				err = ws.SetDeadline(time.Now().Add(wsTimeout))
				if err != nil {
					t.Errorf("Failed to set deadline: %v", err)
					// 发送close消息
					err = websocket.Message.Send(ws, "close")
					close(done)
					return
				}

				var receivedMessage string
				err = websocket.Message.Receive(ws, &receivedMessage)
				if err != nil {
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						println("TestGuestKeepAlive receive message timeout")
					} else {
						t.Errorf("Failed to receive message: %v", err)
						// 发送close消息
						err = websocket.Message.Send(ws, "close")
						close(done)
						return
					}
				}

				// 若收到close消息，则关闭连接
				if receivedMessage == "close" {
					err = ws.Close()
					if err != nil {
						t.Errorf("Failed to close connection: %v", err)
						close(done)
						return
					} else {
						close(done)
						return
					}
				}
			}
		}
	}()

	// 发送消息
	message := guestToken
	err = websocket.Message.Send(ws, message)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// 等待一定时间，确保连接保持打开状态
	for {
		select {
		case <-timeoutDone:
			println("TestGuestKeepAlive timeout")
			return
		case <-done:
			println("TestGuestKeepAlive connection closed")
			return
		case <-time.After(10 * time.Minute): // 等待10分钟
			t.Log("Test completed, connection remains open")
			return
		}
	}
}

func (*ProcessHandlerTestImpl) TestUser0KeepAlive(t *testing.T, infoCtx context.Context, timeoutCtx context.Context) {
	// 从context中获取用户信息
	user0InContext := infoCtx.Value("user0").(model.UserInfo)
	regToken := user0InContext.Token

	// 创建测试服务器
	Reg := maintain.RegMaintainer{}
	server := httptest.NewServer(websocket.Handler(Reg.KeepAlive))
	defer server.Close()

	// 解析WebSocket URL
	wsURL := "ws" + server.URL[len("http"):] + "/echo"

	// 创建WebSocket连接
	ws, err := websocket.Dial(wsURL, "", "http://localhost/")
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer ws.Close()

	// 创建一个通道来阻塞主协程
	done := make(chan struct{})
	timeoutDone := make(chan struct{})

	// 启动一个新的协程来保持连接
	go func() {
		for {
			println("TestUser0KeepAlive running...")

			select { // 使用select来处理超时
			case <-timeoutCtx.Done():
				// 发送close消息
				err = websocket.Message.Send(ws, "close")
				close(timeoutDone)
				return

			default:
				// 设置ws的超时时间
				err = ws.SetDeadline(time.Now().Add(wsTimeout))
				if err != nil {
					t.Errorf("Failed to set deadline: %v", err)
					// 发送close消息
					err = websocket.Message.Send(ws, "close")
					close(done)
					return
				}

				var receivedMessage string
				err = websocket.Message.Receive(ws, &receivedMessage)
				if err != nil {
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						println("TestUser0KeepAlive receive message timeout")
					} else {
						t.Errorf("Failed to receive message: %v", err)
						// 发送close消息
						err = websocket.Message.Send(ws, "close")
						close(done)
						return
					}
				}

				// 若收到close消息，则关闭连接
				if receivedMessage == "close" {
					err = ws.Close()
					if err != nil {
						t.Errorf("Failed to close connection: %v", err)
						close(done)
						return
					} else {
						close(done)
						return
					}
				}
			}
		}
	}()

	// 发送消息
	message := regToken
	err = websocket.Message.Send(ws, message)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// 等待一定时间，确保连接保持打开状态
	for {
		select {
		case <-timeoutDone:
			println("TestUser0KeepAlive timeout")
			return
		case <-done:
			println("TestUser0KeepAlive connection closed")
			return
		case <-time.After(10 * time.Minute): // 等待10分钟
			t.Log("Test completed, connection remains open")
			return
		}
	}
}

func (*ProcessHandlerTestImpl) TestUser1KeepAlive(t *testing.T, infoCtx context.Context, timeoutCtx context.Context) {
	// 从context中获取用户信息
	user0InContext := infoCtx.Value("user1").(model.UserInfo)
	regToken := user0InContext.Token

	// 创建测试服务器
	Reg := maintain.RegMaintainer{}
	server := httptest.NewServer(websocket.Handler(Reg.KeepAlive))
	defer server.Close()

	// 解析WebSocket URL
	wsURL := "ws" + server.URL[len("http"):] + "/echo"

	// 创建WebSocket连接
	ws, err := websocket.Dial(wsURL, "", "http://localhost/")
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer ws.Close()

	// 创建一个通道来阻塞主协程
	done := make(chan struct{})
	timeoutDone := make(chan struct{})

	// 启动一个新的协程来保持连接
	go func() {
		for {
			println("TestUser1KeepAlive running...")

			select { // 使用select来处理超时
			case <-timeoutCtx.Done():
				// 发送close消息
				err = websocket.Message.Send(ws, "close")
				close(timeoutDone)
				return

			default:
				// 设置ws的超时时间
				err = ws.SetDeadline(time.Now().Add(wsTimeout))
				if err != nil {
					t.Errorf("Failed to set deadline: %v", err)
					// 发送close消息
					err = websocket.Message.Send(ws, "close")
					close(done)
					return
				}

				var receivedMessage string
				err = websocket.Message.Receive(ws, &receivedMessage)
				if err != nil {
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						println("TestUser1KeepAlive receive message timeout")
					} else {
						t.Errorf("Failed to receive message: %v", err)
						// 发送close消息
						err = websocket.Message.Send(ws, "close")
						close(done)
						return
					}
				}

				// 若收到close消息，则关闭连接
				if receivedMessage == "close" {
					err = ws.Close()
					if err != nil {
						t.Errorf("Failed to close connection: %v", err)
						close(done)
						return
					} else {
						close(done)
						return
					}
				}
			}
		}
	}()

	// 发送消息
	message := regToken
	err = websocket.Message.Send(ws, message)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// 等待一定时间，确保连接保持打开状态
	for {
		select {
		case <-timeoutDone:
			println("TestUser1KeepAlive timeout")
			return
		case <-done:
			println("TestUser1KeepAlive connection closed")
			return
		case <-time.After(10 * time.Minute): // 等待10分钟
			t.Log("Test completed, connection remains open")
			return
		}
	}
}

func (*HandlerTestImpl) TestCloseRegWebSocketHTTPWithToken(t *testing.T, regToken string) {
	// 创建一个请求
	req, err := http.NewRequest("POST", "/api/CloseRegWebSocket", nil)
	if err != nil {
		t.Fatal(err)
	}
	// 添加一个cookie到请求中
	cookie := &http.Cookie{Name: "token", Value: regToken}
	req.AddCookie(cookie)

	Reg := maintain.RegMaintainer{}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Reg.CloseWebsocketHTTP)

	handler.ServeHTTP(rr, req)
	t.Logf("close reg websocket successfully")

}
