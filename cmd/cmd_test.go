package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dotbrains/awsw/internal/config"
)

func init() {
	// Ensure cfg is set for all tests.
	cfg = config.DefaultConfig()
}

// MockExecutor implements exec.CommandExecutor for testing.
type MockExecutor struct {
	RunFunc            func(name string, args ...string) (string, error)
	RunPassthroughFunc func(name string, args ...string) error
	RunInteractiveFunc func(name string, args ...string) (string, error)
}

func (m *MockExecutor) Run(name string, args ...string) (string, error) {
	if m.RunFunc != nil {
		return m.RunFunc(name, args...)
	}
	return "", nil
}

func (m *MockExecutor) RunPassthrough(name string, args ...string) error {
	if m.RunPassthroughFunc != nil {
		return m.RunPassthroughFunc(name, args...)
	}
	return nil
}

func (m *MockExecutor) RunInteractive(name string, args ...string) (string, error) {
	if m.RunInteractiveFunc != nil {
		return m.RunInteractiveFunc(name, args...)
	}
	return "", nil
}

func TestGetEKSProfiles(t *testing.T) {
	profiles := getEKSProfiles()
	for _, p := range profiles {
		if !p.HasEKS() {
			t.Errorf("getEKSProfiles() returned non-EKS profile: %s", p.Alias)
		}
	}
	if len(profiles) != 2 {
		t.Errorf("getEKSProfiles() returned %d profiles, want 2", len(profiles))
	}
}

func TestSwitchTo_ShellEval_Fish(t *testing.T) {
	// Set shell-eval mode.
	shellEval = true
	shellName = "fish"
	defer func() {
		shellEval = false
		shellName = ""
	}()

	profile := findTestProfile("dev-eks")

	// Capture stdout.
	output := captureStdout(t, func() {
		err := switchTo(profile, nil)
		if err != nil {
			t.Fatalf("switchTo() error: %v", err)
		}
	})

	if !strings.Contains(output, "set -gx AWS_PROFILE dev-eks") {
		t.Errorf("output missing fish export, got: %q", output)
	}
	if !strings.Contains(output, "kubectl config use-context my-cluster-dev") {
		t.Errorf("output missing kubectl context switch, got: %q", output)
	}
}

func TestSwitchTo_ShellEval_Bash(t *testing.T) {
	shellEval = true
	shellName = "bash"
	defer func() {
		shellEval = false
		shellName = ""
	}()

	profile := findTestProfile("dev-eks")

	output := captureStdout(t, func() {
		err := switchTo(profile, nil)
		if err != nil {
			t.Fatalf("switchTo() error: %v", err)
		}
	})

	if !strings.Contains(output, "export AWS_PROFILE=dev-eks") {
		t.Errorf("output missing bash export, got: %q", output)
	}
}

func TestSwitchTo_ShellEval_NoEKS(t *testing.T) {
	shellEval = true
	shellName = "fish"
	defer func() {
		shellEval = false
		shellName = ""
	}()

	profile := findTestProfile("dev-ro")

	output := captureStdout(t, func() {
		err := switchTo(profile, nil)
		if err != nil {
			t.Fatalf("switchTo() error: %v", err)
		}
	})

	if !strings.Contains(output, "set -gx AWS_PROFILE dev-ro") {
		t.Errorf("output missing fish export, got: %q", output)
	}
	if strings.Contains(output, "kubectl") {
		t.Errorf("output should NOT contain kubectl for non-EKS profile, got: %q", output)
	}
}

func TestSwitchTo_DirectMode(t *testing.T) {
	mock := &MockExecutor{
		RunFunc: func(name string, args ...string) (string, error) {
			if name == "kubectl" {
				return "", nil
			}
			return "", nil
		},
	}
	oldExecutor := executor
	executor = mock
	defer func() { executor = oldExecutor }()

	shellEval = false
	profile := findTestProfile("dev-eks")

	output := captureStdout(t, func() {
		err := switchTo(profile, nil)
		if err != nil {
			t.Fatalf("switchTo() error: %v", err)
		}
	})

	if !strings.Contains(output, "→ AWS_PROFILE=dev-eks") {
		t.Errorf("output missing direct mode output, got: %q", output)
	}
	if !strings.Contains(output, "Ready.") {
		t.Errorf("output missing Ready., got: %q", output)
	}
}

func TestSwitchTo_InvalidShell(t *testing.T) {
	shellEval = true
	shellName = "powershell"
	defer func() {
		shellEval = false
		shellName = ""
	}()

	profile := findTestProfile("dev-eks")
	err := switchTo(profile, nil)
	if err == nil {
		t.Error("switchTo() should return error for unsupported shell")
	}
}


