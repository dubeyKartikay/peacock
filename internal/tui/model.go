package tui

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	appconfig "peacock/internal/config"
	"peacock/internal/logs"
)

const (
	minViewportDimension   = 1
	statusLineCount        = 1
	filterLineCount        = 1
	noResultIndex          = -1
	viewportPageDownKey    = "pgdown"
	viewportPageDownAltKey = "ctrl+f"
	viewportPageUpKey      = "pgup"
	viewportPageUpAltKey   = "ctrl+b"
)

type EntryMsg struct {
	Entry logs.Entry
}

type SourceErrMsg struct {
	Err error
}

type SourceDoneMsg struct{}

type model struct {
	sourceName         string
	width              int
	height             int
	cfg                appconfig.Config
	viewport           viewport.Model
	filterInput        textinput.Model
	styles             styles
	entries            []logs.Entry
	paused             bool
	filterActive       bool
	sourceDone         bool
	sourceErr          error
	pendingWhilePaused int
	query              string
}

func NewModel(sourceName string, cfg appconfig.Config) tea.Model {
	input := textinput.New()
	input.Prompt = cfg.Input.FilterPrompt
	input.CharLimit = cfg.Input.FilterCharLimit
	input.Placeholder = cfg.Input.FilterPlaceholder

	vp := viewport.New()
	vp.KeyMap.PageDown.SetKeys(viewportPageDownKey, viewportPageDownAltKey)
	vp.KeyMap.PageUp.SetKeys(viewportPageUpKey, viewportPageUpAltKey)

	return model{
		sourceName:  sourceName,
		cfg:         cfg,
		viewport:    vp,
		filterInput: input,
		styles:      defaultStyles(cfg.Theme),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) appendEntry(entry logs.Entry) model {
	m.entries = append(m.entries, entry)
	if len(m.entries) > m.cfg.Buffer.MaxEntries {
		trim := len(m.entries) - m.cfg.Buffer.MaxEntries
		m.entries = append([]logs.Entry(nil), m.entries[trim:]...)
	}
	return m
}

func (m model) filteredEntryIndexes() []int {
	if m.query == "" {
		return []int{}
	}

	filtered := make([]int, 0, len(m.entries))
	for index, entry := range m.entries {
		if strings.Contains(entry.Search, m.query) {
			filtered = append(filtered, index)
		}
	}
	if len(filtered) == 0 {
		return []int{noResultIndex}
	}

	return filtered
}

func (m model) contentLines() []string {
	width := max(minViewportDimension, m.width-m.styles.panel.GetHorizontalFrameSize()-2)
	if m.query == "" {
		lines := make([]string, 0, len(m.entries))
		for index := range m.entries {
			rendered, renderedHeight := m.styles.renderEntry(m.entries[index], width)
			m.entries[index].SetRenderHeight(renderedHeight) 
			lines = append(lines, rendered)
		}
		return lines
	}

	entryIndexes := m.filteredEntryIndexes()
	if isNoResultFilter(entryIndexes) {
		return nil
	}

	lines := make([]string, 0, len(entryIndexes))
	for _, index := range entryIndexes {
		rendered, renderedHeight := m.styles.renderEntry(m.entries[index], width)
		m.entries[index].SetRenderHeight(renderedHeight) 
		lines = append(lines, rendered)
	}
	return lines
}

func (m model) visibleEntryCount() int {
	if m.query == "" {
		return len(m.entries)
	}

	indexes := m.filteredEntryIndexes()
	if isNoResultFilter(indexes) {
		return 0
	}

	return len(indexes)
}

func isNoResultFilter(indexes []int) bool {
	return len(indexes) == 1 && indexes[0] == noResultIndex
}

func (m model) syncViewport(stickBottom bool) model {
	content := m.contentLines()
	contentHeight := m.totalHeight()
	contentWidth := max(minViewportDimension, m.width-m.styles.panel.GetHorizontalFrameSize())
	m.viewport.SetWidth(contentWidth)
	m.viewport.SetHeight(contentHeight)
	m.filterInput.SetWidth(max(minViewportDimension, m.width-m.styles.filterBar.GetHorizontalFrameSize()-2))
	m.viewport.SetContent(lipgloss.JoinVertical(lipgloss.Left, content...))
	if stickBottom {
		m.viewport.GotoBottom()
	}
	return m
}

func (m model) totalHeight() int {
	filterLines := 0
	if m.filterActive {
		filterLines = filterLineCount
	}
	if !m.cfg.Source.FileFollow {
		total := 0
		for _, e := range m.entries {
			total += e.ContentHeight()
		}
		maxHeight := max(minViewportDimension, m.height)
		return max(minViewportDimension, min(total, maxHeight))
	}
	height := m.height - statusLineCount - filterLines - m.styles.panel.GetVerticalFrameSize()
	return max(minViewportDimension, height)
}

func (m model) contentHeight() int {
	total := 0
	for _, e := range m.entries {
		total += e.ContentHeight()
	}
	return total
}
