package db

import (
	"context"
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
	"github.com/teamyapp/one/config"
	"github.com/teamyapp/one/io"
)

const lowerCaseLetters = "abcdefghijklmnopqrstuvwxyz"
const upperCaseLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const digits = "0123456789"
const specialChars = "!@#$%^&*()-+=?[]"
const dbPasswordLen = 20

func init() {
	rand.Seed(time.Now().UnixNano())
}

const dbType = "postgres"

func Connect(cfg config.Config) (*sql.DB, error) {
	dbSource := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBSSLMode)
	return sql.Open(dbType, dbSource)
}

func WaitUntilReady(sqlDB *sql.DB) {
	for {
		err := sqlDB.Ping()
		if err == nil {
			log.Println("successfully connected to the DB")
			break
		}

		log.Println(err)
		log.Println("fail to connect to the DB")
		log.Println("retry after 5 seconds")
		time.Sleep(5 * time.Second)
	}
}

func With(cfg config.Config, action func(sqlDB *sql.DB) error) error {
	sqlDB, err := Connect(cfg)
	if err != nil {
		return err
	}

	WaitUntilReady(sqlDB)

	defer sqlDB.Close()
	return action(sqlDB)
}

func DefaultWith(action func(sqlDB *sql.DB) error) error {
	cfg, err := config.OneConfigFromEnv()
	if err != nil {
		log.Println(err)
		return err
	}
	return With(cfg, action)
}

func Migrate(migrationDir string, steps int) error {
	return DefaultWith(func(sqlDB *sql.DB) error {
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

func ExecSQL(sqlFileName string) error {
	return DefaultWith(func(sqlDB *sql.DB) error {
		buf, err := ioutil.ReadFile(sqlFileName)
		if err != nil {
			return err
		}

		tx, err := sqlDB.BeginTx(context.Background(), nil)
		if err != nil {
			return err
		}

		_, err = tx.Exec(string(buf))
		if err != nil {
			return err
		}

		err = tx.Commit()
		if err == nil {
			log.Println("successfully seeded DB")
		}
		return err
	})
}

func NewMigration(migrationDir string, fileName string) error {
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

	err := os.MkdirAll(migrationDir, os.ModePerm)
	if err != nil {
		return err
	}

	for _, fileNameFormat := range fileNameFormats {
		fileName = fmt.Sprintf(fileNameFormat, prefix)
		err = io.CreateFileWithLog(filepath.Join(migrationDir, fileName))
		if err != nil {
			return err
		}
	}

	return nil
}

func New(dbName string) {
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

	fmt.Println(strings.TrimSpace(fmt.Sprintf(`
user: %s
password: %s
dbName: %s
SQL:
================================================================================
CREATE DATABASE "%s";
CREATE USER "%s" WITH PASSWORD '%s';
GRANT ALL PRIVILEGES ON DATABASE "%s" TO "%s";
================================================================================
`,
		fullDBName,
		password,
		fullDBName,
		fullDBName,
		fullDBName,
		password,
		fullDBName,
		fullDBName,
	)))
}

func concatenate(src []string) []rune {
	return []rune(strings.Join(src, ""))
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
