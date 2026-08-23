package cmd

import (
	"fmt"
	"os"

	"github.com/smeltery/awsw/internal/config"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Generate AWS config and configure EKS cluster contexts",
	Long: `First-time (or refresh) setup:
1. Writes/updates ~/.aws/config with SSO profiles.
2. Configures ~/.kube/config with EKS cluster contexts defined in your config.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("could not determine home directory: %w", err)
		}

		// Step 1: Generate AWS config.
		if err := config.WriteAWSConfig(cfg, homeDir); err != nil {
			return fmt.Errorf("writing AWS config: %w", err)
		}

		// Step 2: Configure EKS cluster contexts.
		eksProfiles := getEKSProfiles()
		if len(eksProfiles) == 0 {
			return nil
		}

		fmt.Println("✓ Configuring EKS clusters...")
		contextCount := 0

		for _, p := range eksProfiles {
			for _, cluster := range p.EKSClusters {
				fmt.Printf("  %s: configuring cluster %q in %s\n", p.Alias, cluster.Name, cluster.Region)

				err := executor.RunPassthrough("aws", "eks", "update-kubeconfig",
					"--name", cluster.Name,
					"--region", cluster.Region,
					"--profile", p.Alias,
					"--alias", cluster.ContextAlias,
				)
				if err != nil {
					fmt.Fprintf(os.Stderr, "  ⚠ %s: could not update kubeconfig for %s: %v\n", p.Alias, cluster.Name, err)
					continue
				}
				contextCount++
			}
		}

		if contextCount > 0 {
			fmt.Printf("✓ Updated ~/.kube/config with %d contexts\n", contextCount)
		}

		return nil
	},
}

func getEKSProfiles() []config.Profile {
	var profiles []config.Profile
	for _, p := range cfg.Profiles {
		if p.HasEKS() {
			profiles = append(profiles, p)
		}
	}
	return profiles
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
