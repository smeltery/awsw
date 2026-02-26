package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type callerIdentity struct {
	UserID  string `json:"UserId"`
	Account string `json:"Account"`
	ARN     string `json:"Arn"`
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current AWS profile and kubectl context",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		profile := os.Getenv("AWS_PROFILE")
		if profile == "" {
			profile = "(not set)"
		}
		fmt.Printf("AWS Profile:  %s\n", profile)

		// Get caller identity.
		output, err := executor.Run("aws", "sts", "get-caller-identity")
		if err != nil {
			fmt.Printf("Account:      (not authenticated)\n")
		} else {
			var identity callerIdentity
			if err := json.Unmarshal([]byte(output), &identity); err == nil {
				fmt.Printf("Account:      %s\n", identity.Account)
				// Extract role from ARN: ...assumed-role/RoleName/username
				if parts := strings.Split(identity.ARN, "/"); len(parts) >= 2 {
					role := strings.Join(parts[1:], "/")
					fmt.Printf("Role:         %s\n", role)
				}
			}
		}

		// Get current kubectl context.
		kubeCtx, err := executor.Run("kubectl", "config", "current-context")
		if err != nil {
			fmt.Printf("Kube Context: (not set)\n")
		} else {
			fmt.Printf("Kube Context: %s\n", strings.TrimSpace(kubeCtx))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
