package profile

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/torgiren/ksef-cli/pkg/ksef"
)

type Config struct {
	CurrentProfile string `yaml:"currentProfile"`
	Profiles map[string]Profile `yaml:"profiles"`
}

type Profile struct {
	Nip string `yaml:"nip"`
	Api string `yaml:"api"`
	Name string `yaml:"name"`
}

func LoadConfig(configPath string) (Config, error) {
	if configPath == "" {
		slog.Debug("no config file specified, using defaults")
		return Config{}, nil
	}
	file, err := os.Open(configPath)
	if err != nil {
		//slog.Error("could not open config file", "err", err)
		return Config{}, fmt.Errorf("could not open config file: %w", err)
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err = decoder.Decode(&config); err != nil {
		slog.Error("could not parse config file", "err", err)
		return Config{}, fmt.Errorf("could not parse config file: %w", err)
	}

	slog.Debug("config loaded", "currentProfile", config.CurrentProfile, "profilesCount", len(config.Profiles))
	for name, profile := range config.Profiles {
		slog.Debug("profile loaded", "name", name, "nip", profile.Nip, "api", profile.Api)
	}
	return config, nil
}

func SaveConfig(config Config, configPath string) error {
	if configPath == "" {
		slog.Debug("no config file specified, skipping save")
		return nil
	}
	slog.Debug("saving config", "currentProfile", config.CurrentProfile, "profilesCount", len(config.Profiles))
	for name, profile := range config.Profiles {
		slog.Debug("profile to save", "name", name, "nip", profile.Nip, "api", profile.Api)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		slog.Error("could not encode config", "err", err)
		return fmt.Errorf("could not encode config: %w", err)
	}

	os.MkdirAll(filepath.Dir(configPath), 0700)
	if err = os.WriteFile(configPath, data, 0600); err != nil {
		slog.Error("could not write config file", "err", err)
		return fmt.Errorf("could not write config file: %w", err)
	}
	return nil
}

func LoadTokens(profileName string, cacheDir string) (*ksef.Tokens, error) {
	slog.Debug("loading tokens from cache", "profile", profileName, "cacheDir", cacheDir)
	cacheFile := filepath.Join(cacheDir, "profile_"+profileName+".json")
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("cache file not found for Profile: %s", profileName)
	}

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, fmt.Errorf("error reading cache file: %w", err)
	}

	var tokens ksef.Tokens
	if err = json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("error parsing cache file: %w", err)
	}
	slog.Debug("tokens loaded", "accessTokenExpiry", tokens.AccessTokenExpiry, "refreshTokenExpiry", tokens.RefreshTokenExpiry)
	slog.Log(nil, ksef.LevelSecret, "loaded tokens", "accessToken", tokens.AccessToken, "refreshToken", tokens.RefreshToken)

	return &tokens, nil
}

func SaveTokens(profileName string, tokens *ksef.Tokens, cacheDir string) error {
	slog.Debug("saving tokens to cache", "profile", profileName, "cacheDir", cacheDir)
	slog.Debug("token expiry", "accessTokenExpiry", tokens.AccessTokenExpiry, "refreshTokenExpiry", tokens.RefreshTokenExpiry)
	slog.Log(nil, ksef.LevelSecret, "tokens to save", "accessToken", tokens.AccessToken, "refreshToken", tokens.RefreshToken)

	cacheFile := filepath.Join(cacheDir, "profile_"+profileName+".json")
	data, err := json.Marshal(tokens)
	if err != nil {
		return fmt.Errorf("error encoding tokens: %w", err)
	}

	os.MkdirAll(cacheDir, 0700)
	if err = os.WriteFile(cacheFile, data, 0600); err != nil {
		return fmt.Errorf("error writing cache file: %w", err)
	}

	return nil
}
