package tui

import (
	"slices"
	"strings"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	appconfig "github.com/dubeyKartikay/peacock/internal/config"
	"github.com/dubeyKartikay/peacock/internal/logs"
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

type Filters []string

type model struct {
	sourceName      string
	width           int
	height          int
	cfg             appconfig.Config
	viewport        viewport.Model
	filterInput     textinput.Model
	styles          styles
	inBufferEntries []logs.Entry
	visibleEntries  []*logs.Entry
	queuedEntries   []logs.Entry
	paused          bool
	filterActive    bool
	sourceDone      bool
	sourceErr       error
	query           string
	filters         Filters
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

func (m model) appendEntry(entries ...logs.Entry) model {
	m.inBufferEntries = append(m.inBufferEntries, entries...)
	if len(m.inBufferEntries) > m.cfg.Buffer.MaxEntries {
		trim := len(m.inBufferEntries) - m.cfg.Buffer.MaxEntries
		m.inBufferEntries = append([]logs.Entry(nil), m.inBufferEntries[trim:]...)
	}
	return m
}

func (m model) queueEntry(entry logs.Entry) model {
	m.queuedEntries = append(m.queuedEntries, entry)
	if len(m.queuedEntries) > m.cfg.Buffer.MaxEntries {
		trim := len(m.queuedEntries) - m.cfg.Buffer.MaxEntries
		m.queuedEntries = append([]logs.Entry(nil), m.queuedEntries[trim:]...)
	}
	return m
}

func (m model) filteredEntries(limit int) []*logs.Entry {
	maxEntries := len(m.inBufferEntries)
	if limit > 0 {
		maxEntries = min(maxEntries, limit)
	}
	filtered := make([]*logs.Entry, 0, maxEntries)

	if len(m.filters) == 0 {
		start := 0
		if limit > 0 && len(m.inBufferEntries) > limit {
			start = len(m.inBufferEntries) - limit
		}
		entries := m.inBufferEntries[start:]
		for index := range entries {
			filtered = append(filtered, &entries[index])
		}
		return filtered
	}

	for i := len(m.inBufferEntries) - 1; i >= 0; i-- {
		allMatched := true
		for _, f := range m.filters {
			if !strings.Contains(m.inBufferEntries[i].Search, f) {
				allMatched = false
				break
			}
		}
		if allMatched {
			filtered = append(filtered, &m.inBufferEntries[i])
			if limit > 0 && len(filtered) >= limit {
				break
			}
		}
	}
	slices.Reverse(filtered)
	return filtered
}

func (m *model) contentLines(limit int) []string {
	width := max(minViewportDimension, m.width-m.styles.panel.GetHorizontalFrameSize())

	m.visibleEntries = m.filteredEntries(limit)

	lines := make([]string, 0, len(m.visibleEntries))
	for index := range m.visibleEntries {
		rendered, renderedHeight := m.styles.renderEntry(m.visibleEntries[index], width)
		m.visibleEntries[index].SetRenderHeight(renderedHeight)
		lines = append(lines, rendered)
	}
	return lines
}

func (m model) visibleEntryCount() int {
	return len(m.visibleEntries)
}

func isNoResultFilter(indexes []int) bool {
	return len(indexes) == 1 && indexes[0] == noResultIndex
}

func (m model) liveEntryLimit() int {
	return max(minViewportDimension, m.height-m.styles.panel.GetVerticalFrameSize())
}

func (m *model) syncViewport(stickBottom bool) {
	contentWidth := max(minViewportDimension, m.width-m.styles.panel.GetHorizontalFrameSize())
	viewportHeight := m.totalHeight()
	contentLimit := m.liveEntryLimit()
	if m.paused {
		contentLimit = 0
	}
	content := m.contentLines(contentLimit)

	m.viewport.SetWidth(contentWidth)
	m.viewport.SetHeight(viewportHeight)
	m.filterInput.SetWidth(max(minViewportDimension, m.width-m.styles.filterBar.GetHorizontalFrameSize()-2))
	m.viewport.SetContentLines(content)
	if stickBottom {
		m.viewport.GotoBottom()
	}
}

func (m model) totalHeight() int {
	filterLines := 0
	if m.filterActive {
		filterLines = filterLineCount
	}
	if !m.cfg.Source.FileFollow {
		total := 0
		for _, e := range m.inBufferEntries {
			total += e.ContentHeight()
		}
		maxHeight := max(minViewportDimension, m.height-m.styles.panel.GetVerticalFrameSize())
		return max(minViewportDimension, min(total, maxHeight))
	}
	height := m.height - statusLineCount - filterLines - m.styles.panel.GetVerticalFrameSize()
	return max(minViewportDimension, height)
}

func (m model) contentHeight() int {
	total := 0
	for _, e := range m.inBufferEntries {
		total += e.ContentHeight()
	}
	return total
}
