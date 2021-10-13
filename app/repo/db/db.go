package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/teamyapp/teamy-backend/app/config"
)

const dbType = "postgres"

func Connect(cfg config.Config) (*sql.DB, error) {
	dbSource := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DbHost,
		cfg.DbPort,
		cfg.DbUser,
		cfg.DbPassword,
		cfg.DBName)
	return sql.Open(dbType, dbSource)
}
