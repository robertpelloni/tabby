// Package settings provides persistent user settings for Tabby Go.
// Settings are stored in a TOML file in the user's configuration directory.
// This struct covers all the configurable options from the original Tabby terminal.
package settings

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
)

// Settings holds all user-configurable preferences, mirroring the original Tabby settings.
type Settings struct {
	// ---- Color Scheme ----
	ColorScheme string `toml:"color_scheme" json:"color_scheme"` // Name of a built-in color scheme
	// ---- Appearance ----
	FontSize     int     `toml:"font_size" json:"font_size"`
	FontFamily   string  `toml:"font_family" json:"font_family"`
	FallbackFont string  `toml:"fallback_font" json:"fallback_font"`
	FontWeight   int     `toml:"font_weight" json:"font_weight"`
	FontWeightBold int   `toml:"font_weight_bold" json:"font_weight_bold"`
	LineHeight   float64 `toml:"line_height" json:"line_height"`
	LinePadding  int     `toml:"line_padding" json:"line_padding"`
	Ligatures    bool    `toml:"ligatures" json:"ligatures"`
	Theme        string  `toml:"theme" json:"theme"`         // "auto", "light", "dark"
	CSS          string  `toml:"css" json:"css"`             // custom CSS
	Opacity      float64 `toml:"opacity" json:"opacity"`     // window opacity 0.0–1.0
	Spaciness    int     `toml:"spaciness" json:"spaciness"` // UI spacing multiplier 1–3
	Animations   bool    `toml:"animations" json:"animations"`

	// ---- Terminal ----
	Shell       string `toml:"shell" json:"shell"`               // preferred shell (empty for auto-detect)
	Scrollback  int    `toml:"scrollback" json:"scrollback"`     // scrollback buffer lines
	CursorStyle string `toml:"cursor_style" json:"cursor_style"` // "bar", "block", "underline"
	CursorBlink bool   `toml:"cursor_blink" json:"cursor_blink"`
	Bell        string `toml:"bell" json:"bell"` // "off", "visual", "audible"

	// Terminal rendering
	Frontend string `toml:"frontend" json:"frontend"` // "xterm-webgl", "xterm", "block"
	DrawBoldTextInBrightColors bool `toml:"draw_bold_text_in_bright_colors" json:"draw_bold_text_in_bright_colors"`
	MinimumContrastRatio float64 `toml:"minimum_contrast_ratio" json:"minimum_contrast_ratio"`

	// ---- Keyboard ----
	AltIsMeta     bool `toml:"alt_is_meta" json:"alt_is_meta"`
	ScrollOnInput bool `toml:"scroll_on_input" json:"scroll_on_input"`

	// ---- Clipboard ----
	CopyOnSelect          bool `toml:"copy_on_select" json:"copy_on_select"`
	CopyAsHTML            bool `toml:"copy_as_html" json:"copy_as_html"`
	BracketedPaste        bool `toml:"bracketed_paste" json:"bracketed_paste"`
	WarnOnMultilinePaste  bool `toml:"warn_on_multiline_paste" json:"warn_on_multiline_paste"`
	ReplaceNewlinesOnPaste bool `toml:"replace_newlines_on_paste" json:"replace_newlines_on_paste"`
	TrimWhitespaceOnPaste bool  `toml:"trim_whitespace_on_paste" json:"trim_whitespace_on_paste"`

	// ---- Mouse ----
	RightClick      string `toml:"right_click" json:"right_click"`           // "off", "menu", "paste", "clipboard"
	PasteOnMiddleClick bool `toml:"paste_on_middle_click" json:"paste_on_middle_click"`
	WordSeparator   string `toml:"word_separator" json:"word_separator"`

	// ---- Tabs ----
	TabPosition       string `toml:"tab_position" json:"tab_position"`             // "left", "right", "top", "bottom"
	LastTabClosesWindow bool `toml:"last_tab_closes_window" json:"last_tab_closes_window"`
	CycleTabs          bool  `toml:"cycle_tabs" json:"cycle_tabs"`
	HideCloseButton    bool  `toml:"hide_close_button" json:"hide_close_button"`
	ShowTabProfileIcon bool  `toml:"show_tab_profile_icon" json:"show_tab_profile_icon"`

	// ---- Split Panes ----
	PaneResizeStep  float64 `toml:"pane_resize_step" json:"pane_resize_step"`
	FocusFollowsMouse bool  `toml:"focus_follows_mouse" json:"focus_follows_mouse"`

	// ---- Startup / Session ----
	AutoOpen   bool `toml:"auto_open" json:"auto_open"`     // auto-open terminal on app start
	RecoverTabs bool `toml:"recover_tabs" json:"recover_tabs"` // restore tabs on restart

	// ---- Window ----
	Frame              string  `toml:"frame" json:"frame"`                           // "thin", "none", "native"
	Dock               string  `toml:"dock" json:"dock"`                             // "off", "bottom", "top", "left", "right"
	DockHideOnBlur     bool    `toml:"dock_hide_on_blur" json:"dock_hide_on_blur"`
	DockAlwaysOnTop    bool    `toml:"dock_always_on_top" json:"dock_always_on_top"`
	Vibrancy           bool    `toml:"vibrancy" json:"vibrancy"`
	HideTray           bool    `toml:"hide_tray" json:"hide_tray"`

	// ---- SSH ----
	SSHWarnOnClose   bool   `toml:"ssh_warn_on_close" json:"ssh_warn_on_close"`
	SSHVerifyHostKeys bool  `toml:"ssh_verify_host_keys" json:"ssh_verify_host_keys"`
	SSHAgentType     string `toml:"ssh_agent_type" json:"ssh_agent_type"`     // "auto", "pageant", "pipe"
	SSHAgentPath     string `toml:"ssh_agent_path" json:"ssh_agent_path"`
	SSHX11Display    string `toml:"ssh_x11_display" json:"ssh_x11_display"`
	SSHDisableDynamicTitle bool `toml:"ssh_disable_dynamic_title" json:"ssh_disable_dynamic_title"`

	// ---- Serial ----
	SerialBaudRate  int    `toml:"serial_baud_rate" json:"serial_baud_rate"`
	SerialDataBits  int    `toml:"serial_data_bits" json:"serial_data_bits"`
	SerialStopBits  int    `toml:"serial_stop_bits" json:"serial_stop_bits"`
	SerialParity    string `toml:"serial_parity" json:"serial_parity"` // "none", "even", "odd"
	SerialFlowControl string `toml:"serial_flow_control" json:"serial_flow_control"` // "none", "hardware", "software"

	// ---- Windows-specific ----
	UseConPTY    bool `toml:"use_conpty" json:"use_conpty"`
	SetComSpec   bool `toml:"set_comspec" json:"set_comspec"`

	// ---- Misc ----
	Language                 string `toml:"language" json:"language"`
	EnableAnalytics          bool   `toml:"enable_analytics" json:"enable_analytics"`
	EnableAutomaticUpdates   bool   `toml:"enable_automatic_updates" json:"enable_automatic_updates"`
	EnableExperimentalFeatures bool `toml:"enable_experimental_features" json:"enable_experimental_features"`
}

