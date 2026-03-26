package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	doneStateLabel    = "DONE"
	filterLabelFormat = "%s  filter:%q"
	liveStateLabel    = "LIVE"
	loadingText       = "Loading Peacock..."
	pausedStateFormat = "PAUSED (+%d)"
	sourceLabelFormat = "%s  source:%s"
	stateLabelFormat  = "%s  entries:%d  visible:%d"
	statusErrorFormat = "%s  err:%s"
	statusHelpText    = "Space pause  / filter  Esc clear  q quit"
)

func (m model) View() tea.View {
	if m.width == 0 || m.height == 0 {
		return tea.View{Content: loadingText}
	}

	panel := m.styles.panel.Render(m.viewport.View())

	if !m.cfg.Source.FileFollow {
		return tea.View{Content: panel}
	}

	status := m.renderStatus()
	parts := []string{panel, status}
	if m.filterActive {
		parts = append(parts, m.styles.filterBar.Render(m.filterInput.View()))
	}

	return tea.View{Content: lipgloss.JoinVertical(lipgloss.Left, parts...)}
}

func (m model) renderStatus() string {
	state := liveStateLabel
	if m.paused {
		state = fmt.Sprintf(pausedStateFormat, m.pendingWhilePaused)
	} else if m.sourceDone {
		state = doneStateLabel
	}

	left := fmt.Sprintf(stateLabelFormat, state, len(m.entries), m.visibleEntryCount())
	if m.sourceName != "" {
		left = fmt.Sprintf(sourceLabelFormat, left, m.sourceName)
	}
	if m.query != "" {
		left = fmt.Sprintf(filterLabelFormat, left, m.query)
	}
	if m.sourceErr != nil {
		left = fmt.Sprintf(statusErrorFormat, left, m.sourceErr)
	}

	right := statusHelpText
	contentWidth := max(minViewportDimension, m.width-m.styles.status.GetHorizontalFrameSize())
	if lipgloss.Width(right) >= contentWidth {
		right = truncateText(right, contentWidth)
		return m.styles.status.Render(right)
	}

	leftWidth := contentWidth - lipgloss.Width(right)
	if leftWidth < minViewportDimension {
		right = truncateText(right, max(minViewportDimension, contentWidth/2))
		leftWidth = max(minViewportDimension, contentWidth-lipgloss.Width(right)-1)
	}
	left = truncateText(left, leftWidth)
	leftPart := lipgloss.NewStyle().Width(leftWidth).Render(left)

	return m.styles.status.Render(lipgloss.JoinHorizontal(lipgloss.Top, leftPart, right))
}
