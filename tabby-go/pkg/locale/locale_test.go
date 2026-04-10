package locale

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewService(t *testing.T) {
	svc := NewService()
	if svc.GetLocale() != "en-US" {
		t.Errorf("Default locale should be en-US, got %q", svc.GetLocale())
	}
}

func TestSetLocale(t *testing.T) {
	svc := NewService()
	svc.SetLocale("de-DE")
	if svc.GetLocale() != "de-DE" {
		t.Errorf("Expected de-DE, got %q", svc.GetLocale())
	}
}

func TestTranslateNoTranslations(t *testing.T) {
	svc := NewService()
	result := svc.Translate("Hello World")
	if result != "Hello World" {
		t.Errorf("Untranslated key should return key, got %q", result)
	}
}

func TestTranslateWithArgs(t *testing.T) {
	svc := NewService()
	result := svc.Translate("Hello %s", "World")
	if result != "Hello World" {
		t.Errorf("Expected 'Hello World', got %q", result)
	}
}

func TestLoadPO(t *testing.T) {
	svc := NewService()
	po := `
msgid "Hello"
msgstr "Hallo"

msgid "Goodbye"
msgstr "Auf Wiedersehen"

msgid "Welcome %s"
msgstr "Willkommen %s"
`
	svc.LoadPO("de-DE", po)
	svc.SetLocale("de-DE")

	tests := []struct {
		key      string
		expected string
		args     []interface{}
	}{
		{"Hello", "Hallo", nil},
		{"Goodbye", "Auf Wiedersehen", nil},
		{"Welcome %s", "Willkommen User", []interface{}{"User"}},
		{"Unknown", "Unknown", nil},
	}

	for _, tt := range tests {
		result := svc.Translate(tt.key, tt.args...)
		if result != tt.expected {
			t.Errorf("Translate(%q) = %q, want %q", tt.key, result, tt.expected)
		}
	}
}

func TestFallbackToEnUS(t *testing.T) {
	svc := NewService()
	poEN := `
msgid "Hello"
msgstr "Hello"

msgid "Goodbye"
msgstr "Goodbye"
`
	poDE := `
msgid "Hello"
msgstr "Hallo"
`
	svc.LoadPO("en-US", poEN)
	svc.LoadPO("de-DE", poDE)
	svc.SetLocale("de-DE")

	if svc.Translate("Hello") != "Hallo" {
		t.Errorf("Expected 'Hallo', got %q", svc.Translate("Hello"))
	}

	if svc.Translate("Goodbye") != "Goodbye" {
		t.Errorf("Expected 'Goodbye' (fallback), got %q", svc.Translate("Goodbye"))
	}
}

func TestLoadPOFile(t *testing.T) {
	tmpDir := t.TempDir()
	poPath := filepath.Join(tmpDir, "test.po")
	content := `msgid "Save"
msgstr "Speichern"

msgid "Cancel"
msgstr "Abbrechen"
`
	os.WriteFile(poPath, []byte(content), 0644)

	svc := NewService()
	err := svc.LoadPOFile("de-DE", poPath)
	if err != nil {
		t.Fatalf("LoadPOFile failed: %v", err)
	}

	svc.SetLocale("de-DE")
	if svc.Translate("Save") != "Speichern" {
		t.Errorf("Expected 'Speichern', got %q", svc.Translate("Save"))
	}
}

func TestLoadPOFileNonexistent(t *testing.T) {
	svc := NewService()
	err := svc.LoadPOFile("de-DE", "/nonexistent/path.po")
	if err == nil {
		t.Error("Should fail for nonexistent file")
	}
}

func TestTShorthand(t *testing.T) {
	svc := NewService()
	po := `msgid "Hello"
msgstr "Hola"
`
	svc.LoadPO("es-ES", po)
	svc.SetLocale("es-ES")

	if svc.T("Hello") != "Hola" {
		t.Errorf("T() shorthand should work, got %q", svc.T("Hello"))
	}
}

func TestDetectLocaleDefault(t *testing.T) {
	svc := NewService()
	// Without LANG set, should return en-US
	lang := svc.DetectLocale()
	if lang == "" {
		t.Error("DetectLocale should not return empty")
	}
}

func TestDetectLocaleWithEnv(t *testing.T) {
	os.Setenv("LANG", "de_DE.UTF-8")
	defer os.Unsetenv("LANG")

	svc := NewService()
	lang := svc.DetectLocale()
	if lang != "de-DE" {
		t.Errorf("Expected de-DE, got %q", lang)
	}
}

func TestGetAvailableLanguages(t *testing.T) {
	svc := NewService()
	svc.LoadPO("en-US", `msgid "Hello" msgstr "Hello"`)
	svc.LoadPO("de-DE", `msgid "Hello" msgstr "Hallo"`)

	langs := svc.GetAvailableLanguages()
	if len(langs) != 2 {
		t.Errorf("Expected 2 languages, got %d", len(langs))
	}
}

func TestAllLanguages(t *testing.T) {
	if len(AllLanguages) < 20 {
		t.Errorf("Expected at least 20 languages, got %d", len(AllLanguages))
	}

	// Check en-US exists
	found := false
	for _, l := range AllLanguages {
		if l.Code == "en-US" {
			found = true
			break
		}
	}
	if !found {
		t.Error("en-US should be in AllLanguages")
	}
}
