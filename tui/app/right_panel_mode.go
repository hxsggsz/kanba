package app

import (
	"fmt"
	"strings"

	"kanba/tui/diff"
	"kanba/tui/models"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"kanba/git"
)

type RightPanelMode struct{}

func (m *RightPanelMode) Render(model *Model) string {
	if len(model.flatLines) == 0 {
		return ""
	}

	theme := model.CurrentTheme()
	contentVis := model.VisibleLines()
	panelWidth := max(model.width-panelBorderWidth, panelMinWidth)
	content := m.renderSinglePanel(model, panelWidth, contentVis)

	return renderFrame(model, theme, "", content)
}

func (m *RightPanelMode) renderSinglePanel(model *Model, width int, vis int) string {
	total := len(model.flatLines)
	if total == 0 {
		return ""
	}
	if vis <= 0 {
		model.scroller.UpdateScroll(total, vis)
		return ""
	}

	theme := model.CurrentTheme()
	model.scroller.UpdateScroll(total, vis)
	hScroll := model.scroller.HScroll()

	start := model.scroller.Scroll()
	end := min(start+vis, total)

	contentAreaWidth := width - (diff.LineNumColWidth + 3)
	if contentAreaWidth < 0 {
		contentAreaWidth = 0
	}

	selHighlighter := model.buildSelectionHighlighter(width)

	var lines []string
	var selectedTextParts []string

	if si := stickyHeaderIndex(model.flatLines, start); si >= 0 {
		line, selText := model.renderStickyHeader(model.flatLines[si], width, selHighlighter, si, theme)
		lines = append(lines, line)
		if selText != "" {
			selectedTextParts = append(selectedTextParts, selText)
		}
		end = min(start+vis-1, total)
	}

	end, needsBorder := model.reserveLastPanelBorder(start, end)

	for gi := start; gi < end; gi++ {
		fl := model.flatLines[gi]

		if fl.IsHeader {
			lines = append(lines, model.renderFileHeader(fl, width))
		} else if fl.IsHunkHeader {
			lines = append(lines, model.renderHunkHeader(fl, width))
		} else {
			f := model.diffs[fl.FileIdx]
			h := f.Hunks[fl.HunkIdx]
			ln := h.Lines[fl.LineIdx]

			newNum := ""
			if ln.NewLineNum > 0 {
				newNum = fmt.Sprintf("%d", ln.NewLineNum)
			}

			var prefix, content string
			var kind git.LineKind
			var isLeft bool

			switch ln.Kind {
			case git.KindAdded:
				prefix = "+"
				content = ln.NewContent
				kind = git.KindAdded
				isLeft = false
			case git.KindDeleted:
				prefix = "-"
				content = ln.OldContent
				kind = git.KindDeleted
				isLeft = true
			case git.KindModified:
				prefix = "+"
				content = ln.NewContent
				kind = git.KindAdded
				isLeft = false
			default:
				prefix = " "
				content = ln.NewContent
				kind = git.KindContext
				isLeft = false
			}

			content = ansi.Cut(content, hScroll, hScroll+contentAreaWidth)

			prefixStr := fmt.Sprintf("%*s %s ", diff.LineNumColWidth, newNum, prefix)
			line := renderStyledLine(prefixStr, content, width, kind, isLeft, model.highlighter, f.NewPath, theme)
			lines = append(lines, line)
		}
	}

	if needsBorder {
		borderStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.SidebarDir)).
			Background(lipgloss.Color(theme.PanelBg)).
			Render(strings.Repeat("─", width))
		lines = append(lines, borderStyle)
		marginStyle := lipgloss.NewStyle().
			Background(lipgloss.Color(theme.PanelBg)).
			Render(strings.Repeat(" ", width))
		lines = append(lines, marginStyle)
	}

	fillStyle := lipgloss.NewStyle().Background(lipgloss.Color(theme.PanelBg))
	for len(lines) < vis {
		lines = append(lines, fillStyle.Render(strings.Repeat(" ", width)))
	}

	model.selectedText = strings.Join(selectedTextParts, "\n")
	return strings.Join(lines, "\n")
}

func renderStyledLine(prefix, content string, width int, kind git.LineKind, isLeft bool, sh *diff.SyntaxHighlighter, filePath string, theme models.Theme) string {
	bgColor := theme.BgFor(kind, isLeft)
	if bgColor == "" {
		bgColor = theme.PanelBg
	}

	numBg := bgColor
	if kind == git.KindContext {
		numBg = theme.LineNumberBg
	}

	baseStyle := lipgloss.NewStyle().Background(lipgloss.Color(bgColor)).Foreground(lipgloss.Color(theme.ContextFg))

	numStyle := lipgloss.NewStyle()
	if fg := theme.LineNumFg(kind, isLeft); fg != "" {
		numStyle = numStyle.Foreground(lipgloss.Color(fg))
	}
	numStyle = numStyle.Background(lipgloss.Color(numBg))

	prefixRendered := numStyle.Render(prefix)

	var contentRendered string
	if sh != nil {
		contentRendered = sh.HighlightWithStyle(content, filePath, baseStyle, theme)
	} else if bgColor != "" {
		contentRendered = baseStyle.Render(content)
	} else {
		contentRendered = content
	}

	styled := prefixRendered + contentRendered
	vis := lipgloss.Width(styled)
	if vis < width {
		padStyle := baseStyle
		styled += padStyle.Render(strings.Repeat(" ", width-vis))
	}

	return styled
}
