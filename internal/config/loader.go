package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// SSOConfig holds AWS SSO session settings.
type SSOConfig struct {
	StartURL           string `yaml:"start_url"`
	Region             string `yaml:"region"`
	SessionName        string `yaml:"session_name"`
	RegistrationScopes string `yaml:"registration_scopes"`
}

// Config is the top-level awsw configuration.
type Config struct {
	SSO           SSOConfig `yaml:"sso"`
	DefaultRegion string    `yaml:"default_region"`
	Profiles      []Profile `yaml:"profiles"`
}

// FindProfile looks up a profile by alias. Returns nil if not found.
func (c *Config) FindProfile(alias string) *Profile {
	for i := range c.Profiles {
		if c.Profiles[i].Alias == alias {
			return &c.Profiles[i]
		}
	}
	return nil
}

// ProfileAliases returns a list of all profile aliases.
func (c *Config) ProfileAliases() []string {
	aliases := make([]string, len(c.Profiles))
	for i, p := range c.Profiles {
		aliases[i] = p.Alias
	}
	return aliases
}

// ClusterMatch pairs a profile with the specific EKS cluster that matched.
type ClusterMatch struct {
	Profile *Profile
	Cluster *EKSCluster
}

// FindByCluster searches all profiles for an EKS cluster matching the given
// name by ContextAlias first, then by cluster Name.
func (c *Config) FindByCluster(name string) *ClusterMatch {
	// First pass: match on ContextAlias.
	for i := range c.Profiles {
		for j := range c.Profiles[i].EKSClusters {
			if c.Profiles[i].EKSClusters[j].ContextAlias == name {
				return &ClusterMatch{
					Profile: &c.Profiles[i],
					Cluster: &c.Profiles[i].EKSClusters[j],
				}
			}
		}
	}
	// Second pass: match on cluster Name.
	for i := range c.Profiles {
		for j := range c.Profiles[i].EKSClusters {
			if c.Profiles[i].EKSClusters[j].Name == name {
				return &ClusterMatch{
					Profile: &c.Profiles[i],
					Cluster: &c.Profiles[i].EKSClusters[j],
				}
			}
		}
	}
	return nil
}

// ClusterAliases returns a list of all EKS cluster context aliases.
func (c *Config) ClusterAliases() []string {
	var aliases []string
	for _, p := range c.Profiles {
		for _, cl := range p.EKSClusters {
			aliases = append(aliases, cl.ContextAlias)
		}
	}
	return aliases
}

// DefaultConfig returns the built-in default configuration.
func DefaultConfig() *Config {
	return &Config{
		SSO: SSOConfig{
			StartURL:           "https://d-1234567890.awsapps.com/start/#",
			Region:             "us-east-1",
			SessionName:        "default",
			RegistrationScopes: "sso:account:access",
		},
		DefaultRegion: "us-east-1",
		Profiles: []Profile{
			{
				Alias:       "dev-ro",
				AccountName: "dev",
				AccountID:   "111111111111",
				RoleName:    "ReadOnlyAccess",
			},
			{
				Alias:       "dev-eks",
				AccountName: "dev",
				AccountID:   "111111111111",
				RoleName:    "eng-eks",
				EKSClusters: []EKSCluster{
					{Name: "my-cluster-dev", Region: "us-east-1", ContextAlias: "my-cluster-dev"},
				},
			},
			{
				Alias:       "it-eks",
				AccountName: "staging",
				AccountID:   "222222222222",
				RoleName:    "eng-eks",
				EKSClusters: []EKSCluster{
					{Name: "my-cluster-staging", Region: "us-east-2", ContextAlias: "my-cluster-staging"},
				},
			},
			{
				Alias:       "prod-ro",
				AccountName: "production",
				AccountID:   "333333333333",
				RoleName:    "ReadOnlyAccess",
			},
		},
	}
}

// ConfigDir returns the directory for awsw config files.
func ConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(homeDir, ".config", "awsw"), nil
}

// ConfigPath returns the full path to the config file.
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// LoadConfig reads the config from ~/.config/awsw/config.yaml.
// If the file does not exist, it returns DefaultConfig().
func LoadConfig() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return DefaultConfig(), nil
	}
	return LoadConfigFrom(path)
}

// LoadConfigFrom reads the config from the given path.
// If the file does not exist, it returns DefaultConfig().
func LoadConfigFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}

	return &cfg, nil
}

// SaveConfig writes a Config to the given path, creating parent directories.
func SaveConfig(cfg *Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}
