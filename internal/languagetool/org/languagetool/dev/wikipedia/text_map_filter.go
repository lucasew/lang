package wikipedia

// TextMapFilter ports org.languagetool.dev.wikipedia.TextMapFilter.
type TextMapFilter interface {
	// Filter maps wiki markup to plain text (mapping may be identity).
	FilterMapped(text string) *PlainTextMapping
}

// FilterMapped adapts SimpleWikipediaTextFilter to TextMapFilter.
func (f *SimpleWikipediaTextFilter) FilterMapped(text string) *PlainTextMapping {
	if f == nil {
		f = NewSimpleWikipediaTextFilter()
	}
	plain := f.Filter(text)
	return NewPlainTextMappingWithOriginal(plain, text)
}

// Ensure SimpleWikipediaTextFilter implements TextMapFilter.
var _ TextMapFilter = (*SimpleWikipediaTextFilter)(nil)
