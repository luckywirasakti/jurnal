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

	if cfg.OpenAIBaseURL != "" && cfg.OpenAIAPIKey != "" {
		if cfg.OpenAIModel == "" {
			cfg.OpenAIModel = "gpt-4o-mini"
		}
		return cfg, nil
	}

	path, err := getPath()
	if err != nil {
		return cfg, err
	}

	file, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}

	err = json.Unmarshal(file, &cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

// Save writes configuration securely to the local store.
func Save(baseURL, apiKey, model string) error {
	cfg := Schema{
		OpenAIBaseURL: baseURL,
		OpenAIAPIKey:  apiKey,
		OpenAIModel:   model,
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
