package tui

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	appconfig "github.com/dubeyKartikay/peacock/internal/config"
	"github.com/dubeyKartikay/peacock/internal/logs"
)

func TestBatchesCoalesceIntoOneRefresh(t *testing.T) {
	cfg := appconfig.DefaultConfig()
	cfg.Source.FileFollow = true
	m := newSizedModel(cfg)
	before := m.View().Content
	m, cmd := updateModelWithCmd(t, m, EntriesMsg{Entries: []logs.Entry{testEntry("first")}})
	if cmd == nil {
		t.Fatal("expected first live batch to schedule a refresh")
	}
	for _, message := range []string{"second", "third"} {
		var next tea.Cmd
		m, next = updateModelWithCmd(t, m, EntriesMsg{Entries: []logs.Entry{testEntry(message)}})
		if next != nil {
			t.Fatal("expected subsequent batches to share the pending refresh")
		}
	}
	if got := m.View().Content; got != before {
		t.Fatal("ingestion rebuilt the view before the refresh")
	}
	if m.queuedEntries.Len() != 3 || m.inBufferEntries.Len() != 0 {
		t.Fatal("expected entries to wait in the queue until the tick")
	}
	m = updateModel(t, m, cmd())
	if m.flushScheduled || m.queuedEntries.Len() != 0 {
		t.Fatal("expected refresh to clear pending work")
	}
	if got := stripANSI(m.View().Content); !strings.Contains(got, "third") {
		t.Fatalf("expected latest entry in refreshed view, got %q", got)
	}
	m, cmd = updateModelWithCmd(t, m, flushTickMsg{})
	if cmd != nil {
		t.Fatal("expected idle refresh not to start a recurring timer")
	}
	_, cmd = updateModelWithCmd(t, m, EntriesMsg{Entries: []logs.Entry{testEntry("fourth")}})
	if cmd == nil {
		t.Fatal("expected new input after a refresh to schedule another frame")
	}
}

func TestPausedBatchesStayFrozenAndResumeThroughFilters(t *testing.T) {
	cfg := appconfig.DefaultConfig()
	cfg.Source.FileFollow = true
	cfg.Buffer.MaxEntries = 3
	m := newSizedModel(cfg)
	m = applyFilter(t, m, "match")
	m = updateModel(t, m, EntriesMsg{Entries: []logs.Entry{testEntry("match before pause")}})
	m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeySpace})
	before := m.View().Content

	m, cmd := updateModelWithCmd(t, m, EntriesMsg{Entries: []logs.Entry{
		testEntry("evicted"), testEntry("match first"), testEntry("excluded"), testEntry("match last"),
	}})
	if cmd != nil || m.queuedEntries.Len() != 3 {
		t.Fatal("expected paused batches to stay bounded without scheduling a frame")
	}
	m = updateModel(t, m, flushTickMsg{})
	if got := m.View().Content; got != before {
		t.Fatal("pending refresh changed the paused frame")
	}
	m = updateModel(t, m, SourceDoneMsg{})
	if got := m.viewport.View(); strings.Contains(got, "match last") {
		t.Fatal("source completion exposed queued entries while paused")
	}
	m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeySpace})
	if got := strings.Join(ringMessages(m.inBufferEntries.Newest(0)), ","); got != "match first,excluded,match last" {
		t.Fatalf("unexpected resumed batch order: %s", got)
	}
	view := stripANSI(m.View().Content)
	if !strings.Contains(view, "match first") || !strings.Contains(view, "match last") || strings.Contains(view, "excluded") {
		t.Fatalf("expected resumed frame to honor active filters, got %q", view)
	}
}

func TestLiveBatchesKeepNewestEntriesInOrder(t *testing.T) {
	cfg := appconfig.DefaultConfig()
	cfg.Buffer.MaxEntries = 3
	m := newSizedModel(cfg)
	m = updateModel(t, m, EntriesMsg{Entries: []logs.Entry{testEntry("first"), testEntry("second")}})
	m = updateModel(t, m, EntriesMsg{Entries: []logs.Entry{testEntry("third"), testEntry("fourth")}})
	m = updateModel(t, m, flushTickMsg{})
	if got := strings.Join(ringMessages(m.inBufferEntries.Newest(0)), ","); got != "second,third,fourth" {
		t.Fatalf("unexpected live batch order: %s", got)
	}
	if got := stripANSI(m.View().Content); strings.Contains(got, "first") || !strings.Contains(got, "fourth") {
		t.Fatalf("expected refreshed view to show the newest retained entries, got %q", got)
	}
}

