package config

import (
	"os"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	if err := Save(Config{Chrome: "amber"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if got := Load(); got.Chrome != "amber" {
		t.Errorf("Load().Chrome = %q, want %q", got.Chrome, "amber")
	}
}

func TestLoadMissingIsZero(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	if got := Load(); got != (Config{}) {
		t.Errorf("Load() with no file = %+v, want zero Config", got)
	}
}

func TestLoadCorruptIsZero(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	p, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := Load(); got != (Config{}) {
		t.Errorf("Load() on corrupt file = %+v, want zero Config", got)
	}
}
