// Package colorscheme provides terminal color scheme definitions and parsing
// for Tabby Go. It includes built-in schemes matching the original Tabby defaults
// and can parse XRDB-format color scheme files from the community collection.
package colorscheme

// ColorScheme represents an xterm.js-compatible terminal color scheme.
// The 16-color palette maps to xterm's ANSI color slots 0-15.
type ColorScheme struct {
	Name             string   `json:"name" toml:"name"`
	Foreground       string   `json:"foreground" toml:"foreground"`
	Background       string   `json:"background" toml:"background"`
	Cursor           string   `json:"cursor" toml:"cursor"`
	CursorAccent     string   `json:"cursorAccent,omitempty" toml:"cursorAccent,omitempty"`
	Selection        string   `json:"selection,omitempty" toml:"selection,omitempty"`
	SelectionForeground string `json:"selectionForeground,omitempty" toml:"selectionForeground,omitempty"`
	Colors           []string `json:"colors" toml:"colors"` // 16 ANSI colors (0-7 normal, 8-15 bright)
}

// Built-in color schemes matching the original Tabby terminal defaults
// plus popular community schemes.

var BuiltInSchemes = []ColorScheme{
	// ---- Tabby Default (Dark) ----
	{
		Name:       "Tabby Default",
		Foreground: "#cacaca",
		Background: "#171717",
		Cursor:     "#bbbbbb",
		Colors: []string{
			"#000000", "#ff615a", "#b1e969", "#ebd99c",
			"#5da9f6", "#e86aff", "#82fff7", "#dedacf",
			"#313131", "#f58c80", "#ddf88f", "#eee5b2",
			"#a5c7ff", "#ddaaff", "#b7fff9", "#ffffff",
		},
	},
	// ---- Tabby Default Light ----
	{
		Name:       "Tabby Default Light",
		Foreground: "#4d4d4c",
		Background: "#ffffff",
		Cursor:     "#4d4d4c",
		Colors: []string{
			"#000000", "#c82829", "#718c00", "#eab700",
			"#4271ae", "#8959a8", "#3e999f", "#ffffff",
			"#000000", "#c82829", "#718c00", "#eab700",
			"#4271ae", "#8959a8", "#3e999f", "#ffffff",
		},
	},
	// ---- Dracula ----
	{
		Name:       "Dracula",
		Foreground: "#f8f8f2",
		Background: "#1e1f29",
		Cursor:     "#bbbbbb",
		Colors: []string{
			"#000000", "#ff5555", "#50fa7b", "#f1fa8c",
			"#bd93f9", "#ff79c6", "#8be9fd", "#bbbbbb",
			"#555555", "#ff5555", "#50fa7b", "#f1fa8c",
			"#bd93f9", "#ff79c6", "#8be9fd", "#ffffff",
		},
	},
	// ---- Solarized Dark ----
	{
		Name:       "Solarized Dark",
		Foreground: "#708284",
		Background: "#001e27",
		Cursor:     "#708284",
		Colors: []string{
			"#002831", "#d11c24", "#738a05", "#a57706",
			"#2176c7", "#c61c6f", "#259286", "#eae3cb",
			"#001e27", "#bd3613", "#475b62", "#536870",
			"#708284", "#5956ba", "#819090", "#fcf4dc",
		},
	},
	// ---- Solarized Light ----
	{
		Name:       "Solarized Light",
		Foreground: "#52676a",
		Background: "#fdf6e3",
		Cursor:     "#52676a",
		Colors: []string{
			"#002831", "#d11c24", "#738a05", "#a57706",
			"#2176c7", "#c61c6f", "#259286", "#eae3cb",
			"#001e27", "#bd3613", "#475b62", "#536870",
			"#708284", "#5956ba", "#819090", "#fcf4dc",
		},
		Selection:        "#eee8d5",
		SelectionForeground: "#586e75",
	},
	// ---- Monokai ----
	{
		Name:       "Monokai",
		Foreground: "#f8f8f2",
		Background: "#272822",
		Cursor:     "#f8f8f0",
		Colors: []string{
			"#272822", "#f92672", "#a6e22e", "#f4bf75",
			"#66d9ef", "#ae81ff", "#a1efe4", "#f8f8f2",
			"#3b3a32", "#f92672", "#a6e22e", "#f4bf75",
			"#66d9ef", "#ae81ff", "#a1efe4", "#f8f8f0",
		},
	},
	// ---- Nord ----
	{
		Name:       "Nord",
		Foreground: "#d8dee9",
		Background: "#2e3440",
		Cursor:     "#d8dee9",
		Colors: []string{
			"#3b4252", "#bf616a", "#a3be8c", "#ebcb8b",
			"#81a1c1", "#b48ead", "#88c0d0", "#e5e9f0",
			"#4c566a", "#bf616a", "#a3be8c", "#ebcb8b",
			"#81a1c1", "#b48ead", "#8fbcbb", "#eceff4",
		},
	},
	// ---- One Half Dark ----
	{
		Name:       "One Half Dark",
		Foreground: "#dcdfe4",
		Background: "#282c34",
		Cursor:     "#a3b3cc",
		Colors: []string{
			"#282c34", "#e06c75", "#98c379", "#e5c07b",
			"#61afef", "#c678dd", "#56b6c2", "#dcdfe4",
			"#282c34", "#e06c75", "#98c379", "#e5c07b",
			"#61afef", "#c678dd", "#56b6c2", "#dcdfe4",
		},
	},
	// ---- One Half Light ----
	{
		Name:       "One Half Light",
		Foreground: "#383a42",
		Background: "#fafafa",
		Cursor:     "#383a42",
		Colors: []string{
			"#383a42", "#e45649", "#50a14f", "#c18401",
			"#0184bc", "#a626a4", "#0997b3", "#4f5261",
			"#4f5261", "#e06c75", "#98c379", "#e5c07b",
			"#61afef", "#c678dd", "#56b6c2", "#dcdfe4",
		},
	},
	// ---- Gruvbox Dark ----
	{
		Name:       "Gruvbox Dark",
		Foreground: "#ebdbb2",
		Background: "#282828",
		Cursor:     "#ebdbb2",
		Colors: []string{
			"#282828", "#cc241d", "#98971a", "#d79921",
			"#458588", "#b16286", "#689d6a", "#a89984",
			"#928374", "#fb4934", "#b8bb26", "#fabd2f",
			"#83a598", "#d3869b", "#8ec07c", "#ebdbb2",
		},
	},
	// ---- Tokyo Night ----
	{
		Name:       "Tokyo Night",
		Foreground: "#a9b1d6",
		Background: "#1a1b26",
		Cursor:     "#c0caf5",
		Colors: []string{
			"#15161e", "#f7768e", "#9ece6a", "#e0af68",
			"#7aa2f7", "#bb9af7", "#7dcfff", "#a9b1d6",
			"#414868", "#f7768e", "#9ece6a", "#e0af68",
			"#7aa2f7", "#bb9af7", "#7dcfff", "#c0caf5",
		},
	},
	// ---- Catppuccin Mocha ----
	{
		Name:       "Catppuccin Mocha",
		Foreground: "#cdd6f4",
		Background: "#1e1e2e",
		Cursor:     "#f5e0dc",
		Colors: []string{
			"#45475a", "#f38ba8", "#a6e3a1", "#f9e2af",
			"#89b4fa", "#cba6f7", "#94e2d5", "#bac2de",
			"#585b70", "#f38ba8", "#a6e3a1", "#f9e2af",
			"#89b4fa", "#cba6f7", "#94e2d5", "#a6adc8",
		},
	},
	// ---- Catppuccin Latte (Light) ----
	{
		Name:       "Catppuccin Latte",
		Foreground: "#4c4f69",
		Background: "#eff1f5",
		Cursor:     "#dc8a78",
		Colors: []string{
			"#5c5f77", "#d20f39", "#40a02b", "#df8e1d",
			"#1e66f5", "#8839ef", "#179299", "#acb0be",
			"#6c6f85", "#d20f39", "#40a02b", "#df8e1d",
			"#1e66f5", "#8839ef", "#179299", "#bcc0cc",
		},
	},
	// ---- Ayu Dark ----
	{
		Name:       "Ayu Dark",
		Foreground: "#b3b1ad",
		Background: "#0a0e14",
		Cursor:     "#e6b450",
		Colors: []string{
			"#01060e", "#ea6c73", "#91b362", "#f9af4d",
			"#53b8e6", "#fae994", "#90e1c6", "#c7c7c7",
			"#687190", "#ea6c73", "#91b362", "#f9af4d",
			"#53b8e6", "#fae994", "#90e1c6", "#ffffff",
		},
	},
	// ---- Atom One Light ----
	{
		Name:       "Atom One Light",
		Foreground: "#383a42",
		Background: "#fafafa",
		Cursor:     "#383a42",
		Colors: []string{
			"#000000", "#e45649", "#50a14f", "#986801",
			"#4078f2", "#a626a4", "#0184bc", "#a0a1a7",
			"#4f5261", "#e06c75", "#50a14f", "#c18401",
			"#4078f2", "#a626a4", "#0184bc", "#ffffff",
		},
	},
	// ---- Batman ----
	{
		Name:       "Batman",
		Foreground: "#fcfaf2",
		Background: "#1b1d1e",
		Cursor:     "#fcfaf2",
		Colors: []string{
			"#1b1d1e", "#e6dc30", "#1e9e29", "#e6dc30",
			"#3799e6", "#e6dc30", "#1e9e29", "#fcfaf2",
			"#1b1d1e", "#e6dc30", "#1e9e29", "#e6dc30",
			"#3799e6", "#e6dc30", "#1e9e29", "#fcfaf2",
		},
	},
}

