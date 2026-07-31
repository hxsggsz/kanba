package widget

import (
	"fmt"
	"strings"

	models "kanba/tui/models"

	"charm.land/lipgloss/v2"
)

type StatusBar struct {
	fileName   string
	fileIdx    int
	totalFiles int
	width      int
	theme      models.Theme
	copyMsg    string
}

func NewStatusBar(fileName string, fileIdx, totalFiles, width int, theme models.Theme, copyMsg string) *StatusBar {
	return &StatusBar{
		fileName:   fileName,
		fileIdx:    fileIdx,
		totalFiles: totalFiles,
		width:      width,
		theme:      theme,
		copyMsg:    copyMsg,
	}
}

func (s *StatusBar) Render() string {
	left := fmt.Sprintf(" ▸ %s  •  ↑/k ↓/j scroll  •  g/G top/bottom  •  ? help  •  q quit", s.fileName)
	right := ""
	if s.copyMsg != "" {
		right = s.copyMsg
	}

	text := left
	if right != "" {
		avail := s.width - lipgloss.Width(left) - 4
		if avail > len(right) {
			text = left + strings.Repeat(" ", avail-len(right)) + right
		}
	}

	style := WithBorder(
		BgBox(s.theme.SurfaceBg, s.width),
		s.theme.SidebarDir, s.theme.SurfaceBg,
		false, false, true, false,
	).
		Foreground(lipgloss.Color(s.theme.StatusBarFg)).
		Padding(1, 1)
	return style.Render(text)
}
