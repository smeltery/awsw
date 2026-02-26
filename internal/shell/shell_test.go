package shell

import (
	"strings"
	"testing"
)

func TestParseShell_Valid(t *testing.T) {
	tests := []struct {
		input string
		want  Shell
	}{
		{"fish", Fish},
		{"bash", Bash},
		{"zsh", Zsh},
		{"FISH", Fish},
		{"Bash", Bash},
		{"ZSH", Zsh},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseShell(tt.input)
			if err != nil {
				t.Fatalf("ParseShell(%q) error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseShell(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseShell_Invalid(t *testing.T) {
	_, err := ParseShell("powershell")
	if err == nil {
		t.Error("ParseShell(\"powershell\") should return error")
	}
	if !strings.Contains(err.Error(), "unsupported shell") {
		t.Errorf("error = %q, want to contain \"unsupported shell\"", err.Error())
	}
}

func TestExportEnv_Fish(t *testing.T) {
	got := ExportEnv(Fish, "AWS_PROFILE", "dev-eks")
	want := "set -gx AWS_PROFILE dev-eks"
	if got != want {
		t.Errorf("ExportEnv(Fish) = %q, want %q", got, want)
	}
}

func TestExportEnv_Bash(t *testing.T) {
	got := ExportEnv(Bash, "AWS_PROFILE", "dev-eks")
	want := "export AWS_PROFILE=dev-eks"
	if got != want {
		t.Errorf("ExportEnv(Bash) = %q, want %q", got, want)
	}
}

func TestExportEnv_Zsh(t *testing.T) {
	got := ExportEnv(Zsh, "AWS_PROFILE", "dev-eks")
	want := "export AWS_PROFILE=dev-eks"
	if got != want {
		t.Errorf("ExportEnv(Zsh) = %q, want %q", got, want)
	}
}

func TestInitScript_Fish(t *testing.T) {
	script := InitScript(Fish, "/usr/local/bin/awsw")
	if !strings.Contains(script, "function awsw") {
		t.Error("fish init script missing 'function awsw'")
	}
	if !strings.Contains(script, "--shell-eval") {
		t.Error("fish init script missing '--shell-eval'")
	}
	if !strings.Contains(script, "--shell fish") {
		t.Error("fish init script missing '--shell fish'")
	}
	if !strings.Contains(script, "/usr/local/bin/awsw") {
		t.Error("fish init script missing binary path")
	}
}

func TestInitScript_Bash(t *testing.T) {
	script := InitScript(Bash, "/usr/local/bin/awsw")
	if !strings.Contains(script, "awsw()") {
		t.Error("bash init script missing 'awsw()'")
	}
	if !strings.Contains(script, "--shell bash") {
		t.Error("bash init script missing '--shell bash'")
	}
	if !strings.Contains(script, "eval") {
		t.Error("bash init script missing 'eval'")
	}
}

func TestInitScript_Zsh(t *testing.T) {
	script := InitScript(Zsh, "/usr/local/bin/awsw")
	if !strings.Contains(script, "awsw()") {
		t.Error("zsh init script missing 'awsw()'")
	}
	if !strings.Contains(script, "--shell zsh") {
		t.Error("zsh init script missing '--shell zsh'")
	}
}

func TestInitScript_UnsupportedShell(t *testing.T) {
	script := InitScript(Shell("tcsh"), "/usr/local/bin/awsw")
	if script != "" {
		t.Errorf("InitScript for unsupported shell should return empty, got %q", script)
	}
}
