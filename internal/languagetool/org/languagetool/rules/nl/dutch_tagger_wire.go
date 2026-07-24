package nl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	taggingnl "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/nl"
)

// WireDutchTaggerCompoundParts ports DutchTagger's use of
// Dutch.getCompoundAcceptor().getParts — inject DefaultCompoundAcceptor.GetParts
// so tagging/nl never imports rules/nl (avoids import cycle).
func WireDutchTaggerCompoundParts(t *taggingnl.DutchTagger) {
	if t == nil {
		return
	}
	t.GetCompoundParts = func(word string) []string {
		return DefaultCompoundAcceptor.GetParts(word)
	}
}

// NewWiredDutchTagger builds DutchTagger with CompoundAcceptor.getParts wired
// (Java DutchTagger.INSTANCE + CompoundAcceptor.INSTANCE).
func NewWiredDutchTagger(wt tagging.WordTagger) *taggingnl.DutchTagger {
	t := taggingnl.NewDutchTagger(wt)
	WireDutchTaggerCompoundParts(t)
	return t
}
