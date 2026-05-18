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
	liveStateLabel    = "LIVE"
	loadingText       = "Loading Peacock..."
	pausedStateLabel  = "PAUSED"
	sourceLabelFormat = "%s  source:%s"
	stateLabelFormat  = "%s  entries:%d  visible:%d"
	statusErrorFormat = "%s  err:%s"
	statusHelpText    = "pause: space  quit: ctrl+c  filter: /  remove filter: backspace"
)

func (m model) View() tea.View {
	if m.width == 0 || m.height == 0 {
		return tea.View{Content: loadingText}
	}

	panel := m.styles.panel.Render(m.viewport.View())

	if !m.cfg.Source.FileFollow {
		return tea.View{Content: panel + "\n"}
	}

	status := m.renderStatus()
	parts := []string{panel}
	if m.filterActive {
		parts = append(parts, m.styles.filterBar.Render(m.filterInput.View()))
	}
	parts = append(parts, status)
	return tea.View{Content: lipgloss.JoinVertical(lipgloss.Left, parts...)}
}

func (m *model) renderStatus() string {
	statusStyle := m.styles.status

	state := statusStyle.live.Render(liveStateLabel)
	if m.paused {
		state = statusStyle.paused.Render(pausedStateLabel)
	} else if m.sourceDone {
		state = statusStyle.done.Render(doneStateLabel)
	}
	entries := statusStyle.entries.Render(fmt.Sprintf("showing: %d/%d", m.visibleEntryCount(), len(m.inBufferEntries)+len(m.queuedEntries)))
	if m.sourceErr != nil {
		entries = statusStyle.err.Render(entries)
	}
	state = statusStyle.source.Render(state)
	var left string
	if m.sourceName != "" {
		source := statusStyle.source.Render(m.sourceName)
		left = lipgloss.JoinHorizontal(lipgloss.Left, left, source)
	}
	left = lipgloss.JoinHorizontal(lipgloss.Left, state, left, entries)

	for entry := range m.filters {
		filter := statusStyle.filter.Render(m.filters[entry])
		left = lipgloss.JoinHorizontal(lipgloss.Left, left, filter)
	}

	if m.sourceErr != nil {
		sourceErr := statusStyle.source.Render(m.sourceErr.Error())
		left = lipgloss.JoinHorizontal(lipgloss.Left, left, sourceErr)
	}

	right := statusStyle.help.Render(statusHelpText)
	contentWidth := max(minViewportDimension, m.width-statusStyle.bar.GetHorizontalFrameSize())
	if lipgloss.Width(right) >= contentWidth {
		right = truncateText(right, contentWidth)
		return statusStyle.bar.Render(right)
	}

	leftWidth := contentWidth - lipgloss.Width(right)
	if leftWidth < minViewportDimension {
		right = truncateText(right, max(minViewportDimension, contentWidth/2))
		leftWidth = max(minViewportDimension, contentWidth-lipgloss.Width(right)-1)
	}
	if lipgloss.Width(left) > leftWidth {
		left = truncateText(left, leftWidth)
	}
	centerPad := max(0, contentWidth-lipgloss.Width(left)-lipgloss.Width(right))
	center := strings.Repeat(" ", centerPad)
	return statusStyle.bar.Render(lipgloss.JoinHorizontal(lipgloss.Top, left, center, right))
}
