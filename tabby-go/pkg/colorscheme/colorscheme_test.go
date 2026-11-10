package colorscheme

import (
	"testing"
)

func TestBuiltInSchemes(t *testing.T) {
	if len(BuiltInSchemes) == 0 {
		t.Fatalf("expected non-empty BuiltInSchemes")
	}

	for _, s := range BuiltInSchemes {
		if s.Name == "" {
			t.Errorf("found scheme with empty name")
		}
		if s.Foreground == "" {
			t.Errorf("scheme %s missing foreground color", s.Name)
		}
		if s.Background == "" {
			t.Errorf("scheme %s missing background color", s.Name)
		}
		if len(s.Colors) != 16 {
			t.Errorf("scheme %s expected 16 ANSI colors, got %d", s.Name, len(s.Colors))
		}
	}
}

func TestGetBuiltInScheme(t *testing.T) {
	// Should find existing built-in
	s := GetBuiltInScheme("Dracula")
	if s == nil {
		t.Fatalf("expected to find 'Dracula' built-in scheme")
	}
	if s.Background != "#1e1f29" {
		t.Errorf("expected Dracula background '#1e1f29', got %q", s.Background)
	}

	// Should find existing community scheme
	c := GetBuiltInScheme("3024 Day")
	if c == nil {
		t.Fatalf("expected to find community scheme '3024 Day'")
	}
	if c.Name != "3024 Day" {
		t.Errorf("expected community scheme name '3024 Day', got %q", c.Name)
	}

	// Should return nil for non-existent scheme
	missing := GetBuiltInScheme("NonExistentSchemeXYZ")
	if missing != nil {
		t.Errorf("expected nil for non-existent scheme, got %v", missing)
	}
}

func TestGetSchemeNamesAndAllSchemes(t *testing.T) {
	all := AllSchemes()
	names := GetSchemeNames()

	if len(all) == 0 {
		t.Fatalf("expected non-empty AllSchemes()")
	}
	if len(names) != len(all) {
		t.Fatalf("expected GetSchemeNames() length %d to match AllSchemes() length %d", len(names), len(all))
	}

	expectedCount := len(BuiltInSchemes) + len(CommunitySchemes)
	if len(all) != expectedCount {
		t.Errorf("expected %d total schemes, got %d", expectedCount, len(all))
	}
}

func TestToXTermTheme(t *testing.T) {
	s := GetBuiltInScheme("Solarized Light")
	if s == nil {
		t.Fatalf("expected 'Solarized Light' scheme to exist")
	}

	theme := s.ToXTermTheme()
	if theme["foreground"] != s.Foreground {
		t.Errorf("expected foreground %q, got %q", s.Foreground, theme["foreground"])
	}
	if theme["background"] != s.Background {
		t.Errorf("expected background %q, got %q", s.Background, theme["background"])
	}
	if theme["cursor"] != s.Cursor {
		t.Errorf("expected cursor %q, got %q", s.Cursor, theme["cursor"])
	}
	if theme["selectionBackground"] != s.Selection {
		t.Errorf("expected selectionBackground %q, got %q", s.Selection, theme["selectionBackground"])
	}
	if theme["selectionForeground"] != s.SelectionForeground {
		t.Errorf("expected selectionForeground %q, got %q", s.SelectionForeground, theme["selectionForeground"])
	}

	// Check ANSI slots
	slots := []string{
		"black", "red", "green", "yellow",
		"blue", "magenta", "cyan", "white",
		"brightBlack", "brightRed", "brightGreen", "brightYellow",
		"brightBlue", "brightMagenta", "brightCyan", "brightWhite",
	}
	for i, slot := range slots {
		if theme[slot] != s.Colors[i] {
			t.Errorf("expected slot %s color %q, got %q", slot, s.Colors[i], theme[slot])
		}
	}
}

func TestIsLight(t *testing.T) {
	darkScheme := GetBuiltInScheme("Dracula")
	if darkScheme.IsLight() {
		t.Errorf("expected Dracula to be dark (IsLight == false)")
	}

	lightScheme := GetBuiltInScheme("Tabby Default Light")
	if !lightScheme.IsLight() {
		t.Errorf("expected Tabby Default Light to be light (IsLight == true)")
	}

	solarizedLight := GetBuiltInScheme("Solarized Light")
	if !solarizedLight.IsLight() {
		t.Errorf("expected Solarized Light to be light (IsLight == true)")
	}

	// Test isLightColor helper edge cases
	if isLightColor("") {
		t.Errorf("empty string should not be light")
	}
	if isLightColor("invalid") {
		t.Errorf("invalid color string should not be light")
	}
	if !isLightColor("#ffffff") {
		t.Errorf("#ffffff should be light")
	}
	if isLightColor("#000000") {
		t.Errorf("#000000 should not be light")
	}
}
