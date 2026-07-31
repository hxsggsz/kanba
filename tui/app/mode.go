package app

const (
	modeBreakpointDefault  = 160
	modeBreakpointDiffOnly = 100
)

type ViewMode interface {
	Render(m *Model) string
}

type ModeFactory struct{}

func (f *ModeFactory) FromWidth(width int) ViewMode {
	switch {
	case width >= modeBreakpointDefault:
		return &DefaultMode{}
	case width >= modeBreakpointDiffOnly:
		return &DiffOnlyMode{}
	default:
		return &RightPanelMode{}
	}
}
