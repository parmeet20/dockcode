package tui

import (
	"os"
	"runtime"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func HasUnicodeSupport() bool {
	if runtime.GOOS != "windows" {
		return true
	}
	if os.Getenv("WT_SESSION") != "" || os.Getenv("TERM_PROGRAM") != "" {
		return true
	}
	term := strings.ToLower(os.Getenv("TERM"))
	if strings.Contains(term, "xterm") || strings.Contains(term, "256color") || term == "cygwin" {
		return true
	}
	return false
}

var (
	ColorPrimary = lipgloss.AdaptiveColor{Light: "#00AA44", Dark: "#00FF66"}
	ColorText    = lipgloss.AdaptiveColor{Light: "#0A2912", Dark: "#E0FFE9"}
	ColorDim     = lipgloss.AdaptiveColor{Light: "#2D663B", Dark: "#009944"}
	ColorBg      = lipgloss.AdaptiveColor{Light: "#F0FFF4", Dark: "#050B07"}
	ColorPanel   = lipgloss.AdaptiveColor{Light: "#DCF5E3", Dark: "#0D1C13"}
	ColorInput   = lipgloss.AdaptiveColor{Light: "#E6F7EB", Dark: "#12241A"}
	ColorSuccess = lipgloss.AdaptiveColor{Light: "#008833", Dark: "#00FF66"}
	ColorError   = lipgloss.AdaptiveColor{Light: "#CC0033", Dark: "#FF2A55"}
	ColorWarning = lipgloss.AdaptiveColor{Light: "#88A800", Dark: "#CCFF00"}
	ColorTool    = lipgloss.AdaptiveColor{Light: "#008B8B", Dark: "#00FFCC"}
	ColorBorder  = lipgloss.AdaptiveColor{Light: "#99CCAA", Dark: "#004D25"}
)

var (
	AppLogo     = "DockCode"
	IconPending = "◦"
	IconSuccess = "✓"
	IconError   = "✗"
	IconUser    = "▸ You"
	IconAgent   = "◈ Matrix"
	IconInfo    = "◆ Info"
	IconErrMsg  = "✖ Error"
	IconTool    = "⚙ Tool"

	SpinnerFrames = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}
)

func init() {
	if !HasUnicodeSupport() {
		AppLogo = "DockCode"
		IconPending = "-"
		IconSuccess = "[OK]"
		IconError = "[ERR]"
		IconUser = "> You"
		IconAgent = "* Matrix"
		IconInfo = "i Info"
		IconErrMsg = "! Error"
		IconTool = "# Tool"
		SpinnerFrames = []string{"-", "\\", "|", "/"}
	}
}

var (
	StyleBase = lipgloss.NewStyle().
			Foreground(ColorText)

	StyleDim = lipgloss.NewStyle().
			Foreground(ColorDim)

	StylePrimary = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	StyleSuccess = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	StyleError = lipgloss.NewStyle().
			Foreground(ColorError)

	StyleWarning = lipgloss.NewStyle().Foreground(ColorWarning)

	StyleTool = lipgloss.NewStyle().
			Foreground(ColorTool)

	StyleBold = lipgloss.NewStyle().
			Bold(true)

	StyleActiveBorder = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary)

	StyleInactiveBorder = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorBorder)

	StyleActiveTab = lipgloss.NewStyle().
				Background(ColorPrimary).
				Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#050B07"}).
				Bold(true).
				Padding(0, 1)

	StyleInactiveTab = lipgloss.NewStyle().
				Foreground(ColorDim).
				Padding(0, 1)

	StyleStatusBar = lipgloss.NewStyle().
			Background(ColorPanel).
			Foreground(ColorText).
			Padding(0, 1)

	StyleInput = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1)

	StyleInputFocused = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary).
				Padding(0, 1)

	StyleUserPrefix  = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
	StyleAgentPrefix = lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true)
	StyleInfoPrefix  = lipgloss.NewStyle().Foreground(ColorTool)
	StyleErrPrefix   = lipgloss.NewStyle().Foreground(ColorError).Bold(true)
	StyleToolPrefix  = lipgloss.NewStyle().Foreground(ColorTool).Bold(true)
)
var IsDarkMode = true

func ToggleTheme() {
	IsDarkMode = !IsDarkMode
}
