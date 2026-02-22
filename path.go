package x

import (
	"regexp"
	"strings"
)

var (
	invalidChars = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)

	// Reserved device names on Windows.
	//nolint:gochecknoglobals
	reservedNames = map[string]bool{
		"CON": true, "PRN": true, "AUX": true, "NUL": true,
		"COM1": true, "COM2": true, "COM3": true, "COM4": true,
		"COM5": true, "COM6": true, "COM7": true, "COM8": true, "COM9": true,
		"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true,
		"LPT5": true, "LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
	}
)

// SafeFilename returns a safe filename for the given name.
func SafeFilename(name string) string {
	name = strings.TrimSpace(name)

	// Windows does not allow trailing spaces or dots.
	name = strings.TrimRight(name, ". ")

	name = invalidChars.ReplaceAllString(name, "_")

	// Prevent empty filename.
	if name == "" {
		return "_"
	}

	// Check reserved device names (case-insensitive, without extension).
	base := strings.ToUpper(name)
	ext := ""
	if dot := strings.Index(name, "."); dot != -1 {
		base = strings.ToUpper(name[:dot])
		ext = name[dot:]
	}

	if reservedNames[base] {
		name = "_" + base + ext
	}

	return name
}
