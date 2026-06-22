package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/parmeet20/dockercode/concurrency"
	"github.com/parmeet20/dockercode/docker"
	"github.com/parmeet20/dockercode/llm"
	"github.com/parmeet20/dockercode/tools"
)

// ─── TUI message types ────────────────────────────────────────────────────────

// AgentChunkMsg carries a streamed text chunk to the TUI.
type AgentChunkMsg struct{ Text string }

// AgentDoneMsg signals that the agent finished its full turn.
type AgentDoneMsg struct{ Err error }

// ToolStartMsg signals that a tool call has begun.
type ToolStartMsg struct {
	Name string
	Args string
}

// ToolDoneMsg signals that a tool call finished.
type ToolDoneMsg struct {
	Name   string
	Result string
	Err    error
}

// AskUserMsg is sent by the agent to the TUI to surface a user question.
type AskUserMsg struct {
	Question string
	Options  []string
	Fields   []tools.AskUserField
}

// AskUserReply is the TUI's answer back to the blocked agent goroutine.
type AskUserReply struct {
	Answer map[string]string
}

// SidebarRefreshMsg carries fresh Docker state for the sidebar.
type SidebarRefreshMsg struct {
	Containers []docker.Container
	Images     []docker.Image
	Volumes    []docker.Volume
	Networks   []docker.Network
}

// ─── Agent ────────────────────────────────────────────────────────────────────

// Agent manages the LLM chat loop and coordinates with the TUI.
type Agent struct {
	ctx    context.Context
	cancel context.CancelFunc

	program    *tea.Program
	llm        *llm.Client
	docker     *docker.Client
	session    *Session
	memory     *Memory
	tools      *tools.Registry
	supervisor *concurrency.Supervisor

	agentBusy  *atomic.Bool
	tokenCount *atomic.Int64

	mu         sync.Mutex
	askReplyCh chan AskUserReply // non-nil only when blocking on ask_user
}

// NewAgent constructs a new Agent under the given parent context.
func NewAgent(
	parent context.Context,
	llmClient *llm.Client,
	dockerClient *docker.Client,
	session *Session,
	supervisor *concurrency.Supervisor,
	agentBusy *atomic.Bool,
	tokenCount *atomic.Int64,
) *Agent {
	ctx, cancel := context.WithCancel(parent)
	a := &Agent{
		ctx:        ctx,
		cancel:     cancel,
		llm:        llmClient,
		docker:     dockerClient,
		session:    session,
		memory:     NewMemory(session),
		supervisor: supervisor,
		agentBusy:  agentBusy,
		tokenCount: tokenCount,
	}

	// Build tool registry
	reg := tools.NewRegistry(dockerClient)
	a.tools = reg

	// Register all tools
	askUserTool := tools.NewAskUserTool(reg, a.askUser)
	reg.Register(tools.NewDockerCheckTool(reg))
	reg.Register(tools.NewImageListTool(reg))
	reg.Register(tools.NewImagePullTool(reg))
	reg.Register(tools.NewImageRemoveTool(reg))
	reg.Register(tools.NewContainerListTool(reg))
	reg.Register(tools.NewContainerRunTool(reg))
	reg.Register(tools.NewContainerStopTool(reg))
	reg.Register(tools.NewContainerRemoveTool(reg))
	reg.Register(tools.NewContainerLogsTool(reg))
	reg.Register(tools.NewContainerExecTool(reg))
	reg.Register(tools.NewContainerInspectTool(reg))
	reg.Register(tools.NewDockerfileWriteTool(reg))
	reg.Register(tools.NewDockerfileBuildTool(reg))
	reg.Register(tools.NewComposeWriteTool(reg))
	reg.Register(tools.NewNetworkListTool(reg))
	reg.Register(tools.NewVolumeListTool(reg))
	reg.Register(askUserTool)

	return a
}

// SetProgram wires the Bubbletea program for TUI messaging.
func (a *Agent) SetProgram(p *tea.Program) {
	a.program = p
	a.llm.SetProgram(p)
}

