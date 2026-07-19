package uk

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// vowels for apo-needed check (Java "аеєиіїоуюя")
const noDashUkVowels = "аеєиіїоуюяАЕЄИІЇОУЮЯ"

// yod letters that may need apostrophe after consonant prefix (Java "єїюя")
const noDashUkYod = "єїюяЄЇЮЯ"

// DynamicNoDashPrefixReadings ports CompoundTagger.guessOtherTagsInternal no-dash loop.
// Requires len>7, Ukrainian letters; prefix from official noDashPrefixes; right tagged
// as noun|adj|adv without pron; drops noun:inanim v_kly; apo / :bad rules.
func DynamicNoDashPrefixReadings(word string, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	if tagWord == nil || word == "" {
		return nil
	}
	if utf8.RuneCountInString(word) <= 7 || !ukrLettersOnly(word) {
		return nil
	}
	// Java: capitalized ending paradigms first (GuessOtherTagsReadings)
	// this function is only the no-dash prefix arm
	lowerCase := strings.ToLower(word)
	for _, prefix := range NoDashPrefixList() {
		if !strings.HasPrefix(lowerCase, prefix) {
			continue
		}
		// right with original case: word.substring(prefix.length())
		// map prefix length in runes on original
		pr := []rune(prefix)
		wr := []rune(word)
		if len(pr) >= len(wr) {
			continue
		}
		// prefix matched on lower; verify first len(pr) lower equals prefix
		head := string(wr[:len(pr)])
		if strings.ToLower(head) != prefix {
			continue
		}
		right := string(wr[len(pr):])
		apo := ""
		var addTags []string

		if strings.HasPrefix(right, "'") || strings.HasPrefix(right, "’") {
			// strip one apo
			rr := []rune(right)
			right = string(rr[1:])
			apo = "'"
		}
		if utf8.RuneCountInString(right) < 2 {
			continue
		}

		rr := []rune(right)
		prr := []rune(prefix)
		apoNeeded := false
		if strings.ContainsRune(noDashUkYod, rr[0]) && !strings.ContainsRune(noDashUkVowels, prr[len(prr)-1]) {
			apoNeeded = true
		}
		// екс'прес — unexpected apo
		if !apoNeeded && apo != "" {
			// Java: break (stop trying further prefixes)
			break
		}
		if apoNeeded == (apo == "") {
			addTags = append(addTags, ":bad")
		}

		// right length >= 4 && ! isCapitalizedWord(right)
		if utf8.RuneCountInString(right) < 4 || tools.IsCapitalizedWord(right) {
			continue
		}

		rightWdList := tagWord(right)
		if len(rightWdList) == 0 {
			// try lower right
			if low := strings.ToLower(right); low != right {
				rightWdList = tagWord(low)
			}
		}
		// filter PREFIX_NO_DASH_POSTAG_PATTERN: (noun|adj|adv)(?!.*pron).*
		var filtered []tagging.TaggedWord
		for _, tw := range rightWdList {
			if !isPrefixNoDashPOS(tw.PosTag) {
				continue
			}
			// remove noun:inanim … v_kly
			if strings.HasPrefix(tw.PosTag, "noun:inanim") && strings.Contains(tw.PosTag, "v_kly") {
				continue
			}
			filtered = append(filtered, tw)
		}
		if len(filtered) == 0 {
			continue
		}
		// adjust(lemmaPrefix=prefix+apo, addTags…)
		lemmaPrefix := prefix + apo
		adj := Adjust(filtered, lemmaPrefix, "", addTags...)
		return taggedWordsToSurfaceTokens(word, adj)
	}
	return nil
}

// isPrefixNoDashPOS ports PREFIX_NO_DASH_POSTAG_PATTERN without lookaround.
func isPrefixNoDashPOS(pos string) bool {
	if pos == "" {
		return false
	}
	if !strings.HasPrefix(pos, "noun") && !strings.HasPrefix(pos, "adj") && !strings.HasPrefix(pos, "adv") {
		return false
	}
	return !strings.Contains(pos, "pron")
}

// TryNoDashPrefixTags keeps the old API used by uk_tagger missing-hyphen path.
// Prefer DynamicNoDashPrefixReadings for full fidelity; this adapts tagRight callback.
func TryNoDashPrefixTags(surface string, tagRight func(string) []*languagetool.AnalyzedToken) []*languagetool.AnalyzedToken {
	if surface == "" || tagRight == nil {
		return nil
	}
	// Bridge: wrap AnalyzedToken results as TaggedWord lookups via tagWord surface.
	// For missing-hyphen candidates, tagRight already returns readings for right part.
	// Keep previous longest-prefix behaviour for callers that pass raw right tagger.
	lower := strings.ToLower(surface)
	for _, prefix := range NoDashPrefixList() {
		if !strings.HasPrefix(lower, prefix) || len(lower) <= len(prefix) {
			continue
		}
		right := surface[lenPrefixBytes(surface, prefix):]
		if right == "" {
			continue
		}
		r, _ := utf8.DecodeRuneInString(right)
		if !unicode.IsLetter(r) {
			continue
		}
		// If right starts with ' strip for lookup
		lookup := right
		apo := ""
		if strings.HasPrefix(right, "'") || strings.HasPrefix(right, "’") {
			rr := []rune(right)
			lookup = string(rr[1:])
			apo = "'"
		}
		heads := tagRight(lookup)
		if len(heads) == 0 {
			continue
		}
		var out []*languagetool.AnalyzedToken
		for _, h := range heads {
			if h == nil || h.GetPOSTag() == nil {
				continue
			}
			pos := *h.GetPOSTag()
			if !isPrefixNoDashPOS(pos) {
				continue
			}
			if strings.HasPrefix(pos, "noun:inanim") && strings.Contains(pos, "v_kly") {
				continue
			}
			lemma := prefix + apo
			if h.GetLemma() != nil {
				lemma += *h.GetLemma()
			}
			// apo / bad heuristic
			rr := []rune(lookup)
			prr := []rune(prefix)
			if len(rr) > 0 && len(prr) > 0 {
				apoNeeded := strings.ContainsRune(noDashUkYod, rr[0]) && !strings.ContainsRune(noDashUkVowels, prr[len(prr)-1])
				if apoNeeded == (apo == "") {
					pos = AddIfNotContains(pos, ":bad")
				}
			}
			p, l := pos, lemma
			out = append(out, languagetool.NewAnalyzedToken(surface, &p, &l))
		}
		if len(out) > 0 {
			return out
		}
	}
	return nil
}
