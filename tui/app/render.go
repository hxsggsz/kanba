package app

import (
	"fmt"

	models "kanba/tui/models"
	"kanba/tui/widget"

	"charm.land/lipgloss/v2"
)

// renderFrame assembles the common frame shared by every ViewMode: it joins
// an optional sidebar with the mode-specific content, builds the status bar,
// wraps the result to the terminal size, and applies the theme-modal and
// help overlays. sidebar may be "" for modes without one.
func renderFrame(model *Model, theme models.Theme, sidebar string, content string) string {
	scroll := model.scroller.Scroll()
	if scroll >= len(model.flatLines) {
		scroll = max(0, len(model.flatLines)-1)
	}
	cursorFileIdx := model.flatLines[scroll].FileIdx
	f := model.diffs[cursorFileIdx]
	statusBar := widget.NewStatusBar(f.NewPath, cursorFileIdx, len(model.diffs), model.width, theme, model.statusRightMsg())

	body := content
	if sidebar != "" {
		body = lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content)
	}

	result := fmt.Sprintf("%s\n%s", statusBar.Render(), body)
	result = lipgloss.
		NewStyle().
		Width(model.width).
		Height(model.height).
		Background(lipgloss.Color(theme.PanelBg)).
		Render(result)
	result = model.themeModal.Overlay(result, theme.SurfaceBg, theme.SidebarSelected, theme.ContextFg)

	if model.helpActive {
		result = model.helpOverlay(result, theme)
	}

	return result
}
