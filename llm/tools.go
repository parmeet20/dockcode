package llm

import (
	"encoding/json"

	"github.com/openai/openai-go"
)

// MapToolToParam converts an internal tool's name, description, and JSON schema
// to the official openai.ChatCompletionToolParam required by the OpenAI Go SDK.
func MapToolToParam(name, desc string, schema json.RawMessage) openai.ChatCompletionToolParam {
	var params map[string]interface{}
	if len(schema) > 0 {
		_ = json.Unmarshal(schema, &params)
	}
	if params == nil {
		params = map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		}
	}

	return openai.ChatCompletionToolParam{
		Type: "function",
		Function: openai.FunctionDefinitionParam{
			Name:        name,
			Description: openai.String(desc),
			Parameters:  openai.FunctionParameters(params),
		},
	}
}

