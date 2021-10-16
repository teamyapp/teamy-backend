package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var dbName string
var migrationSteps int
var migrationDir string
var migrationFileName string
var seedFilePath string

var rootCmd = &cobra.Command{
	Use: "cli",
}

var dbCmd = &cobra.Command{
	Use: "db",
}

var newDBCmd = &cobra.Command{
	Use:   "new",
	Short: "Generate SQL to create new database",
	Run: func(cmd *cobra.Command, args []string) {
		newDB(dbName)
	},
}

var migrateCmd = &cobra.Command{
	Use: "migrate",
	RunE: func(cmd *cobra.Command, args []string) error {
		return migrateDB(migrationDir, migrationSteps)
	},
}

var seedCmd = &cobra.Command{
	Use: "seed",
	RunE: func(cmd *cobra.Command, args []string) error {
		return execSQL(seedFilePath)
	},
}

var newMigrationCmd = &cobra.Command{
	Use: "new",
	RunE: func(cmd *cobra.Command, args []string) error {
		return newMigration(migrationDir, migrationFileName)
	},
}

func init() {
	newMigrationCmd.Flags().StringVarP(
		&migrationFileName,
		"fileName",
		"f",
		"",
		"name of data migration file")
	newMigrationCmd.MarkFlagRequired("fileName")
	migrateCmd.AddCommand(newMigrationCmd)

	migrateCmd.PersistentFlags().StringVarP(
		&migrationDir,
		"migrationDir",
		"d",
		"app/repo/db/migration",
		"location of DB migration files")
	migrateCmd.Flags().IntVarP(
		&migrationSteps,
		"steps",
		"s",
		0,
		"migrate up if n > 0, and down if n < 0")
	migrateCmd.MarkFlagRequired("steps")
	dbCmd.AddCommand(migrateCmd)

	seedCmd.Flags().StringVarP(
		&seedFilePath,
		"file",
		"f",
		"app/repo/db/seed.sql",
		"location of DB seed SQL")
	dbCmd.AddCommand(seedCmd)

	newDBCmd.Flags().StringVarP(
		&dbName,
		"name",
		"n",
		"",
		"name of new DB")
	newDBCmd.MarkFlagRequired("name")
	dbCmd.AddCommand(newDBCmd)

	rootCmd.AddCommand(dbCmd)
}

func execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
