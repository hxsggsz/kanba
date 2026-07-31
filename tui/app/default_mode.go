package app

import (
	"kanba/tui/widget"
)

type DefaultMode struct{}

func (m *DefaultMode) Render(model *Model) string {
	if len(model.flatLines) == 0 {
		return ""
	}

	theme := model.CurrentTheme()
	sideWidth := widget.CalculateSideWidth(model.width)
	scroll := model.scroller.Scroll()
	if scroll >= len(model.flatLines) {
		scroll = max(0, len(model.flatLines)-1)
	}
	cursorFileIdx := model.flatLines[scroll].FileIdx

	sidebar := widget.NewSidebar(model.diffs, cursorFileIdx, sideWidth, model.height, theme, model.fileStats)
	sidebarStr := sidebar.Render()

	contentVis := model.VisibleLines()
	panelWidth := max(model.width-sideWidth-panelBorderWidth, panelMinWidth)
	content := model.renderContinuous(panelWidth, contentVis)

	return renderFrame(model, theme, sidebarStr, content)
}
