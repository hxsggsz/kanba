package git

// LineType classifies a single line in a unified diff hunk.
type LineType int

const (
	LineContext LineType = iota // context line (present in both old and new)
	LineAdded                   // added line (only in new)
	LineDeleted                 // deleted line (only in old)
)

// Line represents a single line within a diff hunk.
type Line struct {
	Type       LineType
	OldLineNum int // line number in old file, 0 if added
	NewLineNum int // line number in new file, 0 if deleted
	Content    string
}

// Hunk represents a contiguous block of changes in a diff.
type Hunk struct {
	OldStart int    // starting line number in old file
	OldCount int    // number of lines in old file
	NewStart int    // starting line number in new file
	NewCount int    // number of lines in new file
	Header   string // raw @@ header line
	Context  string // trailing context (e.g. "func main() {"), or the range when absent
	Lines    []Line
}

// FileDiff represents the diff for a single file.
type FileDiff struct {
	OldPath  string
	NewPath  string
	Status   string // M (modified), A (added), D (deleted), R (renamed)
	Hunks    []Hunk
	IsBinary bool
	IsNew    bool
	IsDelete bool
	IsRename bool
}
