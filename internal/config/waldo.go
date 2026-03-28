package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Root         string
	ClaudeRoot   string
	ActivePersona string
	S3Bucket     string
	AWSProfile   string
}

// Load reads waldo configuration from ~/.config/waldo and ~/.claude.
func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}

	configRoot := filepath.Join(home, ".config", "waldo")
	claudeRoot := filepath.Join(home, ".claude")

	cfg := &Config{
		Root:       configRoot,
		ClaudeRoot: claudeRoot,
	}

	// Read active persona
	activeFile := filepath.Join(configRoot, ".active")
	if data, err := os.ReadFile(activeFile); err == nil {
		cfg.ActivePersona = string(data)
	}

	// Fallback to ~/.claude/.active for backwards compat
	if cfg.ActivePersona == "" {
		activeFile := filepath.Join(claudeRoot, "personas", ".active")
		if data, err := os.ReadFile(activeFile); err == nil {
			cfg.ActivePersona = string(data)
		}
	}

	// Read settings.json for S3 bucket and AWS profile
	settingsFile := filepath.Join(claudeRoot, "settings.json")
	if data, err := os.ReadFile(settingsFile); err == nil {
		var settings struct {
			Env map[string]string `json:"env"`
		}
		if err := json.Unmarshal(data, &settings); err == nil {
			if bucket, ok := settings.Env["WALDO_S3_BUCKET"]; ok {
				cfg.S3Bucket = bucket
			}
			if profile, ok := settings.Env["AWS_PROFILE"]; ok {
				cfg.AWSProfile = profile
			}
		}
	}

	return cfg, nil
}

// SaveActive updates the active persona file.
func (c *Config) SaveActive(name string) error {
	// Create config root if needed
	if err := os.MkdirAll(c.Root, 0755); err != nil {
		return err
	}

	activeFile := filepath.Join(c.Root, ".active")
	return os.WriteFile(activeFile, []byte(name), 0644)
}

// SaveS3Bucket updates the S3 bucket in settings.json.
func (c *Config) SaveS3Bucket(bucket string) error {
	settingsFile := filepath.Join(c.ClaudeRoot, "settings.json")

	// Read existing settings or create empty
	var settings map[string]interface{}
	if data, err := os.ReadFile(settingsFile); err == nil {
		json.Unmarshal(data, &settings)
	}

	if settings == nil {
		settings = make(map[string]interface{})
	}

	// Ensure env section exists
	env, ok := settings["env"].(map[string]interface{})
	if !ok {
		env = make(map[string]interface{})
		settings["env"] = env
	}

	// Update bucket
	env["WALDO_S3_BUCKET"] = bucket
	c.S3Bucket = bucket

	// Write back
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(settingsFile, data, 0644)
}

// SaveAWSProfile updates the AWS profile in settings.json.
func (c *Config) SaveAWSProfile(profile string) error {
	settingsFile := filepath.Join(c.ClaudeRoot, "settings.json")

	// Read existing settings or create empty
	var settings map[string]interface{}
	if data, err := os.ReadFile(settingsFile); err == nil {
		json.Unmarshal(data, &settings)
	}

	if settings == nil {
		settings = make(map[string]interface{})
	}

	// Ensure env section exists
	env, ok := settings["env"].(map[string]interface{})
	if !ok {
		env = make(map[string]interface{})
		settings["env"] = env
	}

	// Update profile
	env["AWS_PROFILE"] = profile
	c.AWSProfile = profile

	// Write back
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(settingsFile, data, 0644)
}
