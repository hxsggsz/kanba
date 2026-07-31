package widget

import (
	"path/filepath"
	"strings"
	"unicode/utf8"

	models "kanba/tui/models"
	"kanba/tui/diff"

	"charm.land/lipgloss/v2"
	"kanba/git"
)

const (
	statusBarHeight     = 4
	sidebarMinWidth     = 15
	sidebarMaxWidth     = 35
	sidebarDefaultWidth = 25
	sidebarDenominator  = 4
)

type visualEntry struct {
	isDir   bool
	dir     string
	fileIdx int
	file    *git.SideBySideDiff
}

type Sidebar struct {
	files         []git.SideBySideDiff
	fileIdx       int
	width         int
	maxLines      int
	contentHeight int
	theme         models.Theme
	stats         []diff.FileStat
}

func NewSidebar(files []git.SideBySideDiff, fileIdx int, width int, height int, theme models.Theme, stats []diff.FileStat) *Sidebar {
	maxLines := max(height-statusBarHeight, 1)
	return &Sidebar{
		files:    files,
		fileIdx:  fileIdx,
		width:    width,
		maxLines: maxLines,
		theme:    theme,
		stats:    stats,
	}
}

func (s *Sidebar) Render() string {
	entries := buildVisualEntries(s.files)
	visible, _ := visibleWindow(entries, s.fileIdx, s.maxLines)

	var sb strings.Builder
	lineCount := 0
	for _, e := range visible {
		var line string
		if e.isDir {
			line = s.renderDir(e.dir)
		} else {
			line = s.renderFile(e)
		}
		sb.WriteString(line)
		sb.WriteByte('\n')
		lineCount++
	}

	s.contentHeight = lineCount

	content := strings.TrimRight(sb.String(), "\n")
	for i := lineCount; i < s.maxLines; i++ {
		content += "\n"
	}

	style := WithBorder(
		BgBox(s.theme.SurfaceBg, s.width),
		s.theme.SidebarDir, s.theme.SurfaceBg,
		false, true, false, false,
	)
	return style.Render(content)
}

func (s *Sidebar) ContentHeight() int { return s.contentHeight }

func (s *Sidebar) renderDir(dir string) string {
	avail := s.width - 4
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(s.theme.SidebarDir)).
		Background(lipgloss.Color(s.theme.SurfaceBg)).
		Render(" " + truncate(dir, avail-1))
}

func statusFg(status string, theme models.Theme) string {
	switch status {
	case "A":
		return theme.SidebarAdded
	case "D":
		return theme.SidebarDeleted
	default:
		return theme.SidebarModified
	}
}

func (s *Sidebar) renderFile(e visualEntry) string {
	avail := s.width - 4

	_, filename := filepath.Split(e.file.NewPath)
	st := s.stats[e.fileIdx]

	bg := lipgloss.NewStyle().Background(lipgloss.Color(s.theme.SurfaceBg))

	statusStyle := bg.Foreground(lipgloss.Color(statusFg(e.file.Status, s.theme)))
	addStyle := bg.Foreground(lipgloss.Color(s.theme.SidebarAdded))
	delStyle := bg.Foreground(lipgloss.Color(s.theme.SidebarDeleted))

	statsStr := RenderStats(st.Added, st.Deleted, addStyle, delStyle, bg)
	statsW := lipgloss.Width(statsStr)

	if e.fileIdx == s.fileIdx {
		nameW := max(avail - 4 - statsW, 0)
		name := truncate(filename, nameW)

		lineStyle := bg.Foreground(lipgloss.Color(s.theme.SidebarSelected)).Bold(true)
		return lineStyle.Render("▸ ") + statusStyle.Render(e.file.Status) + lineStyle.Render(" " + name) + statsStr
	}

	nameW := max(avail - 4 - statsW, 0)
	name := truncate(filename, nameW)

	lineStyle := bg.Foreground(lipgloss.Color(s.theme.ContextFg))
	return lineStyle.Render("  ") + statusStyle.Render(e.file.Status) + lineStyle.Render(" " + name) + statsStr
}

func CalculateSideWidth(totalWidth int) int {
	if totalWidth <= 0 {
		return sidebarDefaultWidth
	}
	return min(max(totalWidth/sidebarDenominator, sidebarMinWidth), sidebarMaxWidth)
}

func buildVisualEntries(files []git.SideBySideDiff) []visualEntry {
	var entries []visualEntry
	var lastDir string
	for i, f := range files {
		d := dirOf(f.NewPath)
		if d != lastDir {
			entries = append(entries, visualEntry{isDir: true, dir: d})
			lastDir = d
		}
		entries = append(entries, visualEntry{fileIdx: i, file: &files[i]})
	}
	return entries
}

func dirOf(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx < 0 {
		return "./"
	}
	return path[:idx+1]
}

func truncate(s string, maxLen int) string {
	if maxLen < 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return string([]rune(s)[:maxLen])
	}
	return string([]rune(s)[:maxLen-3]) + "."
}

// visibleWindow computes the scroll-window of entries that should be shown
// on screen: it locates the entry for fileIdx, then clamps the window so
// that entry is visible within maxLines rows. It returns the visible slice
// and the index into entries where that slice starts.
func visibleWindow(entries []visualEntry, fileIdx, maxLines int) (visible []visualEntry, start int) {
	selPos := 0
	for i, e := range entries {
		if !e.isDir && e.fileIdx == fileIdx {
			selPos = i
			break
		}
	}
	start = 0
	if selPos >= maxLines {
		start = selPos - maxLines + 1
	}
	visible = entries[start:]
	if len(visible) > maxLines {
		visible = visible[:maxLines]
	}
	return visible, start
}

func LookupSidebarEntry(files []git.SideBySideDiff, fileIdx int, height int, y int) (int, bool) {
	entries := buildVisualEntries(files)
	maxLines := max(height-statusBarHeight, 1)
	visible, _ := visibleWindow(entries, fileIdx, maxLines)
	if y < 0 || y >= len(visible) {
		return 0, false
	}
	e := visible[y]
	if e.isDir {
		return 0, false
	}
	return e.fileIdx, true
}
