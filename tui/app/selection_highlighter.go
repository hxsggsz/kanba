package app

import (
	"strings"

	"kanba/tui/diff"
	"kanba/tui/models"
	"kanba/tui/selection"

	"github.com/charmbracelet/x/ansi"
)

type SelectionHighlighter struct {
	sel         *selection.Selection
	colWidth    int
	prefixWidth int
}

func NewSelectionHighlighter(sel *selection.Selection, colWidth int) *SelectionHighlighter {
	return &SelectionHighlighter{
		sel:         sel,
		colWidth:    colWidth,
		prefixWidth: diff.PrefixWidth,
	}
}

func (h *SelectionHighlighter) visualColumns(flatIdx int) (int, int, bool) {
	normalized := h.sel.Range.Normalized()

	if flatIdx < normalized.StartLine || flatIdx > normalized.EndLine {
		return 0, 0, false
	}

	startCol := normalized.StartCol
	if flatIdx != normalized.StartLine {
		startCol = 0
	}

	endCol := normalized.EndCol
	if flatIdx != normalized.EndLine {
		endCol = h.colWidth
	}

	startCol += h.prefixWidth
	endCol += h.prefixWidth

	if h.sel.Panel == selection.PanelRight {
		startCol += h.colWidth
		endCol += h.colWidth
	}

	return startCol, endCol, true
}

func (h *SelectionHighlighter) ProcessLine(line string, flatIdx int, theme models.Theme) (string, string) {
	startCol, endCol, inRange := h.visualColumns(flatIdx)
	if !inRange {
		return line, ""
	}

	// NOTE: the actual highlighting logic (highlightColumns) lives in
	// view.go rather than on SelectionHighlighter itself. This type is
	// mostly a column-range calculator that delegates rendering; keep
	// that coupling in mind if either side changes independently.
	highlighted := highlightColumns(line, startCol, endCol, theme.CursorBg)
	plainText := extractPlainTextFromRendered(line, startCol, endCol)

	return highlighted, plainText
}

func extractPlainTextFromRendered(line string, startCol, endCol int) string {
	if startCol >= endCol {
		return ""
	}

	visWidth := ansi.StringWidth(line)
	if startCol >= visWidth {
		return ""
	}

	if endCol > visWidth {
		endCol = visWidth
	}

	selected := ansi.Cut(line, startCol, endCol)
	return strings.TrimSpace(ansi.Strip(selected))
}
