package tools

import (
	"context"
	"encoding/json"
)

type NetworkListTool struct{ r *Registry }

func NewNetworkListTool(r *Registry) *NetworkListTool { return &NetworkListTool{r: r} }
func (t *NetworkListTool) Name() string               { return "network_list" }
func (t *NetworkListTool) Description() string        { return "List Docker networks." }
func (t *NetworkListTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{}}`)
}

func (t *NetworkListTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	networks, err := t.r.Docker.ListNetworks(ctx)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(networks)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
