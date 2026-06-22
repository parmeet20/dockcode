package tools

import (
	"context"
	"encoding/json"
)

type VolumeListTool struct{ r *Registry }

func NewVolumeListTool(r *Registry) *VolumeListTool { return &VolumeListTool{r: r} }
func (t *VolumeListTool) Name() string              { return "volume_list" }
func (t *VolumeListTool) Description() string       { return "List Docker volumes." }
func (t *VolumeListTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{}}`)
}

func (t *VolumeListTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	volumes, err := t.r.Docker.ListVolumes(ctx)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(volumes)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
