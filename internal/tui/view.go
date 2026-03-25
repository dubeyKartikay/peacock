package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	doneStateLabel    = "DONE"
	filterLabelFormat = "%s  filter:%q"
	gapMinWidth       = 1
	liveStateLabel    = "LIVE"
	loadingText       = "Loading Peacock..."
	pausedStateFormat = "PAUSED (+%d)"
	sourceLabelFormat = "%s  source:%s"
	stateLabelFormat  = "%s  entries:%d  visible:%d"
	statusErrorFormat = "%s  err:%s"
	statusHelpText    = "Space pause  / filter  Esc clear  q quit"
	viewJoinSeparator = "\n"
)

func (m model) View() tea.View {
	if m.width == 0 || m.height == 0 {
		return tea.View{Content: loadingText}
	}

	panel := m.styles.panel.Render(m.viewport.View())
	status := m.renderStatus()

	parts := []string{panel, status}
	if m.filterActive {
		parts = append(parts, m.styles.filterBar.Render(m.filterInput.View()))
	}

	return tea.View{Content: strings.Join(parts, viewJoinSeparator)}
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
	contentWidth := max(minViewportDimension, m.width-viewportHorizontalTrimWidth)
	if lipgloss.Width(right) >= contentWidth {
		right = truncateText(right, contentWidth)
		return m.styles.status.Render(right)
	}

	maxLeftWidth := contentWidth - lipgloss.Width(right) - 1
	if maxLeftWidth < minViewportDimension {
		right = truncateText(right, max(minViewportDimension, contentWidth/2))
		maxLeftWidth = max(minViewportDimension, contentWidth-lipgloss.Width(right)-gapMinWidth)
	}
	left = truncateText(left, maxLeftWidth)

	gap := contentWidth - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < gapMinWidth {
		gap = gapMinWidth
	}

	return m.styles.status.Render(left + strings.Repeat(" ", gap) + right)
}
