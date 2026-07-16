package nl

import (
	"fmt"
	"regexp"
)

// CompoundFilter ports org.languagetool.rules.nl.CompoundFilter.
type CompoundFilter struct{}

func NewCompoundFilter() *CompoundFilter {
	return &CompoundFilter{}
}

var suggestionElem = regexp.MustCompile(`(?s)<suggestion>.*?</suggestion>`)

// Suggest glues word1..wordN parts into a Dutch compound suggestion.
func (f *CompoundFilter) Suggest(words []string) string {
	return GlueParts(words)
}

// RewriteMessage replaces <suggestion>…</suggestion> with the glued compound.
func (f *CompoundFilter) RewriteMessage(message, compound string) string {
	return suggestionElem.ReplaceAllString(message, "<suggestion>"+compound+"</suggestion>")
}

// SuggestFromArgs reads word1..word5 from a pattern-filter args map.
func (f *CompoundFilter) SuggestFromArgs(args map[string]string) string {
	var words []string
	for i := 1; i < 6; i++ {
		arg, ok := args[fmt.Sprintf("word%d", i)]
		if !ok {
			break
		}
		words = append(words, arg)
	}
	return f.Suggest(words)
}
