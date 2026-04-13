// Package locale provides internationalization support for Tabby's Go backend
// and native UI.
//
// It supports loading PO-format translation files, string interpolation,
// and locale detection from environment variables.
package locale

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

// Language represents a supported locale
type Language struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// AllLanguages is the list of all supported languages
var AllLanguages = []Language{
	{"af-ZA", "Afrikaans"},
	{"id-ID", "Bahasa Indonesia"},
	{"cs-CZ", "Čeština"},
	{"da-DK", "Dansk"},
	{"de-DE", "Deutsch"},
	{"en-GB", "English (UK)"},
	{"en-US", "English (US)"},
	{"es-ES", "Español"},
	{"fr-FR", "Français"},
	{"hr-HR", "Hrvatski"},
	{"it-IT", "Italiano"},
	{"pl-PL", "Polski"},
	{"pt-PT", "Português"},
	{"pt-BR", "Português do Brasil"},
	{"sv-SE", "Svenska"},
	{"tr-TR", "Türkçe"},
	{"bg-BG", "Български"},
	{"ru-RU", "Русский"},
	{"sr-SP", "Српски"},
	{"uk-UA", "Українська"},
	{"ja-JP", "日本語"},
	{"ko-KR", "한국어"},
	{"zh-CN", "中文（简体）"},
	{"zh-TW", "中文 (繁體)"},
}

// Service provides locale/translation services
type Service struct {
	mu           sync.RWMutex
	currentLang  string
	translations map[string]map[string]string // lang -> key -> translation
	fallback     map[string]string            // en-US translations
}

// NewService creates a new locale service
func NewService() *Service {
	return &Service{
		currentLang:  "en-US",
		translations: make(map[string]map[string]string),
		fallback:     make(map[string]string),
	}
}

// SetLocale changes the current language
func (s *Service) SetLocale(lang string) {
	s.mu.Lock()
	s.currentLang = lang
	s.mu.Unlock()
}

// GetLocale returns the current language
func (s *Service) GetLocale() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentLang
}

// DetectLocale detects the locale from environment variables
func (s *Service) DetectLocale() string {
	// Check LANG environment variable
	lang := os.Getenv("LANG")
	if lang != "" {
		lang = strings.ReplaceAll(lang, ".UTF-8", "")
		lang = strings.ReplaceAll(lang, ".utf8", "")
		lang = strings.ReplaceAll(lang, "_", "-")
		for _, l := range AllLanguages {
			if l.Code == lang {
				return l.Code
			}
		}
		// Try just the language part
		if parts := strings.Split(lang, "-"); len(parts) >= 1 {
			prefix := parts[0]
			for _, l := range AllLanguages {
				if strings.HasPrefix(l.Code, prefix) {
					return l.Code
				}
			}
		}
	}

	// Check LC_ALL, LC_MESSAGES
	for _, env := range []string{"LC_ALL", "LC_MESSAGES"} {
		if val := os.Getenv(env); val != "" {
			val = strings.ReplaceAll(val, ".UTF-8", "")
			val = strings.ReplaceAll(val, "_", "-")
			for _, l := range AllLanguages {
				if l.Code == val {
					return l.Code
				}
			}
		}
	}

	return "en-US"
}

// Translate returns the translated string for the given key
func (s *Service) Translate(key string, args ...interface{}) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Try current language first
	if trans, ok := s.translations[s.currentLang]; ok {
		if val, ok := trans[key]; ok && val != "" {
			if len(args) > 0 {
				return fmt.Sprintf(val, args...)
			}
			return val
		}
	}

	// Fall back to English
	if val, ok := s.fallback[key]; ok && val != "" {
		if len(args) > 0 {
			return fmt.Sprintf(val, args...)
		}
		return val
	}

	// Return key as-is
	if len(args) > 0 {
		return fmt.Sprintf(key, args...)
	}
	return key
}

// T is a shorthand for Translate
func (s *Service) T(key string, args ...interface{}) string {
	return s.Translate(key, args...)
}

// LoadPO loads translations from a PO format file
func (s *Service) LoadPO(lang string, data string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	trans := make(map[string]string)
	var currentKey string
	scanner := bufio.NewScanner(strings.NewReader(data))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "msgid ") {
			currentKey = unquotePO(strings.TrimPrefix(line, "msgid "))
		} else if strings.HasPrefix(line, "msgstr ") && currentKey != "" {
			val := unquotePO(strings.TrimPrefix(line, "msgstr "))
			if val != "" {
				trans[currentKey] = val
			}
			currentKey = ""
		}
	}

	s.translations[lang] = trans

	// Store en-US as fallback
	if lang == "en-US" {
		s.fallback = trans
	}
}

// LoadPOFile loads translations from a PO file
func (s *Service) LoadPOFile(lang, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read PO file: %w", err)
	}
	s.LoadPO(lang, string(data))
	return nil
}

// GetAvailableLanguages returns languages that have loaded translations
func (s *Service) GetAvailableLanguages() []Language {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []Language
	for _, lang := range AllLanguages {
		if _, ok := s.translations[lang.Code]; ok {
			result = append(result, lang)
		}
	}
	return result
}

// unquotePO removes surrounding quotes from PO format strings
func unquotePO(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "\"")
	s = strings.TrimSuffix(s, "\"")
	// Handle basic escape sequences
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\t", "\t")
	s = strings.ReplaceAll(s, "\\\"", "\"")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	return s
}
