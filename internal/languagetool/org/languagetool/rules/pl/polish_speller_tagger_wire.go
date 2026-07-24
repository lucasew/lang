package pl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// WirePolishSpellerTagPOS sets TagPOS from a TagWord inject (fail-closed when nil).
// Used by isNotCompound adj / num:comp arms.
func WirePolishSpellerTagPOS(r *MorfologikPolishSpellerRule, tagWord func(token string) []languagetool.TokenTag) {
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
