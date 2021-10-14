package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var migrationSteps int
var migrationDir string

var seedFilePath string

var rootCmd = &cobra.Command{
	Use: "cli",
}

var dbCmd = &cobra.Command{
	Use: "db",
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

func init() {
	migrateCmd.Flags().StringVarP(
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

	rootCmd.AddCommand(dbCmd)
}

func execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
