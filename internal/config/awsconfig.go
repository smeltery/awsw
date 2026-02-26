package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

const awsConfigTemplate = `[sso-session {{ .SessionName }}]
sso_start_url = {{ .StartURL }}
sso_region = {{ .SSORegion }}
sso_registration_scopes = {{ .RegistrationScope }}
{{ range .Profiles }}
[profile {{ .Alias }}]
sso_session = {{ $.SessionName }}
sso_account_id = {{ .AccountID }}
sso_role_name = {{ .RoleName }}
region = {{ $.DefaultRegion }}
{{ end }}`

type awsConfigData struct {
	SessionName       string
	StartURL          string
	SSORegion         string
	RegistrationScope string
	DefaultRegion     string
	Profiles          []Profile
}

// GenerateAWSConfig renders the ~/.aws/config INI content.
func GenerateAWSConfig(cfg *Config) (string, error) {
	data := awsConfigData{
		SessionName:       cfg.SSO.SessionName,
		StartURL:          cfg.SSO.StartURL,
		SSORegion:         cfg.SSO.Region,
		RegistrationScope: cfg.SSO.RegistrationScopes,
		DefaultRegion:     cfg.DefaultRegion,
		Profiles:          cfg.Profiles,
	}

	tmpl, err := template.New("awsconfig").Parse(awsConfigTemplate)
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return buf.String(), nil
}

// WriteAWSConfig writes the generated config to ~/.aws/config,
// backing up any existing file first.
func WriteAWSConfig(cfg *Config, homeDir string) error {
	awsDir := filepath.Join(homeDir, ".aws")
	configPath := filepath.Join(awsDir, "config")
	backupPath := filepath.Join(awsDir, "config.bak")

	// Ensure ~/.aws exists.
	if err := os.MkdirAll(awsDir, 0o700); err != nil {
		return fmt.Errorf("creating %s: %w", awsDir, err)
	}

	// Backup existing config if present.
	if _, err := os.Stat(configPath); err == nil {
		existing, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("reading existing config: %w", err)
		}
		if err := os.WriteFile(backupPath, existing, 0o600); err != nil {
			return fmt.Errorf("writing backup: %w", err)
		}
		fmt.Printf("✓ Backed up %s to %s\n", configPath, backupPath)
	}

	content, err := GenerateAWSConfig(cfg)
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	fmt.Printf("✓ Wrote %d profiles to %s\n", len(cfg.Profiles), configPath)
	return nil
}
