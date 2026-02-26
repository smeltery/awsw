package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/dotbrains/awsw/internal/config"
	awswexec "github.com/dotbrains/awsw/internal/exec"
	"github.com/dotbrains/awsw/internal/shell"
	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	shellEval bool
	shellName string
	executor  awswexec.CommandExecutor = &awswexec.RealExecutor{}
	cfg       *config.Config
)

// SetVersion sets the version string from main.
func SetVersion(v string) {
	version = v
}

// SetExecutor replaces the command executor (for testing).
func SetExecutor(e awswexec.CommandExecutor) {
	executor = e
}

// SetConfig replaces the loaded config (for testing).
func SetConfig(c *config.Config) {
	cfg = c
}

var rootCmd = &cobra.Command{
	Use:   "awsw [alias]",
	Short: "Seamlessly switch between AWS SSO accounts and EKS contexts",
	Long: `awsw — AWS Account & EKS Context Switcher

A lightweight CLI for seamlessly switching between AWS SSO accounts
and EKS cluster contexts.`,
	Version:      version,
	SilenceUsage: true,
	Args:         cobra.MaximumNArgs(1),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cfg != nil {
			return nil // already loaded (e.g. tests)
		}
		var err error
		cfg, err = config.LoadConfig()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var profile *config.Profile
		var cluster *config.EKSCluster

		if len(args) == 1 {
			// Direct alias mode: try profile alias first, then cluster name.
			profile = cfg.FindProfile(args[0])
			if profile == nil {
				// Fallback: resolve by cluster name.
				match := cfg.FindByCluster(args[0])
				if match != nil {
					profile = match.Profile
					cluster = match.Cluster
				} else {
					msg := fmt.Sprintf("unknown profile or cluster: %q\nAvailable profiles: %s",
						args[0], strings.Join(cfg.ProfileAliases(), ", "))
					if clusters := cfg.ClusterAliases(); len(clusters) > 0 {
						msg += fmt.Sprintf("\nAvailable clusters: %s", strings.Join(clusters, ", "))
					}
					return fmt.Errorf("%s", msg)
				}
			}
		} else {
			// Interactive mode via fzf.
			selectedProfile, selectedCluster, err := selectProfileInteractive()
			if err != nil {
				return err
			}
			profile = selectedProfile
			cluster = selectedCluster
		}

		return switchTo(profile, cluster)
	},
}

// fzfEntry ties an fzf display line back to a profile and optional cluster.
type fzfEntry struct {
	label   string
	profile *config.Profile
	cluster *config.EKSCluster
}

// buildFzfEntries returns the list of selectable entries for fzf.
// Profiles with EKS clusters get one line per cluster (keyed by context alias);
// profiles without EKS get a single line keyed by profile alias.
func buildFzfEntries() []fzfEntry {
	var entries []fzfEntry
	for i := range cfg.Profiles {
		p := &cfg.Profiles[i]
		if p.HasEKS() {
			for j := range p.EKSClusters {
				cl := &p.EKSClusters[j]
				label := cl.ContextAlias + "\t" + p.AccountName + " (" + p.AccountID + ") " + p.RoleName
				entries = append(entries, fzfEntry{label: label, profile: p, cluster: cl})
			}
		} else {
			entries = append(entries, fzfEntry{label: p.FzfLabel(), profile: p})
		}
	}
	return entries
}

// buildFzfInput returns the newline-delimited list of labels for fzf.
func buildFzfInput() string {
	entries := buildFzfEntries()
	lines := make([]string, len(entries))
	for i, e := range entries {
		lines[i] = e.label
	}
	return strings.Join(lines, "\n")
}

func selectProfileInteractive() (*config.Profile, *config.EKSCluster, error) {
	input := buildFzfInput()
	entries := buildFzfEntries()

	// Pipe profile list into fzf.
	output, err := executor.RunInteractive("fzf", "--height=10", "--reverse", "--prompt=AWS Profile> ", "--input", "/dev/stdin")
	if err != nil {
		// Try alternative: use echo + pipe approach via sh.
		output, err = executor.Run("sh", "-c", fmt.Sprintf("echo '%s' | fzf --height=10 --reverse --prompt='AWS Profile> '", input))
		if err != nil {
			return nil, nil, fmt.Errorf("fzf selection cancelled or failed: %w", err)
		}
	}

	selected := strings.Fields(strings.TrimSpace(output))[0]

	// Match the selection back to an entry.
	for _, e := range entries {
		key := strings.Fields(e.label)[0]
		if key == selected {
			return e.profile, e.cluster, nil
		}
	}
	return nil, nil, fmt.Errorf("could not resolve selection: %q", output)
}

// switchTo sets the AWS profile and optionally switches kubectl context.
// If cluster is non-nil it targets that specific cluster; otherwise it falls
// back to the first cluster in the profile (preserving existing behaviour).
func switchTo(p *config.Profile, cluster *config.EKSCluster) error {
	// Determine which cluster context to use.
	eksCtx := ""
	if cluster != nil {
		eksCtx = cluster.ContextAlias
	} else if p.HasEKS() {
		eksCtx = p.EKSClusters[0].ContextAlias
	}

	if shellEval {
		// Shell eval mode: output commands for the wrapper to eval.
		sh, err := shell.ParseShell(shellName)
		if err != nil {
			return err
		}
		fmt.Println(shell.ExportEnv(sh, "AWS_PROFILE", p.Alias))
		if eksCtx != "" {
			fmt.Printf("kubectl config use-context %s >/dev/null 2>&1\n", eksCtx)
		}
		return nil
	}

	// Direct execution mode (no shell wrapper).
	fmt.Printf("→ AWS_PROFILE=%s\n", p.Alias)
	if err := os.Setenv("AWS_PROFILE", p.Alias); err != nil {
		return err
	}

	if eksCtx != "" {
		_, err := executor.Run("kubectl", "config", "use-context", eksCtx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠ Could not switch kubectl context to %s: %v\n", eksCtx, err)
		} else {
			fmt.Printf("→ kubectl context: %s\n", eksCtx)
		}
	}

	fmt.Println("Ready.")
	return nil
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&shellEval, "shell-eval", false, "Output shell commands for eval (used by shell wrapper)")
	rootCmd.PersistentFlags().StringVar(&shellName, "shell", "", "Target shell for --shell-eval (fish, bash, zsh)")

	// Hide internal flags from help.
	_ = rootCmd.PersistentFlags().MarkHidden("shell-eval")
	_ = rootCmd.PersistentFlags().MarkHidden("shell")
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
