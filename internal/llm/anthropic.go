// Package llm wraps an LLM behind a tiny interface.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	apiURL       = "https://api.anthropic.com/v1/messages"
	apiVersion   = "2023-06-01"
	defaultModel = "claude-sonnet-4-20250514"
)

const systemPrompt = `You are a task-planning assistant. The user provides text extracted from a document (plan, spec, meeting notes, onboarding doc, etc.).

Your job: turn it into actionable tasks organized under one or more PROJECTS.

Output rules — follow EXACTLY:
- Output ONLY GitHub-flavored Markdown. No preamble, no code fences, no closing remarks.
- First line: "# Tasks" (or a short descriptive H1 if the document implies a clear title).
- For each distinct initiative or theme, add a section:
  ## Project: <short name>
  <!-- status:active source:import created:YYYY-MM-DD -->
- Use ### <group name> only when sub-grouping helps (e.g. "Backlog", "Week 1", "Day 1"). Skip empty groups.
- Optional focus line under a group: "> Focus: <one sentence>"
- Each task: "- [ ] <verb-first actionable task>"
- Prefer concrete, delegable tasks. If the document is a day-by-day timeline, use day/week groups; otherwise use thematic groups.
- Do NOT add per-task HTML comments, IDs, or metadata — only project-level comments as shown.
- Create a separate project when topics are clearly unrelated (e.g. "HubSpot cleanup" vs "Q2 onboarding").`

// Client talks to the Anthropic Messages API.
type Client struct {
	APIKey string
	Model  string
	HTTP   *http.Client
}

// New builds a Client from the environment.
func New() (*Client, error) {
	key := os.Getenv("ANTHROPIC_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY is not set")
	}
	model := os.Getenv("ANTHROPIC_MODEL")
	if model == "" {
		model = defaultModel
	}
	return &Client{
		APIKey: key,
		Model:  model,
		HTTP:   &http.Client{Timeout: 3 * time.Minute},
	}, nil
}

type apiRequest struct {
	Model     string       `json:"model"`
	MaxTokens int          `json:"max_tokens"`
	System    string       `json:"system,omitempty"`
	Messages  []apiMessage `json:"messages"`
}

type apiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type apiResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// GenerateFromDocument returns markdown in the store's project format.
func (c *Client) GenerateFromDocument(ctx context.Context, docText, today string) (string, error) {
	user := fmt.Sprintf(
		"Today is %s.\n\nDocument text:\n\n<document>\n%s\n</document>\n\nGenerate projects and tasks now.",
		today, docText,
	)

	body, err := json.Marshal(apiRequest{
		Model:     c.Model,
		MaxTokens: 8192,
		System:    systemPrompt,
		Messages:  []apiMessage{{Role: "user", Content: user}},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("anthropic-version", apiVersion)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var out apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		if out.Error != nil {
			return "", fmt.Errorf("api error (%s): %s", out.Error.Type, out.Error.Message)
		}
		return "", fmt.Errorf("api returned status %d", resp.StatusCode)
	}

	var sb bytes.Buffer
	for _, block := range out.Content {
		if block.Type == "text" {
			sb.WriteString(block.Text)
		}
	}
	if sb.Len() == 0 {
		return "", fmt.Errorf("model returned no text content")
	}
	return sb.String(), nil
}

// GeneratePlan is deprecated; use GenerateFromDocument.
func (c *Client) GeneratePlan(ctx context.Context, docText, today string) (string, error) {
	return c.GenerateFromDocument(ctx, docText, today)
}
