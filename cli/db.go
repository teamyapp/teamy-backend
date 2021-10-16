package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/teamyapp/teamy-backend/app/config"
	"github.com/teamyapp/teamy-backend/app/repo/db"
)

func migrateDB(migrationDir string, steps int) error {
	return withDB(func(sqlDB *sql.DB) error {
		migrationURI := fmt.Sprintf("file://%s", migrationDir)

		driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
		if err != nil {
			return err
		}

		m, err := migrate.NewWithDatabaseInstance(
			migrationURI,
			"postgres", driver)
		if err != nil {
			return err
		}

		err = m.Steps(steps)
		if errors.Is(err, os.ErrNotExist) {
			log.Println("no change in DB")
			return nil
		}
		if err != nil {
			return err
		}

		log.Println("finish DB migration")
		return nil
	})
}

func execSQL(sqlFileName string) error {
	return withDB(func(sqlDB *sql.DB) error {
		buf, err := ioutil.ReadFile(sqlFileName)
		if err != nil {
			return err
		}
		_, err = sqlDB.Exec(string(buf))
		if err == nil {
			log.Println("successfully seeded DB")
		}
		return err
	})
}

func withDB(action func(sqlDB *sql.DB) error) error {
	cfg, err := config.FromEnv()
	if err != nil {
		log.Println(err)
		return err
	}
	return db.WithDB(cfg, action)
}

func newMigration(migrationDir string, fileName string) error {
	now := time.Now()
	prefix := fmt.Sprintf(
		"%04d%02d%02d%02d%02d%02d_%s",
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		now.Second(),
		fileName)
	fileNameFormats := []string{
		"%s.up.sql",
		"%s.down.sql",
	}

	for _, fileNameFormat := range fileNameFormats {
		fileName = fmt.Sprintf(fileNameFormat, prefix)
		filePath := filepath.Join(migrationDir, fileName)
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}
		file.Close()
	}

	return nil
}
