package pt

import (
	taggingpt "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/pt"
)

// WirePortugueseSpellerTagger ports Java MorfologikPortugueseSpellerRule's use of
// PortugueseTagger.INSTANCE for isValidCliticVerb / dialectAlternative lemma path.
// Sets TagPOS and TagLemma from a single-token Tag() call (fail-closed if tagger nil).
//
// Java POS for clitics is often "V…:P…" from the dict; TagPOS returns those tags as-is.
func WirePortugueseSpellerTagger(r *MorfologikPortugueseSpellerRule, t *taggingpt.PortugueseTagger) {
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
			out = append(out, *rd.GetPOSTag())
		}
		return out
	}
	r.TagLemma = func(word string) []string {
		if word == "" {
			return nil
		}
		atrs := t.Tag([]string{word})
		if len(atrs) == 0 || atrs[0] == nil {
			return nil
		}
		var out []string
		seen := map[string]struct{}{}
		for _, rd := range atrs[0].GetReadings() {
			if rd == nil || rd.GetLemma() == nil {
				continue
			}
			l := *rd.GetLemma()
			if l == "" {
				continue
			}
			if _, ok := seen[l]; ok {
				continue
			}
			seen[l] = struct{}{}
			out = append(out, l)
		}
		return out
	}
}
