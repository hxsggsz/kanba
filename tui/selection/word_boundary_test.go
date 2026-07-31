package selection

import (
	"testing"
)

func TestIsWordChar(t *testing.T) {
	tests := []struct {
		r        rune
		expected bool
	}{
		{'a', true},
		{'Z', true},
		{'0', true},
		{'9', true},
		{'_', true},
		{' ', false},
		{'.', false},
		{'-', false},
		{'!', false},
		{'@', false},
		{'ç', true},
		{'é', true},
		{'中', true},
	}

	for _, tt := range tests {
		t.Run(string(tt.r), func(t *testing.T) {
			if got := IsWordChar(tt.r); got != tt.expected {
				t.Errorf("IsWordChar(%q) = %v, want %v", tt.r, got, tt.expected)
			}
		})
	}
}

func TestFindWordBoundaries(t *testing.T) {
	tests := []struct {
		name          string
		line          string
		col           int
		expectedStart int
		expectedEnd   int
	}{
		{
			name:          "middle of word",
			line:          "hello world",
			col:           2,
			expectedStart: 0,
			expectedEnd:   5,
		},
		{
			name:          "start of word",
			line:          "hello world",
			col:           0,
			expectedStart: 0,
			expectedEnd:   5,
		},
		{
			name:          "end of word",
			line:          "hello world",
			col:           4,
			expectedStart: 0,
			expectedEnd:   5,
		},
		{
			name:          "on space",
			line:          "hello world",
			col:           5,
			expectedStart: 5,
			expectedEnd:   5,
		},
		{
			name:          "single char word",
			line:          "a b",
			col:           0,
			expectedStart: 0,
			expectedEnd:   1,
		},
		{
			name:          "end of line",
			line:          "hello",
			col:           4,
			expectedStart: 0,
			expectedEnd:   5,
		},
		{
			name:          "empty line",
			line:          "",
			col:           0,
			expectedStart: 0,
			expectedEnd:   0,
		},
		{
			name:          "underscores in word",
			line:          "foo_bar baz",
			col:           2,
			expectedStart: 0,
			expectedEnd:   7,
		},
		{
			name:          "underscore at cursor",
			line:          "foo_bar baz",
			col:           3,
			expectedStart: 0,
			expectedEnd:   7,
		},
		{
			name:          "UTF-8 characters",
			line:          "café résumé",
			col:           2,
			expectedStart: 0,
			expectedEnd:   4,
		},
		{
			name:          "UTF-8 middle of word",
			line:          "café résumé",
			col:           8,
			expectedStart: 5,
			expectedEnd:   11,
		},
		{
			name:          "col beyond line length",
			line:          "hello",
			col:           10,
			expectedStart: 0,
			expectedEnd:   5,
		},
		{
			name:          "negative col",
			line:          "hello",
			col:           -1,
			expectedStart: 0,
			expectedEnd:   5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := findWordBoundaries(tt.line, tt.col)
			if start != tt.expectedStart || end != tt.expectedEnd {
				t.Errorf("findWordBoundaries(%q, %d) = (%d, %d), want (%d, %d)",
					tt.line, tt.col, start, end, tt.expectedStart, tt.expectedEnd)
			}
		})
	}
}
