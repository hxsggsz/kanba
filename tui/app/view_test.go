package app

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"kanba/git"
	"kanba/tui/diff"
	"kanba/tui/selection"
)

func setupRenderModel() *Model {
	diffs := []git.SideBySideDiff{
		{NewPath: "main.go", Status: "M",
			Hunks: []git.AlignedHunk{{OldStart: 1, NewStart: 1,
				Header: "@@ -1,20 +1,20 @@ func main() {", Context: "func main() {",
				Lines: hunkLines(20, "stay")}}},
		{NewPath: "util.go", Status: "M",
			Hunks: []git.AlignedHunk{{OldStart: 1, NewStart: 1,
				Header: "@@ -1,20 +1,20 @@ func helper() {", Context: "func helper() {",
				Lines: hunkLines(20, "keep")}}},
	}
	return &Model{
		diffs:        diffs,
		flatLines:    diff.BuildFlatLines(diffs),
		fileStats:    diff.ComputeFileStats(diffs),
		scroller:     diff.NewScroller(),
		highlighter:  diff.NewSyntaxHighlighter(),
		selection:    selection.NewCoordinator(nil),
		width:        80,
		height:       24,
		visibleLines: 20,
	}
}

func hunkLines(n int, text string) []git.AlignedLine {
	lines := make([]git.AlignedLine, n)
	for i := range lines {
		lines[i] = git.AlignedLine{Kind: git.KindContext, OldLineNum: i + 1, NewLineNum: i + 1, OldContent: text, NewContent: text}
	}
	return lines
}

func TestRenderContinuousShowsHunkHeader(t *testing.T) {
	m := setupRenderModel()
	out := m.renderContinuous(m.width, m.visibleLines)

	plain := stripANSI(out)
	if !strings.Contains(plain, "···") {
		t.Errorf("expected hunk header marker ··· in output, got:\n%s", plain)
	}
	if !strings.Contains(plain, "func main() {") {
		t.Errorf("expected hunk context in output, got:\n%s", plain)
	}
	if !strings.Contains(plain, "stay") {
		t.Errorf("expected content line in output, got:\n%s", plain)
	}
}

func TestRenderSinglePanelShowsHunkHeader(t *testing.T) {
	m := setupRenderModel()
	rm := RightPanelMode{}
	out := rm.renderSinglePanel(m, 80, 20)

	plain := stripANSI(out)
	if !strings.Contains(plain, "func main() {") {
		t.Errorf("expected hunk context in single-panel output, got:\n%s", plain)
	}
}

func stripANSI(s string) string {
	var b strings.Builder
	in := false
	for _, r := range s {
		if r == '\x1b' {
			in = true
			continue
		}
		if in {
			if r == 'm' {
				in = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func TestSetupSelectionProviderSkipsHunkHeaders(t *testing.T) {
	m := setupRenderModel()
	m.setupSelectionProvider()

	provider := m.selection.LineContentProvider()
	if provider == nil {
		t.Fatal("expected line content provider to be set")
	}

	if got := provider(0, selection.PanelRight); got != "" {
		t.Errorf("expected empty content for file header, got %q", got)
	}
	if got := provider(1, selection.PanelRight); got != "" {
		t.Errorf("expected empty content for hunk header, got %q", got)
	}
	if got := provider(2, selection.PanelRight); got != "stay" {
		t.Errorf("expected content for content line, got %q", got)
	}
}

func TestHandleSidebarClickSkipsHunkHeaders(t *testing.T) {
	m := setupRenderModel()

	// Sidebar entries for two root files: ["./", main.go, util.go].
	// Row 2 is util.go (file index 1); flat line 24 is its first content line.
	m.handleSidebarClick(tea.MouseClickMsg{X: 5, Y: statusBarHeight + 2})

	if got := m.scroller.Scroll(); got != 24 {
		t.Errorf("expected scroll to first content line of file 1 (index 24), got %d", got)
	}
}
