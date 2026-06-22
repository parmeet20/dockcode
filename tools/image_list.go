package tools

import (
	"context"
	"encoding/json"
)

type ImageListTool struct {
	r *Registry
}

func NewImageListTool(r *Registry) *ImageListTool {
	return &ImageListTool{r: r}
}

func (t *ImageListTool) Name() string {
	return "image_list"
}

func (t *ImageListTool) Description() string {
	return "List all locally available Docker images."
}

func (t *ImageListTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {}
	}`)
}

func (t *ImageListTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	images, err := t.r.Docker.ListImages(ctx)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(images)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
