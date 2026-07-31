package widget

import (
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
)

// BgBox returns a style with the given background color and width set,
// matching the base "Background(...).Width(...)" chain that panel, sidebar
// and statusbar frames all start from.
func BgBox(bgColor string, width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color(bgColor)).
		Width(width)
}

// WithBorder layers a rounded border on top of an existing style, using the
// same border color/background pairing shared by the sidebar and status bar
// frames. top/right/bottom/left select which sides get a border, matching
// lipgloss's Border(...) side arguments.
func WithBorder(style lipgloss.Style, borderColor, borderBg string, top, right, bottom, left bool) lipgloss.Style {
	return style.
		Border(lipgloss.RoundedBorder(), top, right, bottom, left).
		BorderForeground(lipgloss.Color(borderColor)).
		BorderBackground(lipgloss.Color(borderBg))
}

// RenderStats renders the "(+N, -N)" stat segment shared by the panel file
// header and the sidebar file entries. It returns "" when there is nothing
// to show (no additions and no deletions).
func RenderStats(added, deleted int, addStyle, delStyle, normalStyle lipgloss.Style) string {
	if added <= 0 && deleted <= 0 {
		return ""
	}
	var segs []string
	if added > 0 {
		segs = append(segs, addStyle.Render("+"+strconv.Itoa(added)))
	}
	if deleted > 0 {
		segs = append(segs, delStyle.Render("-"+strconv.Itoa(deleted)))
	}
	return normalStyle.Render(" (") + strings.Join(segs, normalStyle.Render(", ")) + normalStyle.Render(")")
}
