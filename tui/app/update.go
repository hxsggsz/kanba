package app

import (
	"log/slog"
	"time"

	"kanba/tui/diff"
	models "kanba/tui/models"
	"kanba/tui/selection"
	"kanba/tui/setting"
	"kanba/tui/widget"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSizeMsg(msg)

	case setting.DiffMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.diffs = msg.Diffs
			m.flatLines = diff.BuildFlatLines(m.diffs)
			m.fileStats = diff.ComputeFileStats(m.diffs)
			m.setupSelectionProvider()
		}
		m.visibleLines = m.height - 4
		return m, nil

	case setting.UpdateCheckMsg:
		if msg.Available {
			m.selfUpdate.state = updateStateAvailable
			m.selfUpdate.version = msg.Version
		}
		return m, nil

	case setting.UpdateInstallMsg:
		if msg.Err != nil {
			m.selfUpdate.state = updateStateFailed
			m.selfUpdate.err = msg.Err.Error()
		} else {
			m.selfUpdate.state = updateStateSucceeded
		}
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)

	case tea.MouseClickMsg:
		return m.handleMouseClick(msg), nil
	case tea.MouseWheelMsg:
		return m.handleMouseWheel(msg)

	case tea.MouseMotionMsg:
		if m.selection != nil {
			panel, line, col := m.mapMouseToContent(msg.X, msg.Y)
			if panel >= 0 {
				m.selection.HandleDrag(panel, line, col)
			}
		}

	case tea.MouseReleaseMsg:
		if m.selection != nil {
			return m, m.selection.HandleRelease()
		}

	case selection.CopyMsg:
		if m.selectedText == "" {
			return m, nil
		}
		if err := selection.CopyToClipboard(m.selectedText); err != nil {
			slog.Warn("failed to copy to clipboard", "error", err)
		}
		m.selectedText = ""
		return m, nil
	}
	return m, nil
}

func (m *Model) handleWindowSizeMsg(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	m.visibleLines = m.height - 4
	m.activeMode = m.modeFactory.FromWidth(msg.Width)
	m.scroller.UpdateScroll(m.TotalLines(), m.visibleLines)
	return m, nil
}

func (m *Model) handleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.err != nil {
		return m, tea.Quit
	}

	if m.themeModal.Active {
		return m.handleThemeModalKeys(msg)
	}

	if m.helpActive {
		switch msg.String() {
		case setting.KeyHelp, setting.KeyBack:
			m.helpActive = false
		}
		return m, nil
	}

	switch msg.String() {
	case setting.KeyQuit, setting.KeyQuitAlt:
		return m, tea.Quit
	case setting.KeyUpdate:
		return m.handleUpdateKey()
	}

	if m.activeMode != nil {
		return m.handleDiffKeys(msg)
	}

	return m, nil
}

func (m *Model) handleUpdateKey() (tea.Model, tea.Cmd) {
	switch m.selfUpdate.state {
	case updateStateAvailable, updateStateFailed:
		m.selfUpdate.state = updateStateUpdating
		m.selfUpdate.err = ""
		return m, setting.UpdateInstallCmd()
	default:
		return m, nil
	}
}

// overlayActive reports whether a modal overlay (theme selector or help) is
// currently capturing input, in which case normal key/mouse handling for the
// diff view should be suppressed.
func (m *Model) overlayActive() bool {
	return m.themeModal.Active || m.helpActive
}

