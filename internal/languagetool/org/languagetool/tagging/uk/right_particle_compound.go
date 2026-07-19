package uk

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// rightPartsWithLeftTag ports CompoundTagger.rightPartsWithLeftTagMap
// (стривай-бо, чекай-но, прийшов-таки, такий-от, такий-то, ішов-єм).
// Values are POS matchers approximating Java regex (RE2-safe).
var rightPartsWithLeftTag = map[string]func(pos string) bool{
	"бо": func(pos string) bool {
		return strings.HasPrefix(pos, "verb") || strings.Contains(pos, "pron") ||
			strings.HasPrefix(pos, "noun") || strings.HasPrefix(pos, "adv") ||
			strings.HasPrefix(pos, "intj") || strings.HasPrefix(pos, "part")
	},
	"но": func(pos string) bool {
		// Java: (verb(?!.*bad).*?:(impr|futr|insert))|intj|adv|part|conj
		if strings.HasPrefix(pos, "verb") && !strings.Contains(pos, "bad") &&
			(strings.Contains(pos, "impr") || strings.Contains(pos, "futr") || strings.Contains(pos, "insert")) {
			return true
		}
		return strings.HasPrefix(pos, "intj") || strings.HasPrefix(pos, "adv") ||
			strings.HasPrefix(pos, "part") || strings.HasPrefix(pos, "conj")
	},
	"от": func(pos string) bool {
		return strings.Contains(pos, "pron") || strings.HasPrefix(pos, "adv") ||
			strings.HasPrefix(pos, "part") || strings.HasPrefix(pos, "verb")
	},
	"то": func(pos string) bool {
		return strings.Contains(pos, "pron") || strings.HasPrefix(pos, "verb") ||
			strings.HasPrefix(pos, "noun") || strings.HasPrefix(pos, "adj") ||
			strings.HasPrefix(pos, "adv") || strings.HasPrefix(pos, "conj")
	},
	"таки": func(pos string) bool {
		return strings.HasPrefix(pos, "verb") || strings.HasPrefix(pos, "adv") ||
			strings.HasPrefix(pos, "adj") || strings.Contains(pos, "pron") ||
			strings.HasPrefix(pos, "part") ||
			(strings.HasPrefix(pos, "noninfl") && strings.Contains(pos, "predic"))
	},
	"єм": func(pos string) bool {
		return strings.HasPrefix(pos, "verb")
	},
}

// DynamicRightParticleReadings ports the rightPartsWithLeftTagMap branch.
// Keeps left POS/lemma when left matches the particle's allowed tag set.
// Fail-closed without left dict tags — no invent.
func DynamicRightParticleReadings(token string, tagWord func(string) []tagging.TaggedWord) []struct{ Lemma, POS string } {
	if tagWord == nil || token == "" || !strings.Contains(token, "-") {
		return nil
	}
	if strings.Count(token, "-") != 1 {
		return nil
	}
	dash := strings.LastIndex(token, "-")
	leftWord := token[:dash]
	rightWord := token[dash+1:]
	matchLeft, ok := rightPartsWithLeftTag[strings.ToLower(rightWord)]
	if !ok {
		return nil
	}

	leftTags := lookupBothCases(leftWord, tagWord)
	if len(leftTags) == 0 {
		return nil
	}
	// Java: ignore left abbr
	for _, tw := range leftTags {
		if strings.Contains(tw.PosTag, "abbr") {
			return nil
		}
	}

	// ignore хто-то / що-то / чи-то
	if strings.EqualFold(rightWord, "то") {
		for _, tw := range leftTags {
			lem := strings.ToLower(tw.Lemma)
			if lem == "" {
				lem = strings.ToLower(leftWord)
			}
			if lem == "хто" || lem == "що" || lem == "чи" {
				return nil
			}
		}
	}

	leftLow := strings.ToLower(leftWord)
	var out []struct{ Lemma, POS string }
	seen := map[string]struct{}{}
	for _, tw := range leftTags {
		pos := tw.PosTag
		if pos == "" {
			continue
		}
		// Java: ignore як + noun
		if strings.EqualFold(leftWord, "як") && strings.Contains(pos, "noun") {
			continue
		}
		// Java: (leftWordLowerCase.equals("дуже") && posTag.contains("adv")) || leftTagRegex
		okMatch := matchLeft(pos)
		if leftLow == "дуже" && strings.Contains(pos, "adv") {
			okMatch = true
		}
		if !okMatch {
			continue
		}
		if strings.EqualFold(rightWord, "єм") && !strings.Contains(pos, ":arch") {
			pos = pos + ":arch"
		}
		lem := tw.Lemma
		if lem == "" {
			lem = leftWord
		}
		key := lem + "|" + pos
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, struct{ Lemma, POS string }{Lemma: lem, POS: pos})
	}
	return out
}