// DefaultSettings returns the default settings, matching original Tabby defaults
// with platform-specific overrides.
func DefaultSettings() Settings {
	s := Settings{
		// Color Scheme
		ColorScheme: "Tabby Default",

		// Appearance
		FontSize:     14,
		FontFamily:   "Cascadia Code,Fira Code,Consolas,Courier New,monospace",
		FallbackFont: "",
		FontWeight:   400,
		FontWeightBold: 700,
		LineHeight:   1.2,
		LinePadding:  0,
		Ligatures:    false,
		Theme:        "dark",
		CSS:          "",
		Opacity:      1.0,
		Spaciness:    1,
		Animations:   true,

		// Terminal
		Scrollback:   25000,
		CursorStyle:  "bar",
		CursorBlink:  true,
		Bell:         "off",
		Frontend:     "xterm-webgl",
		DrawBoldTextInBrightColors: true,
		MinimumContrastRatio: 4.0,

		// Keyboard
		AltIsMeta:     false,
		ScrollOnInput: true,

		// Clipboard
		CopyOnSelect:          false,
		CopyAsHTML:            true,
		BracketedPaste:        true,
		WarnOnMultilinePaste:  true,
		ReplaceNewlinesOnPaste: false,
		TrimWhitespaceOnPaste: true,

		// Mouse
		RightClick:         "menu",
		PasteOnMiddleClick: true,
		WordSeparator:      " ()[]{}'\"",

		// Tabs
		TabPosition:          "left",
		LastTabClosesWindow:  false,
		CycleTabs:            true,
		HideCloseButton:      false,
		ShowTabProfileIcon:   false,

		// Split Panes
		PaneResizeStep:    0.1,
		FocusFollowsMouse: false,

		// Startup
		AutoOpen:    true,
		RecoverTabs: true,

		// Window
		Frame:            "thin",
		Dock:             "off",
		DockHideOnBlur:   false,
		DockAlwaysOnTop:  true,
		Vibrancy:         false,
		HideTray:         false,

		// SSH
		SSHWarnOnClose:   false,
		SSHVerifyHostKeys: true,
		SSHAgentType:     "auto",
		SSHDisableDynamicTitle: true,

		// Serial
		SerialBaudRate:    115200,
		SerialDataBits:    8,
		SerialStopBits:    1,
		SerialParity:      "none",
		SerialFlowControl: "none",

		// Misc
		EnableAnalytics:          true,
		EnableAutomaticUpdates:   true,
		EnableExperimentalFeatures: false,
	}

	// Platform-specific overrides
	switch runtime.GOOS {
	case "windows":
		s.FontFamily = "Consolas"
		s.RightClick = "clipboard"
		s.PasteOnMiddleClick = false
		s.CopyOnSelect = true
		s.UseConPTY = true
	case "darwin":
		s.FontFamily = "Menlo"
		s.PasteOnMiddleClick = true
	case "linux":
		s.FontFamily = "Liberation Mono"
		s.PasteOnMiddleClick = false
	}

	return s
}

