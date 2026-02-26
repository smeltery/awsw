package cmd

import (
	"fmt"
	"os"

	"github.com/dotbrains/awsw/internal/config"
	"github.com/spf13/cobra"
)

var forceOverwrite bool

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage awsw configuration",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a default config file",
	Long: `Creates a config file at ~/.config/awsw/config.yaml with default profiles.
Edit the file to add, remove, or modify profiles and SSO settings.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.ConfigPath()
		if err != nil {
			return err
		}

		// Check if file already exists.
		if _, err := os.Stat(path); err == nil && !forceOverwrite {
			return fmt.Errorf("config file already exists: %s\nUse --force to overwrite", path)
		}

		defaults := config.DefaultConfig()
		if err := config.SaveConfig(defaults, path); err != nil {
			return err
		}

		fmt.Printf("✓ Wrote default config to %s\n", path)
		fmt.Println("Edit the file to customize your profiles and SSO settings.")
		return nil
	},
}

func init() {
	configInitCmd.Flags().BoolVar(&forceOverwrite, "force", false, "Overwrite existing config file")
	configCmd.AddCommand(configInitCmd)
	rootCmd.AddCommand(configCmd)
}
