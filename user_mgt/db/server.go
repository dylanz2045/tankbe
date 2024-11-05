package db

import (
	"context"
	"fmt"
	"os"
	"time"
	"user_mgt/utils"
)

const (
	savePlayer        = `INSERT INTO user_mgt.t_players(id) VALUES($1)`
	checkEmailExist   = `SELECT EXISTS(SELECT 1 FROM user_mgt.t_user WHERE email=$1 AND deleted_at IS NULL)`
	getUserIdByEmail  = `SELECT id FROM user_mgt.t_user WHERE email=$1 AND deleted_at IS NULL`
	verifyPassword    = `SELECT password FROM user_mgt.t_user WHERE id=$1`
	updateLastLoginAt = `UPDATE user_mgt.t_user SET last_login_at=$1 WHERE id=$2`
)

func (guestdbserver *GuestDBServer) SavePlayer(guestid string) error {
	err := utils.ValidateLogger()
	if err != nil {
		return fmt.Errorf("cannot find Logger %v", err)
	}
	if guestid == "" {
		return fmt.Errorf("传入的参数为空,请检查错误")
	}
	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		utils.Logger.Fatalf("开展一个数据库的事务失败：%v", err)
		return err
	}
	_, err = tx.Exec(ctx, savePlayer, guestid)
	if err != nil {
		utils.Logger.Errorf("执行数据库操作失败：%v", err)
		tx.Rollback(ctx)
		return err
	}
	tx.Commit(ctx)
	return nil

}

func (guestdbserver *GuestDBServer) CheckEmailExist(email string) (bool, error) {
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println("Logger is nil")
		os.Exit(-1)
		return false, fmt.Errorf("Logger is nil")
	}
	if email == "" {
		utils.Logger.Errorf("key is nil")
		return false, fmt.Errorf("email is nil")
	}
	var exist bool
	err = pool.QueryRow(context.Background(), checkEmailExist, email).Scan(&exist)
	if err != nil {
		utils.Logger.Errorf("cannot deal with database in checkEmailexist :%v", err)
		return false, fmt.Errorf("cannot deal with database in checkEmailexist :%v", err)
	}
	if exist {
		return true, nil
	} else {
		return false, nil
	}
}
func (guestdbserver *GuestDBServer) GetUserIdByEmail(email string) (string, error) {
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println("Logger is nil")
		os.Exit(-1)
		return "", fmt.Errorf("Logger is nil")
	}
	if email == "" {
		utils.Logger.Errorf("email is nil")
		return "", fmt.Errorf("email is nil")
	}
	var userid string
	err = pool.QueryRow(context.Background(), getUserIdByEmail, email).Scan(&userid)
	if err != nil {
		utils.Logger.Errorf("cannot deal with database in getUserIdByEmail :%v", err)
		return "", fmt.Errorf("cannot deal with database in getUserIdByEmail :%v", err)
	}
	return userid, nil
}

// 验证userid跟密码是否为同一行数据  useid password
func (guestdbserver *GuestDBServer) VerifyPassword(userid string, password string) (bool, error) {
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println("Logger is nil")
		os.Exit(-1)
		return false, fmt.Errorf("Logger is nil")
	}
	if userid == "" {
		utils.Logger.Errorf("userid is nil")
		return false, fmt.Errorf("userid is nil")
	}
	if password == "" {
		utils.Logger.Errorf("password is nil")
		return false, fmt.Errorf("password is nil")
	}
	//加密密码,通过取出加密的密码之后，与数据库的密码进行比较
	encryptPwd, err := Aes.Encrypt(password)
	if err != nil {
		utils.Logger.Errorf("Encrypt is failed :%v", err)
		return false, fmt.Errorf("Encrypt is failed :%v", err)
	}
	//查询用户存在数据库的密码
	var userpassword string
	err = pool.QueryRow(context.Background(), verifyPassword, userid).Scan(&userpassword)
	if err != nil {
		utils.Logger.Errorf("cannot deal with database in VerifyPassword :%v", err)
		return false, fmt.Errorf("cannot deal with database in VerifyPassword :%v", err)
	}
	if userpassword == encryptPwd {
		return true, nil
	} else {
		return false, nil
	}
}

// 更新用户最近登录的时间
func (guestdbserver *GuestDBServer) UpdateLastLoginAt(userid string) error {
	err := utils.CheckLogger()
	if err != nil {
		fmt.Println("Logger is nil")
		os.Exit(-1)
		return fmt.Errorf("Logger is nil")
	}
	if userid == "" {
		utils.Logger.Errorf("userid is nil")
		return fmt.Errorf("userid is nil")
	}
	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		utils.Logger.Fatalf("take a transaction failed：%v", err)
		return err
	}
	_, err = tx.Exec(ctx, updateLastLoginAt, time.Now(), userid)
	if err != nil {
		utils.Logger.Errorf("cannot deal with database in updateLastLoginAt%v", err)
		tx.Rollback(ctx)
		return err
	}
	tx.Commit(ctx)
	return nil
}
