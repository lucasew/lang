package wikipedia

// MarkupAwareWikipediaResult ports
// org.languagetool.dev.wikipedia.MarkupAwareWikipediaResult.
type MarkupAwareWikipediaResult struct {
	OriginalWikiMarkup MediaWikiContent
	AppliedRuleMatches []*AppliedRuleMatch
	InternalErrors     int
}

func NewMarkupAwareWikipediaResult(wiki MediaWikiContent, applied []*AppliedRuleMatch, internalErrors int) *MarkupAwareWikipediaResult {
	return &MarkupAwareWikipediaResult{
		OriginalWikiMarkup: wiki,
		AppliedRuleMatches: applied,
		InternalErrors:     internalErrors,
	}
}

func (r *MarkupAwareWikipediaResult) GetAppliedRuleMatches() []*AppliedRuleMatch {
	return r.AppliedRuleMatches
}
func (r *MarkupAwareWikipediaResult) GetInternalErrorCount() int { return r.InternalErrors }
func (r *MarkupAwareWikipediaResult) GetOriginalWikiMarkup() string {
	return r.OriginalWikiMarkup.GetContent()
}
func (r *MarkupAwareWikipediaResult) GetLastEditTimestamp() string {
	return r.OriginalWikiMarkup.GetTimestamp()
}
