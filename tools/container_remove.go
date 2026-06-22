package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
)

type ContainerRemoveTool struct{ r *Registry }

func NewContainerRemoveTool(r *Registry) *ContainerRemoveTool { return &ContainerRemoveTool{r: r} }
func (t *ContainerRemoveTool) Name() string                   { return "container_remove" }
func (t *ContainerRemoveTool) Description() string            { return "Remove a container." }
func (t *ContainerRemoveTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"name":  {"type": "string", "description": "Container name or ID"},
			"force": {"type": "boolean", "description": "Force remove even if running"}
		},
		"required": ["name"]
	}`)
}

func (t *ContainerRemoveTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	parsed := gjson.ParseBytes(args)
	name := parsed.Get("name").String()
	if name == "" {
		return "", fmt.Errorf("missing required parameter: name")
	}
	force := parsed.Get("force").Bool()
	if err := t.r.Docker.RemoveContainer(ctx, name, force); err != nil {
		return "", err
	}
	return fmt.Sprintf("removed container %s", name), nil
}
