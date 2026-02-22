package x_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/tozd/go/x"
)

func TestSafeFilename(t *testing.T) { //nolint:maintidx
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Normal filenames - should remain unchanged.
		{
			name:     "simple filename",
			input:    "file.txt",
			expected: "file.txt",
		},
		{
			name:     "filename with spaces",
			input:    "my file.txt",
			expected: "my file.txt",
		},
		{
			name:     "filename with underscores",
			input:    "my_file_name.txt",
			expected: "my_file_name.txt",
		},
		{
			name:     "filename with dashes",
			input:    "my-file-name.txt",
			expected: "my-file-name.txt",
		},

		// Empty and whitespace.
		{
			name:     "empty string",
			input:    "",
			expected: "_",
		},
		{
			name:     "only spaces",
			input:    "   ",
			expected: "_",
		},
		{
			name:     "only dots",
			input:    "...",
			expected: "_",
		},
		{
			name:     "spaces and dots",
			input:    " . . ",
			expected: "_",
		},

		// Leading/trailing spaces.
		{
			name:     "leading spaces",
			input:    "  file.txt",
			expected: "file.txt",
		},
		{
			name:     "trailing spaces",
			input:    "file.txt  ",
			expected: "file.txt",
		},
		{
			name:     "leading and trailing spaces",
			input:    "  file.txt  ",
			expected: "file.txt",
		},

		// Trailing dots and spaces (Windows requirement).
		{
			name:     "trailing dot",
			input:    "file.txt.",
			expected: "file.txt",
		},
		{
			name:     "trailing dots",
			input:    "file.txt...",
			expected: "file.txt",
		},
		{
			name:     "trailing space and dot",
			input:    "file.txt .",
			expected: "file.txt",
		},
		{
			name:     "trailing dots and spaces mixed",
			input:    "file.txt. . .",
			expected: "file.txt",
		},

		// Invalid characters.
		{
			name:     "less than",
			input:    "file<name.txt",
			expected: "file_name.txt",
		},
		{
			name:     "greater than",
			input:    "file>name.txt",
			expected: "file_name.txt",
		},
		{
			name:     "colon",
			input:    "file:name.txt",
			expected: "file_name.txt",
		},
		{
			name:     "double quote",
			input:    `file"name.txt`,
			expected: "file_name.txt",
		},
		{
			name:     "forward slash",
			input:    "file/name.txt",
			expected: "file_name.txt",
		},
		{
			name:     "backslash",
			input:    `file\name.txt`,
			expected: "file_name.txt",
		},
		{
			name:     "pipe",
			input:    "file|name.txt",
			expected: "file_name.txt",
		},
		{
			name:     "question mark",
			input:    "file?name.txt",
			expected: "file_name.txt",
		},
		{
			name:     "asterisk",
			input:    "file*name.txt",
			expected: "file_name.txt",
		},
		{
			name:     "null character",
			input:    "file\x00name.txt",
			expected: "file_name.txt",
		},
		{
			name:     "control characters",
			input:    "file\x01\x02\x1Fname.txt",
			expected: "file___name.txt",
		},
		{
			name:     "multiple invalid characters",
			input:    `file<>:"/\|?*name.txt`,
			expected: "file_________name.txt",
		},

		// Reserved device names - exact match (case-insensitive).
		{
			name:     "CON uppercase",
			input:    "CON",
			expected: "_CON",
		},
		{
			name:     "con lowercase",
			input:    "con",
			expected: "_CON",
		},
		{
			name:     "Con mixed case",
			input:    "Con",
			expected: "_CON",
		},
		{
			name:     "PRN",
			input:    "PRN",
			expected: "_PRN",
		},
		{
			name:     "AUX",
			input:    "AUX",
			expected: "_AUX",
		},
		{
			name:     "NUL",
			input:    "NUL",
			expected: "_NUL",
		},
		{
			name:     "COM1",
			input:    "COM1",
			expected: "_COM1",
		},
		{
			name:     "COM9",
			input:    "COM9",
			expected: "_COM9",
		},
		{
			name:     "LPT1",
			input:    "LPT1",
			expected: "_LPT1",
		},
		{
			name:     "LPT9",
			input:    "LPT9",
			expected: "_LPT9",
		},

		// Reserved names with extensions.
		{
			name:     "CON with extension",
			input:    "CON.txt",
			expected: "_CON.txt",
		},
		{
			name:     "con.txt lowercase",
			input:    "con.txt",
			expected: "_CON.txt",
		},
		{
			name:     "PRN.log",
			input:    "PRN.log",
			expected: "_PRN.log",
		},
		{
			name:     "AUX.dat",
			input:    "AUX.dat",
			expected: "_AUX.dat",
		},
		{
			name:     "NUL.txt",
			input:    "NUL.txt",
			expected: "_NUL.txt",
		},
		{
			name:     "COM1.txt",
			input:    "COM1.txt",
			expected: "_COM1.txt",
		},
		{
			name:     "LPT1.txt",
			input:    "LPT1.txt",
			expected: "_LPT1.txt",
		},

		// Reserved names with multiple dots.
		{
			name:     "CON with multiple extensions",
			input:    "CON.tar.gz",
			expected: "_CON.tar.gz",
		},

		// Not reserved - similar but different.
		{
			name:     "CON2 not reserved",
			input:    "CON2",
			expected: "CON2",
		},
		{
			name:     "COM10 not reserved",
			input:    "COM10",
			expected: "COM10",
		},
		{
			name:     "LPT0 not reserved",
			input:    "LPT0",
			expected: "LPT0",
		},
		{
			name:     "ACON not reserved",
			input:    "ACON",
			expected: "ACON",
		},
		{
			name:     "CONA not reserved",
			input:    "CONA",
			expected: "CONA",
		},
		{
			name:     "PRNT not reserved",
			input:    "PRNT",
			expected: "PRNT",
		},

		// Complex mixed scenarios.
		{
			name:     "leading/trailing spaces with invalid chars",
			input:    "  file<name>  ",
			expected: "file_name_",
		},
		{
			name:     "reserved name with spaces",
			input:    " CON ",
			expected: "_CON",
		},
		{
			name:     "reserved name with trailing dot",
			input:    "CON.",
			expected: "_CON",
		},
		{
			name:     "reserved name with trailing space",
			input:    "CON ",
			expected: "_CON",
		},
		{
			name:     "invalid chars and trailing dot",
			input:    "file|name.",
			expected: "file_name",
		},
		{
			name:     "all invalid becomes underscore",
			input:    "<>:|",
			expected: "____",
		},
		{
			name:     "spaces with invalid chars",
			input:    " <> ",
			expected: "__",
		},

		// Edge cases with dots.
		{
			name:     "filename starting with dot",
			input:    ".hidden",
			expected: ".hidden",
		},
		{
			name:     "multiple dots in filename",
			input:    "file.name.txt",
			expected: "file.name.txt",
		},
		{
			name:     "dot in middle with trailing dot",
			input:    "file.name.",
			expected: "file.name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := x.SafeFilename(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
