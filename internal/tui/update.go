package tui

import (
	tea "charm.land/bubbletea/v2"
)

const (
	keyFilterMode   = "/"
	keyGoToBottom   = "G"
	keyGoToTop      = "g"
	keySpaceLiteral = "space"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m = m.syncViewport(true)
		return m, nil
	case EntryMsg:
		if m.paused {
      m = m.queueEntry(msg.Entry)
			return m, nil
		}else{
			m = m.appendEntry(msg.Entry)
		}
		m = m.syncViewport(true)
		return m, nil
	case SourceErrMsg:
		m.sourceErr = msg.Err
		return m, nil
	case SourceDoneMsg:
		m.sourceDone = true
		if !m.cfg.Source.FileFollow {
			m = m.syncViewport(true)
			return m, tea.Quit
		}
		return m, nil
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	if m.filterActive {
		switch msg.Code {
			case tea.KeyEsc:
				m.filterActive = false
				m.filterInput.Blur()
				m = m.syncViewport(true)
				return m, nil
		case tea.KeyEnter:
				m.filterActive = false
				m.query = m.filterInput.Value()
				m = m.syncViewport(true)
				return m, nil
		}
		var cmd tea.Cmd
		m.filterInput, cmd = m.filterInput.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case keySpaceLiteral:
		m.paused = !m.paused
		if !m.paused {
			m = m.appendEntry(m.queuedEntries...)
			m.queuedEntries = nil
			m = m.syncViewport(true)
		}
		return m, nil
	case keyFilterMode:
		m.filterActive = true
		m.filterInput.SetValue(m.query)
		m.filterInput.CursorEnd()
		m = m.syncViewport(false)
		cmd := m.filterInput.Focus()
		return m, cmd
	case keyGoToTop:
		m.viewport.GotoTop()
		return m, nil
	case keyGoToBottom :
		m.viewport.GotoBottom()
		return m, nil
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}
