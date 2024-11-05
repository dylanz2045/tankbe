package usermgt

import (
	"user_mgt/user_mgt/auth"
	"user_mgt/user_mgt/db"
	"user_mgt/user_mgt/maintain"
)

func Run() {
	auth.Init()
	db.Init()
	maintain.Init()
}
