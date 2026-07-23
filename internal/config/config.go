package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Schema defines the structure of Jurnal configuration.
type Schema struct {
	OpenAIBaseURL string `json:"openai_base_url"`
	OpenAIAPIKey  string `json:"openai_api_key"`
	OpenAIModel   string `json:"openai_model"`
	DefaultBranch string `json:"default_branch,omitempty"`
}

func getPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".jurnal.json"), nil
}

// Load retrieves configuration from environment variables or local store.
func Load() (Schema, error) {
	var cfg Schema

	// Prioritize environment variables for non-interactive environments (CI, scripts)
	cfg.OpenAIBaseURL = os.Getenv("OPENAI_BASE_URL")
	cfg.OpenAIAPIKey = os.Getenv("OPENAI_API_KEY")
	cfg.OpenAIModel = os.Getenv("OPENAI_MODEL")
	cfg.DefaultBranch = os.Getenv("JURNAL_DEFAULT_BRANCH")

	if cfg.OpenAIBaseURL == "" || cfg.OpenAIAPIKey == "" {
		if path, err := getPath(); err == nil {
			if file, err := os.ReadFile(path); err == nil {
				_ = json.Unmarshal(file, &cfg)
			}
		}
	}

	if cfg.OpenAIModel == "" {
		cfg.OpenAIModel = "gpt-4o-mini"
	}
	if cfg.DefaultBranch == "" {
		cfg.DefaultBranch = "main"
	}

	return cfg, nil
}

// Save writes configuration securely to the local store.
func Save(baseURL, apiKey, model, defaultBranch string) error {
	if defaultBranch == "" {
		defaultBranch = "main"
	}
	cfg := Schema{
		OpenAIBaseURL: baseURL,
		OpenAIAPIKey:  apiKey,
		OpenAIModel:   model,
		DefaultBranch: defaultBranch,
	}

	path, err := getPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
