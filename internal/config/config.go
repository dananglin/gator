package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	CurrentUsername string   `json:"currentUsername"`
	DBConfig        DBConfig `json:"database"`
}

type DBConfig struct {
	URL string `json:"url"`
}

func NewConfig() (Config, error) {
	path, err := configFilePath()
	if err != nil {
		return Config{}, fmt.Errorf("unable to get the path to the configuration file: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("unable to read %s: %w", path, err)
	}

	var cfg Config

	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("unable to decode the JSON data: %w", err)
	}

	return cfg, nil
}

func (c *Config) SetUser(user string) error {
	c.CurrentUsername = user

	return write(*c)
}

func configFilePath() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("unable to get the user's home config directory: %w", err)
	}

	path := filepath.Join(userConfigDir, "gator", "config.json")

	return path, nil
}

func write(cfg Config) error {
	path, err := configFilePath()
	if err != nil {
		return fmt.Errorf("unable to get the path to the configuration file: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("unable to create %s: %w", path, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")

	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("unable to save the config to file: %w", err)
	}

	return nil
}
