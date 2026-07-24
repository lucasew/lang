package uk

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// nameSuffixes ports CompoundTagger.NAME_SUFFIX (Мустафа-ага, …).
var nameSuffixes = map[string]struct{}{
	"ага": {}, "ефенді": {}, "бек": {}, "заде": {},
	"огли": {}, "сан": {}, "кизи": {}, "сенсей": {},
}

// DynamicNameSuffixReadings ports CompoundTagger NAME_SUFFIX branch.
// Left must have POS containing "name" (fname/lname/…) from wordTagger; right is
// a fixed official suffix. Lemma gets "-" + right (Java PosTagHelper.adjust).
// Fail-closed without dict name tags — no invent prop:fname from surface alone.
func DynamicNameSuffixReadings(token string, tagWord func(string) []tagging.TaggedWord) []struct{ Lemma, POS string } {
	if tagWord == nil || token == "" || !strings.Contains(token, "-") {
		return nil
	}
	if strings.Count(token, "-") != 1 {
		return nil
	}
	dash := strings.LastIndex(token, "-")
	if dash <= 0 || dash == len(token)-1 {
		return nil
	}
	leftWord := token[:dash]
	rightWord := token[dash+1:]
	// Java: NAME_SUFFIX.contains(rightWord) — list is lowercase
	if _, ok := nameSuffixes[rightWord]; !ok {
		// also accept lower right
		if _, ok2 := nameSuffixes[strings.ToLower(rightWord)]; !ok2 {
			return nil
		}
		rightWord = strings.ToLower(rightWord)
	}

	tws := tagWord(leftWord)
	if len(tws) == 0 {
		low := strings.ToLower(leftWord)
		if low != leftWord {
			tws = tagWord(low)
		}
	}
	if len(tws) == 0 {
		return nil
	}

	var out []struct{ Lemma, POS string }
	seen := map[string]struct{}{}
	for _, tw := range tws {
		pos := tw.PosTag
		// Java: hasPosTagPart(..., "name")
		if pos == "" || !strings.Contains(pos, "name") {
			continue
		}
		lem := tw.Lemma
		if lem == "" {
			lem = leftWord
		}
		// PosTagHelper.adjust(leftWdList, null, "-" + rightWord)
		lemma := lem + "-" + rightWord
		key := lemma + "|" + pos
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, struct{ Lemma, POS string }{Lemma: lemma, POS: pos})
	}
	return out
}
