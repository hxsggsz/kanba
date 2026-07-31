package selection

import (
	"testing"
	"time"
)

func TestNewCoordinator(t *testing.T) {
	c := NewCoordinator(nil)
	if c == nil {
		t.Fatal("NewCoordinator returned nil")
	}
	if _, ok := c.state.(IdleState); !ok {
		t.Errorf("initial state = %T, want IdleState", c.state)
	}
}

func TestHandleClick_IdleToSelecting(t *testing.T) {
	c := NewCoordinator(nil)
	cmd := c.HandleClick(PanelLeft, 3, 5)

	if cmd != nil {
		t.Errorf("expected nil cmd, got %v", cmd)
	}
	if _, ok := c.state.(SelectingState); !ok {
		t.Errorf("state = %T, want SelectingState", c.state)
	}
	sel := c.CurrentSelection()
	if sel == nil {
		t.Fatal("CurrentSelection returned nil")
	}
	if sel.Panel != PanelLeft {
		t.Errorf("Panel = %v, want PanelLeft", sel.Panel)
	}
	if sel.Range != (Range{StartLine: 3, StartCol: 5, EndLine: 3, EndCol: 5}) {
		t.Errorf("Range = %+v, want {StartLine:3 StartCol:5 EndLine:3 EndCol:5}", sel.Range)
	}
}

func TestHandleDrag_ExtendsSelection(t *testing.T) {
	c := NewCoordinator(nil)
	c.HandleClick(PanelLeft, 0, 0)
	c.HandleDrag(PanelLeft, 5, 10)

	sel := c.CurrentSelection()
	if sel == nil {
		t.Fatal("CurrentSelection returned nil")
	}
	if sel.Range.EndLine != 5 || sel.Range.EndCol != 10 {
		t.Errorf("Range.End = (%d,%d), want (5,10)", sel.Range.EndLine, sel.Range.EndCol)
	}
}

func TestHandleDrag_MultipleDrags(t *testing.T) {
	c := NewCoordinator(nil)
	c.HandleClick(PanelLeft, 1, 2)
	c.HandleDrag(PanelLeft, 3, 4)
	c.HandleDrag(PanelLeft, 7, 8)

	sel := c.CurrentSelection()
	if sel == nil {
		t.Fatal("CurrentSelection returned nil")
	}
	if sel.Range != (Range{StartLine: 1, StartCol: 2, EndLine: 7, EndCol: 8}) {
		t.Errorf("Range = %+v, want {StartLine:1 StartCol:2 EndLine:7 EndCol:8}", sel.Range)
	}
}

func TestHandleRelease_SelectingToSelected(t *testing.T) {
	c := NewCoordinator(nil)
	c.HandleClick(PanelLeft, 0, 0)
	c.HandleDrag(PanelLeft, 3, 5)
	cmd := c.HandleRelease()

	if cmd == nil {
		t.Error("expected non-nil cmd for valid selection")
	}
	if _, ok := c.state.(SelectedState); !ok {
		t.Errorf("state = %T, want SelectedState", c.state)
	}
	sel := c.CurrentSelection()
	if sel == nil {
		t.Fatal("CurrentSelection returned nil")
	}
	if sel.Range != (Range{StartLine: 0, StartCol: 0, EndLine: 3, EndCol: 5}) {
		t.Errorf("Range = %+v, want {StartLine:0 StartCol:0 EndLine:3 EndCol:5}", sel.Range)
	}
}

func TestHandleClick_SelectedStartsNewSelection(t *testing.T) {
	c := NewCoordinator(nil)
	c.HandleClick(PanelLeft, 0, 0)
	c.HandleDrag(PanelLeft, 3, 5)
	c.HandleRelease()

	// Now in SelectedState; clicking should start new selection
	c.HandleClick(PanelRight, 10, 20)
	if _, ok := c.state.(SelectingState); !ok {
		t.Errorf("state = %T, want SelectingState", c.state)
	}
	sel := c.CurrentSelection()
	if sel == nil {
		t.Fatal("CurrentSelection returned nil")
	}
	if sel.Panel != PanelRight {
		t.Errorf("Panel = %v, want PanelRight", sel.Panel)
	}
	if sel.Range != (Range{StartLine: 10, StartCol: 20, EndLine: 10, EndCol: 20}) {
		t.Errorf("Range = %+v, want {StartLine:10 StartCol:20 EndLine:10 EndCol:20}", sel.Range)
	}
}

