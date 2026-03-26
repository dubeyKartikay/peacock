package app

import (
	"testing"
)

func TestNonQueryEnvironmentRemovesQueryRelatedVariables(t *testing.T) {
	input := []string{
		"TERM=wezterm",
		"TERM_PROGRAM=WezTerm",
		"WT_SESSION=abc",
		"SSH_TTY=/dev/pts/1",
		"HOME=/tmp/home",
		"LANG=en_US.UTF-8",
	}

	got := nonQueryEnvironment(input)

	has := func(keyPrefix string) bool {
		for _, value := range got {
			if len(value) >= len(keyPrefix) && value[:len(keyPrefix)] == keyPrefix {
				return true
			}
		}
		return false
	}

	if has("WT_SESSION=") {
		t.Fatal("expected WT_SESSION to be removed")
	}
	if has("SSH_TTY=") {
		t.Fatal("expected SSH_TTY to be removed")
	}
	if has("TERM=wezterm") {
		t.Fatal("expected original TERM to be removed")
	}
	if has("TERM_PROGRAM=WezTerm") {
		t.Fatal("expected original TERM_PROGRAM to be removed")
	}
	if !has("TERM=xterm-256color") {
		t.Fatal("expected TERM override")
	}
	if !has("TERM_PROGRAM=Apple_Terminal") {
		t.Fatal("expected TERM_PROGRAM override")
	}
	if !has("HOME=/tmp/home") {
		t.Fatal("expected unrelated variables to remain")
	}
}
