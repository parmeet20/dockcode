package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tidwall/gjson"
)

type DockerfileWriteTool struct{ r *Registry }

func NewDockerfileWriteTool(r *Registry) *DockerfileWriteTool { return &DockerfileWriteTool{r: r} }
func (t *DockerfileWriteTool) Name() string                   { return "dockerfile_write" }
func (t *DockerfileWriteTool) Description() string            { return "Write a Dockerfile to disk." }
func (t *DockerfileWriteTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path":    {"type": "string", "description": "File path to write the Dockerfile, e.g. ./Dockerfile"},
			"content": {"type": "string", "description": "Full Dockerfile content"}
		},
		"required": ["path", "content"]
	}`)
}

func (t *DockerfileWriteTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
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
		return "", fmt.Errorf("failed to write Dockerfile: %w", err)
	}
	return fmt.Sprintf("written to %s", path), nil
}