func TestStatusCmd_Authenticated(t *testing.T) {
	mock := &MockExecutor{
		RunFunc: func(name string, args ...string) (string, error) {
			if name == "aws" {
				return `{"UserId":"AROA:nick","Account":"111111111111","Arn":"arn:aws:sts::111111111111:assumed-role/eng-eks/nick"}`, nil
			}
			if name == "kubectl" {
				return "dev-eks\n", nil
			}
			return "", nil
		},
	}
	oldExecutor := executor
	executor = mock
	defer func() { executor = oldExecutor }()

	t.Setenv("AWS_PROFILE", "dev-eks")

	output := captureStdout(t, func() {
		err := statusCmd.RunE(statusCmd, nil)
		if err != nil {
			t.Fatalf("status error: %v", err)
		}
	})

	if !strings.Contains(output, "AWS Profile:  dev-eks") {
		t.Errorf("missing profile, got: %q", output)
	}
	if !strings.Contains(output, "Account:      111111111111") {
		t.Errorf("missing account, got: %q", output)
	}
	if !strings.Contains(output, "Kube Context: dev-eks") {
		t.Errorf("missing kube context, got: %q", output)
	}
}

func TestStatusCmd_NotAuthenticated(t *testing.T) {
	mock := &MockExecutor{
		RunFunc: func(name string, args ...string) (string, error) {
			return "", fmt.Errorf("not logged in")
		},
	}
	oldExecutor := executor
	executor = mock
	defer func() { executor = oldExecutor }()

	t.Setenv("AWS_PROFILE", "")

	output := captureStdout(t, func() {
		err := statusCmd.RunE(statusCmd, nil)
		if err != nil {
			t.Fatalf("status error: %v", err)
		}
	})

	if !strings.Contains(output, "(not set)") {
		t.Errorf("missing (not set), got: %q", output)
	}
	if !strings.Contains(output, "(not authenticated)") {
		t.Errorf("missing (not authenticated), got: %q", output)
	}
}

func TestLoginCmd_Success(t *testing.T) {
	called := false
	mock := &MockExecutor{
		RunFunc: func(name string, args ...string) (string, error) {
			// STS check fails → not authenticated.
			return "", fmt.Errorf("not logged in")
		},
		RunPassthroughFunc: func(name string, args ...string) error {
			called = true
			if name != "aws" {
				t.Errorf("expected aws command, got %q", name)
			}
			expectedArgs := []string{"sso", "login", "--sso-session", "default"}
			for i, a := range expectedArgs {
				if i >= len(args) || args[i] != a {
					t.Errorf("arg[%d] = %q, want %q", i, args[i], a)
				}
			}
			return nil
		},
	}
	oldExecutor := executor
	executor = mock
	defer func() { executor = oldExecutor }()

	output := captureStdout(t, func() {
		err := loginCmd.RunE(loginCmd, nil)
		if err != nil {
			t.Fatalf("login error: %v", err)
		}
	})

	if !called {
		t.Error("RunPassthrough was not called")
	}
	if !strings.Contains(output, "Logged in") {
		t.Errorf("missing success message, got: %q", output)
	}
}

func TestLoginCmd_Failure(t *testing.T) {
	mock := &MockExecutor{
		RunFunc: func(name string, args ...string) (string, error) {
			return "", fmt.Errorf("not logged in")
		},
		RunPassthroughFunc: func(name string, args ...string) error {
			return fmt.Errorf("browser error")
		},
	}
	oldExecutor := executor
	executor = mock
	defer func() { executor = oldExecutor }()

	_ = captureStdout(t, func() {
		err := loginCmd.RunE(loginCmd, nil)
		if err == nil {
			t.Error("login should return error on failure")
		}
		if !strings.Contains(err.Error(), "SSO login failed") {
			t.Errorf("error = %q, want to contain 'SSO login failed'", err.Error())
		}
	})
}

func TestRootCmd_InvalidAlias(t *testing.T) {
	err := rootCmd.RunE(rootCmd, []string{"nonexistent"})
	if err == nil {
		t.Error("root command should return error for invalid alias")
	}
	if !strings.Contains(err.Error(), "unknown profile or cluster") {
		t.Errorf("error = %q, want to contain 'unknown profile or cluster'", err.Error())
	}
	// Should include cluster names in the hint.
	if !strings.Contains(err.Error(), "Available clusters") {
		t.Errorf("error = %q, want to contain 'Available clusters'", err.Error())
	}
}

