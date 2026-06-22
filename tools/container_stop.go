package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
)

type ContainerStopTool struct{ r *Registry }

func NewContainerStopTool(r *Registry) *ContainerStopTool { return &ContainerStopTool{r: r} }
func (t *ContainerStopTool) Name() string                 { return "container_stop" }
func (t *ContainerStopTool) Description() string          { return "Stop a running container." }
func (t *ContainerStopTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"name": {"type": "string", "description": "Container name or ID"}
		},
		"required": ["name"]
	}`)
}

func (t *ContainerStopTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	name := gjson.ParseBytes(args).Get("name").String()
	if name == "" {
		return "", fmt.Errorf("missing required parameter: name")
	}
	if err := t.r.Docker.StopContainer(ctx, name); err != nil {
		return "", err
	}
	return fmt.Sprintf("stopped container %s", name), nil
}
