package pt

import (
	taggingpt "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/pt"
)

// WirePortugueseSpellerTagger ports Java MorfologikPortugueseSpellerRule's use of
// PortugueseTagger.INSTANCE for isValidCliticVerb / dialectAlternative lemma path.
// Sets TagReadings (same-reading pairs), TagPOS, and TagLemma from Tag() (fail-closed if nil).
//
// Java POS for clitics is often "V…:P…" from the dict; tags are returned as-is.
func WirePortugueseSpellerTagger(r *MorfologikPortugueseSpellerRule, t *taggingpt.PortugueseTagger) {
	if r == nil || t == nil {
		return
	}
	r.TagReadings = func(word string) []PortugueseSpellerReading {
		if word == "" {
			return nil
		}
		atrs := t.Tag([]string{word})
		if len(atrs) == 0 || atrs[0] == nil {
			return nil
		}
		var out []PortugueseSpellerReading
		for _, rd := range atrs[0].GetReadings() {
			if rd == nil {
				continue
			}
			pos, lem := "", ""
			if rd.GetPOSTag() != nil {
				pos = *rd.GetPOSTag()
			}
			if rd.GetLemma() != nil {
				lem = *rd.GetLemma()
			}
			if pos == "" && lem == "" {
				continue
			}
			out = append(out, PortugueseSpellerReading{Lemma: lem, POS: pos})
		}
		return out
	}
	r.TagPOS = func(word string) []string {
		rds := r.TagReadings(word)
		if len(rds) == 0 {
			return nil
		}
		var out []string
		for _, rd := range rds {
			if rd.POS != "" {
				out = append(out, rd.POS)
			}
		}
		return out
	}
	r.TagLemma = func(word string) []string {
		rds := r.TagReadings(word)
		if len(rds) == 0 {
			return nil
		}
		var out []string
		seen := map[string]struct{}{}
		for _, rd := range rds {
			if rd.Lemma == "" {
				continue
			}
			if _, ok := seen[rd.Lemma]; ok {
				continue
			}
			seen[rd.Lemma] = struct{}{}
			out = append(out, rd.Lemma)
		}
		return out
	}
}
