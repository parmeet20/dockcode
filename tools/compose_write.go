package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tidwall/gjson"
)

type ComposeWriteTool struct{ r *Registry }

func NewComposeWriteTool(r *Registry) *ComposeWriteTool { return &ComposeWriteTool{r: r} }
func (t *ComposeWriteTool) Name() string                { return "compose_write" }
func (t *ComposeWriteTool) Description() string         { return "Write a docker-compose.yml to disk." }
func (t *ComposeWriteTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path":    {"type": "string", "description": "File path to write, e.g. ./docker-compose.yml"},
			"content": {"type": "string", "description": "Full docker-compose YAML content"}
		},
		"required": ["path", "content"]
	}`)
}

func (t *ComposeWriteTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	parsed := gjson.ParseBytes(args)
	path := parsed.Get("path").String()
	content := parsed.Get("content").String()
	if path == "" || content == "" {
		return "", fmt.Errorf("missing required parameters: path and content")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", fmt.Errorf("failed to create directories: %w", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write compose file: %w", err)
	}
	return fmt.Sprintf("written to %s", path), nil
}
