package pg_conn

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
	"user_mgt/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

// 定义一个数据库的连接池，存放在go后端的内存中，可以在需要的时候访问他，获取一个连接
// 也需要进行配置，配置这个连接池的大小
func Init() bool {
	// 获取调用者信息(包名/文件名:行数)
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		if utils.Logger == nil {
			fmt.Println("zap.L() create utils.Logger failed, please check zap utils.Logger")
			os.Exit(-1)
		}
		utils.Logger.Error("获取调用者信息失败,初始化连接数据库失败")
		return false
	}

	if line == 0 {
		if utils.Logger == nil {
			fmt.Println("zap.L() create utils.Logger failed, please check zap utils.Logger")
			os.Exit(-1)
		}
		utils.Logger.Error("获取调用行数失败,初始化连接数据库失败")
		return false
	}

	if file == "" {
		if utils.Logger == nil {
			fmt.Println("zap.L() create utils.Logger failed, please check zap utils.Logger")
			os.Exit(-1)
		}
		utils.Logger.Error("获取调用文件路径失败,初始化连接数据库失败")
		return false
	}

	callFrom := fmt.Sprintf("%s/%s:%d", filepath.Base(filepath.Dir(file)), filepath.Base(file), line)
	if callFrom == "" {
		if utils.Logger == nil {
			fmt.Println("zap.L() create utils.Logger failed, please check zap utils.Logger")
			os.Exit(-1)
		}
		utils.Logger.Error("格式化调用者信息失败,初始化连接数据库失败")
	}
	// 组建连接数据库的语句
	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", "tankuser", "cst4Ever", "localhost", "6900", "tankdb")

	//创建连接池的配置
	poolconfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		utils.Logger.Errorf("'" + callFrom + "'进行初始化连接数据库(" + connStr + ")配置失败,错误详情:" + err.Error())
		return false
	}
	//进行配置
	poolconfig.MaxConns = 15
	poolconfig.MaxConnLifetime = 5 * time.Minute

	pools, err := pgxpool.NewWithConfig(context.Background(), poolconfig)
	if err != nil {
		utils.Logger.Fatalf("创建数据库连接池失败")
		os.Exit(-1)
		return false
	}
	dbpool = pools
	//检查此时数据库连接池是否正常运行
	conn, err := dbpool.Acquire(context.Background())
	if err != nil {
		utils.Logger.Fatalf("从连接池中创建并添加连接失败 :%v", err)
		os.Exit(-1)
		return false
	}
	//测试连接是否正常运行
	var version string
	err = conn.QueryRow(context.Background(), "SELECT VERSION()").Scan(&version)
	if err != nil {
		utils.Logger.Fatalf("查询数据库版本失败，获取连接无法正常查询 :%v", err)
		os.Exit(-1)
		return false
	}
	utils.Logger.Infof("此时数据库版本为：%s", version)
	conn.Release()
	utils.Logger.Info("'成功初始化连接数据库(" + connStr + ")")
	return true
}

func Getpgpool() *pgxpool.Pool {
	//检查此时数据库连接池是否正常运行
	conn, err := dbpool.Acquire(context.Background())
	if err != nil {
		utils.Logger.Fatalf("从连接池中创建并添加连接失败 :%v", err)
		os.Exit(-1)
		return nil
	}
	//测试连接是否正常运行
	var version string
	err = conn.QueryRow(context.Background(), "SELECT VERSION()").Scan(&version)
	if err != nil {
		utils.Logger.Fatalf("查询数据库版本失败，获取连接无法正常查询 :%v", err)
		os.Exit(-1)
		return nil
	}
	utils.Logger.Infof("此时数据库版本为：%s", version)
	conn.Release()
	utils.Logger.Info("数据库连接池有效")
	return dbpool
}
