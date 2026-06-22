package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
)

type ContainerLogsTool struct{ r *Registry }

func NewContainerLogsTool(r *Registry) *ContainerLogsTool { return &ContainerLogsTool{r: r} }
func (t *ContainerLogsTool) Name() string                 { return "container_logs" }
func (t *ContainerLogsTool) Description() string          { return "Get container logs." }
func (t *ContainerLogsTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"name": {"type": "string", "description": "Container name or ID"},
			"tail": {"type": "integer", "description": "Number of log lines to return (default 50)"}
		},
		"required": ["name"]
	}`)
}

func (t *ContainerLogsTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	parsed := gjson.ParseBytes(args)
	name := parsed.Get("name").String()
	if name == "" {
		return "", fmt.Errorf("missing required parameter: name")
	}
	tail := int(parsed.Get("tail").Int())
	if tail <= 0 {
		tail = 50
	}
	logs, err := t.r.Docker.GetContainerLogs(ctx, name, tail)
	if err != nil {
		return "", err
	}
	return logs, nil
}
