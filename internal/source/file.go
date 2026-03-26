package source

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/hpcloud/tail"

	appconfig "peacock/internal/config"
)

const (
	fileEventBufferSize = 128
	fileMustExist       = true
)

type source struct {
	name   string
	events chan Event
}


type tailedFileSource struct {
	source
	tail         *tail.Tail
}

type fileSource struct {
  source
	reader io.ReadCloser
}


func NewFileSource(path string, cfg appconfig.SourceConfig) (Source, error) {
	file,err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file %q: %w", path, err)
	}
	pos,_ :=getPositionOfNthLineFromEnd(file, cfg.FileTailLines)
	file.Seek(int64(pos),io.SeekStart)
	src := &fileSource{
		source: source{
			name:   path,
			events: make(chan Event, fileEventBufferSize),
		},
		reader: file,
	}
	go src.stream()
	return src, nil
}

func NewTailedFileSource(path string, cfg appconfig.SourceConfig) (Source, error) {
	var t *tail.Tail
	var err error
	file,err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file %q: %w", path, err)
	}
	defer file.Close()
	pos ,_ :=getPositionOfNthLineFromEnd(file, cfg.FileTailLines)
	t, err = tail.TailFile(path, tail.Config{
		Follow:        true,
		ReOpen:        cfg.FileReopen,
		MustExist:     fileMustExist,
		Poll:     cfg.FilePoll,
		Location: &tail.SeekInfo{
			Offset: pos,
			Whence: io.SeekStart,
		},
		Logger: tail.DiscardingLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("tail file %q: %w", path, err)
	}

	src := &tailedFileSource{
		source: source{
			name:   path,
			events: make(chan Event, fileEventBufferSize),
		},
		tail:         t,
	}
	go src.stream()
	return src, nil
}

func (s *source) Name() string {
	return s.name
}

func (s *source) Events() <-chan Event {
	return s.events
}

func (s *tailedFileSource) Close() error {
	if s.tail == nil {
		return nil
	}
	if err := s.tail.Stop(); err != nil {
		return err
	}
	s.tail.Cleanup()
	return nil
}

func (s *fileSource) Close() error {
	if s.reader == nil {
		return nil
	}
	if err := s.reader.Close(); err != nil {
		return err
	}
	return nil
}

func (s *fileSource) stream() {
	defer close(s.events)
	scanner := bufio.NewScanner(s.reader)
	for scanner.Scan() {
		text := scanner.Text()
		s.events <- Event{Line: &text}
	}
	if err := scanner.Err(); err != nil {
		s.events <- Event{Err: fmt.Errorf("scan file: %w", err)}
	}
	s.events <- Event{Done: true}
}

func (s *tailedFileSource) stream() {
	defer close(s.events)
	if s.tail == nil {
		s.events <- Event{Done: true}
		return
	}
	for line := range s.tail.Lines {
		if line == nil {
			continue
		}
		if line.Err != nil {
			s.events <- Event{Err: fmt.Errorf("tail line: %w", line.Err)}
			continue
		}
		text := line.Text
		s.events <- Event{Line: &text}
	}
	s.events <- Event{Done: true}
}

func getPositionOfNthLineFromEnd(file *os.File, n int) (int64,error) {
	const chunkSize = 512

	size, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}
	if size == 0 || n == 0 {
		return size, nil
	}

	// Skip a trailing newline so it doesn't count as an extra empty line.
	end := make([]byte, 1)
	if _, err := file.ReadAt(end, size-1); err != nil {
		return 0,err
	}
	scanEnd := size
	if end[0] == '\n' {
		scanEnd--
	}

	remaining := scanEnd
	newlines := 0
	buf := make([]byte, chunkSize)

	for remaining > 0 {
		chunkLen := int64(chunkSize)
		if remaining < chunkLen {
			chunkLen = remaining
		}
		remaining -= chunkLen

		if _, err := file.Seek(remaining, io.SeekStart); err != nil {
			return 0,err
		}
		if _, err := io.ReadFull(file, buf[:chunkLen]); err != nil {
			return 0,err
		}

		chunk := buf[:chunkLen]

		// Fast path: not enough newlines in this chunk to reach n.
		lineC := bytes.Count(chunk, []byte{'\n'})
		if newlines+lineC < n {
			newlines += lineC
			continue
		}

		// Slow path: find the exact newline position within this chunk.
		for i := chunkLen - 1; i >= 0; i-- {
			if chunk[i] == '\n' {
				newlines++
				if newlines == n {
					return remaining + i + 1, nil
				}
			}
		}
	}

	return 0, nil
}
