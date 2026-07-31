package selection

type State interface {
	HandleClick(panel PanelSide, line, col int) State
	HandleDrag(panel PanelSide, line, col int) State
	HandleRelease() State
	HandleDoubleClick(panel PanelSide, line, col int, boundaries WordBoundary) State
	Clear() State
}

type WordBoundary struct {
	Start int
	End   int
}

// newSelecting builds a SelectingState anchored at the given position.
func newSelecting(panel PanelSide, line, col int) SelectingState {
	return SelectingState{
		Selection: Selection{
			Panel: panel,
			Range: Range{
				StartLine: line,
				StartCol:  col,
				EndLine:   line,
				EndCol:    col,
			},
		},
	}
}

// newSelected builds a SelectedState covering the given word boundary.
func newSelected(panel PanelSide, line int, b WordBoundary) SelectedState {
	return SelectedState{
		Selection: Selection{
			Panel: panel,
			Range: Range{
				StartLine: line,
				StartCol:  b.Start,
				EndLine:   line,
				EndCol:    b.End,
			},
		},
	}
}

// IdleState - no selection active
type IdleState struct{}

func (IdleState) HandleClick(panel PanelSide, line, col int) State {
	return newSelecting(panel, line, col)
}

func (IdleState) HandleDrag(panel PanelSide, line, col int) State {
	return IdleState{}
}

func (IdleState) HandleRelease() State {
	return IdleState{}
}

func (IdleState) HandleDoubleClick(panel PanelSide, line, col int, boundaries WordBoundary) State {
	return newSelected(panel, line, boundaries)
}

func (IdleState) Clear() State {
	return IdleState{}
}

// SelectingState - mouse down, dragging
type SelectingState struct {
	Selection Selection
}

func (SelectingState) HandleClick(panel PanelSide, line, col int) State {
	return newSelecting(panel, line, col)
}

func (st SelectingState) HandleDrag(panel PanelSide, line, col int) State {
	st.Selection.Range.EndLine = line
	st.Selection.Range.EndCol = col
	return st
}

func (st SelectingState) HandleRelease() State {
	return SelectedState{Selection: st.Selection}
}

func (st SelectingState) HandleDoubleClick(panel PanelSide, line, col int, boundaries WordBoundary) State {
	return newSelected(panel, line, boundaries)
}

func (st SelectingState) Clear() State {
	return IdleState{}
}

// SelectedState - selection complete, awaiting copy or new click
type SelectedState struct {
	Selection Selection
}

func (SelectedState) HandleClick(panel PanelSide, line, col int) State {
	return newSelecting(panel, line, col)
}

func (st SelectedState) HandleDrag(panel PanelSide, line, col int) State {
	st.Selection.Range.EndLine = line
	st.Selection.Range.EndCol = col
	return st
}

func (st SelectedState) HandleRelease() State {
	return st
}

func (st SelectedState) HandleDoubleClick(panel PanelSide, line, col int, boundaries WordBoundary) State {
	return newSelected(panel, line, boundaries)
}

func (st SelectedState) Clear() State {
	return IdleState{}
}
