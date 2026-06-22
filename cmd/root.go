package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/parmeet20/dockercode/agent"
	"github.com/parmeet20/dockercode/concurrency"
	"github.com/parmeet20/dockercode/config"
	"github.com/parmeet20/dockercode/docker"
	"github.com/parmeet20/dockercode/llm"
	"github.com/parmeet20/dockercode/tui"
)

var rootCmd = &cobra.Command{
	Use:   "dockercode",
	Short: "🐳 DockerCode — AI-powered Docker management TUI",
	RunE: func(cmd *cobra.Command, args []string) error {
		return startApp(cmd.Context())
	},
}

// Execute starts the command-line parser under the given context.
func Execute(ctx context.Context) error {
	_ = EnsureGoBinInPath()
	return rootCmd.ExecuteContext(ctx)
}

func startApp(ctx context.Context) error {
	// ── Config ────────────────────────────────────────────────────────────────
	cfgManager, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	// ── Onboarding ────────────────────────────────────────────────────────────
	if !cfgManager.ConfigExists() {
		return runOnboarding(ctx, cfgManager)
	}

	if err := cfgManager.Load(); err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	cfg := cfgManager.Get()

	// ── Docker client ─────────────────────────────────────────────────────────
	dockerClient, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("docker: %w", err)
	}

	// ── LLM client ────────────────────────────────────────────────────────────
	llmClient := llm.NewClient(cfg.APIURL, cfg.APIToken, cfg.Model)

	// ── Sessions ──────────────────────────────────────────────────────────────
	home, _ := os.UserHomeDir()
	sessionsDir := filepath.Join(home, ".dockercode", "sessions")
	_ = os.MkdirAll(sessionsDir, 0755)

	sessIdx, err := agent.NewSessionIndex(sessionsDir)
	if err != nil {
		return fmt.Errorf("session index: %w", err)
	}

	sess, err := agent.NewSession(ctx, sessionsDir)
	if err != nil {
		return fmt.Errorf("session: %w", err)
	}

	// Update session index with new session
	_ = sessIdx.Upsert(agent.SessionSummary{
		ID:        sess.ID,
		Title:     "New Session",
		UpdatedAt: time.Now().Format(time.RFC3339),
	})

	// ── Supervisor ────────────────────────────────────────────────────────────
	sup := concurrency.NewSupervisor()

	// ── App context (child of root, for graceful shutdown) ────────────────────
	appCtx, appCancel := context.WithCancel(ctx)

	// ── Build TUI model ───────────────────────────────────────────────────────
	m := tui.NewModel(appCtx, appCancel, cfgManager, dockerClient, llmClient, sess, sessIdx, sup)

	// ── Run Bubbletea ─────────────────────────────────────────────────────────
	p := tea.NewProgram(
		&m,
		tea.WithAltScreen(),
	)

	// Wire program reference so goroutines can send messages
	m.SetProgram(p)

	// Suppress unused variable warning — atomic used by sidebar refresher
	var _ atomic.Bool

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui: %w", err)
	}

	return nil
}

func runOnboarding(ctx context.Context, cfgManager *config.Manager) error {
	m := tui.NewOnboardingModel(ctx, cfgManager)
	p := tea.NewProgram(m, tea.WithAltScreen())

	_, err := p.Run()
	if err != nil {
		return fmt.Errorf("onboarding: %w", err)
	}

	// If config was saved by the wizard, launch the main app.
	if cfgManager.ConfigExists() {
		return startApp(ctx)
	}
	return nil
}
