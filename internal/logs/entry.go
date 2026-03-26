package logs

type Entry struct {
	Raw       string
	Parsed    bool
	Level     string
	Timestamp string
	Message   string
	Caller    string
	Context   []Field
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
	Level string
}
