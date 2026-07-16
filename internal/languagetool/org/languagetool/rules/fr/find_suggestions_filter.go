package fr

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"

// FindSuggestionsFilter wraps the core FindSuggestionsFilter for fr.
// Speller/tagger wiring is deferred (pluggable hooks on the embedded filter).
type FindSuggestionsFilter struct {
	*rules.FindSuggestionsFilter
}

func NewFindSuggestionsFilter() *FindSuggestionsFilter {
	return &FindSuggestionsFilter{FindSuggestionsFilter: rules.NewFindSuggestionsFilter()}
}
