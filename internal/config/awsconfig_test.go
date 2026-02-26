package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateAWSConfig(t *testing.T) {
	cfg := DefaultConfig()
	content, err := GenerateAWSConfig(cfg)
	if err != nil {
		t.Fatalf("GenerateAWSConfig() error: %v", err)
	}

	// Verify SSO session block.
	if !strings.Contains(content, "[sso-session default]") {
		t.Error("missing [sso-session default]")
	}
	if !strings.Contains(content, "sso_start_url = "+cfg.SSO.StartURL) {
		t.Error("missing sso_start_url")
	}
	if !strings.Contains(content, "sso_region = "+cfg.SSO.Region) {
		t.Error("missing sso_region")
	}

	// Verify all profiles are present.
	for _, p := range cfg.Profiles {
		if !strings.Contains(content, "[profile "+p.Alias+"]") {
			t.Errorf("missing [profile %s]", p.Alias)
		}
		if !strings.Contains(content, "sso_account_id = "+p.AccountID) {
			t.Errorf("missing sso_account_id for %s", p.Alias)
		}
		if !strings.Contains(content, "sso_role_name = "+p.RoleName) {
			t.Errorf("missing sso_role_name for %s", p.Alias)
		}
	}
}

func TestWriteAWSConfig_NewFile(t *testing.T) {
	cfg := DefaultConfig()
	tmpDir := t.TempDir()

	err := WriteAWSConfig(cfg, tmpDir)
	if err != nil {
		t.Fatalf("WriteAWSConfig() error: %v", err)
	}

	configPath := filepath.Join(tmpDir, ".aws", "config")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("reading config: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "[sso-session default]") {
		t.Error("config missing sso-session block")
	}
	if !strings.Contains(content, "[profile dev-eks]") {
		t.Error("config missing dev-eks profile")
	}
}

func TestWriteAWSConfig_BackupsExisting(t *testing.T) {
	cfg := DefaultConfig()
	tmpDir := t.TempDir()
	awsDir := filepath.Join(tmpDir, ".aws")
	if err := os.MkdirAll(awsDir, 0o700); err != nil {
		t.Fatal(err)
	}

	// Write a pre-existing config.
	existingContent := "[profile old]\nregion = us-west-2\n"
	configPath := filepath.Join(awsDir, "config")
	if err := os.WriteFile(configPath, []byte(existingContent), 0o600); err != nil {
		t.Fatal(err)
	}

	err := WriteAWSConfig(cfg, tmpDir)
	if err != nil {
		t.Fatalf("WriteAWSConfig() error: %v", err)
	}

	// Verify backup was created.
	backupPath := filepath.Join(awsDir, "config.bak")
	backup, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("reading backup: %v", err)
	}
	if string(backup) != existingContent {
		t.Errorf("backup content = %q, want %q", string(backup), existingContent)
	}

	// Verify new config was written.
	newConfig, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(newConfig), "[sso-session default]") {
		t.Error("new config missing sso-session block")
	}
}
