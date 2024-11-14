package maintain

import (
	"fmt"
	"os"
	"testing"
	"time"
	"user_mgt/common/pg_conn"
	"user_mgt/user_mgt/db"
	"user_mgt/user_mgt/maintain"
	jwtutils "user_mgt/user_mgt/test/jwt"
	"user_mgt/utils"

	"golang.org/x/exp/rand"
)

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// 高并发下的用户在线人数峰值
// 最高峰在4300。2024/11/13
func TestPeakOnlineUsers(t *testing.T) {
	// 重定向 stdout 和 stderr，屏蔽print输出
	// devNull, _ := os.Open(os.DevNull)
	// os.Stdout = devNull
	// os.Stderr = devNull
	err := utils.Initlogger()
	if err != nil {
		fmt.Printf("logger init error:%v", err)
		os.Exit(-1)
	}
	if ok := pg_conn.Init(); !ok {
		utils.Logger.Fatalf("初始化数据库失败")
		os.Exit(-1)
	}
	db.Init()
	maintain.Init()

	// 初始化测试组件
	handlerTest := NewHandlerTest()
	serviceTest := NewServiceTest()

	// 设置测试参数
	targetUserAmount := 4200               // 测试目标在线用户数量
	regIdDigits := 12                      // 注册用户ID位数
	sleepTimeAfterTest := 10 * time.Second // 测试结束后的缓冲时间

	// 生成所有用户的token并存入数组
	userTokens := make([]string, targetUserAmount)
	for i := 0; i < targetUserAmount; i++ {
		userTokens[i] = jwtutils.TestGenerateToken(t, generateRandomString(regIdDigits))
	}

	// 定时启动 goroutine，逐批处理
	batchSize := 100                        // 每批次启动的数量
	batchInterval := 200 * time.Millisecond // 每批次的间隔时间

	for i := 0; i < targetUserAmount; i += batchSize {
		end := i + batchSize
		if end > targetUserAmount {
			end = targetUserAmount
		}

		// 启动一批 goroutine
		for j := i; j < end; j++ {
			go handlerTest.TestRegKeepAliveWithToken(t, userTokens[j])
		}

		// 等待批次间隔时间
		time.Sleep(batchInterval)
	}
	// 等待所有 goroutine 启动完成
	fmt.Printf("倒计时 %d 秒开始...\n", int(sleepTimeAfterTest.Seconds()))
	for i := int(sleepTimeAfterTest.Seconds()); i > 0; i-- {
		fmt.Printf("\r剩余时间: %d 秒", i)
		time.Sleep(1 * time.Second)
	}
	fmt.Println("\n倒计时结束!") // 缓冲时间以确保所有 goroutine 保持在线
	// 检查在线用户数量是否达到目标值
	actualOnlineUsersAmount := serviceTest.TestGetOnlineUsersAmount(t)
	if actualOnlineUsersAmount != targetUserAmount {
		// 关闭所有用户在线状态
		for _, token := range userTokens {
			handlerTest.TestCloseRegWebSocketHTTPWithToken(t, token)
		}

		t.Errorf("FAIL! Actual online users amount: %d, target online users amount: %d", actualOnlineUsersAmount, targetUserAmount)

		return
	}

	// 关闭所有用户在线状态
	for _, token := range userTokens {
		handlerTest.TestCloseRegWebSocketHTTPWithToken(t, token)
	}

	t.Logf("SUCCESS! Actual online users amount: %d, target online users amount: %d", actualOnlineUsersAmount, targetUserAmount)

	return
}
