package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/parmeet20/dockcode/agent"
)

// SessionBrowserMsg is sent when the user selects a session to open.
type SessionBrowserMsg struct{ SessionID string }

// SessionBrowserCloseMsg is sent when the user dismisses the browser.
type SessionBrowserCloseMsg struct{}

// SessionBrowserModel is the full-screen session browser.
type SessionBrowserModel struct {
	index    *agent.SessionIndex
	sessions []agent.SessionSummary
	selected int
	search   textinput.Model
	width    int
	height   int
}

// NewSessionBrowserModel creates the browser.
func NewSessionBrowserModel(index *agent.SessionIndex) SessionBrowserModel {
	si := textinput.New()
	si.Placeholder = "Search sessions..."
	si.Width = 50
	si.Focus()

	m := SessionBrowserModel{
		index:  index,
		search: si,
	}
	m.reload()
	return m
}

func (m *SessionBrowserModel) reload() {
	all := m.index.List()
	q := strings.ToLower(m.search.Value())
	if q == "" {
		m.sessions = all
	} else {
		var filtered []agent.SessionSummary
		for _, s := range all {
			if strings.Contains(strings.ToLower(s.Title), q) {
				filtered = append(filtered, s)
			}
		}
		m.sessions = filtered
	}
	if m.selected >= len(m.sessions) {
		m.selected = len(m.sessions) - 1
	}
	if m.selected < 0 {
		m.selected = 0
	}
}

func (m SessionBrowserModel) Init() tea.Cmd { return nil }

func (m SessionBrowserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, func() tea.Msg { return SessionBrowserCloseMsg{} }

		case "q":
			if m.search.Value() == "" {
				return m, func() tea.Msg { return SessionBrowserCloseMsg{} }
			}

		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}

		case "down", "j":
			if m.selected < len(m.sessions)-1 {
				m.selected++
			}

		case "enter":
			if len(m.sessions) > 0 {
				id := m.sessions[m.selected].ID
				return m, func() tea.Msg { return SessionBrowserMsg{SessionID: id} }
			}

		case "n":
			return m, func() tea.Msg { return SessionBrowserMsg{SessionID: "new"} }

		default:
			var cmd tea.Cmd
			m.search, cmd = m.search.Update(msg)
			m.reload()
			return m, cmd
		}
	}
	return m, nil
}

func (m SessionBrowserModel) View() string {
	browserTitle := "📋  Sessions"
	if !HasUnicodeSupport() {
		browserTitle = "Sessions"
	}
	header := StylePrimary.Render(browserTitle) + "  " +
		StyleDim.Render("Enter=open  N=new  Q=back")

	searchBox := StyleInputFocused.Width(60).Render(m.search.View())

	var rows []string
	rows = append(rows, StyleBold.Render(
		fmt.Sprintf("  %-40s  %-12s  %s", "Title", "Updated", "Tokens"),
	))
	rows = append(rows, StyleDim.Render(strings.Repeat("─", 70)))

	for i, s := range m.sessions {
		title := s.Title
		if len(title) > 40 {
			title = title[:37] + "..."
		}
		date := s.UpdatedAt
		if len(date) > 10 {
			date = date[:10]
		}
		tokens := fmt.Sprintf("%d", s.TokensIn+s.TokensOut)
		row := fmt.Sprintf("  %-40s  %-12s  %s", title, date, tokens)
		arrow := "▸ "
		if !HasUnicodeSupport() {
			arrow = "> "
		}
		if i == m.selected {
			rows = append(rows, lipgloss.NewStyle().
				Foreground(ColorDim).
				Bold(true).
				Render(arrow+row))
		} else {
			rows = append(rows, StyleBase.Render("  "+row))
		}
	}

	if len(m.sessions) == 0 {
		rows = append(rows, StyleDim.Render("  No sessions found."))
	}

	body := strings.Join(rows, "\n")
	content := header + "\n\n" + searchBox + "\n\n" + body

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorDim).
		Padding(1, 2).
		Width(m.width - 4).
		Height(m.height - 4).
		Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