func TestRootCmd_ValidAlias(t *testing.T) {
	mock := &MockExecutor{
		RunFunc: func(name string, args ...string) (string, error) {
			return "", nil
		},
	}
	oldExecutor := executor
	executor = mock
	defer func() { executor = oldExecutor }()

	shellEval = false

	output := captureStdout(t, func() {
		err := rootCmd.RunE(rootCmd, []string{"dev-ro"})
		if err != nil {
			t.Fatalf("root error: %v", err)
		}
	})

	if !strings.Contains(output, "→ AWS_PROFILE=dev-ro") {
		t.Errorf("output missing profile switch, got: %q", output)
	}
}

func TestSwitchTo_DirectMode_KubectlError(t *testing.T) {
	mock := &MockExecutor{
		RunFunc: func(name string, args ...string) (string, error) {
			if name == "kubectl" {
				return "", fmt.Errorf("context not found")
			}
			return "", nil
		},
	}
	oldExecutor := executor
	executor = mock
	defer func() { executor = oldExecutor }()

	shellEval = false
	profile := findTestProfile("dev-eks")

	output := captureStdout(t, func() {
		err := switchTo(profile, nil)
		if err != nil {
			t.Fatalf("switchTo() should not return error on kubectl failure, got: %v", err)
		}
	})

	if !strings.Contains(output, "→ AWS_PROFILE=dev-eks") {
		t.Errorf("output missing profile switch, got: %q", output)
	}
}

func TestBuildFzfInput(t *testing.T) {
	input := buildFzfInput()
	lines := strings.Split(input, "\n")
	// 2 non-EKS profiles + 1 cluster per EKS profile (2 EKS profiles) = 4 lines
	if len(lines) != 4 {
		t.Fatalf("buildFzfInput() returned %d lines, want 4", len(lines))
	}
	for _, line := range lines {
		if !strings.Contains(line, "\t") {
			t.Errorf("line missing tab separator: %q", line)
		}
	}
	// First field of each line should be a valid alias or cluster context alias.
	for _, line := range lines {
		key := strings.Fields(line)[0]
		if cfg.FindProfile(key) == nil && cfg.FindByCluster(key) == nil {
			t.Errorf("fzf input contains unresolvable key: %q", key)
		}
	}
}

func TestBuildFzfEntries_ClusterPerLine(t *testing.T) {
	entries := buildFzfEntries()
	// EKS profiles should produce cluster-level entries.
	var clusterEntries int
	for _, e := range entries {
		if e.cluster != nil {
			clusterEntries++
			key := strings.Fields(e.label)[0]
			if key != e.cluster.ContextAlias {
				t.Errorf("cluster entry key = %q, want %q", key, e.cluster.ContextAlias)
			}
		}
	}
	if clusterEntries != 2 {
		t.Errorf("expected 2 cluster entries, got %d", clusterEntries)
	}
}

func TestInitCmd_Fish(t *testing.T) {
	output := captureStdout(t, func() {
		err := initCmd.RunE(initCmd, []string{"fish"})
		if err != nil {
			t.Fatalf("init fish error: %v", err)
		}
	})
	if !strings.Contains(output, "function awsw") {
		t.Errorf("init fish output missing function definition, got: %q", output)
	}
	if !strings.Contains(output, "--shell-eval") {
		t.Errorf("init fish output missing --shell-eval, got: %q", output)
	}
}

func TestInitCmd_Bash(t *testing.T) {
	output := captureStdout(t, func() {
		err := initCmd.RunE(initCmd, []string{"bash"})
		if err != nil {
			t.Fatalf("init bash error: %v", err)
		}
	})
	if !strings.Contains(output, "awsw()") {
		t.Errorf("init bash output missing function definition, got: %q", output)
	}
}

func TestInitCmd_InvalidShell(t *testing.T) {
	err := initCmd.RunE(initCmd, []string{"tcsh"})
	if err == nil {
		t.Error("init should return error for unsupported shell")
	}
}

