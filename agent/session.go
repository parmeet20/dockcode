package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// SessionMeta holds session metadata stored in meta.json.
type SessionMeta struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	TokensIn  int       `json:"tokens_in"`
	TokensOut int       `json:"tokens_out"`
	Tags      []string  `json:"tags"`
}

// ChatEntry is a single message in the chat log.
type ChatEntry struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Session represents a conversation session with its backing files.
type Session struct {
	mu      sync.RWMutex
	ID      string
	Dir     string
	Meta    SessionMeta
	agentMD string
	chatLog []ChatEntry
	dirty   bool

	// auto-save ticker
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
}

// NewSession creates a new session directory, writes initial files, and starts auto-save.
func NewSession(parent context.Context, baseDir string) (*Session, error) {
	id := fmt.Sprintf("session-%d", time.Now().UnixNano())
	dir := filepath.Join(baseDir, id)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create session dir: %w", err)
	}

	ctx, cancel := context.WithCancel(parent)
	s := &Session{
		ID:  id,
		Dir: dir,
		Meta: SessionMeta{
			ID:        id,
			Title:     "New Session",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Tags:      []string{},
		},
		agentMD: fmt.Sprintf("# Agent Memory — %s\n\n## Context\nNew session.\n\n## Docker Status\nUnknown.\n", id),
		chatLog: []ChatEntry{},
		ctx:     ctx,
		cancel:  cancel,
		done:    make(chan struct{}),
	}

	if err := s.Save(); err != nil {
		cancel()
		return nil, err
	}

	go s.autoSave()
	return s, nil
}

// LoadSession loads an existing session from disk and starts auto-save.
func LoadSession(parent context.Context, dir string) (*Session, error) {
	ctx, cancel := context.WithCancel(parent)
	s := &Session{
		Dir:    dir,
		ctx:    ctx,
		cancel: cancel,
		done:   make(chan struct{}),
	}

	// Read meta.json
	metaData, err := os.ReadFile(filepath.Join(dir, "meta.json"))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to read meta.json: %w", err)
	}
	if err := json.Unmarshal(metaData, &s.Meta); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to parse meta.json: %w", err)
	}
	s.ID = s.Meta.ID

	// Read agent.md
	agentData, err := os.ReadFile(filepath.Join(dir, "agent.md"))
	if err == nil {
		s.agentMD = string(agentData)
	}

	// Read chat.md lines (simple format)
	chatData, err := os.ReadFile(filepath.Join(dir, "chat.md"))
	if err == nil {
		s.chatLog = parseChatMD(string(chatData))
	}

	go s.autoSave()
	return s, nil
}

// autoSave runs in background and saves every 30 seconds.
func (s *Session) autoSave() {
	defer close(s.done)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.mu.RLock()
			dirty := s.dirty
			s.mu.RUnlock()
			if dirty {
				_ = s.Save()
			}
		}
	}
}

// AppendChat appends a message to the chat log and marks session dirty.
func (s *Session) AppendChat(role, content string, toolCalls interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.chatLog = append(s.chatLog, ChatEntry{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
	s.Meta.UpdatedAt = time.Now()
	s.dirty = true
}

// UpdateAgentMD replaces the agent memory markdown.
func (s *Session) UpdateAgentMD(content string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.agentMD = content
	s.dirty = true
}

// GetAgentMD returns the current agent memory content.
func (s *Session) GetAgentMD() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.agentMD
}

// GetChatLog returns a copy of the current chat log.
func (s *Session) GetChatLog() []ChatEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ChatEntry, len(s.chatLog))
	copy(out, s.chatLog)
	return out
}

// SetTitle sets the session title.
func (s *Session) SetTitle(title string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Meta.Title = title
	s.Meta.UpdatedAt = time.Now()
	s.dirty = true
}

// GetMeta returns a copy of the session metadata.
func (s *Session) GetMeta() SessionMeta {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Meta
}

// Save writes all session files atomically.
func (s *Session) Save() error {
	s.mu.RLock()
	meta := s.Meta
	agentMD := s.agentMD
	chatLog := make([]ChatEntry, len(s.chatLog))
	copy(chatLog, s.chatLog)
	s.mu.RUnlock()

	// Write meta.json
	metaData, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	if err := atomicWrite(filepath.Join(s.Dir, "meta.json"), metaData); err != nil {
		return err
	}

	// Write agent.md
	if err := atomicWrite(filepath.Join(s.Dir, "agent.md"), []byte(agentMD)); err != nil {
		return err
	}

	// Write chat.md
	chatMD := formatChatMD(chatLog)
	if err := atomicWrite(filepath.Join(s.Dir, "chat.md"), []byte(chatMD)); err != nil {
		return err
	}

	s.mu.Lock()
	s.dirty = false
	s.mu.Unlock()

	return nil
}

// Stop cancels auto-save and waits for the goroutine to exit.
func (s *Session) Stop() {
	s.cancel()
	<-s.done
}

// atomicWrite writes to a .tmp file and renames it to the target path.
func atomicWrite(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// formatChatMD converts the chat log into a markdown-formatted string.
func formatChatMD(log []ChatEntry) string {
	var sb strings.Builder
	sb.WriteString("# Chat Log\n\n")
	for _, e := range log {
		sb.WriteString(fmt.Sprintf("## %s [%s]\n\n%s\n\n---\n\n",
			strings.ToUpper(e.Role), e.Timestamp.Format(time.RFC3339), e.Content))
	}
	return sb.String()
}

// parseChatMD reconstructs a minimal chat log from the markdown file.
func parseChatMD(content string) []ChatEntry {
	var log []ChatEntry
	blocks := strings.Split(content, "\n\n---\n\n")
	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" || !strings.HasPrefix(block, "## ") {
			continue
		}
		// Split header and body
		lines := strings.SplitN(block, "\n", 2)
		if len(lines) < 2 {
			continue
		}
		header := lines[0]
		body := strings.TrimSpace(lines[1])

		// Parse header: "## ROLE [TIMESTAMP]"
		header = strings.TrimPrefix(header, "## ")
		parts := strings.SplitN(header, " [", 2)
		role := strings.ToLower(strings.TrimSpace(parts[0]))

		var timestamp time.Time
		if len(parts) > 1 {
			timeStr := strings.TrimSuffix(parts[1], "]")
			if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
				timestamp = t
			} else {
				timestamp = time.Now()
			}
		} else {
			timestamp = time.Now()
		}

		log = append(log, ChatEntry{
			Role:      role,
			Content:   body,
			Timestamp: timestamp,
		})
	}
	return log
}
