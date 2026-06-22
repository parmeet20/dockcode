package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
)

type ContainerRunTool struct{ r *Registry }

func NewContainerRunTool(r *Registry) *ContainerRunTool { return &ContainerRunTool{r: r} }
func (t *ContainerRunTool) Name() string                { return "container_run" }
func (t *ContainerRunTool) Description() string {
	return "Create and start a Docker container."
}
func (t *ContainerRunTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"image":   {"type": "string", "description": "Docker image to run"},
			"name":    {"type": "string", "description": "Container name"},
			"ports":   {"type": "array", "items": {"type": "string"}, "description": "Port mappings e.g. [\"80:80\"]"},
			"env":     {"type": "object", "description": "Environment variables as key-value pairs"},
			"volumes": {"type": "array", "items": {"type": "string"}, "description": "Volume mounts e.g. [\"/host:/container\"]"},
			"restart": {"type": "string", "enum": ["no","always","unless-stopped","on-failure"], "description": "Restart policy"},
			"detach":  {"type": "boolean", "description": "Run in background (always true)"}
		},
		"required": ["image"]
	}`)
}

func (t *ContainerRunTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	parsed := gjson.ParseBytes(args)

	image := parsed.Get("image").String()
	if image == "" {
		return "", fmt.Errorf("missing required parameter: image")
	}
	name := parsed.Get("name").String()
	restart := parsed.Get("restart").String()
	if restart == "" {
		restart = "no"
	}

	var ports []string
	for _, p := range parsed.Get("ports").Array() {
		ports = append(ports, p.String())
	}

	env := map[string]string{}
	parsed.Get("env").ForEach(func(k, v gjson.Result) bool {
		env[k.String()] = v.String()
		return true
	})

	var volumes []string
	for _, v := range parsed.Get("volumes").Array() {
		volumes = append(volumes, v.String())
	}

	id, err := t.r.Docker.RunContainer(ctx, image, name, ports, env, volumes, restart, true)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("started container %s (id: %s)", name, id[:12]), nil
}
