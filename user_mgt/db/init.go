package db

import (
	"fmt"
	"user_mgt/common/pg_conn"
	"user_mgt/utils"
)

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

// 如果在错误的情况下，必须关闭这个连接池，防止数据库被攻击
func Close() {
	pool.Close()
}
