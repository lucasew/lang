package uk

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// Ports UkrainianTagger.additionalTags / getAnalyzedTokens alt rewrites.
// All paths require dictionary hits (fail closed; no invent).

// YI_PATTERN: consonant + ї → consonant + і
var yiPatternRE = regexp.MustCompile(`(?i)([бвгґджзклмнпрстфхцчшщ])ї`)

// Java filter2 for missing apostrophe: exclude bad/arch/alt/abbr/slang/subst/short/long.
var missingApoExcluded = []string{":bad", ":arch", ":alt", ":abbr", ":slang", ":subst", ":short", ":long"}

// AltTagAdjustReadings ports additionalTags CAPS_INSIDE / з→с / ї→і and
// getAnalyzedTokens convertTokens (ґ/ія/тер/льо/сьвя/сьві/ьск) when untagged.
func AltTagAdjustReadings(word string, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	if tagWord == nil || word == "" {
		return nil
	}
	if rs := capsInsideAltReadings(word, tagWord); len(rs) > 0 {
		return rs
	}
	if rs := zLabialAltReadings(word, tagWord); len(rs) > 0 {
		return rs
	}
	if rs := yiToIAltReadings(word, tagWord); len(rs) > 0 {
		return rs
	}
	if rs := convertTokenAltReadings(word, tagWord); len(rs) > 0 {
		return rs
	}
	return nil
}

// capsInsideAltReadings: length>5, mixed-case interior, tag lower + :alt.
func capsInsideAltReadings(word string, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	if tagging.UTF16Len(word) <= 5 {
		return nil
	}
	// Java UNICODE_CASE CAPS_INSIDE — require an interior upper after a lower Cyrillic.
	if !hasCapsInside(word) {
		return nil
	}
	lower := strings.ToLower(word)
	return taggedWithExtra(word, tagWord(lower), ":alt", nil)
}

func hasCapsInside(word string) bool {
	// Java CAPS_INSIDE_WORD: … lower-cyrillic, upper-cyrillic, lower-cyrillic …
	rs := []rune(word)
	for i := 0; i+2 < len(rs); i++ {
		a, b, c := rs[i], rs[i+1], rs[i+2]
		if unicode.IsLower(a) && unicode.IsUpper(b) && unicode.IsLower(c) &&
			isCyrLetter(a) && isCyrLetter(b) && isCyrLetter(c) {
			return true
		}
	}
	return false
}

func isCyrLetter(r rune) bool {
	switch {
	case r >= 'а' && r <= 'я', r >= 'А' && r <= 'Я':
		return true
	case r == 'і' || r == 'ї' || r == 'є' || r == 'ґ' || r == 'І' || r == 'Ї' || r == 'Є' || r == 'Ґ':
		return true
	}
	return false
}

// zLabialAltReadings: з/З before кптфх → с/С + :alt; lemma maps с back to з.
func zLabialAltReadings(word string, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	if tagging.UTF16Len(word) <= 5 {
		return nil
	}
	rs := []rune(word)
	if len(rs) < 2 {
		return nil
	}
	first := rs[0]
	if first != 'з' && first != 'З' {
		return nil
	}
	sec := unicode.ToLower(rs[1])
	if !strings.ContainsRune("кптфх", sec) {
		return nil
	}
	// replace first з/З with с/С
	adj := make([]rune, len(rs))
	copy(adj, rs)
	if first == 'з' {
		adj[0] = 'с'
	} else {
		adj[0] = 'С'
	}
	adjusted := string(adj)
	wdList := tagWord(adjusted)
	if len(wdList) == 0 {
		// try both cases like Java tagBothCases — lower of adjusted
		if low := strings.ToLower(adjusted); low != adjusted {
			wdList = tagWord(low)
		}
	}
	if len(wdList) == 0 {
		return nil
	}
	// map lemma: ^с → з, ^С → З
	var out []*languagetool.AnalyzedToken
	for _, tw := range wdList {
		lemma := tw.Lemma
		if strings.HasPrefix(lemma, "с") {
			lemma = "з" + lemma[len("с"):]
		} else if strings.HasPrefix(lemma, "С") {
			lemma = "З" + lemma[len("С"):]
		}
		pos := AddIfNotContains(tw.PosTag, ":alt")
		p, l := pos, lemma
		out = append(out, languagetool.NewAnalyzedToken(word, &p, &l))
	}
	return out
}

// yiToIAltReadings: дївчина-style consonant+ї → і + :alt.
func yiToIAltReadings(word string, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	if tagging.UTF16Len(word) <= 3 || !strings.Contains(strings.ToLower(word), "ї") {
		return nil
	}
	adjusted := yiPatternRE.ReplaceAllString(word, "${1}і")
	if adjusted == word {
		return nil
	}
	return taggedWithExtra(word, tagWord(adjusted), ":alt", nil)
}

