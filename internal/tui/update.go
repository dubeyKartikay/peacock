package tui

import (
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/dubeyKartikay/peacock/internal/logs"
)

const (
	keyFilterMode       = "/"
	keyGoToBottom       = "G"
	keyGoToTop          = "g"
	keyRemoveLastFilter = "backspace"
	keySpaceLiteral     = "space"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.syncViewport(!m.paused)
	case EntryMsg:
		return m.queueEntries([]logs.Entry{msg.Entry})
	case EntriesMsg:
		if len(msg.Entries) == 0 && msg.Err == nil && !msg.Done {
			return m, m.retrySourceRead()
		}
		m, cmd = m.queueEntries(msg.Entries)
		if !msg.Done && msg.Err == nil {
			return m, tea.Batch(cmd, m.readSource)
		}
		if msg.Err != nil {
			m.sourceErr = msg.Err
		}
		if !m.paused {
			m.syncViewport(true)
		}
		if msg.Done {
			m.sourceDone = true
			if !m.cfg.Source.FileFollow {
				cmd = tea.Quit
			}
		} else {
			cmd = tea.Batch(cmd, m.readSource)
		}
	case flushTickMsg:
		m.flushScheduled = false
		if m.queuedEntries.Len() == 0 || m.paused {
			return m, nil
		}
		m.syncViewport(true)
	case SourceErrMsg:
		m.sourceErr = msg.Err
	case SourceDoneMsg:
		m.sourceDone = true
		if !m.paused {
			m.syncViewport(true)
		}
		if !m.cfg.Source.FileFollow {
			cmd = tea.Quit
		}
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		m, cmd = m.handleKey(msg)
	case tea.MouseWheelMsg:
		if msg.Button == tea.MouseWheelUp && !msg.Mod.Contains(tea.ModShift) {
			m.pauseForScroll()
		}
		m.viewport, cmd = m.viewport.Update(msg)
	default:
		m.viewport, cmd = m.viewport.Update(msg)
	}
	m.cachedView = m.renderView()
	return m, cmd
}

func (m model) retrySourceRead() tea.Cmd {
	if m.readSource == nil {
		return nil
	}
	readSource := m.readSource
	return tea.Tick(m.cfg.Input.BatchInterval, func(time.Time) tea.Msg {
		return readSource()
	})
}

func (m model) queueEntries(entries []logs.Entry) (model, tea.Cmd) {
	if len(entries) == 0 {
		return m, nil
	}
	m.queuedEntries.AppendBatch(entries)
	if m.paused || m.flushScheduled {
		return m, nil
	}
	m.flushScheduled = true
	return m, tea.Tick(m.cfg.Input.BatchInterval, func(time.Time) tea.Msg {
		return flushTickMsg{}
	})
}

func (m *model) pauseForScroll() {
	if !m.paused {
		m.paused = true
		m.syncViewport(true)
	}
}

func (m model) handleKey(msg tea.KeyPressMsg) (model, tea.Cmd) {
	if m.filterActive {
		switch msg.Code {
		case tea.KeyEsc:
			m.filterActive = false
			m.filterInput.Blur()
			m.syncViewport(!m.paused)
			return m, nil
		case tea.KeyEnter:
			m.filterActive = false
			q := m.filterInput.Value()
			m.filters = append(m.filters, q)
			m.syncViewport(!m.paused)
			return m, nil
		}
		var cmd tea.Cmd
		m.filterInput, cmd = m.filterInput.Update(msg)
		return m, cmd
	}

	if msg.String() == keyGoToTop || key.Matches(msg, m.viewport.KeyMap.Up, m.viewport.KeyMap.PageUp, m.viewport.KeyMap.HalfPageUp) {
		m.pauseForScroll()
	}

	switch msg.String() {
	case keySpaceLiteral:
		if !m.paused {
			m.flushQueuedEntries()
		}
		m.paused = !m.paused
		m.syncViewport(true)
		return m, nil
	case keyFilterMode:
		m.filterActive = true
		m.filterInput.SetValue("")
		m.filterInput.CursorEnd()
		m.syncViewport(false)
		cmd := m.filterInput.Focus()
		return m, cmd
	case keyGoToTop:
		m.viewport.GotoTop()
		return m, nil
	case keyGoToBottom:
		m.viewport.GotoBottom()
		return m, nil
	case keyRemoveLastFilter:
		if len(m.filters) > 0 {
			m.filters = m.filters[:len(m.filters)-1]
			m.syncViewport(!m.paused)
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}
