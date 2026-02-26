package config

// EKSCluster represents a single EKS cluster associated with a profile.
type EKSCluster struct {
	Name         string `yaml:"name"`
	Region       string `yaml:"region"`
	ContextAlias string `yaml:"context_alias"`
}

// Profile represents an AWS SSO account/role combination.
type Profile struct {
	Alias       string       `yaml:"alias"`
	AccountName string       `yaml:"account_name"`
	AccountID   string       `yaml:"account_id"`
	RoleName    string       `yaml:"role_name"`
	EKSClusters []EKSCluster `yaml:"eks_clusters,omitempty"`
}

// HasEKS returns true if the profile has any EKS clusters configured.
func (p Profile) HasEKS() bool {
	return len(p.EKSClusters) > 0
}

// FzfLabel returns a formatted string for display in fzf.
func (p Profile) FzfLabel() string {
	return p.Alias + "\t" + p.AccountName + " (" + p.AccountID + ") " + p.RoleName
}
