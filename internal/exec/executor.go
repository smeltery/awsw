package exec

import (
	"bytes"
	"os"
	"os/exec"
)

// CommandExecutor abstracts running external commands for testability.
type CommandExecutor interface {
	// Run executes a command and returns combined stdout output.
	Run(name string, args ...string) (string, error)
	// RunPassthrough executes a command with stdin/stdout/stderr connected to the terminal.
	RunPassthrough(name string, args ...string) error
	// RunInteractive executes a command with full TTY access (for fzf).
	RunInteractive(name string, args ...string) (string, error)
}

// RealExecutor implements CommandExecutor using os/exec.
type RealExecutor struct{}

// Run executes a command and returns its stdout.
func (r *RealExecutor) Run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return stdout.String(), err
}

// RunPassthrough executes a command with stdin/stdout/stderr connected to the terminal.
func (r *RealExecutor) RunPassthrough(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunInteractive executes a command with TTY access and captures stdout.
// Used for fzf where we need both TTY interaction and to capture the selection.
func (r *RealExecutor) RunInteractive(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	return stdout.String(), err
}