func (m *Model) handleDiffKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case setting.KeyUp, setting.KeyUpAlt:
		m.scroller.MoveUp()

	case setting.KeyDown, setting.KeyDownAlt:
		m.scroller.MoveDown(m.TotalLines(), m.visibleLines)

	case setting.KeyTop:
		m.scroller.GoToTop()

	case setting.KeyBottom:
		m.scroller.GoToBottom(m.TotalLines(), m.visibleLines)

	case setting.KeyLeft, setting.KeyLeftAlt:
		m.scroller.ScrollLeft()

	case setting.KeyRight, setting.KeyRightAlt:
		m.scroller.ScrollRight()

	case setting.KeyLeftWord:
		m.scroller.ScrollLeftFast()

	case setting.KeyRightWord:
		m.scroller.ScrollRightFast()

	case setting.KeyHome:
		m.scroller.ScrollHome()

	case setting.KeyEnd:
		sideWidth := widget.CalculateSideWidth(m.width)
		panelWidth := max(m.width-sideWidth-panelBorderWidth, panelMinWidth)
		colWidth := (panelWidth - 3) / 2
		prefixWidth := diff.LineNumColWidth + 3
		maxContent := 0
		for _, f := range m.diffs {
			w := maxFileContentWidth(f)
			if w > maxContent {
				maxContent = w
			}
		}
		m.scroller.ScrollEnd(max(0, maxContent-(colWidth-prefixWidth)))

	case setting.KeyHelp:
		m.helpActive = true

	case setting.KeyTheme:
		m.themeModal.Active = true
		m.themeModal.SyncCursor(m.themeModal.Selected)

	case setting.KeyCopyPath:
		if len(m.flatLines) == 0 {
			break
		}
		scroll := m.scroller.Scroll()
		if scroll >= len(m.flatLines) {
			scroll = max(0, len(m.flatLines)-1)
		}
		cursorFileIdx := m.flatLines[scroll].FileIdx
		path := m.diffs[cursorFileIdx].NewPath
		if err := selection.CopyToClipboard(path); err != nil {
			slog.Warn("failed to copy path to clipboard", "error", err)
		}
		m.copyNotice.msg = " ✓ " + path
		m.copyNotice.till = time.Now().Add(3 * time.Second)
		return m, nil
	}

	return m, nil
}

type scrollUpMsg struct{}
type scrollDownMsg struct{}

func (m *Model) scrollUp(lines int) tea.Cmd {
	return func() tea.Msg {
		for i := 0; i < lines; i++ {
			m.scroller.MoveUp()
		}
		return scrollUpMsg{}
	}
}

func (m *Model) scrollDown(lines int) tea.Cmd {
	return func() tea.Msg {
		for i := 0; i < lines; i++ {
			m.scroller.MoveDown(m.TotalLines(), m.visibleLines)
		}
		return scrollDownMsg{}
	}
}

func (m *Model) handleMouseWheel(msg tea.MouseWheelMsg) (tea.Model, tea.Cmd) {
	if len(m.flatLines) == 0 {
		return m, nil
	}
	if m.themeModal.Active {
		switch msg.Button {
		case tea.MouseWheelUp:
			m.themeModal.MoveUp()
		case tea.MouseWheelDown:
			m.themeModal.MoveDown()
		}
		return m, nil
	}
	if m.helpActive {
		return m, nil
	}

	switch msg.Button {
	case tea.MouseWheelUp:
		if msg.Mod.Contains(tea.ModShift) {
			m.scroller.ScrollLeftFast()
			return m, nil
		}
		lines := 1
		if msg.Mod.Contains(tea.ModAlt) {
			lines = 5
		}
		return m, m.scrollUp(lines)
	case tea.MouseWheelDown:
		if msg.Mod.Contains(tea.ModShift) {
			m.scroller.ScrollRightFast()
			return m, nil
		}
		lines := 1
		if msg.Mod.Contains(tea.ModAlt) {
			lines = 5
		}
		return m, m.scrollDown(lines)
	case tea.MouseWheelLeft:
		m.scroller.ScrollLeftFast()
		return m, nil
	case tea.MouseWheelRight:
		m.scroller.ScrollRightFast()
		return m, nil
	}
	return m, nil
}

