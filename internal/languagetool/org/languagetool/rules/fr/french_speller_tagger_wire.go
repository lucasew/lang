package fr

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	taggingfr "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/fr"
)

// WireFrenchSpellerTagger ports Java MorfologikFrenchSpellerRule's use of
// FrenchTagger.INSTANCE in findSuggestion. Sets TagPOS from a single-token Tag().
func WireFrenchSpellerTagger(r *MorfologikFrenchSpellerRule, t *taggingfr.FrenchTagger) {
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

// WireFrenchSpellerTagPOS sets TagPOS from a JLanguageTool-style TagWord inject
// (BinaryPOSTagWord / MapWordTagger bridge). Fail-closed when tagWord nil.
func WireFrenchSpellerTagPOS(r *MorfologikFrenchSpellerRule, tagWord func(token string) []languagetool.TokenTag) {
	if r == nil || tagWord == nil {
		return
	}
	r.TagPOS = func(word string) []string {
		tags := tagWord(word)
		if len(tags) == 0 {
			return nil
		}
		out := make([]string, 0, len(tags))
		for _, t := range tags {
			if t.POS != "" {
				out = append(out, t.POS)
			}
		}
		return out
	}
}
