package app

type DiffOnlyMode struct{}

func (m *DiffOnlyMode) Render(model *Model) string {
	if len(model.flatLines) == 0 {
		return ""
	}

	theme := model.CurrentTheme()
	contentVis := model.VisibleLines()
	panelWidth := max(model.width-panelBorderWidth, panelMinWidth)
	content := model.renderContinuous(panelWidth, contentVis)

	return renderFrame(model, theme, "", content)
}
