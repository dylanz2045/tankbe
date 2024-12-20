package pg_conn

import (
	"user_mgt/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	dbpool *pgxpool.Pool
	config utils.Config
)
