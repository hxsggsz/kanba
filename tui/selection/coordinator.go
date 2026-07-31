package selection

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// Coordinator manages selection state and mouse event routing.
type Coordinator struct {
	state      State
	clickCount int
	lastClick  time.Time
	lastX      int
	lastY      int

	onCopy         func(string) tea.Cmd
	getLineContent func(line int, panel PanelSide) string
}

// NewCoordinator creates a new selection coordinator.
func NewCoordinator(onCopy func(string) tea.Cmd) *Coordinator {
	return &Coordinator{
		state:  IdleState{},
		onCopy: onCopy,
	}
}

// SetLineContentProvider sets the callback used to retrieve line content
// for word boundary detection during double-click.
func (c *Coordinator) SetLineContentProvider(fn func(line int, panel PanelSide) string) {
	c.getLineContent = fn
}

// LineContentProvider returns the current line content provider (may be nil).
func (c *Coordinator) LineContentProvider() func(line int, panel PanelSide) string {
	return c.getLineContent
}

// HandleClick processes a mouse click.
func (c *Coordinator) HandleClick(panel PanelSide, line, col int) tea.Cmd {
	now := time.Now()
	samePosition := abs(col-c.lastX) <= 2 && abs(line-c.lastY) <= 2

	if samePosition && now.Sub(c.lastClick) < 300*time.Millisecond {
		c.clickCount++
	} else {
		c.clickCount = 1
	}

	c.lastClick = now
	c.lastX = col
	c.lastY = line

	if c.clickCount >= 2 {
		c.clickCount = 0

		var boundary WordBoundary
		if c.getLineContent != nil {
			content := c.getLineContent(line, panel)
			start, end := findWordBoundaries(content, col)
			boundary = WordBoundary{Start: start, End: end}
		} else {
			boundary = WordBoundary{Start: col, End: col}
		}

		c.state = c.state.HandleDoubleClick(panel, line, col, boundary)
		return c.copyIfSelected()
	}

	c.state = c.state.HandleClick(panel, line, col)
	return nil
}

// HandleDrag processes mouse drag.
func (c *Coordinator) HandleDrag(panel PanelSide, line, col int) {
	c.state = c.state.HandleDrag(panel, line, col)
}

// HandleRelease processes mouse release.
func (c *Coordinator) HandleRelease() tea.Cmd {
	c.state = c.state.HandleRelease()
	return c.copyIfSelected()
}

func (c *Coordinator) copyIfSelected() tea.Cmd {
	sel := c.CurrentSelection()
	if sel == nil || sel.Range.IsEmpty() {
		return nil
	}
	return DelayedCopyCmd()
}

// Clear resets the selection.
func (c *Coordinator) Clear() {
	c.state = c.state.Clear()
}

// CurrentSelection returns the active selection (if any).
func (c *Coordinator) CurrentSelection() *Selection {
	switch st := c.state.(type) {
	case SelectingState:
		sel := st.Selection
		return &sel
	case SelectedState:
		sel := st.Selection
		return &sel
	default:
		return nil
	}
}

// HasSelection returns true if there's an active non-empty selection.
func (c *Coordinator) HasSelection() bool {
	sel := c.CurrentSelection()
	if sel == nil {
		return false
	}
	return !sel.Range.IsEmpty()
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
