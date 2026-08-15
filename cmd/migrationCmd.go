package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(migrateCmd)
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Manage Database Migrations",
	Long:  "Manage Database Migrations for Backend Portal",
	Run:   executeMigration,
}

func executeMigration(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		panic("missing argument: ./{bin-file} [goose-command]")
	}

	// Init config
	cfg, secret, err := config.LoadConfig(cfgFile, scrtFile)
	if err != nil {
		log.Fatalf("Unable to load configuration and secret: %v", err)
	}

	// Build DSN Connection
	dbConnection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local&multiStatements=true",
		secret.MySQLSecret.Service.Username,
		secret.MySQLSecret.Service.Password,
		cfg.MySQLConfig.Host,
		cfg.MySQLConfig.Port,
		secret.MySQLSecret.Service.Database)

	// Init Variable
	migrationDir := migrationFile
	gooseCommand := args[0]

	// Check db connection
	db, err := goose.OpenDBWithDriver(cfg.MySQLConfig.Dialect, dbConnection)
	if err != nil {
		panic(fmt.Sprintf("goose: failed to open DB: %v\n", err))
	}

	defer func() {
		if err := db.Close(); err != nil {
			panic(fmt.Sprintf("goose: failed to close DB: %v\n", err))
		}
	}()

	// Make spesify CREATE new migration file
	if gooseCommand == "create" && len(args) >= 3 {
		migrationType := args[len(args)-1] // 'sql' or 'go'
		migrationName := args[len(args)-2] // the name of the migration
		args = append(args[:len(args)-2], migrationName, migrationType)
		args = args[1:]
	}

	// executing actual goose
	if err := goose.RunWithOptionsContext(context.Background(), gooseCommand, db, migrationDir, args, goose.WithAllowMissing()); err != nil {
		log.Fatalf("goose %v: %v", gooseCommand, err)
	}
}
