package presets

import (
	"os"
	"testing"
)

func withTempHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	return home
}

func TestSaveLoadRoundTrip(t *testing.T) {
	withTempHome(t)

	in := Preset{
		Name:        "My Look",
		FitMode:     2,
		CustomWidth: 60,
		Theme:       "amber",
		Ramp:        "braille",
		Brightness:  10,
		Contrast:    1.4,
		Invert:      true,
	}
	if err := Save(in); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := Load("My Look")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Name != "My-Look" {
		t.Errorf("Name = %q, want sanitized %q", got.Name, "My-Look")
	}
	if got.Theme != "amber" || got.Ramp != "braille" || got.CustomWidth != 60 || !got.Invert {
		t.Errorf("round-trip mismatch: %+v", got)
	}
}

func TestListIncludesBuiltinsAndFiles(t *testing.T) {
	withTempHome(t)

	if err := Save(Preset{Name: "zeta", Theme: "matrix", Ramp: "blocks"}); err != nil {
		t.Fatal(err)
	}
	names, err := List()
	if err != nil {
		t.Fatal(err)
	}
	has := func(n string) bool {
		for _, x := range names {
			if x == n {
				return true
			}
		}
		return false
	}
	if !has("matrix") || !has("noir") {
		t.Errorf("List missing built-ins: %v", names)
	}
	if !has("zeta") {
		t.Errorf("List missing on-disk preset: %v", names)
	}
}

func TestLoadUnknownIsNotExist(t *testing.T) {
	withTempHome(t)
	if _, err := Load("does-not-exist"); !os.IsNotExist(err) {
		t.Errorf("err = %v, want os.ErrNotExist", err)
	}
}
