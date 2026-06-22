package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
)

type ContainerInspectTool struct{ r *Registry }

func NewContainerInspectTool(r *Registry) *ContainerInspectTool {
	return &ContainerInspectTool{r: r}
}
func (t *ContainerInspectTool) Name() string        { return "container_inspect" }
func (t *ContainerInspectTool) Description() string { return "Get full details of a container." }
func (t *ContainerInspectTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"name": {"type": "string", "description": "Container name or ID"}
		},
		"required": ["name"]
	}`)
}

func (t *ContainerInspectTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	name := gjson.ParseBytes(args).Get("name").String()
	if name == "" {
		return "", fmt.Errorf("missing required parameter: name")
	}
	return t.r.Docker.InspectContainer(ctx, name)
}
