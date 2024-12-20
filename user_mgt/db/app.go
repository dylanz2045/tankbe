package db

import (
	"user_mgt/user_mgt/aes"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	pool  *pgxpool.Pool
	query Query
	Aes   = aes.NewAes()
)

type Query struct {
	getUserNickName string
	getAvatar       string
}

type GuestDBServer struct {
}

type GuestDBHandle interface {
	SavePlayer(string) error
	CheckEmailExist(string) (bool, error)
	GetUserIdByEmail(string) (string, error)
	VerifyPassword(string, string) (bool, error)
	UpdateLastLoginAt(string) error
}

type RegDBServer struct {
}
type RegDBHandle interface {
	GetAvatar(string) (string, error)
}

func NewGuestDBServer() GuestDBHandle {
	return &GuestDBServer{}
}

func NewRegDBServer() RegDBHandle {
	return &RegDBServer{}
}
