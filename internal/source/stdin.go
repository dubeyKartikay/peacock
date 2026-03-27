package source

import (
	"bufio"
	"fmt"
	"io"
	"os"

	appconfig "github.com/dubeyKartikay/peacock/internal/config"
)

const (
	defaultScannerBufferCapacity = 0
	stdinSourceName              = "stdin"
	stdinEventBufferSize         = 128
)

type stdinSource struct {
	events chan Event
	reader io.ReadCloser
	cfg    appconfig.InputConfig
}

func NewStdinSource(file *os.File, cfg appconfig.InputConfig) Source {
	src := &stdinSource{
		events: make(chan Event, stdinEventBufferSize),
		reader: file,
		cfg:    cfg,
	}
	go src.stream()
	return src
}

func (s *stdinSource) Name() string {
	return stdinSourceName
}

func (s *stdinSource) Events() <-chan Event {
	return s.events
}

func (s *stdinSource) Close() error {
	return nil
}

func (s *stdinSource) stream() {
	defer close(s.events)

	scanner := newScanner(s.reader, s.cfg)
	for scanner.Scan() {
		line := scanner.Text()
		s.events <- Event{Line: &line}
	}
	if err := scanner.Err(); err != nil {
		s.events <- Event{Err: fmt.Errorf("read stdin: %w", err)}
	}
	s.events <- Event{Done: true}
}

func newScanner(r io.Reader, cfg appconfig.InputConfig) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, defaultScannerBufferCapacity, cfg.ScannerInitialBufferBytes)
	scanner.Buffer(buf, cfg.ScannerMaxBufferBytes)
	return scanner
}
