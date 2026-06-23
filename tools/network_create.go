package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
)

type NetworkCreateTool struct{ r *Registry }

func NewNetworkCreateTool(r *Registry) *NetworkCreateTool {
	return &NetworkCreateTool{r: r}
}

func (t *NetworkCreateTool) Name() string {
	return "network_create"
}

func (t *NetworkCreateTool) Description() string {
	return "Create a Docker network."
}

func (t *NetworkCreateTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type":"object",
		"properties":{
			"name":{
				"type":"string",
				"description":"Network name"
			},
			"driver":{
				"type":"string",
				"description":"bridge, overlay, host, macvlan",
				"default":"bridge"
			}
		},
		"required":["name"]
	}`)
}

func (t *NetworkCreateTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	parsed := gjson.ParseBytes(args)

	name := parsed.Get("name").String()
	driver := parsed.Get("driver").String()

	if driver == "" {
		driver = "bridge"
	}

	id, err := t.r.Docker.CreateNetwork(ctx, name, driver)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		`{"id":"%s","name":"%s","driver":"%s"}`,
		id,
		name,
		driver,
	), nil
}
