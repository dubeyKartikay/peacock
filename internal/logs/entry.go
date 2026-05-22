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
	renderCache   renderCache
}
type renderCache struct {
	rendered      bool
	renderedText  string
	renderHeight  int
	viewportWidth int
}

func (r *renderCache) cacheRenderedString(viewportWidth int, rendered string, height int) {
	r.rendered = true
	r.renderedText = rendered
	r.renderHeight = height
	r.viewportWidth = viewportWidth
}

func (r *renderCache) reset() {
	r.rendered = false
}

func (r renderCache) checkIfCached(viewportWidth int) bool {
	return r.rendered && r.viewportWidth == viewportWidth
}


func (e *Entry) GetCachedRender(viewportWidth int,invalidate bool) (string, bool) {
	if(invalidate){
		e.renderCache.reset()
	}
	if e.renderCache.checkIfCached(viewportWidth) {
		return e.renderCache.renderedText, true
	}
	return "", false
}

func (e Entry) GetCachedHeight(viewportWidth int) (int, bool) {
	if e.renderCache.checkIfCached(viewportWidth) {
		return e.renderCache.renderHeight, true
	}
	return 0, false
}

func (e *Entry) CacheRender(viewportWidth int, rendered string, height int) {
	e.renderCache.cacheRenderedString(viewportWidth, rendered, height)
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
