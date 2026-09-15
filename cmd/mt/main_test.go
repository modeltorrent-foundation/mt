package main

import "testing"

// These cover the stable CLI plumbing only. Verb behavior (get/seed/pack/
// catalog) is asserted by the internal package tests, so implementing those
// verbs in later waves does not churn these.

func TestRunHelp(t *testing.T) {
	if code := run([]string{"--help"}); code != 0 {
		t.Errorf("help exit = %d want 0", code)
	}
}

func TestRunNoArgs(t *testing.T) {
	if code := run(nil); code != 2 {
		t.Errorf("no-args exit = %d want 2", code)
	}
}

func TestRunUnknown(t *testing.T) {
	if code := run([]string{"frobnicate"}); code != 2 {
		t.Errorf("unknown exit = %d want 2", code)
	}
}
