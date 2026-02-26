package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var forceLogin bool

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate to AWS SSO",
	Long:  "Runs 'aws sso login' against the shared SSO session. A single login grants access to all accounts/roles.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Quick check: if already authenticated, skip the browser flow.
		if !forceLogin {
			if _, err := executor.Run("aws", "sts", "get-caller-identity"); err == nil {
				fmt.Printf("✓ Already authenticated to SSO session %q\n", cfg.SSO.SessionName)
				fmt.Println("  Use --force to re-authenticate.")
				return nil
			}
		}

		fmt.Println("Opening browser for SSO authentication...")
		err := executor.RunPassthrough("aws", "sso", "login", "--sso-session", cfg.SSO.SessionName)
		if err != nil {
			return fmt.Errorf("SSO login failed: %w", err)
		}
		fmt.Printf("✓ Logged in to SSO session %q\n", cfg.SSO.SessionName)
		return nil
	},
}

func init() {
	loginCmd.Flags().BoolVar(&forceLogin, "force", false, "Force re-authentication even if already logged in")
	rootCmd.AddCommand(loginCmd)
}
