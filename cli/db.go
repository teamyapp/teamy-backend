package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/teamyapp/teamy-backend/app/config"
	"github.com/teamyapp/teamy-backend/app/repo/db"
)

const lowerCaseLetters = "abcdefghijklmnopqrstuvwxyz"
const upperCaseLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const digits = "0123456789"
const specialChars = "!@#$%^&*()-+=?[]"
const dbPasswordLen = 20

func init() {
	rand.Seed(time.Now().UnixNano())
}

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

func newDB(dbName string) {
	alphabet := concatenate([]string{
		lowerCaseLetters,
		upperCaseLetters,
		digits,
		specialChars,
	})
	dbNamePostfixAlphabet := concatenate([]string{lowerCaseLetters, upperCaseLetters, digits})
	dbNamePostfix := randString(dbNamePostfixAlphabet, 5)
	fullDBName := fmt.Sprintf("%s-%s", dbName, dbNamePostfix)
	password := randString(alphabet, dbPasswordLen)

	createDBSQL := strings.TrimSpace(fmt.Sprintf(`
CREATE DATABASE "%s";
CREATE USER "%s" WITH PASSWORD '%s';
GRANT ALL PRIVILEGES ON DATABASE "%s" TO "%s";
`,
		fullDBName,
		fullDBName,
		password,
		fullDBName,
		fullDBName,
	))
	fmt.Println(createDBSQL)
}

func concatenate(src []string) []rune {
	return []rune(strings.Join([]string{lowerCaseLetters, upperCaseLetters, digits}, ""))
}

func randString(alphabet []rune, length int) string {
	alphabetEndIndex := len(alphabet) - 1
	result := make([]rune, length)
	for i := 0; i < length; i++ {
		randomIndex := randInt(0, alphabetEndIndex)
		result[i] = alphabet[randomIndex]
	}
	return string(result)
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min+1)
}