// ConfigDir returns the platform-specific configuration directory for Tabby Go.
// On Windows: %APPDATA%\Tabby
// On macOS: $HOME/Library/Application Support/Tabby
// On Linux: $HOME/.config/Tabby
func ConfigDir() (string, error) {
	var configDir string
	switch {
	case os.Getenv("APPDATA") != "":
		configDir = filepath.Join(os.Getenv("APPDATA"), "Tabby")
	case os.Getenv("HOME") != "":
		if runtime.GOOS == "darwin" {
			configDir = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Tabby")
		} else {
			configDir = filepath.Join(os.Getenv("HOME"), ".config", "Tabby")
		}
	default:
		configDir = "./tabby-config"
	}
	return configDir, nil
}

// SettingsFile returns the full path to the settings file.
func SettingsFile() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "settings.toml"), nil
}

// LoadSettings loads settings from disk, or returns defaults if file doesn't exist or is invalid.
func LoadSettings() (Settings, error) {
	settings := DefaultSettings()
	file, err := SettingsFile()
	if err != nil {
		return settings, err
	}
	data, err := os.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return settings, nil
		}
		return settings, err
	}
	if err := toml.Unmarshal(data, &settings); err != nil {
		return settings, err
	}
	return settings, nil
}

// SaveSettings saves the given settings to disk.
func SaveSettings(s Settings) error {
	file, err := SettingsFile()
	if err != nil {
		return err
	}
	dir := filepath.Dir(file)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := toml.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(file, data, 0o644)
}

// ResetSettings resets settings to defaults and saves them.
func ResetSettings() error {
	return SaveSettings(DefaultSettings())
}

// ListColorSchemes returns the names of available color schemes.
func ListColorSchemes() []string {
	return []string{
		"Tabby Default",
		"Tabby Default Light",
		"Dracula",
		"Solarized Dark",
		"Solarized Light",
		"Monokai",
		"Nord",
		"One Half Dark",
		"One Half Light",
		"Gruvbox Dark",
		"Tokyo Night",
		"Catppuccin Mocha",
		"Catppuccin Latte",
		"Ayu Dark",
		"Atom One Light",
		"Batman",
	}
}
