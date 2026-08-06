package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags.
	cfgFile       string
	scrtFile      string
	migrationFile string
	cronJob       string
	command       string
	date          string
	startDate     string
	endDate       string

	rootCmd = &cobra.Command{
		Use:   "cobra-cli",
		Short: "A generator for Cobra based Applications",
		Long: `Cobra is a CLI library for Go that empowers applications.
			This application is a tool to generate the needed files
			to quickly create a Cobra application.`,
	}
)

// Execute executes the Go function.
//
// It has no parameters and returns an error.
func Execute() error {
	return rootCmd.Execute()
}

// init initializes the application configuration.
//
// No parameters.
// No return type.
func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config.yaml)")
	rootCmd.PersistentFlags().StringVar(&scrtFile, "secret", "", "secret file (default is $HOME/.secret.yaml)")
	rootCmd.PersistentFlags().StringVar(&migrationFile, "migration", "", "migration file (default is $HOME/migration/*.sql)")
	rootCmd.PersistentFlags().StringVar(&cronJob, "cronJob", "", "cronJob (example like calculate-all-merchant-eod-balance)")
	rootCmd.PersistentFlags().StringVar(&date, "date", "", "Date in YYYY-mm-dd HH:MM:SS format")
	rootCmd.PersistentFlags().StringP("author", "a", "Repo author", "author name for copyright attribution")
	rootCmd.PersistentFlags().Bool("viper", true, "use Viper for configuration")
	if err := viper.BindPFlag("author", rootCmd.PersistentFlags().Lookup("author")); err != nil {
		cobra.CheckErr(err)
	}
	if err := viper.BindPFlag("useViper", rootCmd.PersistentFlags().Lookup("viper")); err != nil {
		cobra.CheckErr(err)
	}
	viper.SetDefault("author", "Repository Author <repo_author@email.id>")
}

// initConfig initializes the configuration for the application.
//
// No parameters.
func initConfig() {
	// Find current working directory (project root).
	cwd, err := os.Getwd()
	cobra.CheckErr(err)

	if cfgFile == "" {
		cfgFile = filepath.Join(cwd, ".config.yaml")
	}

	if scrtFile == "" {
		scrtFile = filepath.Join(cwd, ".secret.yaml")
	}

	if migrationFile == "" {
		migrationFile = filepath.Join(cwd, "/migration/")
	}
}
