package es

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	tagginges "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/es"
)

// WireSpanishSpellerTagger ports SpanishTagger.INSTANCE for orderSuggestions
// pronoun reordering and digit-split isTagged checks.
func WireSpanishSpellerTagger(r *MorfologikSpanishSpellerRule, t *tagginges.SpanishTagger) {
	if r == nil || t == nil {
		return
	}
	r.TagPOS = func(word string) []string {
		if word == "" {
			return nil
		}
		atrs := t.Tag([]string{word})
		if len(atrs) == 0 || atrs[0] == nil {
			return nil
		}
		var out []string
		for _, rd := range atrs[0].GetReadings() {
			if rd == nil || rd.GetPOSTag() == nil {
				continue
			}
			if p := *rd.GetPOSTag(); p != "" {
				out = append(out, p)
			}
		}
		return out
	}
}

// WireSpanishSpellerTagPOS sets TagPOS from a TagWord inject (fail-closed when nil).
func WireSpanishSpellerTagPOS(r *MorfologikSpanishSpellerRule, tagWord func(token string) []languagetool.TokenTag) {
	if r == nil || tagWord == nil {
		return
	}
	r.TagPOS = func(word string) []string {
		tags := tagWord(word)
		out := make([]string, 0, len(tags))
		for _, t := range tags {
			if t.POS != "" {
				out = append(out, t.POS)
			}
		}
		return out
	}
}
