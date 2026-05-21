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
	inBufferEntries entryRing
	visibleEntries  []*logs.Entry
	queuedEntries   entryRing
	paused          bool
	filterActive    bool
	sourceDone      bool
	sourceErr       error
	query           string
	filters         Filters
}

func NewModel(sourceName string, cfg appconfig.Config) model {
	input := textinput.New()
	input.Prompt = cfg.Input.FilterPrompt
	input.CharLimit = cfg.Input.FilterCharLimit
	input.Placeholder = cfg.Input.FilterPlaceholder

	vp := viewport.New()
	vp.KeyMap.PageDown.SetKeys(viewportPageDownKey, viewportPageDownAltKey)
	vp.KeyMap.PageUp.SetKeys(viewportPageUpKey, viewportPageUpAltKey)

	return model{
		sourceName:      sourceName,
		cfg:             cfg,
		viewport:        vp,
		filterInput:     input,
		styles:          defaultStyles(cfg.Theme),
		inBufferEntries: newEntryRing(cfg.Buffer.MaxEntries),
		queuedEntries:   newEntryRing(cfg.Buffer.MaxEntries),
	}
}

func (m model) WithFilters(filters ...string) model {
	m.filters = filters
	m.filterActive = len(filters) > 0
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) appendEntry(entries ...logs.Entry) model {
	m.inBufferEntries.AppendBatch(entries)
	return m
}

func (m model) queueEntry(entry logs.Entry) model {
	m.queuedEntries.Append(entry)
	return m
}

func (m model) filteredEntries(limit int) []*logs.Entry {
	maxEntries := min(m.inBufferEntries.Len(), limit)
	filtered := make([]*logs.Entry, 0, maxEntries)

	if len(m.filters) == 0 {
		return append(filtered, m.inBufferEntries.Newest(limit)...)
	}

	m.inBufferEntries.ReverseRange(func(entry *logs.Entry) bool {

		if limit <= 0 || len(filtered) >= limit-1 {
			return false
		}
		for _, filter := range m.filters {
			if !strings.Contains(entry.Search, filter) {
				return true
			}
		}
		filtered = append(filtered, entry)
		return true
	})
	slices.Reverse(filtered)
	return filtered
}

func (m *model) contentLines(limit int) []string {
	width := max(minViewportDimension, m.width-m.styles.panel.GetHorizontalFrameSize())

	m.visibleEntries = m.filteredEntries(limit)

	lines := make([]string, 0, len(m.visibleEntries))
	for index := range m.visibleEntries {
		rendered := m.styles.renderEntry(m.visibleEntries[index], width)
		lines = append(lines, rendered)
	}
	return lines
}


func (m model) liveEntryLimit() int {
	if(m.paused){
		return m.inBufferEntries.Len()
	}
	return max(minViewportDimension, m.height-m.styles.panel.GetVerticalFrameSize())
}

func (m *model) syncViewport(stickBottom bool) {
	contentWidth := max(minViewportDimension, m.width-m.styles.panel.GetHorizontalFrameSize())
	contentLimit := m.liveEntryLimit()
	content := m.contentLines(contentLimit)
	viewportHeight := m.totalHeight(contentWidth)

	m.viewport.SetWidth(contentWidth)
	m.viewport.SetHeight(viewportHeight)
	m.filterInput.SetWidth(max(minViewportDimension, m.width-m.styles.filterBar.GetHorizontalFrameSize()-2))
	m.viewport.SetContentLines(content)
	if stickBottom {
		m.viewport.GotoBottom()
	}
}

func (m model) totalHeight(width int) int {
	filterLines := 0
	if m.filterActive {
		filterLines = filterLineCount
	}
	if !m.cfg.Source.FileFollow {
		total := 0
		m.inBufferEntries.Range(func(entry *logs.Entry) bool {
			height, ok := entry.GetCachedHeight(width)
			if !ok {
				m.styles.renderEntry(entry, width)
				height, _ = entry.GetCachedHeight(width)
			}
			total += height
			return true
		})
		maxHeight := max(minViewportDimension, m.height-m.styles.panel.GetVerticalFrameSize())
		return max(minViewportDimension, min(total, maxHeight))
	}
	height := m.height - statusLineCount - filterLines - m.styles.panel.GetVerticalFrameSize()
	return max(minViewportDimension, height)
}

func (m model) contentHeight(width int) int {
	total := 0
	m.inBufferEntries.Range(func(entry *logs.Entry) bool {
		height, ok := entry.GetCachedHeight(width)
		if !ok {
			m.styles.renderEntry(entry, width)
			height, _ = entry.GetCachedHeight(width)
		}
		total += height
		return true
	})
	return total
}
