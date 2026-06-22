package tools

import (
	"context"
	"encoding/json"
)

type DockerCheckTool struct {
	r *Registry
}

func NewDockerCheckTool(r *Registry) *DockerCheckTool {
	return &DockerCheckTool{r: r}
}

func (t *DockerCheckTool) Name() string {
	return "docker_status"
}

func (t *DockerCheckTool) Description() string {
	return "Check if Docker daemon is running. Always call this first."
}

func (t *DockerCheckTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {}
	}`)
}

func (t *DockerCheckTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	err := t.r.Docker.Ping(ctx)
	if err != nil {
		return "stopped", nil
	}
	return "running", nil
}
