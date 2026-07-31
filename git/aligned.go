package git

type LineKind int

const (
	KindContext LineKind = iota
	KindAdded
	KindDeleted
	KindModified
)

type AlignedLine struct {
	OldLineNum  int
	NewLineNum  int
	OldContent  string
	NewContent  string
	Kind        LineKind
}

type AlignedHunk struct {
	OldStart int
	NewStart int
	Header   string
	Context  string
	Lines    []AlignedLine
}

type SideBySideDiff struct {
	OldPath string
	NewPath string
	Status  string
	Hunks   []AlignedHunk
}

type LineAligner interface {
	Align(hunks []Hunk) []AlignedHunk
}

type UnifiedAligner struct{}

func (a *UnifiedAligner) Align(hunks []Hunk) []AlignedHunk {
	result := make([]AlignedHunk, len(hunks))

	for i, h := range hunks {
		ah := AlignedHunk{OldStart: h.OldStart, NewStart: h.NewStart, Header: h.Header, Context: h.Context}
		var pending []Line

		for _, ln := range h.Lines {
			switch ln.Type {
			case LineContext:
				ah.Lines = append(ah.Lines, flushDeleted(pending)...)
				pending = nil
				ah.Lines = append(ah.Lines, AlignedLine{
					OldLineNum: ln.OldLineNum,
					NewLineNum: ln.NewLineNum,
					Kind:       KindContext,
					OldContent: ln.Content,
					NewContent: ln.Content,
				})

			case LineDeleted:
				pending = append(pending, ln)

			case LineAdded:
				if len(pending) > 0 {
					d := pending[0]
					pending = pending[1:]
					ah.Lines = append(ah.Lines, AlignedLine{
						OldLineNum: d.OldLineNum,
						NewLineNum: ln.NewLineNum,
						Kind:       KindModified,
						OldContent: d.Content,
						NewContent: ln.Content,
					})
				} else {
					ah.Lines = append(ah.Lines, AlignedLine{
						NewLineNum: ln.NewLineNum,
						Kind:       KindAdded,
						NewContent: ln.Content,
					})
				}
			}
		}
		ah.Lines = append(ah.Lines, flushDeleted(pending)...)
		result[i] = ah
	}
	return result
}

func flushDeleted(pending []Line) []AlignedLine {
	out := make([]AlignedLine, len(pending))
	for i, d := range pending {
		out[i] = AlignedLine{
			OldLineNum: d.OldLineNum,
			Kind:       KindDeleted,
			OldContent: d.Content,
		}
	}
	return out
}
