package logs

type Entry struct {
	Raw           string
	Parsed        bool
	Level         Part
	Timestamp     Part
	Message       Part
	Caller        Part
	Context       Part
	Search        string
	rendered      bool
	renderedText  string
	renderHeight  int
	viewportWidth int
}

func (e *Entry) CacheRenderedString(viewportWidth int, rendered string, height int) {
	e.rendered = true
	e.renderedText = rendered
	e.renderHeight = height
	e.viewportWidth = viewportWidth
}

func (e Entry) GetCachedRender(viewportWidth int) (string, bool) {
	if e.rendered && e.viewportWidth == viewportWidth {
		return e.renderedText, true
	}
	return "", false
}

func (e Entry) GetCachedHeight(viewportWidth int) (int, bool) {
	if e.rendered && e.viewportWidth == viewportWidth {
		return e.renderHeight, true
	}
	return 0, false
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
	Kind       PartKind
	Text       string
	highlights []Highlight
}