func TestDoubleClick_Detection(t *testing.T) {
	c := NewCoordinator(nil)
	c.getLineContent = func(line int, panel PanelSide) string {
		return "hello world"
	}

	// First click
	c.HandleClick(PanelLeft, 0, 0)

	// Second click at same position within 300ms
	c.HandleClick(PanelLeft, 0, 0)

	if _, ok := c.state.(SelectedState); !ok {
		t.Errorf("state after double-click = %T, want SelectedState", c.state)
	}
}

func TestDoubleClick_UsesWordBoundaries(t *testing.T) {
	c := NewCoordinator(nil)
	c.getLineContent = func(line int, panel PanelSide) string {
		return "hello world"
	}

	// First click
	c.HandleClick(PanelLeft, 0, 6)

	// Double-click at position 6 (the 'w' in "world")
	c.HandleClick(PanelLeft, 0, 6)

	sel := c.CurrentSelection()
	if sel == nil {
		t.Fatal("CurrentSelection returned nil after double-click")
	}
	// "world" starts at col 6, ends at col 11 (exclusive)
	if sel.Range.StartCol != 6 || sel.Range.EndCol != 11 {
		t.Errorf("Range = %+v, want StartCol=6 EndCol=11 (word 'world')", sel.Range)
	}
}

func TestDoubleClick_DifferentPositionResetsCount(t *testing.T) {
	c := NewCoordinator(nil)

	// First click at position A
	c.HandleClick(PanelLeft, 0, 0)

	// Click at a distant position — resets click count
	c.HandleClick(PanelLeft, 10, 10)

	if _, ok := c.state.(SelectingState); !ok {
		t.Errorf("state after distant click = %T, want SelectingState", c.state)
	}
}

func TestDoubleClick_TooFarApartResetsCount(t *testing.T) {
	c := NewCoordinator(nil)

	// First click
	c.HandleClick(PanelLeft, 0, 0)

	// Second click more than 2 cells away
	c.HandleClick(PanelLeft, 0, 5)

	// Should not be a double-click (too far apart)
	if _, ok := c.state.(SelectedState); ok {
		t.Error("should not be SelectedState (clicks too far apart)")
	}
}

func TestClear_ResetsToIdle(t *testing.T) {
	c := NewCoordinator(nil)
	c.HandleClick(PanelLeft, 0, 0)
	c.HandleDrag(PanelLeft, 3, 5)

	c.Clear()

	if _, ok := c.state.(IdleState); !ok {
		t.Errorf("state after Clear = %T, want IdleState", c.state)
	}
}

func TestCurrentSelection_Idle(t *testing.T) {
	c := NewCoordinator(nil)
	if sel := c.CurrentSelection(); sel != nil {
		t.Errorf("CurrentSelection in Idle = %+v, want nil", sel)
	}
}

func TestCurrentSelection_Selecting(t *testing.T) {
	c := NewCoordinator(nil)
	c.HandleClick(PanelLeft, 2, 3)

	sel := c.CurrentSelection()
	if sel == nil {
		t.Fatal("CurrentSelection returned nil")
	}
	if sel.Range != (Range{StartLine: 2, StartCol: 3, EndLine: 2, EndCol: 3}) {
		t.Errorf("Range = %+v, want {StartLine:2 StartCol:3 EndLine:2 EndCol:3}", sel.Range)
	}
}