func (m *Model) handleThemeModalKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.themeModal.IsFocused {
		switch msg.String() {
		case setting.KeyBack:
			if m.themeModal.FilterQuery != "" {
				m.themeModal.Filter("")
				m.themeModal.FocusList()
			} else {
				m.themeModal.Close()
			}
		case "enter":
			m.themeModal.Select()
		case "tab", setting.KeyDown, setting.KeyUp:
			m.themeModal.MoveDown()
		case "backspace":
			m.themeModal.HandleBackspace()
		default:
			if text := msg.Key().Text; text != "" {
				for _, r := range text {
					m.themeModal.HandleRune(r)
				}
			}
		}
	} else {
		switch msg.String() {
		case "/":
			m.themeModal.FocusInput()
		case setting.KeyUp, setting.KeyUpAlt:
			m.themeModal.MoveUp()
		case setting.KeyDown, setting.KeyDownAlt:
			m.themeModal.MoveDown()
		case "enter":
			m.themeModal.Select()
		case setting.KeyTheme, setting.KeyBack, setting.KeyQuitAlt:
			m.themeModal.Close()
		}
	}
	return m, nil
}

// handleMouseClick dispatches a click to the modal overlay, sidebar, or
// diff content, depending on where it landed and what's currently active.
func (m *Model) handleMouseClick(msg tea.MouseClickMsg) *Model {
	if len(m.flatLines) == 0 {
		return m
	}

	if m.themeModal.Active {
		return m.handleThemeModalClick(msg)
	}

	if m.helpActive {
		return m.handleHelpOverlayClick(msg)
	}

	if msg.Y < statusBarHeight {
		return m
	}

	sideWidth := widget.CalculateSideWidth(m.width)
	if msg.X < sideWidth {
		return m.handleSidebarClick(msg)
	}

	return m.handleContentClick(msg)
}

// handleThemeModalClick routes a click while the theme modal is open: a
// click on a list row selects it, a click outside the modal closes it.
func (m *Model) handleThemeModalClick(msg tea.MouseClickMsg) *Model {
	x, y := msg.X, msg.Y
	theme := m.CurrentTheme()

	fg := m.themeModal.Render(theme.SurfaceBg, theme.SidebarSelected, theme.ContextFg)
	fgWidth, fgHeight := lipgloss.Size(fg)
	modalX := max(0, (m.width-fgWidth)/2)
	modalY := max(0, (m.height-fgHeight)/2)

	if x >= modalX && x < modalX+fgWidth && y >= modalY && y < modalY+fgHeight {
		relY := y - modalY
		n := m.themeModal.VisibleCount()
		start := models.ListStartLine
		if relY >= start && relY < start+n {
			idx := m.themeModal.ScrollOffset + relY - start
			m.themeModal.Cursor = idx
			m.themeModal.Select()
			m.themeModal.SyncCursor(m.themeModal.Selected)
		}
	} else {
		m.themeModal.Close()
	}
	return m
}

// handleHelpOverlayClick closes the help overlay when the click lands
// outside of it; clicks inside are ignored.
func (m *Model) handleHelpOverlayClick(msg tea.MouseClickMsg) *Model {
	x, y := msg.X, msg.Y
	theme := m.CurrentTheme()

	fg := m.helpContent(theme)
	fgWidth, fgHeight := lipgloss.Size(fg)
	modalX := max(0, (m.width-fgWidth)/2)
	modalY := max(0, m.height/2-fgHeight/2)

	if !(x >= modalX && x < modalX+fgWidth && y >= modalY && y < modalY+fgHeight) {
		m.helpActive = false
	}
	return m
}

// handleSidebarClick jumps the viewport to the file whose sidebar entry was
// clicked.
func (m *Model) handleSidebarClick(msg tea.MouseClickMsg) *Model {
	contentY := msg.Y - statusBarHeight

	fileIdx, ok := widget.LookupSidebarEntry(m.diffs, m.flatLines[m.scroller.Scroll()].FileIdx, m.height, contentY)
	if !ok {
		return m
	}
	for i, fl := range m.flatLines {
		if fl.FileIdx == fileIdx && !fl.IsHeader && !fl.IsHunkHeader {
			m.scroller.SetScroll(i, len(m.flatLines), m.visibleLines)
			return m
		}
	}
	return m
}

