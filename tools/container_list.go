package tools

import (
	"context"
	"encoding/json"

	"github.com/tidwall/gjson"
)

type ContainerListTool struct {
	r *Registry
}

func NewContainerListTool(r *Registry) *ContainerListTool {
	return &ContainerListTool{r: r}
}

func (t *ContainerListTool) Name() string        { return "container_list" }
func (t *ContainerListTool) Description() string { return "List Docker containers." }
func (t *ContainerListTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"all": {"type": "boolean", "description": "If true, include stopped containers"}
		}
	}`)
}

func (t *ContainerListTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	parsed := gjson.ParseBytes(args)
	all := parsed.Get("all").Bool()
	containers, err := t.r.Docker.ListContainers(ctx, all)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(containers)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