func TestSourceDoneFlushesLastFrameBeforeQuit(t *testing.T) {
	for _, follow := range []bool{false, true} {
		cfg := appconfig.DefaultConfig()
		cfg.Source.FileFollow = follow
		m := newSizedModel(cfg)
		m = updateModel(t, m, EntriesMsg{Entries: []logs.Entry{testEntry("final entry")}})
		m, cmd := updateModelWithCmd(t, m, SourceDoneMsg{})
		if !strings.Contains(stripANSI(m.View().Content), "final entry") || m.queuedEntries.Len() != 0 {
			t.Fatalf("follow=%v: expected final entry to render before completion", follow)
		}
		if (cmd == nil) != follow {
			t.Fatalf("follow=%v: unexpected quit command", follow)
		}
		if !follow {
			if _, ok := cmd().(tea.QuitMsg); !ok {
				t.Fatal("expected non-follow completion to quit")
			}
		}
	}
}

func TestInteractionsRefreshPendingEntriesImmediately(t *testing.T) {
	for _, msg := range []tea.Msg{
		tea.WindowSizeMsg{Width: 100, Height: 24},
		tea.KeyPressMsg{Code: '/', Text: "/"},
	} {
		m := newSizedModel(appconfig.DefaultConfig())
		m = updateModel(t, m, EntriesMsg{Entries: []logs.Entry{testEntry("pending entry")}})
		m = updateModel(t, m, msg)
		if !strings.Contains(stripANSI(m.View().Content), "pending entry") {
			t.Fatalf("interaction %T did not render pending input", msg)
		}
	}
}

func TestQuitLeavesPendingEntriesAndDisplayedFrameUnchanged(t *testing.T) {
	for _, state := range []string{"live", "paused", "filtering"} {
		t.Run(state, func(t *testing.T) {
			m := newSizedModel(appconfig.DefaultConfig())
			if state == "paused" {
				m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeySpace})
			}
			if state == "filtering" {
				m = updateModel(t, m, tea.KeyPressMsg{Code: '/', Text: "/"})
			}
			before := m.View().Content
			m = updateModel(t, m, EntriesMsg{Entries: []logs.Entry{testEntry("pending entry")}})
			m, cmd := updateModelWithCmd(t, m, tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
			if m.View().Content != before || m.queuedEntries.Len() != 1 || m.inBufferEntries.Len() != 0 {
				t.Fatal("quit should leave pending input and the displayed frame unchanged")
			}
			if cmd == nil {
				t.Fatal("expected an immediate quit command")
			}
			if _, ok := cmd().(tea.QuitMsg); !ok {
				t.Fatal("expected an immediate quit message")
			}
		})
	}
}

func TestOrdinaryKeysKeepPendingEntriesQueuedUntilTick(t *testing.T) {
	for _, key := range []tea.KeyPressMsg{
		{Code: 'x', Text: "x"}, {Code: 'G', Text: "G"},
		{Code: tea.KeyPgDown}, {Code: tea.KeyBackspace},
	} {
		t.Run(key.String(), func(t *testing.T) {
			m := newSizedModel(appconfig.DefaultConfig())
			m = updateModel(t, m, EntriesMsg{Entries: []logs.Entry{testEntry("pending entry")}})
			m = updateModel(t, m, key)
			if m.queuedEntries.Len() != 1 || m.inBufferEntries.Len() != 0 {
				t.Fatal("ordinary key flushed pending input before the tick")
			}
			m = updateModel(t, m, flushTickMsg{})
			if !strings.Contains(m.View().Content, "pending entry") {
				t.Fatal("expected queued entry to render on the tick")
			}
		})
	}
}