// handleContentClick handles a click inside the diff content area: on a
// file header it copies the file path, otherwise it either extends the
// active text selection or jumps the scroll position proportionally
// (scrollbar-style click).
func (m *Model) handleContentClick(msg tea.MouseClickMsg) *Model {
	contentY := msg.Y - statusBarHeight

	panel, line, col := m.mapMouseToContent(msg.X, msg.Y)
	isClickInContent := panel >= 0
	if isClickInContent && line < len(m.flatLines) && m.flatLines[line].IsHeader {
		return m.copyFileHeaderPath(line)
	}

	if isClickInContent && m.selection != nil {
		m.selection.HandleClick(panel, line, col)
		return m
	}

	return m.jumpScrollToClick(contentY)
}

// copyFileHeaderPath copies the file path for the header at flat-line index
// line to the clipboard and shows a transient status-bar confirmation.
func (m *Model) copyFileHeaderPath(line int) *Model {
	path := m.diffs[m.flatLines[line].FileIdx].NewPath
	if err := selection.CopyToClipboard(path); err != nil {
		slog.Warn("failed to copy path to clipboard", "error", err)
	}
	m.copyNotice.msg = " ✓ " + path
	m.copyNotice.till = time.Now().Add(3 * time.Second)
	return m
}

// jumpScrollToClick treats a click as a scrollbar-proportional jump: the
// relative vertical position of the click within the content area maps to
// the corresponding scroll offset.
func (m *Model) jumpScrollToClick(contentY int) *Model {
	if m.visibleLines <= 0 {
		return m
	}
	maxScroll := m.TotalLines() - m.visibleLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	targetScroll := int(float64(contentY) / float64(m.visibleLines) * float64(maxScroll))
	m.scroller.SetScroll(targetScroll, m.TotalLines(), m.visibleLines)
	return m
}

// mapMouseToContent maps mouse coordinates to panel, line, and column.
// Returns panel (-1 if outside content area), flat line index, and content column.
func (m *Model) mapMouseToContent(x, y int) (selection.PanelSide, int, int) {
	if y < statusBarHeight {
		return -1, 0, 0
	}

	contentY := y - statusBarHeight
	sideWidth := widget.CalculateSideWidth(m.width)

	isInsideSidebar := x < sideWidth
	if isInsideSidebar {
		return -1, 0, 0
	}

	// Determine panel
	panelWidth := max(m.width-sideWidth-panelBorderWidth, panelMinWidth)
	colWidth := panelWidth / 2
	contentX := x - sideWidth
	isLeftPanel := contentX < colWidth

	var panel selection.PanelSide
	var panelLeft int
	if isLeftPanel {
		panel = selection.PanelLeft
		panelLeft = 0
	} else {
		panel = selection.PanelRight
		panelLeft = colWidth
	}

	// Map y to flat line index
	start := m.scroller.Scroll()
	vis := m.VisibleLines()
	total := m.TotalLines()

	// Account for sticky header at the top
	if si := stickyHeaderIndex(m.flatLines, start); si >= 0 {
		sh := m.lineVisualHeight(m.flatLines[si])
		if contentY < sh {
			return selection.PanelLeft, si, 0
		}
		contentY -= sh
		vis-- // sticky header takes one flat line slot
	}

	end := min(start+vis, total)
	visualRow := 0
	targetLine := end - 1
	for gi := start; gi < end; gi++ {
		fl := m.flatLines[gi]
		h := m.lineVisualHeight(fl)
		clickFallsInLine := visualRow+h > contentY
		if clickFallsInLine {
			targetLine = gi
			break
		}
		visualRow += h
	}

	isOutOfBounds := targetLine >= len(m.flatLines)
	if isOutOfBounds {
		return -1, 0, 0
	}

	// Map x to content column
	prefixWidth := diff.LineNumColWidth + 3
	contentCol := contentX - panelLeft - prefixWidth + m.scroller.HScroll()
	if contentCol < 0 {
		contentCol = 0
	}

	return panel, targetLine, contentCol
}

// lineVisualHeight returns the visual height of a flat line in rows.
func (m *Model) lineVisualHeight(fl diff.FlatLine) int {
	if !fl.IsHeader {
		return 1
	}
	if fl.FileIdx > 0 {
		return 4
	}
	return 3
}
