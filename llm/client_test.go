package llm

import (
	"testing"
)

func TestNormalizeBaseURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "https://api.openai.com/v1"},
		{"https://api.openai.com/v1/", "https://api.openai.com/v1"},
		{"https://integrate.api.nvidia.com", "https://integrate.api.nvidia.com/v1"},
		{"https://integrate.api.nvidia.com/", "https://integrate.api.nvidia.com/v1"},
		{"https://integrate.api.nvidia.com/v1", "https://integrate.api.nvidia.com/v1"},
		{"https://api.nvidia.com", "https://api.nvidia.com/v1"},
	}

	for _, tt := range tests {
		got := normalizeBaseURL(tt.input)
		if got != tt.expected {
			t.Errorf("normalizeBaseURL(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestMapMessagesToParam(t *testing.T) {
	messages := []Message{
		{Role: "system", Content: "System prompt"},
		{Role: "user", Content: "Hello"},
		{
			Role:    "assistant",
			Content: "",
			ToolCalls: []ToolCall{
				{ID: "call_123", Name: "docker_ps", Args: "{}"},
			},
		},
	}

	params := mapMessagesToParam(messages)
	if len(params) != 3 {
		t.Fatalf("expected 3 params, got %d", len(params))
	}
}
