package logs

type Entry struct {
	Raw       string
	Parsed    bool
	Level     Part
	Timestamp Part
	Message   Part
	Caller    Part
	Context   Part
	Search    string
	renderHeight int
}

func (e Entry) ContentHeight() int {
	return e.renderHeight
}

func (e *Entry) SetRenderHeight(height int) {
	e.renderHeight = height
}

type Field struct {
	Key   string
	Value string
}

type Highlight struct {
	Start int
	End   int
}

type PartKind int

const (
	PartRaw PartKind = iota
	PartTimestamp
	PartLevel
	PartMessage
	PartCaller
	PartContext
)

type Part struct {
	Kind  PartKind
	Text  string
	highlights []Highlight
}