// GetBuiltInScheme returns a built-in color scheme by name, or nil if not found.
func GetBuiltInScheme(name string) *ColorScheme {
	for i := range BuiltInSchemes {
		if BuiltInSchemes[i].Name == name {
			return &BuiltInSchemes[i]
		}
	}
	// Also check community schemes
	for i := range CommunitySchemes {
		if CommunitySchemes[i].Name == name {
			return &CommunitySchemes[i]
		}
	}
	return nil
}

// GetSchemeNames returns the names of all available color schemes.
func GetSchemeNames() []string {
	all := AllSchemes()
	names := make([]string, len(all))
	for i, s := range all {
		names[i] = s.Name
	}
	return names
}

// ToXTermTheme converts a ColorScheme into an xterm.js-compatible theme map.
// This is used by the frontend to apply the color scheme to the terminal.
func (cs *ColorScheme) ToXTermTheme() map[string]string {
	theme := map[string]string{
		"foreground": cs.Foreground,
		"background": cs.Background,
		"cursor":     cs.Cursor,
	}
	if cs.CursorAccent != "" {
		theme["cursorAccent"] = cs.CursorAccent
	}
	if cs.Selection != "" {
		theme["selectionBackground"] = cs.Selection
	}
	if cs.SelectionForeground != "" {
		theme["selectionForeground"] = cs.SelectionForeground
	}
	// Map 16 ANSI colors to xterm.js slot names
	colorSlots := []string{
		"black", "red", "green", "yellow",
		"blue", "magenta", "cyan", "white",
		"brightBlack", "brightRed", "brightGreen", "brightYellow",
		"brightBlue", "brightMagenta", "brightCyan", "brightWhite",
	}
	for i, name := range colorSlots {
		if i < len(cs.Colors) {
			theme[name] = cs.Colors[i]
		}
	}
	return theme
}

// IsLight returns true if the color scheme has a light background.
// This is determined by comparing the perceived luminance of the background color.
func (cs *ColorScheme) IsLight() bool {
	return isLightColor(cs.Background)
}

// isLightColor determines if a hex color is "light" based on perceived luminance.
func isLightColor(hex string) bool {
	if len(hex) < 7 || hex[0] != '#' {
		return false
	}
	r := hexByte(hex, 1)
	g := hexByte(hex, 3)
	b := hexByte(hex, 5)
	// Perceived luminance formula
	luminance := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 255.0
	return luminance > 0.5
}

func hexByte(hex string, start int) uint8 {
	if len(hex) < start+2 {
		return 0
	}
	var val uint8
	for i := 0; i < 2; i++ {
		c := hex[start+i]
		val <<= 4
		switch {
		case c >= '0' && c <= '9':
			val |= c - '0'
		case c >= 'a' && c <= 'f':
			val |= c - 'a' + 10
		case c >= 'A' && c <= 'F':
			val |= c - 'A' + 10
		}
	}
	return val
}
