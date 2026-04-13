package theme

import (
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	mgr := NewManager()
	if len(mgr.ListThemes()) < 1 {
		t.Error("Should have default themes")
	}
}

func TestSetGetTheme(t *testing.T) {
	mgr := NewManager()
	err := mgr.SetTheme("default")
	if err != nil {
		t.Fatalf("SetTheme failed: %v", err)
	}
	theme := mgr.GetTheme()
	if theme.Name != "default" {
		t.Errorf("Expected 'default', got %q", theme.Name)
	}
}

func TestSetThemeNotFound(t *testing.T) {
	mgr := NewManager()
	err := mgr.SetTheme("nonexistent")
	if err == nil {
		t.Error("Should error for nonexistent theme")
	}
}

func TestGetActiveColorSchemeDark(t *testing.T) {
	mgr := NewManager()
	mgr.SetTheme("default")
	mgr.SetMode(ThemeModeDark)
	scheme := mgr.GetActiveColorScheme()
	if scheme.Background == "" {
		t.Error("Should have dark background color")
	}
}

func TestGetActiveColorSchemeLight(t *testing.T) {
	mgr := NewManager()
	mgr.SetTheme("default")
	mgr.SetMode(ThemeModeLight)
	scheme := mgr.GetActiveColorScheme()
	if scheme.Background == "" {
		t.Error("Should have light background color")
	}
}

func TestSetGetMode(t *testing.T) {
	mgr := NewManager()
	mgr.SetMode(ThemeModeLight)
	if mgr.GetMode() != ThemeModeLight {
		t.Error("Mode should be light")
	}
	mgr.SetMode(ThemeModeAuto)
	if mgr.GetMode() != ThemeModeAuto {
		t.Error("Mode should be auto")
	}
}

func TestRegisterTheme(t *testing.T) {
	mgr := NewManager()
	custom := Theme{
		Name:        "custom-test",
		DisplayName: "Custom Test",
		ColorSchemeDark: ColorScheme{
			Foreground: "#ffffff", Background: "#000000",
			Colors: [16]string{},
		},
	}
	mgr.RegisterTheme(custom)

	err := mgr.SetTheme("custom-test")
	if err != nil {
		t.Fatalf("Should find custom theme: %v", err)
	}
	if mgr.GetTheme().DisplayName != "Custom Test" {
		t.Error("Custom theme not loaded correctly")
	}
}

func TestRegisterThemeReplace(t *testing.T) {
	mgr := NewManager()
	mgr.RegisterTheme(Theme{Name: "default", DisplayName: "Replaced"})
	theme := mgr.GetTheme()
	if theme.DisplayName != "Replaced" {
		t.Error("Should have replaced existing theme")
	}
}

func TestCustomCSS(t *testing.T) {
	mgr := NewManager()
	mgr.SetCustomCSS("body { color: red; }")
	if mgr.GetCustomCSS() != "body { color: red; }" {
		t.Error("Custom CSS not stored")
	}
}

func TestOnChange(t *testing.T) {
	mgr := NewManager()
	var changedName string
	mgr.OnChange(func(theme Theme) {
		changedName = theme.Name
	})
	mgr.SetTheme("compact")
	if changedName != "compact" {
		t.Errorf("Expected 'compact', got %q", changedName)
	}
}

func TestSaveLoadThemes(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "themes.json")

	mgr := NewManager()
	err := mgr.SaveThemesToFile(path)
	if err != nil {
		t.Fatalf("SaveThemesToFile failed: %v", err)
	}

	mgr2 := NewManager()
	err = mgr2.LoadThemesFromFile(path)
	if err != nil {
		t.Fatalf("LoadThemesFromFile failed: %v", err)
	}

	themes := mgr2.ListThemes()
	if len(themes) < 3 {
		t.Errorf("Expected at least 3 themes, got %d", len(themes))
	}
}

func TestLoadThemesNonexistent(t *testing.T) {
	mgr := NewManager()
	err := mgr.LoadThemesFromFile("/nonexistent/themes.json")
	if err == nil {
		t.Error("Should fail for nonexistent file")
	}
}

func TestParseHexColor(t *testing.T) {
	c, err := ParseHexColor("#ff8040")
	if err != nil {
		t.Fatalf("ParseHexColor failed: %v", err)
	}
	if c.R != 255 || c.G != 128 || c.B != 64 {
		t.Errorf("Expected (255,128,64), got (%d,%d,%d)", c.R, c.G, c.B)
	}
}

func TestParseHexColorInvalid(t *testing.T) {
	_, err := ParseHexColor("invalid")
	if err == nil {
		t.Error("Should fail for invalid color")
	}
}

func TestParseHexColorShort(t *testing.T) {
	_, err := ParseHexColor("#fff")
	if err == nil {
		t.Error("Should fail for short color")
	}
}

func TestLuminance(t *testing.T) {
	// White should have luminance ~1.0
	white, _ := ParseHexColor("#ffffff")
	l := Luminance(white)
	if l < 0.9 {
		t.Errorf("White luminance should be ~1.0, got %f", l)
	}

	// Black should have luminance ~0.0
	black, _ := ParseHexColor("#000000")
	l = Luminance(black)
	if l > 0.1 {
		t.Errorf("Black luminance should be ~0.0, got %f", l)
	}
}

func TestIsDark(t *testing.T) {
	mgr := NewManager()
	mgr.SetTheme("default")
	scheme := mgr.GetActiveColorScheme()
	// Default dark scheme should be dark
	mgr.SetMode(ThemeModeDark)
	scheme = mgr.GetActiveColorScheme()
	if !IsDark(scheme) {
		t.Error("Dark color scheme should be detected as dark")
	}
}

func TestColorSchemeHas16Colors(t *testing.T) {
	mgr := NewManager()
	mgr.SetTheme("default")
	scheme := mgr.GetActiveColorScheme()
	if len(scheme.Colors) != 16 {
		t.Errorf("Color scheme should have 16 colors, got %d", len(scheme.Colors))
	}
}
