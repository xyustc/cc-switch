package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var validNameRe = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// Profile represents a named set of Claude settings to apply.
type Profile struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Settings    map[string]interface{} `json:"settings"`
}

// Profiles is the top-level structure stored in ~/.cc-switch/profiles.json.
type Profiles struct {
	Profiles []Profile `json:"profiles"`
	Active   string    `json:"active"`
}

func profilesPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cc-switch", "profiles.json")
}

// LoadProfiles reads ~/.cc-switch/profiles.json.
// Returns empty Profiles if the file does not exist.
// Exits on parse error.
func LoadProfiles() (*Profiles, error) {
	path := profilesPath()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Profiles{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read profiles: %w", err)
	}
	var p Profiles
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parse profiles.json: %w", err)
	}
	return &p, nil
}

// SaveProfiles writes the Profiles struct to disk with 0600 permissions.
func SaveProfiles(p *Profiles) error {
	path := profilesPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create .cc-switch dir: %w", err)
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal profiles: %w", err)
	}
	return atomicWrite(path, data, 0600)
}

// ValidateName checks profile name constraints.
func ValidateName(name string, existing []Profile, editingName string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if !validNameRe.MatchString(name) {
		return fmt.Errorf("name must match [a-zA-Z0-9_-]")
	}
	for _, p := range existing {
		if p.Name == name && p.Name != editingName {
			return fmt.Errorf("name %q already exists", name)
		}
	}
	return nil
}

// AddProfile adds a new profile. Returns error if name is invalid or duplicate.
func AddProfile(profiles *Profiles, p Profile) error {
	if err := ValidateName(p.Name, profiles.Profiles, ""); err != nil {
		return err
	}
	profiles.Profiles = append(profiles.Profiles, p)
	return SaveProfiles(profiles)
}

// UpdateProfile replaces the profile with oldName.
func UpdateProfile(profiles *Profiles, oldName string, updated Profile) error {
	if err := ValidateName(updated.Name, profiles.Profiles, oldName); err != nil {
		return err
	}
	for i, p := range profiles.Profiles {
		if p.Name == oldName {
			profiles.Profiles[i] = updated
			if profiles.Active == oldName {
				profiles.Active = updated.Name
			}
			return SaveProfiles(profiles)
		}
	}
	return fmt.Errorf("profile %q not found", oldName)
}

// DeleteProfile removes the profile by name.
// If it was active, sets active to "".
func DeleteProfile(profiles *Profiles, name string) error {
	newList := make([]Profile, 0, len(profiles.Profiles))
	found := false
	for _, p := range profiles.Profiles {
		if p.Name == name {
			found = true
			continue
		}
		newList = append(newList, p)
	}
	if !found {
		return fmt.Errorf("profile %q not found", name)
	}
	profiles.Profiles = newList
	if profiles.Active == name {
		profiles.Active = ""
	}
	return SaveProfiles(profiles)
}
