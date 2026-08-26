package uk

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// DynamicIntjRedupReadings ports CompoundTagger.doGuessMultiHyphens intj/noninfl
// reduplication (а-а, га-га, гей-гей-гей). Requires wordTagger hits on parts —
// fail-closed without dict (no invent intj list).
func DynamicIntjRedupReadings(token string, tagWord func(string) []tagging.TaggedWord) []struct{ Lemma, POS string } {
	if tagWord == nil || token == "" || !strings.Contains(token, "-") {
		return nil
	}
	// Normalize dash variants
	t := strings.ReplaceAll(token, "–", "-")
	t = strings.ReplaceAll(t, "—", "-")
	parts := strings.Split(strings.ToLower(t), "-")
	if len(parts) < 2 {
		return nil
	}
	for _, p := range parts {
		if p == "" {
			return nil
		}
	}

	// unique parts (Java LinkedHashSet)
	uniq := uniqueStrings(parts)
	lemma := strings.ToLower(t)

	// set.size()==1: all parts equal (а-а, гей-гей-гей)
	if len(uniq) == 1 {
		// Java special: lowerWord.equals("ла") → intj (ла-ла…)
		if uniq[0] == "ла" {
			return []struct{ Lemma, POS string }{{Lemma: lemma, POS: "intj"}}
		}
		pos := firstIntjOrNoninflPOS(tagWord, uniq[0])
		if pos == "" {
			return nil
		}
		return []struct{ Lemma, POS string }{{Lemma: lemma, POS: pos}}
	}

	// set.size()==2: both sides intj or both noninfl onomat/predic
	if len(uniq) == 2 {
		posL := firstIntjOrNoninflPOS(tagWord, uniq[0])
		posR := firstIntjOrNoninflPOS(tagWord, uniq[1])
		if posL == "" || posR == "" {
			return nil
		}
		// Java: both intj OR both noninfl.*(onomat|predic)
		if isIntjPOS(posL) && isIntjPOS(posR) {
			return []struct{ Lemma, POS string }{{Lemma: lemma, POS: posR}}
		}
		if isNoninflOnomat(posL) && isNoninflOnomat(posR) {
			return []struct{ Lemma, POS string }{{Lemma: lemma, POS: posR}}
		}
	}
	return nil
}

func uniqueStrings(parts []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, p := range parts {
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}

func firstIntjOrNoninflPOS(tagWord func(string) []tagging.TaggedWord, w string) string {
	tws := tagWord(w)
	if len(tws) == 0 {
		// try title-case surface of part (Га from Га-га is lowercased already)
		return ""
	}
	for _, tw := range tws {
		if isIntjPOS(tw.PosTag) || isNoninflOnomat(tw.PosTag) {
			return tw.PosTag
		}
	}
	return ""
}

func isIntjPOS(pos string) bool {
	return strings.HasPrefix(pos, "intj")
}

func isNoninflOnomat(pos string) bool {
	// Java NONINFL_PATTERN: noninfl.*(onomat|predic).*
	if !strings.HasPrefix(pos, "noninfl") {
		return false
	}
	return strings.Contains(pos, "onomat") || strings.Contains(pos, "predic")
}
