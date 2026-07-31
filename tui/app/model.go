package app

import (
	"time"

	models "kanba/tui/models"
	"kanba/tui/diff"
	"kanba/tui/selection"
	"kanba/tui/setting"

	tea "charm.land/bubbletea/v2"
	"kanba/git"
)

const (
	statusBarHeight  = 4
	panelBorderWidth = 0
	panelMinWidth    = 10
)

type Model struct {
	diffs       []git.SideBySideDiff
	flatLines   []diff.FlatLine
	fileStats   []diff.FileStat
	scroller    *diff.Scroller
	loading     bool
	err         error
	width       int
	height      int

	repoPath     string
	gitArgs      []string
	highlighter  *diff.SyntaxHighlighter
	themeModal   *models.Modal
	helpActive   bool

	selection    *selection.Coordinator
	selectedText string

	activeMode   ViewMode
	modeFactory  *ModeFactory
	visibleLines int

	copyNotice copyNotice

	version    string
	selfUpdate selfUpdate
}

// copyNotice tracks the transient "copied to clipboard" status-bar message.
type copyNotice struct {
	msg  string
	till time.Time
}

// selfUpdate tracks the state of the in-place self-update checker/installer.
type selfUpdate struct {
	state   updateState
	version string
	err     string
}

type updateState int

const (
	updateStateNone updateState = iota
	updateStateAvailable
	updateStateUpdating
	updateStateSucceeded
	updateStateFailed
)

func (m *Model) TotalLines() int {
	return len(m.flatLines)
}

func (m *Model) CurrentTheme() models.Theme {
	if m.themeModal != nil {
		return models.GetTheme(m.themeModal.Selected)
	}
	return models.GetTheme("rose-pine")
}

func New(repoPath string, gitArgs []string, themeName string, version string) *Model {
	themeItems := make([]models.ModalItem, 0, len(models.Themes))
	for _, k := range models.SortedThemeKeys() {
		t := models.Themes[k]
		themeItems = append(themeItems, models.ModalItem{Key: k, Label: t.Name})
	}
	themeModal := models.NewModal("Theme", themeItems)

	if _, ok := models.Themes[themeName]; !ok {
		themeName = models.SortedThemeKeys()[0]
	}
	themeModal.Selected = themeName

	factory := &ModeFactory{}

	return &Model{
		repoPath:    repoPath,
		gitArgs:     gitArgs,
		loading:     true,
		scroller:    diff.NewScroller(),
		highlighter: diff.NewSyntaxHighlighter(),
		themeModal:  themeModal,
		selection:   selection.NewCoordinator(nil),
		modeFactory: factory,
		activeMode:  factory.FromWidth(80),
		version:     version,
	}
}

func (m *Model) Init() tea.Cmd {
	cmds := []tea.Cmd{setting.GitDiffCmd(m.repoPath, m.gitArgs)}
	if m.version != "dev" {
		cmds = append(cmds, setting.UpdateCheckCmd(m.version))
	}
	return tea.Batch(cmds...)
}

func (m *Model) VisibleLines() int {
	return m.height - statusBarHeight
}

// statusRightMsg returns the message shown right-aligned in the status bar.
// A transient copy confirmation takes priority over the persistent update
// notice; the update notice reappears once the copy message expires.
func (m *Model) statusRightMsg() string {
	if m.copyNotice.msg != "" {
		return m.copyNotice.msg
	}

	switch m.selfUpdate.state {
	case updateStateAvailable:
		return " " + m.selfUpdate.version + " available — press u to update"
	case updateStateUpdating:
		return " Updating..."
	case updateStateSucceeded:
		return " Updated to " + m.selfUpdate.version + " — restart kanba to apply"
	case updateStateFailed:
		return " Update failed: " + m.selfUpdate.err
	default:
		return ""
	}
}

func (m *Model) setupSelectionProvider() {
	if m.selection == nil {
		return
	}
	m.selection.SetLineContentProvider(func(flatLineIdx int, panel selection.PanelSide) string {
		if flatLineIdx < 0 || flatLineIdx >= len(m.flatLines) {
			return ""
		}
		fl := m.flatLines[flatLineIdx]
		if fl.IsHeader || fl.IsHunkHeader {
			return ""
		}
		f := m.diffs[fl.FileIdx]
		h := f.Hunks[fl.HunkIdx]
		ln := h.Lines[fl.LineIdx]
		if f.Status == "A" {
			return ln.NewContent
		}
		if panel == selection.PanelRight {
			return ln.NewContent
		}
		return ln.OldContent
	})
}
