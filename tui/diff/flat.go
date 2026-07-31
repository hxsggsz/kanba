package diff

import "kanba/git"

type FlatLine struct {
	IsHeader     bool
	IsHunkHeader bool
	FileIdx      int
	HunkIdx      int
	LineIdx      int
}

type FileStat struct {
	Added   int
	Deleted int
}

func BuildFlatLines(diffs []git.SideBySideDiff) []FlatLine {
	var lines []FlatLine
	for fi := range diffs {
		lines = append(lines, FlatLine{IsHeader: true, FileIdx: fi})
		for hi, h := range diffs[fi].Hunks {
			if h.OldStart != 0 || h.NewStart != 1 {
				lines = append(lines, FlatLine{IsHunkHeader: true, FileIdx: fi, HunkIdx: hi})
			}
			for li := range h.Lines {
				lines = append(lines, FlatLine{FileIdx: fi, HunkIdx: hi, LineIdx: li})
			}
		}
	}
	return lines
}

func ComputeFileStats(diffs []git.SideBySideDiff) []FileStat {
	stats := make([]FileStat, len(diffs))
	for fi, f := range diffs {
		for _, h := range f.Hunks {
			for _, ln := range h.Lines {
				switch ln.Kind {
				case git.KindAdded:
					stats[fi].Added++
				case git.KindDeleted:
					stats[fi].Deleted++
				case git.KindModified:
					stats[fi].Added++
					stats[fi].Deleted++
				}
			}
		}
	}
	return stats
}
