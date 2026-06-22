package tui

import "time"

// Spinner tracks the current frame index for the animated spinner.
type Spinner struct {
	frame  int
	frames []string
}

// NewSpinner creates a spinner with the theme's custom frames.
func NewSpinner() Spinner {
	return Spinner{frames: SpinnerFrames}
}

// Tick advances the spinner to the next frame and returns a tea.Cmd for scheduling.
func (s *Spinner) Tick() {
	s.frame = (s.frame + 1) % len(s.frames)
}

// View returns the current spinner frame styled in ColorTool.
func (s Spinner) View() string {
	return StyleTool.Render(s.frames[s.frame])
}

// TickInterval is the delay between spinner frame advances.
const TickInterval = 80 * time.Millisecond
