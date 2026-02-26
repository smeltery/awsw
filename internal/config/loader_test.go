package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.SSO.StartURL == "" {
		t.Error("SSO.StartURL is empty")
	}
	if cfg.SSO.Region == "" {
		t.Error("SSO.Region is empty")
	}
	if cfg.SSO.SessionName == "" {
		t.Error("SSO.SessionName is empty")
	}
	if cfg.DefaultRegion == "" {
		t.Error("DefaultRegion is empty")
	}
	if len(cfg.Profiles) == 0 {
		t.Error("Profiles is empty")
	}
}

func TestLoadConfigFrom_FileNotExist(t *testing.T) {
	cfg, err := LoadConfigFrom("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("LoadConfigFrom() error: %v", err)
	}

	// Should fall back to defaults.
	defaults := DefaultConfig()
	if len(cfg.Profiles) != len(defaults.Profiles) {
		t.Errorf("got %d profiles, want %d", len(cfg.Profiles), len(defaults.Profiles))
	}
}

func TestLoadConfigFrom_ValidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")

	yaml := `sso:
  start_url: https://example.com/start
  region: eu-west-1
  session_name: test-session
  registration_scopes: sso:account:access
default_region: eu-west-1
profiles:
  - alias: staging
    account_name: staging
    account_id: "999999999999"
    role_name: AdminAccess
    eks_clusters:
      - name: staging-cluster
        region: eu-west-1
        context_alias: staging-cluster
`
	if err := os.WriteFile(path, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfigFrom(path)
	if err != nil {
		t.Fatalf("LoadConfigFrom() error: %v", err)
	}

	if cfg.SSO.StartURL != "https://example.com/start" {
		t.Errorf("SSO.StartURL = %q, want %q", cfg.SSO.StartURL, "https://example.com/start")
	}
	if cfg.SSO.Region != "eu-west-1" {
		t.Errorf("SSO.Region = %q, want %q", cfg.SSO.Region, "eu-west-1")
	}
	if cfg.SSO.SessionName != "test-session" {
		t.Errorf("SSO.SessionName = %q, want %q", cfg.SSO.SessionName, "test-session")
	}
	if cfg.DefaultRegion != "eu-west-1" {
		t.Errorf("DefaultRegion = %q, want %q", cfg.DefaultRegion, "eu-west-1")
	}
	if len(cfg.Profiles) != 1 {
		t.Fatalf("got %d profiles, want 1", len(cfg.Profiles))
	}

	p := cfg.Profiles[0]
	if p.Alias != "staging" {
		t.Errorf("Alias = %q, want %q", p.Alias, "staging")
	}
	if p.AccountID != "999999999999" {
		t.Errorf("AccountID = %q, want %q", p.AccountID, "999999999999")
	}
	if !p.HasEKS() {
		t.Error("HasEKS() should be true")
	}
}

func TestLoadConfigFrom_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(path, []byte("{{invalid yaml"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfigFrom(path)
	if err == nil {
		t.Error("LoadConfigFrom() should return error for invalid YAML")
	}
}

func TestSaveConfig_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")

	original := &Config{
		SSO: SSOConfig{
			StartURL:           "https://test.example.com/start",
			Region:             "us-west-2",
			SessionName:        "my-session",
			RegistrationScopes: "sso:account:access",
		},
		DefaultRegion: "us-west-2",
		Profiles: []Profile{
			{
				Alias: "alpha", AccountName: "alpha-acct", AccountID: "111111111111", RoleName: "Dev",
				EKSClusters: []EKSCluster{{Name: "alpha-cluster", Region: "us-west-2", ContextAlias: "alpha-cluster"}},
			},
			{Alias: "beta", AccountName: "beta-acct", AccountID: "222222222222", RoleName: "ReadOnly"},
		},
	}

	if err := SaveConfig(original, path); err != nil {
		t.Fatalf("SaveConfig() error: %v", err)
	}

	loaded, err := LoadConfigFrom(path)
	if err != nil {
		t.Fatalf("LoadConfigFrom() error: %v", err)
	}

	if loaded.SSO.StartURL != original.SSO.StartURL {
		t.Errorf("SSO.StartURL = %q, want %q", loaded.SSO.StartURL, original.SSO.StartURL)
	}
	if loaded.SSO.Region != original.SSO.Region {
		t.Errorf("SSO.Region = %q, want %q", loaded.SSO.Region, original.SSO.Region)
	}
	if loaded.DefaultRegion != original.DefaultRegion {
		t.Errorf("DefaultRegion = %q, want %q", loaded.DefaultRegion, original.DefaultRegion)
	}
	if len(loaded.Profiles) != len(original.Profiles) {
		t.Fatalf("got %d profiles, want %d", len(loaded.Profiles), len(original.Profiles))
	}
	for i, p := range loaded.Profiles {
		if p.Alias != original.Profiles[i].Alias {
			t.Errorf("Profiles[%d].Alias = %q, want %q", i, p.Alias, original.Profiles[i].Alias)
		}
		if p.HasEKS() != original.Profiles[i].HasEKS() {
			t.Errorf("Profiles[%d].HasEKS() = %v, want %v", i, p.HasEKS(), original.Profiles[i].HasEKS())
		}
	}
}

