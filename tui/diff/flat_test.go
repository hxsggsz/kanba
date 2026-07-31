package diff

import (
	"testing"

	"kanba/git"
)

func TestBuildFlatLinesIncludesHunkHeaders(t *testing.T) {
	diffs := []git.SideBySideDiff{{
		NewPath: "main.go",
		Status:  "M",
		Hunks: []git.AlignedHunk{
			{OldStart: 1, NewStart: 1, Header: "@@ -1,3 +1,3 @@ func main() {", Context: "func main() {",
				Lines: []git.AlignedLine{{Kind: git.KindContext, OldContent: "a", NewContent: "a"}}},
			{OldStart: 4, NewStart: 4, Header: "@@ -4,2 +4,2 @@ func helper() {", Context: "func helper() {",
				Lines: []git.AlignedLine{{Kind: git.KindContext, OldContent: "b", NewContent: "b"}}},
		},
	}}

	lines := BuildFlatLines(diffs)

	if len(lines) != 5 {
		t.Fatalf("expected 5 flat lines, got %d: %+v", len(lines), lines)
	}
	if !lines[0].IsHeader {
		t.Error("expected line 0 to be the file header")
	}
	if !lines[1].IsHunkHeader || lines[1].HunkIdx != 0 {
		t.Errorf("expected line 1 to be hunk header 0, got %+v", lines[1])
	}
	if lines[2].IsHunkHeader || lines[2].LineIdx != 0 {
		t.Errorf("expected line 2 to be content line 0 of hunk 0, got %+v", lines[2])
	}
	if !lines[3].IsHunkHeader || lines[3].HunkIdx != 1 {
		t.Errorf("expected line 3 to be hunk header 1, got %+v", lines[3])
	}
	if lines[4].IsHunkHeader || lines[4].HunkIdx != 1 {
		t.Errorf("expected line 4 to be content of hunk 1, got %+v", lines[4])
	}
}

func TestBuildFlatLinesSkipsWholeFileAdditions(t *testing.T) {
	diffs := []git.SideBySideDiff{{
		NewPath: "newfile.go",
		Status:  "A",
		Hunks: []git.AlignedHunk{{
			OldStart: 0, NewStart: 1, Header: "@@ -0,0 +1,2 @@",
			Lines: []git.AlignedLine{
				{Kind: git.KindAdded, NewLineNum: 1, NewContent: "one"},
				{Kind: git.KindAdded, NewLineNum: 2, NewContent: "two"},
			},
		}},
	}}

	lines := BuildFlatLines(diffs)

	if len(lines) != 3 {
		t.Fatalf("expected 3 flat lines (header + 2 content), got %d: %+v", len(lines), lines)
	}
	for i, l := range lines {
		if l.IsHunkHeader {
			t.Errorf("expected no hunk header for whole-file addition, got %+v at index %d", l, i)
		}
	}
	if !lines[1].IsHunkHeader && lines[1].LineIdx != 0 {
		t.Errorf("expected line 1 to be content line 0, got %+v", lines[1])
	}
}
