package agent

import (
	"testing"
)

func TestParseChatMD(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []ChatEntry
	}{
		{
			name: "Simple chat",
			content: `# Chat Log

## USER [2026-06-23T12:00:00Z]

hello

---

## ASSISTANT [2026-06-23T12:00:01Z]

hi

---

`,
			want: []ChatEntry{
				{
					Role:    "user",
					Content: "hello",
				},
				{
					Role:    "assistant",
					Content: "hi",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseChatMD(tt.content)
			if len(got) != len(tt.want) {
				t.Fatalf("parseChatMD() got %d entries, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i].Role != tt.want[i].Role {
					t.Errorf("got[%d].Role = %q, want %q", i, got[i].Role, tt.want[i].Role)
				}
				if got[i].Content != tt.want[i].Content {
					t.Errorf("got[%d].Content = %q, want %q", i, got[i].Content, tt.want[i].Content)
				}
			}
		})
	}
}

func TestParseChatMD_WindowsLineEndings(t *testing.T) {
	content := "# Chat Log\r\n\r\n## USER [2026-06-23T12:00:00Z]\r\n\r\nhello\r\n\r\n---\r\n\r\n## ASSISTANT [2026-06-23T12:00:01Z]\r\n\r\nhi\r\n\r\n---\r\n\r\n"
	got := parseChatMD(content)
	if len(got) != 2 {
		t.Fatalf("Windows line endings: expected 2 entries, got %d. Entries: %+v", len(got), got)
	}
}