func TestSaveConfig_CreatesDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "nested", "dir", "config.yaml")

	cfg := DefaultConfig()
	if err := SaveConfig(cfg, path); err != nil {
		t.Fatalf("SaveConfig() error: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Errorf("config file not created: %v", err)
	}
}

func TestConfigPath(t *testing.T) {
	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath() error: %v", err)
	}
	if !filepath.IsAbs(path) {
		t.Errorf("ConfigPath() returned non-absolute path: %q", path)
	}
	if filepath.Base(path) != "config.yaml" {
		t.Errorf("ConfigPath() filename = %q, want %q", filepath.Base(path), "config.yaml")
	}
}

func TestLoadConfig_FallsBackToDefaults(t *testing.T) {
	// Point HOME at a temp dir with no config file.
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	defaults := DefaultConfig()
	if len(cfg.Profiles) != len(defaults.Profiles) {
		t.Errorf("got %d profiles, want %d", len(cfg.Profiles), len(defaults.Profiles))
	}
}

func TestLoadConfig_ReadsFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Write a custom config.
	cfgDir := filepath.Join(tmpDir, ".config", "awsw")
	if err := os.MkdirAll(cfgDir, 0o700); err != nil {
		t.Fatal(err)
	}
	yaml := `sso:
  start_url: https://custom.example.com
  region: ap-southeast-1
  session_name: custom
  registration_scopes: sso:account:access
default_region: ap-southeast-1
profiles:
  - alias: custom-profile
    account_name: custom
    account_id: "333333333333"
    role_name: CustomRole
`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	if cfg.SSO.StartURL != "https://custom.example.com" {
		t.Errorf("SSO.StartURL = %q, want custom URL", cfg.SSO.StartURL)
	}
	if len(cfg.Profiles) != 1 {
		t.Fatalf("got %d profiles, want 1", len(cfg.Profiles))
	}
	if cfg.Profiles[0].Alias != "custom-profile" {
		t.Errorf("Alias = %q, want %q", cfg.Profiles[0].Alias, "custom-profile")
	}
}

func TestWriteAWSConfig_CustomConfig(t *testing.T) {
	cfg := &Config{
		SSO: SSOConfig{
			StartURL:           "https://custom.example.com",
			Region:             "eu-west-1",
			SessionName:        "my-session",
			RegistrationScopes: "sso:account:access",
		},
		DefaultRegion: "eu-west-1",
		Profiles: []Profile{
			{Alias: "staging", AccountName: "staging", AccountID: "999999999999", RoleName: "Admin"},
		},
	}

	tmpDir := t.TempDir()
	err := WriteAWSConfig(cfg, tmpDir)
	if err != nil {
		t.Fatalf("WriteAWSConfig() error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmpDir, ".aws", "config"))
	if err != nil {
		t.Fatalf("reading config: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "[sso-session my-session]") {
		t.Error("missing custom sso-session")
	}
	if !strings.Contains(content, "[profile staging]") {
		t.Error("missing custom profile")
	}
	if !strings.Contains(content, "region = eu-west-1") {
		t.Error("missing custom region")
	}
}

func TestSaveConfig_WriteError(t *testing.T) {
	tmpDir := t.TempDir()
	// Make directory read-only so WriteFile fails.
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	if err := os.MkdirAll(readOnlyDir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(readOnlyDir, 0o700) })

	path := filepath.Join(readOnlyDir, "config.yaml")
	err := SaveConfig(DefaultConfig(), path)
	if err == nil {
		t.Error("SaveConfig() should return error for read-only directory")
	}
}

func TestLoadConfigFrom_ReadError(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")

	// Create file, then make it unreadable.
	if err := os.WriteFile(path, []byte("test"), 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o600) })

	_, err := LoadConfigFrom(path)
	if err == nil {
		t.Error("LoadConfigFrom() should return error for unreadable file")
	}
}

