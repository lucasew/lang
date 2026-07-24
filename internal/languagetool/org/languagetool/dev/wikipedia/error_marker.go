package wikipedia

// ErrorMarker ports org.languagetool.dev.wikipedia.ErrorMarker.
type ErrorMarker struct {
	StartMarker string
	EndMarker   string
}

func NewErrorMarker(start, end string) ErrorMarker {
	return ErrorMarker{StartMarker: start, EndMarker: end}
}

func DefaultErrorMarker() ErrorMarker {
	// Java default uses <<span>> to avoid clashes with <span> in original markup
	return NewErrorMarker(`<<span class="error">>`, `<</span>>`)
}

func (e ErrorMarker) GetStartMarker() string { return e.StartMarker }
func (e ErrorMarker) GetEndMarker() string   { return e.EndMarker }
