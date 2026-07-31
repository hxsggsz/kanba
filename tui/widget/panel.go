package widget

import (
	"strings"

	models "kanba/tui/models"
	"kanba/tui/diff"

	"charm.land/lipgloss/v2"
	"kanba/git"
)

type Panel struct {
	width       int
	theme       models.Theme
	diffs       []git.SideBySideDiff
	flatLines   []diff.FlatLine
	fileStats   []diff.FileStat
	scroller    *diff.Scroller
	highlighter *diff.SyntaxHighlighter
}

func NewPanel() *Panel {
	return &Panel{}
}

func (p *Panel) Width(w int) *Panel                         { p.width = w; return p }
func (p *Panel) Theme(t models.Theme) *Panel                 { p.theme = t; return p }
func (p *Panel) Diffs(d []git.SideBySideDiff) *Panel         { p.diffs = d; return p }
func (p *Panel) FlatLines(f []diff.FlatLine) *Panel          { p.flatLines = f; return p }
func (p *Panel) FileStats(s []diff.FileStat) *Panel          { p.fileStats = s; return p }
func (p *Panel) Scroller(s *diff.Scroller) *Panel            { p.scroller = s; return p }
func (p *Panel) Highlighter(h *diff.SyntaxHighlighter) *Panel { p.highlighter = h; return p }

func (p *Panel) Render(vis int) string {
	total := len(p.flatLines)
	if total == 0 {
		return ""
	}
	if vis <= 0 {
		p.scroller.UpdateScroll(total, vis)
		return ""
	}

	p.scroller.UpdateScroll(total, vis)
	hScroll := p.scroller.HScroll()

	start := p.scroller.Scroll()
	end := min(start+vis, total)

	innerWidth := max(p.width-2, 0)

	type fileBlock struct {
		lines []string
	}
	var blocks []fileBlock
	var cur []string
	curFile := -1

	flush := func() {
		if cur != nil {
			blocks = append(blocks, fileBlock{lines: cur})
			cur = nil
		}
	}

	for gi := start; gi < end; gi++ {
		fl := p.flatLines[gi]

		if fl.FileIdx != curFile {
			flush()
			curFile = fl.FileIdx
		}

		var line string
		if fl.IsHeader {
			line = p.renderFileHeader(fl, innerWidth)
		} else {
			f := p.diffs[fl.FileIdx]
			h := f.Hunks[fl.HunkIdx]
			ln := h.Lines[fl.LineIdx]
			fmtr := diff.DefaultFormatters[ln.Kind]

			singlePanel := f.Status == "A" || f.Status == "D"
			singlePanelLeft := f.Status == "D"
			colWidth := innerWidth
			if !singlePanel {
				colWidth = (innerWidth - 3) / 2
			}

			line = diff.RenderAlignedLine(fmtr, ln, colWidth, p.highlighter, f.NewPath, hScroll, singlePanel, singlePanelLeft, p.theme)
		}

		cur = append(cur, line)
	}
	flush()

	var rendered []string
	borderStyle := BgBox(p.theme.PanelBg, p.width)

	for _, b := range blocks {
		content := strings.Join(b.lines, "\n")
		rendered = append(rendered, borderStyle.Render(content))
	}

	return strings.Join(rendered, "\n")
}

func (p *Panel) renderFileHeader(fl diff.FlatLine, colWidth int) string {
	f := p.diffs[fl.FileIdx]
	stats := p.fileStats[fl.FileIdx]

	bgColor := p.theme.PanelHeaderBg

	bg := lipgloss.NewStyle().Background(lipgloss.Color(bgColor))
	normalStyle := bg.Foreground(lipgloss.Color(p.theme.ContextFg))
	addStyle := bg.Foreground(lipgloss.Color(p.theme.SidebarAdded))
	delStyle := bg.Foreground(lipgloss.Color(p.theme.SidebarDeleted))

	var segs []string
	segs = append(segs, normalStyle.Render(" "+f.NewPath))

	if statsStr := RenderStats(stats.Added, stats.Deleted, addStyle, delStyle, normalStyle); statsStr != "" {
		segs = append(segs, statsStr)
	}

	text := strings.Join(segs, "")

	style := lipgloss.NewStyle().
		Background(lipgloss.Color(bgColor)).
		MarginBackground(lipgloss.Color(bgColor)).
		Padding(1, 1).
		Width(colWidth)
	return style.Render(text)
}
