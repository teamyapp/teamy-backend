package sqldb

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/teamyapp/cloud/libs/io"
)

const dbType = "postgres"

const DefaultMigrationRoot = "app/dao/sqldb/migrations"
const DefaultSeedFile = "app/dao/sqldb/seed.sql"
const MigrateAll = 0

const lowerCaseLetters = "abcdefghijklmnopqrstuvwxyz"
const upperCaseLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const digits = "0123456789"
const specialChars = "!@#$%^&*()-+=?[]"
const dbPasswordLen = 20

type Config struct {
	DBHost     string `envconfig:"DB_HOST" default:"localhost"`
	DBPort     int    `envconfig:"DB_PORT" default:"5432"`
	DBUser     string `envconfig:"DB_USER"`
	DBName     string `envconfig:"DB_NAME" default:"teamy"`
	DBPassword string `envconfig:"DB_PASSWORD"`
	DBSSLMode  string `envconfig:"DB_SSL_MODE" default:"require"`
}

func Use(cfg Config, action func(sqlDB *sql.DB) error) error {
	sqlDB, err := connect(cfg)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	waitUntilReady(sqlDB)
	return action(sqlDB)
}

func MigrateUp(sqlDB *sql.DB, migrationRoot string, steps int) error {
	return migrateDB(sqlDB, migrationRoot, migrate.Up, steps)
}

func MigrateDown(sqlDB *sql.DB, migrationRoot string, steps int) error {
	return migrateDB(sqlDB, migrationRoot, migrate.Down, steps)
}

func NewMigration(migrationDir string, fileName string) (string, error) {
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

	err := os.MkdirAll(migrationDir, os.ModePerm)
	if err != nil {
		return "", err
	}

	fileName = fmt.Sprintf("%s.sql", prefix)
	fullFilePath := filepath.Join(migrationDir, fileName)
	err = io.CreateFileWithLog(fullFilePath)
	if err != nil {
		return "", err
	}

	return fullFilePath, nil
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

func ExecSQL(sqlDB *sql.DB, sqlFileName string) error {
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
}

func waitUntilReady(sqlDB *sql.DB) {
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

func connect(cfg Config) (*sql.DB, error) {
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

func migrateDB(
	db *sql.DB,
	migrationRoot string,
	migrateDirection migrate.MigrationDirection,
	steps int,
) error {
	migrations := &migrate.FileMigrationSource{
		Dir: migrationRoot,
	}
	_, err := migrate.ExecMax(db, dbType, migrations, migrateDirection, steps)
	if err == nil {
		log.Println("migration finished")
	}
	return err
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