// convertTokenAltReadings ports convertTokens chain from getAnalyzedTokens.
func convertTokenAltReadings(word string, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	type repl struct {
		from, to, tag string
		// singleRune: also replace uppercase form of from/to
		singleRune bool
		// startsWith: only if word starts with from (case-insensitive for match body)
		startsWith bool
		// endsWith: only if word ends with from
		endsWith bool
		// skipIf: optional predicates
		skipIf func(string) bool
	}
	chain := []repl{
		{from: "ґ", to: "г", tag: ":alt", singleRune: true},
		{from: "ія", to: "іа", tag: ":alt"},
		{from: "тер", to: "тр", tag: ":alt", endsWith: true},
		{from: "льо", to: "ло", tag: ":alt"},
		{from: "сьвя", to: "свя", tag: ":arch", startsWith: true},
		{from: "сьві", to: "сві", tag: ":arch", startsWith: true},
		{from: "ьск", to: "ьськ", tag: ":bad", skipIf: func(w string) bool {
			return strings.HasSuffix(w, "ская") || w == "Комсомольском"
		}},
	}
	for _, r := range chain {
		if r.skipIf != nil && r.skipIf(word) {
			continue
		}
		if r.startsWith && !strings.HasPrefix(strings.ToLower(word), strings.ToLower(r.from)) {
			continue
		}
		if r.endsWith && !strings.HasSuffix(strings.ToLower(word), strings.ToLower(r.from)) {
			continue
		}
		if !strings.Contains(word, r.from) && !(r.singleRune && strings.Contains(word, strings.ToUpper(r.from))) {
			// for multi-char from, also check case-insensitive via lower for contains of lower from
			if !strings.Contains(strings.ToLower(word), strings.ToLower(r.from)) {
				continue
			}
		}
		adjusted := word
		if r.singleRune {
			adjusted = strings.ReplaceAll(adjusted, r.from, r.to)
			adjusted = strings.ReplaceAll(adjusted, strings.ToUpper(r.from), strings.ToUpper(r.to))
		} else {
			// case-preserving simple replace of lowercase form only (Java replace is case-sensitive)
			if !strings.Contains(word, r.from) {
				continue
			}
			adjusted = strings.ReplaceAll(word, r.from, r.to)
		}
		if adjusted == word {
			continue
		}
		wdList := tagWord(adjusted)
		if len(wdList) == 0 {
			if low := strings.ToLower(adjusted); low != adjusted {
				wdList = tagWord(low)
			}
		}
		if len(wdList) == 0 {
			continue
		}
		// lemma: replace dictStr back to str (Java lemmaFunction)
		var out []*languagetool.AnalyzedToken
		for _, tw := range wdList {
			lemma := tw.Lemma
			if r.singleRune {
				lemma = strings.ReplaceAll(lemma, r.to, r.from)
				lemma = strings.ReplaceAll(lemma, strings.ToUpper(r.to), strings.ToUpper(r.from))
			} else {
				lemma = strings.ReplaceAll(lemma, r.to, r.from)
			}
			pos := AddIfNotContains(tw.PosTag, r.tag)
			p, l := pos, lemma
			out = append(out, languagetool.NewAnalyzedToken(word, &p, &l))
		}
		return out
	}
	return nil
}

// taggedWithExtra builds AnalyzedTokens with optional extra POS tag and lemma map.
func taggedWithExtra(surface string, words []tagging.TaggedWord, extra string, lemmaMap func(string) string) []*languagetool.AnalyzedToken {
	if len(words) == 0 {
		return nil
	}
	var out []*languagetool.AnalyzedToken
	for _, tw := range words {
		lemma := tw.Lemma
		if lemmaMap != nil {
			lemma = lemmaMap(lemma)
		}
		pos := tw.PosTag
		if extra != "" {
			pos = AddIfNotContains(pos, extra)
		}
		p, l := pos, lemma
		out = append(out, languagetool.NewAnalyzedToken(surface, &p, &l))
	}
	return out
}

// filterMissingApoTags ports PosTagHelper.filter2 exclude bad/arch/alt/…
func filterMissingApoTags(words []tagging.TaggedWord) []tagging.TaggedWord {
	var out []tagging.TaggedWord
	for _, w := range words {
		if w.PosTag == "" {
			continue
		}
		skip := false
		for _, ex := range missingApoExcluded {
			if strings.Contains(w.PosTag, ex) {
				skip = true
				break
			}
		}
		if !skip {
			out = append(out, w)
		}
	}
	return out
}

// AnalyzeAllCapitalizedAdj ports UkrainianTagger.analyzeAllCapitamizedAdj:
// hyphenated all-capitalized parts → lower adj dictionary hits only.
func AnalyzeAllCapitalizedAdj(word string, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	if tagWord == nil || word == "" || strings.HasSuffix(word, "-") {
		return nil
	}
	dash := strings.Index(word, "-")
	if dash <= 1 {
		return nil
	}
	parts := strings.Split(word, "-")
	if len(parts) < 2 {
		return nil
	}
	for _, p := range parts {
		if p == "" || !isCapitalizedUK(p) {
			return nil
		}
	}
	lower := strings.ToLower(word)
	wdList := tagWord(lower)
	if !HasPosTagPart2(wdList, "adj") {
		return nil
	}
	var out []*languagetool.AnalyzedToken
	for _, tw := range wdList {
		if !strings.HasPrefix(tw.PosTag, "adj") {
			continue
		}
		p, l := tw.PosTag, tw.Lemma
		out = append(out, languagetool.NewAnalyzedToken(word, &p, &l))
	}
	return out
}

// isCapitalizedUK ports LemmaHelper.isCapitalized for Ukrainian letters.
func isCapitalizedUK(s string) bool {
	rs := []rune(s)
	if len(rs) == 0 {
		return false
	}
	if !unicode.IsUpper(rs[0]) {
		return false
	}
	for _, r := range rs[1:] {
		if unicode.IsLetter(r) && !unicode.IsLower(r) {
			return false
		}
	}
	return true
}