func TestFilterTypingDoesNotFlushPendingEntries(t *testing.T) {
	m := newSizedModel(appconfig.DefaultConfig())
	m = updateModel(t, m, tea.KeyPressMsg{Code: '/', Text: "/"})
	before := m.viewport.View()
	m = updateModel(t, m, EntriesMsg{Entries: []logs.Entry{testEntry("pending entry")}})
	m = typeText(t, m, "timeout")
	if m.filterInput.Value() != "timeout" || m.queuedEntries.Len() != 1 || m.viewport.View() != before {
		t.Fatal("typing should update the filter input without flushing the log viewport")
	}
}

func TestPauseIncludesAlreadyQueuedEntriesInSnapshot(t *testing.T) {
	m := newSizedModel(appconfig.DefaultConfig())
	m = updateModel(t, m, EntriesMsg{Entries: []logs.Entry{testEntry("before pause")}})
	m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeySpace})
	if !m.paused || m.queuedEntries.Len() != 0 || !strings.Contains(m.View().Content, "before pause") {
		t.Fatal("expected pause to freeze a snapshot containing already-received entries")
	}
}

func TestSourceCommandRunsAgainUntilDone(t *testing.T) {
	reads := 0
	read := func() tea.Msg {
		reads++
		return EntriesMsg{Entries: []logs.Entry{testEntry("from source")}, Done: reads == 2}
	}
	cfg := appconfig.DefaultConfig()
	cfg.Source.FileFollow = true
	m := NewModel("stdin", cfg, read).(model)
	m = updateModel(t, m, tea.WindowSizeMsg{Width: 80, Height: 20})
	m.flushScheduled = true
	m, cmd := updateModelWithCmd(t, m, m.Init()())
	if reads != 1 || cmd == nil || m.inBufferEntries.Len() != 0 {
		t.Fatal("expected a queued batch and the next source read")
	}
	m, cmd = updateModelWithCmd(t, m, cmd())
	if reads != 2 || !m.sourceDone || cmd != nil || m.queuedEntries.Len() != 0 {
		t.Fatal("expected source completion to flush entries and stop reading")
	}
	if !strings.Contains(stripANSI(m.View().Content), "from source") {
		t.Fatal("expected the final source batch to render immediately")
	}
}

func TestEmptySourceBatchWaitsForTickBeforeReadingAgain(t *testing.T) {
	reads := 0
	read := func() tea.Msg {
		reads++
		return EntriesMsg{Entries: []logs.Entry{testEntry("arrived later")}}
	}
	cfg := appconfig.DefaultConfig()
	cfg.Input.BatchInterval = time.Millisecond
	m := NewModel("stdin", cfg, read).(model)
	m = updateModel(t, m, tea.WindowSizeMsg{Width: 80, Height: 20})
	before := m.View().Content
	m, cmd := updateModelWithCmd(t, m, EntriesMsg{})
	if cmd == nil || reads != 0 || m.flushScheduled || m.View().Content != before {
		t.Fatal("empty batch should schedule a retry without reading or refreshing immediately")
	}
	msg := cmd()
	if reads != 1 || len(msg.(EntriesMsg).Entries) != 1 {
		t.Fatal("expected the retry tick to read newly available input")
	}
	m = updateModel(t, m, msg)
	if m.queuedEntries.Len() != 1 {
		t.Fatal("expected the retry result to enter the normal batch flow")
	}
}

func TestSourceBatchErrorFlushesEntriesBeforeShowingError(t *testing.T) {
	m := newSizedModel(appconfig.DefaultConfig())
	m = updateModel(t, m, EntriesMsg{Entries: []logs.Entry{testEntry("before error")}, Err: errTestSource})
	if m.sourceErr != errTestSource || m.queuedEntries.Len() != 0 || !strings.Contains(m.View().Content, "before error") {
		t.Fatal("expected source batch entries to flush at the error boundary")
	}
}

func TestRefreshTickUsesConfiguredInterval(t *testing.T) {
	cfg := appconfig.DefaultConfig()
	cfg.Input.BatchInterval = 50 * time.Millisecond
	m := newSizedModel(cfg)
	started := time.Now()
	_, cmd := updateModelWithCmd(t, m, EntriesMsg{Entries: []logs.Entry{testEntry("sparse entry")}})
	if _, ok := cmd().(flushTickMsg); !ok {
		t.Fatal("expected a Bubble Tea refresh message")
	}
	if time.Since(started) < cfg.Input.BatchInterval {
		t.Fatal("refresh tick did not wait for the configured interval")
	}
}