func TestSetupCmd_UpdateKubeconfigSuccess(t *testing.T) {
	var capturedArgs [][]string
	mock := &MockExecutor{
		RunPassthroughFunc: func(name string, args ...string) error {
			if name == "aws" {
				capturedArgs = append(capturedArgs, args)
			}
			return nil
		},
	}
	oldExecutor := executor
	executor = mock
	defer func() { executor = oldExecutor }()

	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	output := captureStdout(t, func() {
		err := setupCmd.RunE(setupCmd, nil)
		if err != nil {
			t.Fatalf("setup error: %v", err)
		}
	})

	if !strings.Contains(output, "Configuring EKS clusters") {
		t.Errorf("missing configuring message, got: %q", output)
	}
	if !strings.Contains(output, "Updated ~/.kube/config") {
		t.Errorf("missing update message, got: %q", output)
	}

	// Verify eks update-kubeconfig was called with correct args.
	if len(capturedArgs) != 2 {
		t.Fatalf("expected 2 eks update-kubeconfig calls, got %d", len(capturedArgs))
	}
}

func TestSetupCmd_UpdateKubeconfigError(t *testing.T) {
	mock := &MockExecutor{
		RunPassthroughFunc: func(name string, args ...string) error {
			for _, a := range args {
				if a == "update-kubeconfig" {
					return fmt.Errorf("kubeconfig write failed")
				}
			}
			return nil
		},
	}
	oldExecutor := executor
	executor = mock
	defer func() { executor = oldExecutor }()

	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Should not return error — kubeconfig update failures are warnings.
	_ = captureStdout(t, func() {
		err := setupCmd.RunE(setupCmd, nil)
		if err != nil {
			t.Fatalf("setup error: %v", err)
		}
	})
}

func TestSetupCmd_Success(t *testing.T) {
	mock := &MockExecutor{
		RunPassthroughFunc: func(name string, args ...string) error {
			return nil
		},
	}
	oldExecutor := executor
	executor = mock
	defer func() { executor = oldExecutor }()

	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	output := captureStdout(t, func() {
		err := setupCmd.RunE(setupCmd, nil)
		if err != nil {
			t.Fatalf("setup error: %v", err)
		}
	})

	if !strings.Contains(output, "Wrote 4 profiles") {
		t.Errorf("missing profiles message, got: %q", output)
	}
	if !strings.Contains(output, "Configuring EKS clusters") {
		t.Errorf("missing configuring message, got: %q", output)
	}
}

func TestSetVersion(t *testing.T) {
	SetVersion("1.2.3")
	if version != "1.2.3" {
		t.Errorf("version = %q, want %q", version, "1.2.3")
	}
	SetVersion("dev")
}

func TestSetExecutor(t *testing.T) {
	mock := &MockExecutor{}
	oldExecutor := executor
	SetExecutor(mock)
	if executor != mock {
		t.Error("SetExecutor did not set executor")
	}
	executor = oldExecutor
}

func TestSetConfig(t *testing.T) {
	custom := &config.Config{
		SSO: config.SSOConfig{
			StartURL:    "https://example.com",
			Region:      "eu-west-1",
			SessionName: "custom",
		},
		DefaultRegion: "eu-west-1",
		Profiles: []config.Profile{
			{Alias: "test", AccountName: "test", AccountID: "111111111111", RoleName: "Admin"},
		},
	}
	oldCfg := cfg
	SetConfig(custom)
	if cfg != custom {
		t.Error("SetConfig did not set cfg")
	}
	cfg = oldCfg
}

func TestConfigInitCmd_NewFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	output := captureStdout(t, func() {
		err := configInitCmd.RunE(configInitCmd, nil)
		if err != nil {
			t.Fatalf("config init error: %v", err)
		}
	})

	if !strings.Contains(output, "Wrote default config") {
		t.Errorf("missing success message, got: %q", output)
	}

	// Verify the file was actually created.
	path := filepath.Join(tmpDir, ".config", "awsw", "config.yaml")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config file not created: %v", err)
	}
}

func TestConfigInitCmd_ExistsNoForce(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Pre-create the config file.
	cfgDir := filepath.Join(tmpDir, ".config", "awsw")
	if err := os.MkdirAll(cfgDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}

	forceOverwrite = false
	err := configInitCmd.RunE(configInitCmd, nil)
	if err == nil {
		t.Error("config init should return error when file exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error = %q, want to contain 'already exists'", err.Error())
	}
}

func TestConfigInitCmd_ExistsWithForce(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Pre-create the config file.
	cfgDir := filepath.Join(tmpDir, ".config", "awsw")
	if err := os.MkdirAll(cfgDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}

	forceOverwrite = true
	defer func() { forceOverwrite = false }()

	output := captureStdout(t, func() {
		err := configInitCmd.RunE(configInitCmd, nil)
		if err != nil {
			t.Fatalf("config init --force error: %v", err)
		}
	})

	if !strings.Contains(output, "Wrote default config") {
		t.Errorf("missing success message, got: %q", output)
	}
}

