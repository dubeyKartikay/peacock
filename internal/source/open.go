package source

import (
	"fmt"
	"os"

	appconfig "github.com/dubeyKartikay/peacock/internal/config"
	"golang.org/x/term"
)

const (
	usageMessage = "usage: peacock [file]"
)

type Event struct {
	Line *string
	Err  error
	Done bool
}

type Source interface {
	Name() string
	Events() <-chan Event
	Close() error
}

func Open(inputPath string, stdin *os.File, cfg *appconfig.Config) (Source, error) {
	if inputPath != "" {
		if cfg.Source.FileFollow {
			return NewTailedFileSource(inputPath, cfg.Source)
		}
		return NewFileSource(inputPath, cfg.Source)
	}

	if stdin != nil && !term.IsTerminal(int(stdin.Fd())) {
		cfg.Source.FileFollow = true
		return NewStdinSource(stdin, cfg.Input), nil
	}

	return nil, fmt.Errorf(usageMessage)
}
