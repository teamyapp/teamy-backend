package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

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

func WaitUntilReady(sqlDB *sql.DB) {
	for {
		err := sqlDB.Ping()
		if err == nil {
			log.Println("successfully connected to the DB")
			break
		}

		log.Println("fail to connect to the DB")
		log.Println("retry after 5 seconds")
		time.Sleep(5 * time.Second)
	}
}