func TestPersistentPreRunE_LoadsConfig(t *testing.T) {
	// Clear cfg to test the loader.
	oldCfg := cfg
	cfg = nil
	defer func() { cfg = oldCfg }()

	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	err := rootCmd.PersistentPreRunE(rootCmd, nil)
	if err != nil {
		t.Fatalf("PersistentPreRunE error: %v", err)
	}
	if cfg == nil {
		t.Error("cfg should be set after PersistentPreRunE")
	}
}

func TestPersistentPreRunE_SkipsIfAlreadyLoaded(t *testing.T) {
	// cfg is already set from init().
	err := rootCmd.PersistentPreRunE(rootCmd, nil)
	if err != nil {
		t.Fatalf("PersistentPreRunE error: %v", err)
	}
}

func TestRootCmd_ClusterArg(t *testing.T) {
	mock := &MockExecutor{
		RunFunc: func(name string, args ...string) (string, error) {
			return "", nil
		},
	}
	oldExecutor := executor
	executor = mock
	defer func() { executor = oldExecutor }()

	shellEval = false

	// Use the cluster name (not the profile alias).
	output := captureStdout(t, func() {
		err := rootCmd.RunE(rootCmd, []string{"my-cluster-dev"})
		if err != nil {
			t.Fatalf("root error: %v", err)
		}
	})

	// Should resolve to the dev-eks profile.
	if !strings.Contains(output, "→ AWS_PROFILE=dev-eks") {
		t.Errorf("output missing profile switch, got: %q", output)
	}
	if !strings.Contains(output, "→ kubectl context: my-cluster-dev") {
		t.Errorf("output missing kubectl context, got: %q", output)
	}
}

func TestSwitchTo_SpecificCluster(t *testing.T) {
	shellEval = true
	shellName = "fish"
	defer func() {
		shellEval = false
		shellName = ""
	}()

	profile := findTestProfile("it-eks")
	cluster := &profile.EKSClusters[0]

	output := captureStdout(t, func() {
		err := switchTo(profile, cluster)
		if err != nil {
			t.Fatalf("switchTo() error: %v", err)
		}
	})

	if !strings.Contains(output, "set -gx AWS_PROFILE it-eks") {
		t.Errorf("output missing profile export, got: %q", output)
	}
	if !strings.Contains(output, "kubectl config use-context my-cluster-staging") {
		t.Errorf("output missing specific cluster context, got: %q", output)
	}
}

func TestLoginCmd_AlreadyAuthenticated(t *testing.T) {
	mock := &MockExecutor{
		RunFunc: func(name string, args ...string) (string, error) {
			// sts get-caller-identity succeeds = already logged in.
			return `{"Account":"111111111111"}`, nil
		},
	}
	oldExecutor := executor
	executor = mock
	defer func() { executor = oldExecutor }()

	forceLogin = false
	defer func() { forceLogin = false }()

	output := captureStdout(t, func() {
		err := loginCmd.RunE(loginCmd, nil)
		if err != nil {
			t.Fatalf("login error: %v", err)
		}
	})

	if !strings.Contains(output, "Already authenticated") {
		t.Errorf("missing already-authenticated message, got: %q", output)
	}
}

func TestLoginCmd_ForceReauth(t *testing.T) {
	called := false
	mock := &MockExecutor{
		RunFunc: func(name string, args ...string) (string, error) {
			return `{"Account":"111111111111"}`, nil
		},
		RunPassthroughFunc: func(name string, args ...string) error {
			called = true
			return nil
		},
	}
	oldExecutor := executor
	executor = mock
	defer func() { executor = oldExecutor }()

	forceLogin = true
	defer func() { forceLogin = false }()

	output := captureStdout(t, func() {
		err := loginCmd.RunE(loginCmd, nil)
		if err != nil {
			t.Fatalf("login error: %v", err)
		}
	})

	if !called {
		t.Error("--force should trigger SSO login even when already authenticated")
	}
	if !strings.Contains(output, "Logged in") {
		t.Errorf("missing login message, got: %q", output)
	}
}

// --- helpers ---

func findTestProfile(alias string) *config.Profile {
	p := cfg.FindProfile(alias)
	if p == nil {
		panic("test profile not found: " + alias)
	}
	return p
}

// captureStdout redirects os.Stdout to capture printed output.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf strings.Builder
	b := make([]byte, 1024)
	for {
		n, err := r.Read(b)
		if n > 0 {
			buf.Write(b[:n])
		}
		if err != nil {
			break
		}
	}
	return buf.String()
}
