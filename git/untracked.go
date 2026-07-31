package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type LsFilesCommand struct{}

func (c *LsFilesCommand) Args() []string {
	return []string{"ls-files", "--others", "--exclude-standard"}
}

func (c *LsFilesCommand) Parse(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	return strings.Split(raw, "\n"), nil
}

func UntrackedToSideBySideDiff(repoPath, filePath string) (SideBySideDiff, error) {
	content, err := os.ReadFile(filepath.Join(repoPath, filePath))
	if err != nil {
		return SideBySideDiff{}, fmt.Errorf("read untracked file %q: %w", filePath, err)
	}

	text := strings.TrimRight(string(content), "\n")
	var lines []string
	if text == "" {
		lines = []string{""}
	} else {
		lines = strings.Split(text, "\n")
	}

	aligned := make([]AlignedLine, len(lines))
	for i, l := range lines {
		aligned[i] = AlignedLine{
			NewLineNum: i + 1,
			Kind:       KindAdded,
			NewContent: l,
		}
	}

	return SideBySideDiff{
		NewPath: filePath,
		Status:  "A",
		Hunks: []AlignedHunk{{
			NewStart: 1,
			Header:   fmt.Sprintf("@@ -0,0 +1,%d @@", len(lines)),
			Lines:    aligned,
		}},
	}, nil
}