// Run executes the full agent loop for one user message.
// It must be called from within a tea.Cmd goroutine.
func (a *Agent) Run(userMsg string) {
	a.agentBusy.Store(true)
	defer a.agentBusy.Store(false)

	// Record user message
	a.session.AppendChat("user", userMsg, nil)

	// ─── Validate Credentials before calling LLM ──────────────────────────────────
	pingCtx, pingCancel := context.WithTimeout(a.ctx, 5*time.Second)
	if err := a.llm.ValidateCredentials(pingCtx); err != nil {
		pingCancel()
		a.program.Send(AgentDoneMsg{
			Err: fmt.Errorf("LLM API is unreachable or misconfigured. Please check your settings with /config.\nError: %s", err.Error()),
		})
		return
	}
	pingCancel()

	// Build context from session history
	history := a.session.GetChatLog()
	messages := BuildContext(a.memory, history[:len(history)-1], userMsg)
	schemas := a.tools.Schemas()

	isFirst := len(history) == 1

	for {
		// Check cancellation before LLM call
		select {
		case <-a.ctx.Done():
			a.program.Send(AgentDoneMsg{Err: a.ctx.Err()})
			return
		default:
		}

		deltaCh := a.llm.ChatStream(a.ctx, messages, schemas)

		dispatcher := NewStreamDispatcher(
			// onText
			func(text string) {
				a.program.Send(AgentChunkMsg{Text: text})
			},
			// onToolStart
			func(id, name string) {
				a.program.Send(ToolStartMsg{Name: name})
			},
			// onToolChunk — accumulate silently
			nil,
			// onToolEnd — no-op here, tool dispatch notifies separately
			nil,
		)

		result := dispatcher.Run(a.ctx, deltaCh)

		if result.Error != nil {
			a.program.Send(AgentDoneMsg{Err: result.Error})
			return
		}

		// Update token count
		a.tokenCount.Add(int64(llm.EstimateTokens(result.FullText)))

		// No tool calls → final assistant response
		if len(result.ToolCalls) == 0 {
			a.session.AppendChat("assistant", result.FullText, nil)
			a.program.Send(AgentDoneMsg{})

			// Kick off auto-title after first response
			if isFirst {
				go GenerateTitle(a.supervisor, a.ctx, a.llm, a.session, userMsg)
			}

			// Post-run sidebar refresh (best-effort, one-shot)
			go a.refreshSidebar()
			return
		}

		// Execute all tool calls
		toolResults := make([]llm.ToolResult, 0, len(result.ToolCalls))
		for _, tc := range result.ToolCalls {
			a.program.Send(ToolStartMsg{Name: tc.Name, Args: tc.Args})

			var argsRaw json.RawMessage
			if tc.Args != "" {
				argsRaw = json.RawMessage(tc.Args)
			} else {
				argsRaw = json.RawMessage(`{}`)
			}

			toolResult, err := a.tools.Dispatch(a.ctx, tc.Name, argsRaw)
			if err != nil {
				a.program.Send(ToolDoneMsg{Name: tc.Name, Err: err})
				toolResult = fmt.Sprintf("error: %s", err.Error())
			} else {
				a.program.Send(ToolDoneMsg{Name: tc.Name, Result: toolResult})
			}

			toolResults = append(toolResults, llm.ToolResult{
				ToolCallID: tc.ID,
				Content:    toolResult,
			})
		}

		// Append assistant + tool results and loop
		messages = AppendToolRound(messages, result.FullText, result.ToolCalls, toolResults)

		// Check cancellation before next iteration
		select {
		case <-a.ctx.Done():
			a.program.Send(AgentDoneMsg{Err: a.ctx.Err()})
			return
		default:
		}
	}
}

// askUser blocks the agent goroutine until the TUI replies or context is cancelled.
func (a *Agent) askUser(ctx context.Context, args tools.AskUserArgs) (map[string]string, error) {
	ch := make(chan AskUserReply, 1)
	a.mu.Lock()
	a.askReplyCh = ch
	a.mu.Unlock()

	a.program.Send(AskUserMsg{
		Question: args.Question,
		Options:  args.Options,
		Fields:   args.Fields,
	})

	select {
	case reply := <-ch:
		return reply.Answer, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// SubmitAskUserReply is called by the TUI when the user answers an ask_user prompt.
func (a *Agent) SubmitAskUserReply(reply AskUserReply) {
	a.mu.Lock()
	ch := a.askReplyCh
	a.askReplyCh = nil
	a.mu.Unlock()

	if ch != nil {
		select {
		case ch <- reply:
		default:
		}
	}
}

// refreshSidebar does a one-shot refresh of all Docker state after agent completes.
func (a *Agent) refreshSidebar() {
	ctx, cancel := context.WithCancel(a.ctx)
	defer cancel()

	var (
		containers []docker.Container
		images     []docker.Image
		volumes    []docker.Volume
		networks   []docker.Network
	)

	_ = concurrency.RunGroup(ctx, 2_000_000_000, // 2 second timeout
		func(ctx context.Context) error {
			var e error
			containers, e = a.docker.ListContainers(ctx, true)
			return e
		},
		func(ctx context.Context) error {
			var e error
			images, e = a.docker.ListImages(ctx)
			return e
		},
		func(ctx context.Context) error {
			var e error
			volumes, e = a.docker.ListVolumes(ctx)
			return e
		},
		func(ctx context.Context) error {
			var e error
			networks, e = a.docker.ListNetworks(ctx)
			return e
		},
	)

	if a.program != nil {
		a.program.Send(SidebarRefreshMsg{
			Containers: containers,
			Images:     images,
			Volumes:    volumes,
			Networks:   networks,
		})
	}
}

// Stop cancels the agent context. The run goroutine is managed by bubbletea, not us.
func (a *Agent) Stop() {
	a.cancel()
}