func TestCurrentSelection_Selected(t *testing.T) {
	c := NewCoordinator(nil)
	c.HandleClick(PanelLeft, 0, 0)
	c.HandleDrag(PanelLeft, 2, 4)
	c.HandleRelease()

	sel := c.CurrentSelection()
	if sel == nil {
		t.Fatal("CurrentSelection returned nil")
	}
	if sel.Range != (Range{StartLine: 0, StartCol: 0, EndLine: 2, EndCol: 4}) {
		t.Errorf("Range = %+v, want {StartLine:0 StartCol:0 EndLine:2 EndCol:4}", sel.Range)
	}
}

func TestHasSelection_Idle(t *testing.T) {
	c := NewCoordinator(nil)
	if c.HasSelection() {
		t.Error("HasSelection = true in Idle, want false")
	}
}

func TestHasSelection_ZeroLength(t *testing.T) {
	c := NewCoordinator(nil)
	c.HandleClick(PanelLeft, 3, 5)

	// Selection is a single point (start == end), which IsEmpty() == true
	if c.HasSelection() {
		t.Error("HasSelection = true for zero-length selection, want false")
	}
}

func TestHasSelection_NonEmpty(t *testing.T) {
	c := NewCoordinator(nil)
	c.HandleClick(PanelLeft, 0, 0)
	c.HandleDrag(PanelLeft, 2, 5)

	if !c.HasSelection() {
		t.Error("HasSelection = false for non-empty selection, want true")
	}
}

func TestHasSelection_Selected(t *testing.T) {
	c := NewCoordinator(nil)
	c.HandleClick(PanelLeft, 0, 0)
	c.HandleDrag(PanelLeft, 2, 5)
	c.HandleRelease()

	if !c.HasSelection() {
		t.Error("HasSelection = false in SelectedState, want true")
	}
}

func TestDoubleClick_WithoutLineContentProvider(t *testing.T) {
	c := NewCoordinator(nil)

	// First click
	c.HandleClick(PanelLeft, 0, 5)

	// Double-click without content provider
	c.HandleClick(PanelLeft, 0, 5)

	if _, ok := c.state.(SelectedState); !ok {
		t.Errorf("state = %T, want SelectedState", c.state)
	}
	// Without content provider, boundaries default to (col, col)
	sel := c.CurrentSelection()
	if sel == nil {
		t.Fatal("CurrentSelection returned nil")
	}
	if sel.Range.StartCol != 5 || sel.Range.EndCol != 5 {
		t.Errorf("Range = %+v, want StartCol=5 EndCol=5 (no content provider)", sel.Range)
	}
}

func TestCoordinatorFullLifecycle(t *testing.T) {
	c := NewCoordinator(nil)

	// Click to start
	c.HandleClick(PanelLeft, 1, 2)
	if c.HasSelection() {
		t.Error("HasSelection should be false for zero-length selection")
	}

	// Drag to extend
	c.HandleDrag(PanelLeft, 3, 4)
	if !c.HasSelection() {
		t.Error("HasSelection should be true after drag")
	}

	// Release
	c.HandleRelease()
	if _, ok := c.state.(SelectedState); !ok {
		t.Errorf("after release: state = %T, want SelectedState", c.state)
	}

	// Clear
	c.Clear()
	if _, ok := c.state.(IdleState); !ok {
		t.Errorf("after clear: state = %T, want IdleState", c.state)
	}
}

func TestDoubleClick_WindowBoundaryFarApart(t *testing.T) {
	c := NewCoordinator(nil)
	c.getLineContent = func(line int, panel PanelSide) string { return "test" }

	c.HandleClick(PanelLeft, 0, 0)

	// Simulate time passing beyond 300ms by manipulating lastClick
	c.lastClick = time.Now().Add(-400 * time.Millisecond)

	c.HandleClick(PanelLeft, 0, 0)

	// Should be a new single click, not double-click
	if _, ok := c.state.(SelectedState); ok {
		t.Error("clicks 400ms apart should not be detected as double-click")
	}
	if _, ok := c.state.(SelectingState); !ok {
		t.Errorf("state = %T, want SelectingState", c.state)
	}
}
