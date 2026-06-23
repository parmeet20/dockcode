package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Manager handles concurrent read and write access to AppConfig, persisted on disk.
type Manager struct {
	mu   sync.RWMutex
	path string
	cfg  *AppConfig
}

// NewManager creates a Manager configured to save/load config from ~/.dockcode/config.json
func NewManager() (*Manager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not find home directory: %w", err)
	}

	dir := filepath.Join(home, ".dockcode")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("could not create config directory: %w", err)
	}

	path := filepath.Join(dir, "config.json")
	return &Manager{
		path: path,
		cfg:  DefaultConfig(),
	}, nil
}

// ConfigExists checks if the configuration file already exists on disk
func (m *Manager) ConfigExists() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, err := os.Stat(m.path)
	return err == nil
}

// Load reads the configuration from disk. If the file does not exist, it stays with defaults.
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.path)
	if os.IsNotExist(err) {
		// Use defaults
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config JSON: %w", err)
	}

	m.cfg = &cfg
	return nil
}

// Save writes the current configuration back to disk atomically
func (m *Manager) Save() error {
	m.mu.RLock()
	cfg := *m.cfg // shallow copy while read-locked
	m.mu.RUnlock()

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Atomic write: write to temp file first, then rename
	tmpFile := m.path + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write temp config file: %w", err)
	}

	if err := os.Rename(tmpFile, m.path); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to rename temp config file: %w", err)
	}

	return nil
}

// Get returns a copy of the current configuration (thread-safe RLock)
func (m *Manager) Get() AppConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return *m.cfg
}

// Update mutates the configuration using a callback and saves it to disk (thread-safe Lock)
func (m *Manager) Update(fn func(*AppConfig)) error {
	m.mu.Lock()
	fn(m.cfg)
	m.mu.Unlock() // unlock before writing to disk to avoid holding lock during I/O

	return m.Save()
}
