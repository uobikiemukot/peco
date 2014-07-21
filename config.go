package peco

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nsf/termbox-go"
)

var homedirFunc = homedir

// Config holds all the data that can be configured in the
// external configuran file
type Config struct {
	Action        map[string][]string `json:"Action"`
	// Keymap used to be directly responsible for dispatching
	// events against user input, but since then this has changed
	// into something that just records the user's config input
	Keymap        map[string]string `json:"Keymap"`
	Matcher       string   `json:"Matcher"`
	Style         StyleSet `json:"Style"`
	CustomMatcher map[string][]string
	Prompt        string   `json:"Prompt"`
}

// NewConfig creates a new Config
func NewConfig() *Config {
	return &Config{
		Keymap:  make(map[string]string),
		Matcher: IgnoreCaseMatch,
		Style:   NewStyleSet(),
		Prompt:  "QUERY>",
	}
}

// ReadFilename reads the config from the given file, and
// does the appropriate processing, if any
func (c *Config) ReadFilename(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(c)
	if err != nil {
		return err
	}

	return nil
}

var (
	stringToFg = map[string]termbox.Attribute{
		"default": termbox.ColorDefault,
		"black":   termbox.ColorBlack,
		"red":     termbox.ColorRed,
		"green":   termbox.ColorGreen,
		"yellow":  termbox.ColorYellow,
		"blue":    termbox.ColorBlue,
		"magenta": termbox.ColorMagenta,
		"cyan":    termbox.ColorCyan,
		"white":   termbox.ColorWhite,
	}
	stringToBg = map[string]termbox.Attribute{
		"on_default": termbox.ColorDefault,
		"on_black":   termbox.ColorBlack,
		"on_red":     termbox.ColorRed,
		"on_green":   termbox.ColorGreen,
		"on_yellow":  termbox.ColorYellow,
		"on_blue":    termbox.ColorBlue,
		"on_magenta": termbox.ColorMagenta,
		"on_cyan":    termbox.ColorCyan,
		"on_white":   termbox.ColorWhite,
	}
	stringToFgAttr = map[string]termbox.Attribute{
		"bold":      termbox.AttrBold,
		"underline": termbox.AttrUnderline,
		"blink":     termbox.AttrReverse,
	}
	stringToBgAttr = map[string]termbox.Attribute{
		"blink": termbox.AttrBold,
	}
)

// StyleSet holds styles for various sections
type StyleSet struct {
	Basic          Style `json:"Basic"`
	SavedSelection Style `json:"SavedSelection"`
	Selected       Style `json:"Selected"`
	Query          Style `json:"Query"`
	Matched        Style `json:"Matched"`
}

// NewStyleSet creates a new StyleSet struct
func NewStyleSet() StyleSet {
	return StyleSet{
		Basic:          Style{fg: termbox.ColorDefault, bg: termbox.ColorDefault},
		SavedSelection: Style{fg: termbox.ColorBlack | termbox.AttrBold, bg: termbox.ColorCyan},
		Selected:       Style{fg: termbox.ColorDefault | termbox.AttrUnderline, bg: termbox.ColorMagenta},
		Query:          Style{fg: termbox.ColorDefault, bg: termbox.ColorDefault},
		Matched:        Style{fg: termbox.ColorCyan, bg: termbox.ColorDefault},
	}
}

// Style describes termbox styles
type Style struct {
	fg termbox.Attribute
	bg termbox.Attribute
}

// UnmarshalJSON satisfies json.RawMessage.
func (s *Style) UnmarshalJSON(buf []byte) error {
	raw := []string{}
	if err := json.Unmarshal(buf, &raw); err != nil {
		return err
	}
	*s = *stringsToStyle(raw)
	return nil
}

func stringsToStyle(raw []string) *Style {
	style := &Style{
		fg: termbox.ColorDefault,
		bg: termbox.ColorDefault,
	}

	for _, s := range raw {
		fg, ok := stringToFg[s]
		if ok {
			style.fg = fg
		}

		bg, ok := stringToBg[s]
		if ok {
			style.bg = bg
		}
	}

	for _, s := range raw {
		fg_attr, ok := stringToFgAttr[s]
		if ok {
			style.fg |= fg_attr
		}

		bg_attr, ok := stringToBgAttr[s]
		if ok {
			style.bg |= bg_attr
		}
	}

	return style
}

var _locateRcfileIn = locateRcfileIn

func locateRcfileIn(dir string) (string, error) {
	const basename = "config.json"
	file := filepath.Join(dir, basename)
	if _, err := os.Stat(file); err != nil {
		return "", err
	}
	return file, nil
}

// LocateRcfile attempts to find the config file in various locations
func LocateRcfile() (string, error) {
	// http://standards.freedesktop.org/basedir-spec/basedir-spec-latest.html
	//
	// Try in this order:
	//	  $XDG_CONFIG_HOME/peco/config.json
	//    $XDG_CONFIG_DIR/peco/config.json (where XDG_CONFIG_DIR is listed in $XDG_CONFIG_DIRS)
	//	  ~/.peco/config.json

	home, uErr := homedirFunc()

	// Try dir supplied via env var
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		file, err := _locateRcfileIn(filepath.Join(dir, "peco"))
		if err == nil {
			return file, nil
		}
	} else if uErr == nil { // silently ignore failure for homedir()
		// Try "default" XDG location, is user is available
		file, err := _locateRcfileIn(filepath.Join(home, ".config", "peco"))
		if err == nil {
			return file, nil
		}
	}

	// this standard does not take into consideration windows (duh)
	// while the spec says use ":" as the separator, Go provides us
	// with filepath.ListSeparator, so use it
	if dirs := os.Getenv("XDG_CONFIG_DIRS"); dirs != "" {
		for _, dir := range strings.Split(dirs, fmt.Sprintf("%c", filepath.ListSeparator)) {
			file, err := _locateRcfileIn(filepath.Join(dir, "peco"))
			if err == nil {
				return file, nil
			}
		}
	}

	if uErr == nil { // silently ignore failure for homedir()
		file, err := _locateRcfileIn(filepath.Join(home, ".peco"))
		if err == nil {
			return file, nil
		}
	}

	return "", fmt.Errorf("error: Config file not found")
}
