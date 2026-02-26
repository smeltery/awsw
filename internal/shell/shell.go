package shell

import (
	"fmt"
	"strings"
)

// Shell represents a supported shell type.
type Shell string

const (
	Fish Shell = "fish"
	Bash Shell = "bash"
	Zsh  Shell = "zsh"
)

// ParseShell parses a shell name string into a Shell type.
func ParseShell(s string) (Shell, error) {
	switch strings.ToLower(s) {
	case "fish":
		return Fish, nil
	case "bash":
		return Bash, nil
	case "zsh":
		return Zsh, nil
	default:
		return "", fmt.Errorf("unsupported shell: %q (supported: fish, bash, zsh)", s)
	}
}

// ExportEnv returns the shell command to export an environment variable.
func ExportEnv(sh Shell, key, value string) string {
	switch sh {
	case Fish:
		return fmt.Sprintf("set -gx %s %s", key, value)
	default:
		return fmt.Sprintf("export %s=%s", key, value)
	}
}

// InitScript returns the shell wrapper function that users source in their rc file.
// The wrapper intercepts awsw calls, runs the binary with --shell-eval for commands
// that need env var changes, and eval's the output.
func InitScript(sh Shell, binaryName string) string {
	switch sh {
	case Fish:
		return fmt.Sprintf(`function awsw
    set -l cmd (command %s --shell-eval --shell fish $argv)
    if test $status -eq 0
        eval $cmd
    end
end`, binaryName)
	case Bash:
		return fmt.Sprintf(`awsw() {
    local cmd
    cmd=$(command %s --shell-eval --shell bash "$@")
    if [ $? -eq 0 ]; then
        eval "$cmd"
    fi
}`, binaryName)
	case Zsh:
		return fmt.Sprintf(`awsw() {
    local cmd
    cmd=$(command %s --shell-eval --shell zsh "$@")
    if [ $? -eq 0 ]; then
        eval "$cmd"
    fi
}`, binaryName)
	default:
		return ""
	}
}
