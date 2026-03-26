package tui

import (
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	appconfig "peacock/internal/config"
	"peacock/internal/logs"
)

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func TestPauseBuffersNewEntriesUntilResume(t *testing.T) {
	m := newSizedModel()
	m = updateModel(t, m, EntryMsg{Entry: logs.ParseLine(`{"level":"info","time":"2026-03-25T12:00:00Z","message":"first"}`)})

	m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeySpace})
	if !m.paused {
		t.Fatal("expected model to pause after space")
	}
	pausedView := stripANSI(m.viewport.View())

	m = updateModel(t, m, EntryMsg{Entry: logs.ParseLine(`{"level":"error","time":"2026-03-25T12:00:01Z","message":"second"}`)})
	if m.pendingWhilePaused != 1 {
		t.Fatalf("expected 1 buffered entry, got %d", m.pendingWhilePaused)
	}
	if got := stripANSI(m.viewport.View()); got != pausedView {
		t.Fatalf("expected paused viewport to stay frozen\nwant: %q\n got: %q", pausedView, got)
	}

	m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeySpace})
	if m.paused {
		t.Fatal("expected model to resume after second space")
	}
	if m.pendingWhilePaused != 0 {
		t.Fatalf("expected pending count to reset, got %d", m.pendingWhilePaused)
	}
	if got := stripANSI(m.viewport.View()); !strings.Contains(got, "second") {
		t.Fatalf("expected resumed viewport to include buffered entry, got %q", got)
	}
}

func TestSlashFilterUsesLiteralSubstringMatching(t *testing.T) {
	m := newSizedModel()
	m = updateModel(t, m, EntryMsg{Entry: logs.ParseLine(`{"level":"info","time":"2026-03-25T12:00:00Z","message":"health check ok"}`)})
	m = updateModel(t, m, EntryMsg{Entry: logs.ParseLine(`{"level":"error","time":"2026-03-25T12:00:01Z","message":"database timeout"}`)})

	m = updateModel(t, m, tea.KeyPressMsg{Code: '/', Text: "/"})
	m = updateModel(t, m, tea.KeyPressMsg{Code: 't', Text: "t"})
	m = updateModel(t, m, tea.KeyPressMsg{Code: 'i', Text: "i"})
	m = updateModel(t, m, tea.KeyPressMsg{Code: 'm', Text: "m"})
	m = updateModel(t, m, tea.KeyPressMsg{Code: 'e', Text: "e"})
	m = updateModel(t, m, tea.KeyPressMsg{Code: 'o', Text: "o"})
	m = updateModel(t, m, tea.KeyPressMsg{Code: 'u', Text: "u"})
	m = updateModel(t, m, tea.KeyPressMsg{Code: 't', Text: "t"})

	if !m.filterActive {
		t.Fatal("expected filter mode to be active")
	}
	if m.query != "timeout" {
		t.Fatalf("expected query timeout, got %q", m.query)
	}
	filtered := m.filteredEntryIndexes()
	if got, want := len(filtered), 1; got != want {
		t.Fatalf("expected %d filtered entry, got %d", want, got)
	}
	if filtered[0] != 1 {
		t.Fatalf("expected filtered index 1, got %d", filtered[0])
	}
	if m.visibleEntries[filtered[0]].Message != "database timeout" {
		t.Fatalf("unexpected filtered message %q", m.visibleEntries[filtered[0]].Message)
	}

	m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeyEsc})
	if m.filterActive || m.query != "" {
		t.Fatalf("expected escape to clear filter, got active=%v query=%q", m.filterActive, m.query)
	}
	if filtered := m.filteredEntryIndexes(); len(filtered) != 0 {
		t.Fatalf("expected no active filter to return empty indexes, got %#v", filtered)
	}
}

func TestActiveFilterWithNoMatchesUsesSentinelIndex(t *testing.T) {
	m := newSizedModel()
	m = updateModel(t, m, EntryMsg{Entry: logs.ParseLine(`{"level":"info","time":"2026-03-25T12:00:00Z","message":"health check ok"}`)})

	m.query = "zzz"
	filtered := m.filteredEntryIndexes()
	if got, want := len(filtered), 1; got != want {
		t.Fatalf("expected %d sentinel index, got %d", want, got)
	}
	if filtered[0] != noResultIndex {
		t.Fatalf("expected sentinel index %d, got %d", noResultIndex, filtered[0])
	}
	if got := m.visibleEntryCount(); got != 0 {
		t.Fatalf("expected zero visible entries, got %d", got)
	}
}

func newSizedModel() model {
	m := NewModel("stdin", appconfig.DefaultConfig()).(model)
	m.width = 80
	m.height = 20
	return m.syncViewport(true)
}

func updateModel(t *testing.T, m model, msg tea.Msg) model {
	t.Helper()
	updated, _ := m.Update(msg)
	result, ok := updated.(model)
	if !ok {
		t.Fatal("expected concrete model result")
	}
	return result
}

func stripANSI(text string) string {
	return ansiPattern.ReplaceAllString(text, "")
}
