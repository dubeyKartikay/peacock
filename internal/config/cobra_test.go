package config

import (
	"testing"

	flag "github.com/spf13/pflag"
)

func TestReadFlagsPreservesConfigWhenFlagsAreNotProvided(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Source.FileFollow = true
	cfg.Source.FileTailLines = 25

	flags := flag.NewFlagSet("peacock", flag.ContinueOnError)
	RegisterFlags(flags)

	ReadFlags(&cfg, flags)

	if !cfg.Source.FileFollow {
		t.Fatal("expected existing follow mode to be preserved")
	}
	if cfg.Source.FileTailLines != 25 {
		t.Fatalf("expected existing tail line count 25, got %d", cfg.Source.FileTailLines)
	}
}

func TestReadFlagsAppliesExplicitFollowAndLineCount(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Source.FileTailLines = 25

	flags := flag.NewFlagSet("peacock", flag.ContinueOnError)
	RegisterFlags(flags)
	if err := flags.Parse([]string{"--follow", "--lines", "3"}); err != nil {
		t.Fatalf("parse flags: %v", err)
	}

	ReadFlags(&cfg, flags)

	if !cfg.Source.FileFollow {
		t.Fatal("expected --follow to enable follow mode")
	}
	if cfg.Source.FileTailLines != 3 {
		t.Fatalf("expected --lines to set tail line count 3, got %d", cfg.Source.FileTailLines)
	}
}
