package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
)

type ImageRemoveTool struct {
	r *Registry
}

func NewImageRemoveTool(r *Registry) *ImageRemoveTool {
	return &ImageRemoveTool{r: r}
}

func (t *ImageRemoveTool) Name() string        { return "image_remove" }
func (t *ImageRemoveTool) Description() string { return "Remove a local Docker image." }
func (t *ImageRemoveTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"image": {"type": "string", "description": "Image name or ID to remove"},
			"force": {"type": "boolean", "description": "Force remove even if used by containers"}
		},
		"required": ["image"]
	}`)
}

func (t *ImageRemoveTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	parsed := gjson.ParseBytes(args)
	imageName := parsed.Get("image").String()
	if imageName == "" {
		return "", fmt.Errorf("missing required parameter: image")
	}
	force := parsed.Get("force").Bool()
	if err := t.r.Docker.RemoveImage(ctx, imageName, force); err != nil {
		return "", err
	}
	return fmt.Sprintf("removed image %s", imageName), nil
}
