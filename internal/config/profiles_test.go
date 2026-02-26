package config

import "testing"

func TestFindProfile_ValidAlias(t *testing.T) {
	cfg := DefaultConfig()
	tests := []struct {
		alias       string
		wantAccount string
		wantRole    string
	}{
		{"dev-ro", "111111111111", "ReadOnlyAccess"},
		{"dev-eks", "111111111111", "eng-eks"},
		{"it-eks", "222222222222", "eng-eks"},
		{"prod-ro", "333333333333", "ReadOnlyAccess"},
	}

	for _, tt := range tests {
		t.Run(tt.alias, func(t *testing.T) {
			p := cfg.FindProfile(tt.alias)
			if p == nil {
				t.Fatalf("FindProfile(%q) returned nil", tt.alias)
			}
			if p.AccountID != tt.wantAccount {
				t.Errorf("AccountID = %q, want %q", p.AccountID, tt.wantAccount)
			}
			if p.RoleName != tt.wantRole {
				t.Errorf("RoleName = %q, want %q", p.RoleName, tt.wantRole)
			}
		})
	}
}

func TestFindProfile_InvalidAlias(t *testing.T) {
	cfg := DefaultConfig()
	p := cfg.FindProfile("nonexistent")
	if p != nil {
		t.Errorf("FindProfile(\"nonexistent\") = %v, want nil", p)
	}
}

func TestProfileAliases(t *testing.T) {
	cfg := DefaultConfig()
	aliases := cfg.ProfileAliases()
	if len(aliases) != len(cfg.Profiles) {
		t.Fatalf("ProfileAliases() returned %d aliases, want %d", len(aliases), len(cfg.Profiles))
	}

	want := map[string]bool{"dev-ro": true, "dev-eks": true, "it-eks": true, "prod-ro": true}
	for _, a := range aliases {
		if !want[a] {
			t.Errorf("unexpected alias %q", a)
		}
	}
}

func TestFzfLabel(t *testing.T) {
	p := Profile{
		Alias:       "dev-eks",
		AccountName: "dev",
		AccountID:   "111111111111",
		RoleName:    "eng-eks",
	}
	label := p.FzfLabel()
	if label != "dev-eks\tdev (111111111111) eng-eks" {
		t.Errorf("FzfLabel() = %q, want %q", label, "dev-eks\tdev (111111111111) eng-eks")
	}
}

func TestHasEKS(t *testing.T) {
	cfg := DefaultConfig()
	tests := []struct {
		alias   string
		wantEKS bool
	}{
		{"dev-ro", false},
		{"dev-eks", true},
		{"it-eks", true},
		{"prod-ro", false},
	}

	for _, tt := range tests {
		t.Run(tt.alias, func(t *testing.T) {
			p := cfg.FindProfile(tt.alias)
			if p == nil {
				t.Fatalf("FindProfile(%q) returned nil", tt.alias)
			}
			if p.HasEKS() != tt.wantEKS {
				t.Errorf("HasEKS() = %v, want %v", p.HasEKS(), tt.wantEKS)
			}
		})
	}
}
