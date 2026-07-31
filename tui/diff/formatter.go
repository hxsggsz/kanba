package diff

import (
	"fmt"
	"strconv"
	"strings"

	models "kanba/tui/models"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"kanba/git"
)

// LineNumColWidth is the fixed width reserved for a line-number column.
// PrefixWidth below encodes the full prefix layout: line number + space +
// 1-char kind prefix ("+"/"-"/" ") + space, i.e. LineNumColWidth+3.
const LineNumColWidth = 4

// PrefixWidth is the single source of truth for how wide the rendered
// "<linenum> <prefix> " gutter is. Anything that needs to skip past the
// gutter to reach line content (in this package or elsewhere) must use
// this constant instead of re-deriving LineNumColWidth+3 independently.
const PrefixWidth = LineNumColWidth + 3

// LineFormatter describes how a diff line is rendered on each side of the
// panel. Styling (colors/background) is computed independently in
// renderStyledLine via theme.BgFor/theme.LineNumFg, not through this
// interface.
type LineFormatter interface {
	LeftContent(ln git.AlignedLine) string
	RightContent(ln git.AlignedLine) string
	LeftPrefix(ln git.AlignedLine) string
	RightPrefix(ln git.AlignedLine) string
}

// lineSpec is the data-driven description of how a given git.LineKind is
// laid out: which side(s) carry content and what single-character prefix
// each side shows.
type lineSpec struct {
	leftHasContent  bool
	rightHasContent bool
	leftPrefix      string
	rightPrefix     string
}

func (s lineSpec) LeftContent(ln git.AlignedLine) string {
	if !s.leftHasContent {
		return ""
	}
	return ln.OldContent
}

func (s lineSpec) RightContent(ln git.AlignedLine) string {
	if !s.rightHasContent {
		return ""
	}
	return ln.NewContent
}

func (s lineSpec) LeftPrefix(git.AlignedLine) string  { return s.leftPrefix }
func (s lineSpec) RightPrefix(git.AlignedLine) string { return s.rightPrefix }

var lineSpecs = map[git.LineKind]lineSpec{
	git.KindContext:  {leftHasContent: true, rightHasContent: true, leftPrefix: " ", rightPrefix: " "},
	git.KindAdded:    {leftHasContent: false, rightHasContent: true, leftPrefix: " ", rightPrefix: "+"},
	git.KindDeleted:  {leftHasContent: true, rightHasContent: false, leftPrefix: "-", rightPrefix: " "},
	git.KindModified: {leftHasContent: true, rightHasContent: true, leftPrefix: "-", rightPrefix: "+"},
}

var DefaultFormatters = NewDefaultFormatters()

func NewDefaultFormatters() map[git.LineKind]LineFormatter {
	formatters := make(map[git.LineKind]LineFormatter, len(lineSpecs))
	for kind, spec := range lineSpecs {
		formatters[kind] = spec
	}
	return formatters
}

func RenderAlignedLine(f LineFormatter, ln git.AlignedLine, colWidth int, sh *SyntaxHighlighter, filePath string, hScroll int, singlePanel bool, singlePanelLeft bool, theme models.Theme) string {
	oldNum := ""
	if ln.OldLineNum > 0 {
		oldNum = strconv.Itoa(ln.OldLineNum)
	}
	newNum := ""
	if ln.NewLineNum > 0 {
		newNum = strconv.Itoa(ln.NewLineNum)
	}

	leftContent := f.LeftContent(ln)
	rightContent := f.RightContent(ln)

	contentAreaWidth := colWidth - PrefixWidth
	if contentAreaWidth < 0 {
		contentAreaWidth = 0
	}
	leftContent = ansi.Cut(leftContent, hScroll, hScroll+contentAreaWidth)
	rightContent = ansi.Cut(rightContent, hScroll, hScroll+contentAreaWidth)

	if singlePanel {
		if singlePanelLeft {
			prefix := fmt.Sprintf("%*s %s ", LineNumColWidth, oldNum, f.LeftPrefix(ln))
			return renderStyledLine(prefix, leftContent, colWidth, ln.Kind, true, sh, filePath, theme)
		}
		prefix := fmt.Sprintf("%*s %s ", LineNumColWidth, newNum, f.RightPrefix(ln))
		return renderStyledLine(prefix, rightContent, colWidth, ln.Kind, false, sh, filePath, theme)
	}

	leftPrefix := fmt.Sprintf("%*s %s ", LineNumColWidth, oldNum, f.LeftPrefix(ln))
	rightPrefix := fmt.Sprintf("%*s %s ", LineNumColWidth, newNum, f.RightPrefix(ln))

	leftRendered := renderStyledLine(leftPrefix, leftContent, colWidth, ln.Kind, true, sh, filePath, theme)
	rightRendered := renderStyledLine(rightPrefix, rightContent, colWidth, ln.Kind, false, sh, filePath, theme)

	return leftRendered + rightRendered
}

func renderStyledLine(prefix, content string, width int, kind git.LineKind, isLeft bool, sh *SyntaxHighlighter, filePath string, theme models.Theme) string {
	bgColor := theme.BgFor(kind, isLeft)
	if bgColor == "" {
		bgColor = theme.PanelBg
	}

	numBg := bgColor
	if kind == git.KindContext {
		numBg = theme.LineNumberBg
	}

	numStyle := lipgloss.NewStyle()
	if fg := theme.LineNumFg(kind, isLeft); fg != "" {
		numStyle = numStyle.Foreground(lipgloss.Color(fg))
	}
	numStyle = numStyle.Background(lipgloss.Color(numBg))

	prefixRendered := numStyle.Render(prefix)

	var contentRendered string
	if sh != nil {
		baseStyle := lipgloss.NewStyle().Background(lipgloss.Color(bgColor)).Foreground(lipgloss.Color(theme.ContextFg))
		contentRendered = sh.HighlightWithStyle(content, filePath, baseStyle, theme)
	} else {
		contentRendered = content
	}

	styled := prefixRendered + contentRendered
	vis := lipgloss.Width(styled)
	if vis > width {
		styled = ansi.Truncate(styled, width, "")
		vis = width
	}
	if vis < width {
		padStyle := lipgloss.NewStyle().Background(lipgloss.Color(bgColor))
		styled += padStyle.Render(strings.Repeat(" ", width-vis))
	}

	return styled
}
