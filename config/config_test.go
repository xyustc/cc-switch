package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// --- settings tests ---

func TestLoadSettings_FileNotExist(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	m, err := LoadSettings()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m) != 0 {
		t.Fatalf("expected empty map, got %v", m)
	}
}

func TestMergeSettings_TopLevelReplace(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	claudeDir := filepath.Join(home, ".claude")
	os.MkdirAll(claudeDir, 0700)

	original := map[string]interface{}{
		"env":                   map[string]interface{}{"TOKEN": "old", "BASE_URL": "old"},
		"permissions":           map[string]interface{}{"allow": []interface{}{}},
		"alwaysThinkingEnabled": false,
	}
	writeJSON(t, settingsPath(), original)

	profile := Profile{
		Name: "test",
		Settings: map[string]interface{}{
			"env": map[string]interface{}{"TOKEN": "new", "BASE_URL": "new"},
		},
	}
	if err := ApplyProfile(profile); err != nil {
		t.Fatalf("ApplyProfile: %v", err)
	}

	result, _ := LoadSettings()
	env := result["env"].(map[string]interface{})
	if env["TOKEN"] != "new" {
		t.Errorf("TOKEN should be 'new', got %v", env["TOKEN"])
	}
	// Untouched fields must remain
	if result["alwaysThinkingEnabled"] != false {
		t.Errorf("alwaysThinkingEnabled should remain false")
	}
	if result["permissions"] == nil {
		t.Errorf("permissions should remain")
	}
}

func TestMergeSettings_EmptyProfile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	claudeDir := filepath.Join(home, ".claude")
	os.MkdirAll(claudeDir, 0700)

	original := map[string]interface{}{"env": map[string]interface{}{"TOKEN": "keep"}}
	writeJSON(t, settingsPath(), original)

	profile := Profile{Name: "empty", Settings: map[string]interface{}{}}
	if err := ApplyProfile(profile); err != nil {
		t.Fatalf("ApplyProfile: %v", err)
	}

	result, _ := LoadSettings()
	env := result["env"].(map[string]interface{})
	if env["TOKEN"] != "keep" {
		t.Errorf("TOKEN should remain 'keep', got %v", env["TOKEN"])
	}
}

func TestAtomicWrite_FailSafe(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	original := []byte(`{"original":true}`)
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	// Write to a read-only directory to simulate failure
	roDir := filepath.Join(dir, "readonly")
	os.MkdirAll(roDir, 0500)
	roPath := filepath.Join(roDir, "file.json")
	err := atomicWrite(roPath, []byte(`{"new":true}`), 0600)
	if err == nil {
		t.Fatal("expected error writing to read-only dir")
	}
	// Original file in writable dir must be untouched
	data, _ := os.ReadFile(path)
	if string(data) != string(original) {
		t.Errorf("original file was modified: %s", data)
	}
}

func TestPruneBackups_KeepLatest10(t *testing.T) {
	dir := t.TempDir()
	// Create 13 backup files
	for i := 0; i < 13; i++ {
		name := filepath.Join(dir, "settings-2026010"+string(rune('0'+i))+"-120000.json")
		os.WriteFile(name, []byte("{}"), 0600)
	}
	if err := pruneBackups(dir, 10); err != nil {
		t.Fatalf("pruneBackups: %v", err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 10 {
		t.Errorf("expected 10 backups, got %d", len(entries))
	}
}

// --- profile tests ---

func TestProfileName_Validation(t *testing.T) {
	existing := []Profile{{Name: "existing"}}

	if err := ValidateName("", existing, ""); err == nil {
		t.Error("empty name should fail")
	}
	if err := ValidateName("bad name!", existing, ""); err == nil {
		t.Error("name with special chars should fail")
	}
	if err := ValidateName("existing", existing, ""); err == nil {
		t.Error("duplicate name should fail")
	}
	if err := ValidateName("existing", existing, "existing"); err != nil {
		t.Errorf("editing same name should pass: %v", err)
	}
	if err := ValidateName("valid-name_1", existing, ""); err != nil {
		t.Errorf("valid name should pass: %v", err)
	}
}

func TestDeleteActiveProfile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	os.MkdirAll(filepath.Join(home, ".cc-switch"), 0700)

	profiles := &Profiles{
		Profiles: []Profile{{Name: "a"}, {Name: "b"}},
		Active:   "a",
	}
	if err := DeleteProfile(profiles, "a"); err != nil {
		t.Fatalf("DeleteProfile: %v", err)
	}
	if profiles.Active != "" {
		t.Errorf("active should be empty after deleting active profile, got %q", profiles.Active)
	}
	if len(profiles.Profiles) != 1 || profiles.Profiles[0].Name != "b" {
		t.Errorf("unexpected profiles: %v", profiles.Profiles)
	}
}

func TestDeleteNonActiveProfile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	os.MkdirAll(filepath.Join(home, ".cc-switch"), 0700)

	profiles := &Profiles{
		Profiles: []Profile{{Name: "a"}, {Name: "b"}},
		Active:   "a",
	}
	if err := DeleteProfile(profiles, "b"); err != nil {
		t.Fatalf("DeleteProfile: %v", err)
	}
	if profiles.Active != "a" {
		t.Errorf("active should remain 'a', got %q", profiles.Active)
	}
}

// helper
func writeJSON(t *testing.T, path string, v interface{}) {
	t.Helper()
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}
