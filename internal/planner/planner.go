package planner

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/luckywirasakti/jurnal/internal/config"
)

// CommitBatch groups files with their semantic commit message.
type CommitBatch struct {
	Files   []string `json:"files"`
	Message string   `json:"message"`
}

// StagingPlan defines branch renaming recommendation and commit batch plans.
type StagingPlan struct {
	ProposedBranch string        `json:"proposed_branch"`
	Commits        []CommitBatch `json:"commits"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatPayload struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
	Stream      bool          `json:"stream"`
}

type chatChoice struct {
	Message chatMessage `json:"message"`
}

type chatResponse struct {
	Choices []chatChoice `json:"choices"`
}

// CallAPI performs completion request to the OpenAI endpoint.
func CallAPI(systemPrompt, userPrompt string) (string, error) {
	cfg, err := config.Load()
	if err != nil {
		return "", err
	}
	if cfg.OpenAIBaseURL == "" {
		return "", errors.New("OpenAI Base URL is not configured")
	}

	url := fmt.Sprintf("%s/chat/completions", strings.TrimSuffix(cfg.OpenAIBaseURL, "/"))
	
	payload := chatPayload{
		Model: cfg.OpenAIModel,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.2,
		MaxTokens:   1500,
		Stream:      false,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	if cfg.OpenAIAPIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.OpenAIAPIKey))
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	var chatResp chatResponse
	err = json.NewDecoder(resp.Body).Decode(&chatResp)
	if err != nil {
		return "", err
	}

	if len(chatResp.Choices) == 0 {
		return "", errors.New("empty choices in API response")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// BuildStagingPlan requests LLM analysis of current working directory changes.
func BuildStagingPlan(gitStatus, gitDiff string) (*StagingPlan, error) {
	systemPrompt := `You are an expert developer workflow planner.
Review the unstaged git changes (status and diff).
Create a structured git staging plan in English.

Rules:
1. Propose a branch name (format: feat/xxx, fix/xxx, docs/xxx, refactor/xxx).
2. Group the modified files into logical batches (commits) so that each commit has a clear single purpose.
3. For each batch, define:
   - The files to include (relative paths only).
   - A concise, conventional commit message (format: <type>: <short summary>).
4. Output your plan STRICTLY in JSON format. Do not write markdown, do not write explanations.

JSON Output Schema:
{
  "proposed_branch": "feat/user-auth",
  "commits": [
    {
      "files": ["src/db.go", "migrations/001.sql"],
      "message": "feat: add user table and migration schema"
    },
    {
      "files": ["src/auth.go"],
      "message": "feat: implement login endpoint"
    }
  ]
}`

	userPrompt := fmt.Sprintf("Git Status:\n%s\n\nGit Diff:\n%s", gitStatus, gitDiff)
	rawPlan, err := CallAPI(systemPrompt, userPrompt)
	if err != nil {
		return nil, err
	}

	start := strings.Index(rawPlan, "{")
	end := strings.LastIndex(rawPlan, "}")
	if start == -1 || end == -1 || start >= end {
		return nil, fmt.Errorf("no valid JSON object found in response: %s", rawPlan)
	}
	cleaned := rawPlan[start : end+1]

	var plan StagingPlan
	err = json.Unmarshal([]byte(cleaned), &plan)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON plan: %w. Raw output was: %s", err, cleaned)
	}

	return &plan, nil
}