// DynamicDualPropReadings ports capitalized dual proper compounds:
// Київ-Прага, Карпа-Хансен, Україна-ЄС (Java CompoundTagger both-sides upper branch).
// Fail-closed without matching dict POS on both sides.
func DynamicDualPropReadings(token string, tagWord func(string) []tagging.TaggedWord) []struct{ Lemma, POS string } {
	if tagWord == nil || token == "" || !strings.Contains(token, "-") {
		return nil
	}
	if strings.Count(token, "-") != 1 {
		return nil
	}
	dash := strings.LastIndex(token, "-")
	leftWord := token[:dash]
	rightWord := token[dash+1:]
	if leftWord == "" || rightWord == "" {
		return nil
	}
	// both sides start uppercase
	lr, rr := []rune(leftWord), []rune(rightWord)
	if !unicode.IsUpper(lr[0]) || !unicode.IsUpper(rr[0]) {
		return nil
	}

	leftTags := lookupBothCases(leftWord, tagWord)
	rightTags := lookupBothCases(rightWord, tagWord)
	if len(leftTags) == 0 || len(rightTags) == 0 {
		return nil
	}

	has := func(tags []tagging.TaggedWord, pred func(string) bool) bool {
		for _, tw := range tags {
			if pred(tw.PosTag) {
				return true
			}
		}
		return false
	}
	// GEO_V_NAZ: noun:inanim:.:v_naz.*:geo.*
	geoNaz := func(pos string) bool {
		return strings.HasPrefix(pos, "noun:inanim:") && strings.Contains(pos, "v_naz") && strings.Contains(pos, "geo")
	}
	// FNAME: noun:anim:[mf].*fname.*
	fname := func(pos string) bool {
		return strings.HasPrefix(pos, "noun:anim:") && strings.Contains(pos, "fname")
	}
	// LNAME_V_NAZ / LNAME_V_ROD
	lnameNaz := func(pos string) bool {
		return strings.HasPrefix(pos, "noun:anim:") && strings.Contains(pos, "v_naz") && strings.Contains(pos, "lname")
	}
	lnameRod := func(pos string) bool {
		return strings.HasPrefix(pos, "noun:anim:") && strings.Contains(pos, "v_rod") && strings.Contains(pos, "lname")
	}
	// NAME: noun:anim:.*name.*
	anyName := func(pos string) bool {
		return strings.HasPrefix(pos, "noun:anim:") && strings.Contains(pos, "name")
	}
	// PROP_V_NAZ: noun:inanim:.:v_naz.*prop.*
	propNaz := func(pos string) bool {
		return strings.HasPrefix(pos, "noun:inanim:") && strings.Contains(pos, "v_naz") && strings.Contains(pos, "prop")
	}

	if has(leftTags, geoNaz) && has(rightTags, geoNaz) {
		return []struct{ Lemma, POS string }{{Lemma: token, POS: "noninfl:prop:geo"}}
	}
	if has(leftTags, fname) && has(rightTags, fname) {
		// Java returns tagMatch of filtered fname — use shared FullTagMatch path
		// simplified: noninfl dual fname not exact; emit tagMatch-style noun readings
		var out []struct{ Lemma, POS string }
		for _, lt := range leftTags {
			if !fname(lt.PosTag) {
				continue
			}
			for _, rt := range rightTags {
				if !fname(rt.PosTag) {
					continue
				}
				lem := lt.Lemma
				if lem == "" {
					lem = leftWord
				}
				rlem := rt.Lemma
				if rlem == "" {
					rlem = rightWord
				}
				out = append(out, struct{ Lemma, POS string }{Lemma: lem + "-" + rlem, POS: lt.PosTag})
			}
		}
		return out
	}
	if has(leftTags, lnameNaz) && has(rightTags, lnameNaz) {
		return []struct{ Lemma, POS string }{{Lemma: token, POS: "noninfl:prop:lname"}}
	}
	if has(leftTags, lnameRod) && has(rightTags, lnameRod) {
		return []struct{ Lemma, POS string }{{Lemma: token, POS: "noninfl:prop:lname"}}
	}
	// bad: both name but not handled above → leave untagged (Java returns null)
	if has(leftTags, anyName) && has(rightTags, anyName) {
		return nil
	}
	if has(leftTags, propNaz) && has(rightTags, propNaz) {
		return []struct{ Lemma, POS string }{{Lemma: token, POS: "noninfl:prop"}}
	}
	return nil
}
