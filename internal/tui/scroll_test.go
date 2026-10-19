package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	appconfig "github.com/dubeyKartikay/peacock/internal/config"
	"github.com/dubeyKartikay/peacock/internal/logs"
)

func TestUpwardNavigationPausesWithoutShowingPendingEntries(t *testing.T) {
	for _, test := range []struct {
		name string
		msg  tea.Msg
		last int
	}{
		{"wheel", tea.MouseWheelMsg{Button: tea.MouseWheelUp}, 47},
		{"up", tea.KeyPressMsg{Code: tea.KeyUp}, 49},
		{"k", tea.KeyPressMsg{Code: 'k', Text: "k"}, 49},
		{"page-up", tea.KeyPressMsg{Code: tea.KeyPgUp}, 33},
		{"ctrl-b", tea.KeyPressMsg{Code: 'b', Mod: tea.ModCtrl}, 33},
		{"half-page-up", tea.KeyPressMsg{Code: 'u', Text: "u"}, 42},
		{"ctrl-u", tea.KeyPressMsg{Code: 'u', Mod: tea.ModCtrl}, 42},
		{"top", tea.KeyPressMsg{Code: 'g', Text: "g"}, 17},
	} {
		t.Run(test.name, func(t *testing.T) {
			m := scrollModel(t)
			m = updateModel(t, m, EntriesMsg{Entries: []logs.Entry{testEntry("pending")}})
			m = updateModel(t, m, test.msg)
			view := stripANSI(m.viewport.View())
			if !m.paused || !strings.Contains(view, fmt.Sprintf("entry-%03d", test.last)) || strings.Contains(view, fmt.Sprintf("entry-%03d", test.last+1)) {
				t.Fatalf("expected paused history ending at entry %d, got %q", test.last, view)
			}
			if m.queuedEntries.Len() != 1 || m.inBufferEntries.Len() != 50 {
				t.Fatal("scrolling exposed pending input")
			}
			before := m.View().Content
			m = updateModel(t, m, EntriesMsg{Entries: []logs.Entry{testEntry("arrived while paused")}})
			m = updateModel(t, m, flushTickMsg{})
			if m.View().Content != before || m.queuedEntries.Len() != 2 {
				t.Fatal("incoming entries or the pending tick changed the paused view")
			}
		})
	}
}

func TestMouseScrollResumesOnlyOnSpaceAndPreservesFilters(t *testing.T) {
	m := scrollModel(t)
	m = applyFilter(t, m, "entry-")
	m = updateModel(t, m, tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	offset := m.viewport.YOffset()
	m = updateModel(t, m, tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	if m.viewport.YOffset() != offset-m.viewport.MouseWheelDelta {
		t.Fatal("scrolling while paused reset the viewport")
	}
	m = updateModel(t, m, EntriesMsg{Entries: []logs.Entry{testEntry("excluded"), testEntry("entry-051")}})
	for _, msg := range []tea.Msg{
		tea.MouseWheelMsg{Button: tea.MouseWheelDown},
		tea.KeyPressMsg{Code: 'G', Text: "G"},
	} {
		m = updateModel(t, m, msg)
		if !m.paused || strings.Contains(m.viewport.View(), "entry-051") {
			t.Fatal("downward navigation resumed live input")
		}
	}
	m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeySpace})
	view := stripANSI(m.viewport.View())
	if m.paused || m.queuedEntries.Len() != 0 || !strings.Contains(view, "entry-051") || strings.Contains(view, "excluded") {
		t.Fatalf("expected Space to resume through the active filter, got %q", view)
	}
}

func TestOtherNavigationDoesNotPause(t *testing.T) {
	for _, msg := range []tea.Msg{
		tea.MouseWheelMsg{Button: tea.MouseWheelDown},
		tea.MouseWheelMsg{Button: tea.MouseWheelLeft},
		tea.MouseWheelMsg{Button: tea.MouseWheelRight},
		tea.MouseWheelMsg{Button: tea.MouseWheelUp, Mod: tea.ModShift},
		tea.KeyPressMsg{Code: tea.KeyDown},
		tea.KeyPressMsg{Code: tea.KeyPgDown},
		tea.KeyPressMsg{Code: 'G', Text: "G"},
	} {
		m := updateModel(t, scrollModel(t), msg)
		if m.paused {
			t.Fatalf("%v unexpectedly paused live input", msg)
		}
	}
	m := updateModel(t, scrollModel(t), tea.KeyPressMsg{Code: '/', Text: "/"})
	m = typeText(t, m, "gku")
	m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeyUp})
	if m.paused || m.filterInput.Value() != "gku" {
		t.Fatal("filter editing triggered history navigation")
	}
}

func TestMouseReportingOnlyEnabledForInteractiveSources(t *testing.T) {
	for _, follow := range []bool{false, true} {
		cfg := appconfig.DefaultConfig()
		cfg.Source.FileFollow = follow
		m := newSizedModel(cfg)
		if got := m.View().MouseMode; (got == tea.MouseModeCellMotion) != follow {
			t.Fatalf("follow=%v: unexpected mouse mode %v", follow, got)
		}
	}
}

func scrollModel(t *testing.T) model {
	t.Helper()
	cfg := appconfig.DefaultConfig()
	cfg.Source.FileFollow = true
	m := newSizedModel(cfg)
	entries := make([]logs.Entry, 50)
	for i := range entries {
		entries[i] = testEntry(fmt.Sprintf("entry-%03d", i+1))
	}
	m = updateModel(t, m, EntriesMsg{Entries: entries})
	return updateModel(t, m, flushTickMsg{})
}
