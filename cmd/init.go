package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dotbrains/awsw/internal/shell"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init <shell>",
	Short: "Output shell wrapper function for eval",
	Long: `Output a shell function that wraps the awsw binary.
Add this to your shell's rc file:

  fish:  awsw init fish | source
  bash:  eval "$(awsw init bash)"
  zsh:   eval "$(awsw init zsh)"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sh, err := shell.ParseShell(args[0])
		if err != nil {
			return err
		}

		// Resolve the binary path.
		binaryPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("could not determine binary path: %w", err)
		}
		binaryPath, err = filepath.EvalSymlinks(binaryPath)
		if err != nil {
			return fmt.Errorf("could not resolve symlinks: %w", err)
		}

		fmt.Println(shell.InitScript(sh, binaryPath))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
