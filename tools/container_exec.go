package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
)

type ContainerExecTool struct{ r *Registry }

func NewContainerExecTool(r *Registry) *ContainerExecTool { return &ContainerExecTool{r: r} }
func (t *ContainerExecTool) Name() string                 { return "container_exec" }
func (t *ContainerExecTool) Description() string {
	return "Execute a command inside a running container."
}
func (t *ContainerExecTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"name":    {"type": "string", "description": "Container name or ID"},
			"command": {"type": "string", "description": "Command to execute inside the container"}
		},
		"required": ["name", "command"]
	}`)
}

func (t *ContainerExecTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	parsed := gjson.ParseBytes(args)
	name := parsed.Get("name").String()
	cmd := parsed.Get("command").String()
	if name == "" || cmd == "" {
		return "", fmt.Errorf("missing required parameters: name and command")
	}
	return t.r.Docker.ExecCommand(ctx, name, cmd)
}
