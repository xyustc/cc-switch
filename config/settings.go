package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func settingsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "settings.json")
}

func backupDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cc-switch", "backups")
}

// LoadSettings reads ~/.claude/settings.json.
// Returns empty map if the file does not exist.
func LoadSettings() (map[string]interface{}, error) {
	data, err := os.ReadFile(settingsPath())
	if os.IsNotExist(err) {
		return map[string]interface{}{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read settings.json: %w", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse settings.json: %w", err)
	}
	return m, nil
}

// ApplyProfile merges profile.Settings top-level keys into current settings
// and writes the result back atomically. Creates a backup first.
func ApplyProfile(profile Profile) error {
	current, err := LoadSettings()
	if err != nil {
		return err
	}

	// Backup before writing (non-fatal if it fails)
	if berr := backupSettings(current); berr != nil {
		fmt.Fprintf(os.Stderr, "warning: backup failed: %v\n", berr)
	}

	// Top-level field replacement
	for k, v := range profile.Settings {
		current[k] = v
	}

	data, err := json.MarshalIndent(current, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}
	return atomicWrite(settingsPath(), data, 0600)
}

// backupSettings writes a timestamped copy to ~/.cc-switch/backups/ and prunes old ones.
func backupSettings(settings map[string]interface{}) error {
	dir := backupDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create backup dir: %w", err)
	}
	ts := time.Now().Format("20060102-150405")
	dest := filepath.Join(dir, fmt.Sprintf("settings-%s.json", ts))
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	if err := atomicWrite(dest, data, 0600); err != nil {
		return err
	}
	return pruneBackups(dir, 10)
}

// pruneBackups keeps only the most recent `keep` backup files.
func pruneBackups(dir string, keep int) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "settings-") {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(files) // lexicographic = chronological for our timestamp format
	for len(files) > keep {
		if err := os.Remove(files[0]); err != nil {
			return err
		}
		files = files[1:]
	}
	return nil
}

// atomicWrite writes data to path via a temp file + rename.
func atomicWrite(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() {
		tmp.Close()
		os.Remove(tmpName) // no-op if rename succeeded
	}()
	if err := tmp.Chmod(perm); err != nil {
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("atomic rename: %w", err)
	}
	return nil
}