func TestConfigDir(t *testing.T) {
	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() error: %v", err)
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("ConfigDir() returned non-absolute path: %q", dir)
	}
	if filepath.Base(dir) != "awsw" {
		t.Errorf("ConfigDir() base = %q, want %q", filepath.Base(dir), "awsw")
	}
}

func TestFindByCluster_ContextAlias(t *testing.T) {
	cfg := DefaultConfig()
	match := cfg.FindByCluster("my-cluster-dev")
	if match == nil {
		t.Fatal("FindByCluster() returned nil for valid context alias")
	}
	if match.Profile.Alias != "dev-eks" {
		t.Errorf("Profile.Alias = %q, want %q", match.Profile.Alias, "dev-eks")
	}
	if match.Cluster.Name != "my-cluster-dev" {
		t.Errorf("Cluster.Name = %q, want %q", match.Cluster.Name, "my-cluster-dev")
	}
}

func TestFindByCluster_Name(t *testing.T) {
	// Create a config where context alias differs from cluster name.
	cfg := &Config{
		Profiles: []Profile{
			{
				Alias: "test", AccountName: "test", AccountID: "111", RoleName: "Admin",
				EKSClusters: []EKSCluster{
					{Name: "real-name", Region: "us-east-1", ContextAlias: "friendly-alias"},
				},
			},
		},
	}
	// Should match by name when alias doesn't match.
	match := cfg.FindByCluster("real-name")
	if match == nil {
		t.Fatal("FindByCluster() returned nil for valid cluster name")
	}
	if match.Cluster.ContextAlias != "friendly-alias" {
		t.Errorf("Cluster.ContextAlias = %q, want %q", match.Cluster.ContextAlias, "friendly-alias")
	}
}

func TestFindByCluster_NotFound(t *testing.T) {
	cfg := DefaultConfig()
	match := cfg.FindByCluster("nonexistent")
	if match != nil {
		t.Errorf("FindByCluster() should return nil for unknown cluster, got profile %q", match.Profile.Alias)
	}
}

func TestClusterAliases(t *testing.T) {
	cfg := DefaultConfig()
	aliases := cfg.ClusterAliases()
	if len(aliases) != 2 {
		t.Fatalf("ClusterAliases() returned %d aliases, want 2", len(aliases))
	}
	expected := map[string]bool{"my-cluster-dev": true, "my-cluster-staging": true}
	for _, a := range aliases {
		if !expected[a] {
			t.Errorf("unexpected cluster alias: %q", a)
		}
	}
}

func TestFindByCluster_ContextAliasPrecedence(t *testing.T) {
	// If a cluster's Name matches another cluster's ContextAlias,
	// the ContextAlias match should win.
	cfg := &Config{
		Profiles: []Profile{
			{
				Alias: "p1", AccountName: "a", AccountID: "111", RoleName: "R",
				EKSClusters: []EKSCluster{
					{Name: "underlying-name", Region: "us-east-1", ContextAlias: "shared-name"},
				},
			},
			{
				Alias: "p2", AccountName: "b", AccountID: "222", RoleName: "R",
				EKSClusters: []EKSCluster{
					{Name: "shared-name", Region: "us-east-1", ContextAlias: "other-alias"},
				},
			},
		},
	}
	match := cfg.FindByCluster("shared-name")
	if match == nil {
		t.Fatal("FindByCluster() returned nil")
	}
	// Should match p1 (ContextAlias) not p2 (Name).
	if match.Profile.Alias != "p1" {
		t.Errorf("expected ContextAlias match (p1), got %q", match.Profile.Alias)
	}
}

func TestGenerateAWSConfig_CustomConfig(t *testing.T) {
	cfg := &Config{
		SSO: SSOConfig{
			StartURL:           "https://custom.example.com",
			Region:             "eu-west-1",
			SessionName:        "my-session",
			RegistrationScopes: "sso:account:access",
		},
		DefaultRegion: "eu-west-1",
		Profiles: []Profile{
			{Alias: "staging", AccountName: "staging", AccountID: "999999999999", RoleName: "Admin"},
		},
	}

	content, err := GenerateAWSConfig(cfg)
	if err != nil {
		t.Fatalf("GenerateAWSConfig() error: %v", err)
	}

	if !strings.Contains(content, "[sso-session my-session]") {
		t.Error("missing custom sso-session")
	}
	if !strings.Contains(content, "sso_start_url = https://custom.example.com") {
		t.Error("missing custom start_url")
	}
	if !strings.Contains(content, "[profile staging]") {
		t.Error("missing custom profile")
	}
	if !strings.Contains(content, "region = eu-west-1") {
		t.Error("missing custom region")
	}
}
