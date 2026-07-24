package de

import (
	"regexp"
	"strings"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// prefixesNonSeparableVerbs ports GermanTagger.prefixesNonSeparableVerbs.
var prefixesNonSeparableVerbs = []string{"be", "emp", "ent", "er", "hinter", "miss", "un", "ver", "zer"}

var rePrefixesNonSeparable = regexp.MustCompile(`^(be|emp|ent|er|hinter|miss|un|ver|zer)`)

// notAVerb ports GermanTagger.notAVerb (false-positive guards for prefix stripping).
var notAVerb = map[string]struct{}{
	"angebot": {}, "anteil": {}, "aufenthalt": {}, "ausdruck": {}, "auswärtsspiel": {},
	"beispiel": {}, "bereich": {}, "besondere": {}, "daring": {}, "einfach": {}, "einfachst": {},
	"endkasten": {}, "freibetrag": {}, "grautöne": {}, "grüntöne": {}, "großherzöge": {},
	"großteil": {}, "hochhaus": {}, "klarerweise": {}, "maßnahme": {}, "mitglieder": {},
	"nachricht": {}, "nebenfach": {}, "niederlage": {}, "nothing": {}, "notscheid": {},
	"preisver": {}, "reinweiß": {}, "schwarzweiß": {}, "schwarzgrau": {}, "schwarzgrün": {},
	"schwarztöne": {}, "unbesiegt": {}, "unmenge": {}, "unrat": {}, "unver": {},
	"verrückterweise": {}, "versonnen": {}, "vorlieb": {}, "vorteil": {}, "warmweiß": {},
	"wohldefiniert": {}, "wohlergehen": {}, "wohlgemerkt": {}, "zuende": {}, "zuhause": {},
	"zumal": {}, "zuver": {}, "darauf": {}, "einmal": {}, "kleinkram": {}, "hochsicher": {},
	"ehering": {}, "freitag": {}, "großmeister": {}, "handwerk": {}, "herpes": {}, "nachfolger": {},
}

// elative prefixes for unknown-word last-part tagging (Java bitter|dunkel|…).
var reElativePrefix = regexp.MustCompile(`^(bitter|dunkel|erz|extra|früh|gemein|grund|hyper|lau|mega|minder|stock|super|tod|ultra|u[nr]|voll)`)

// startsWithAnyPrefix reports whether s starts with any of prefixes (longest-first for multi).
func startsWithAnyPrefix(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

func isNotAVerb(word string) bool {
	_, ok := notAVerb[strings.ToLower(word)]
	return ok
}

// startsWithNotAVerb: Java startsWithAny(word.toLowerCase(), notAVerb) — prefix of full word.
func startsWithNotAVerb(wordLower string) bool {
	for n := range notAVerb {
		if strings.HasPrefix(wordLower, n) {
			return true
		}
	}
	return false
}

// isTitleOrLower ports Java:
//
//	word.equals(word.substring(0,1).toUpperCase()+word.substring(1).toLowerCase())
//	  || word.equals(word.toLowerCase())
//
// substring uses UTF-16 units (Java String).
func isTitleOrLower(word string) bool {
	if word == "" {
		return false
	}
	if word == strings.ToLower(word) {
		return true
	}
	return word == utf16FirstUpperRestLower(word)
}

// isFirstCharLowerRestUnchanged ports Java:
//
//	word.equals(word.substring(0, 1).toLowerCase() + word.substring(1))
func isFirstCharLowerRestUnchanged(word string) bool {
	if word == "" {
		return false
	}
	return word == utf16FirstLowerRest(word)
}

func utf16FirstUpperRestLower(word string) string {
	u := utf16.Encode([]rune(word))
	if len(u) == 0 {
		return word
	}
	first := string(utf16.Decode(u[:1]))
	rest := string(utf16.Decode(u[1:]))
	return strings.ToUpper(first) + strings.ToLower(rest)
}

func utf16FirstLowerRest(word string) string {
	u := utf16.Encode([]rune(word))
	if len(u) == 0 {
		return word
	}
	first := string(utf16.Decode(u[:1]))
	rest := string(utf16.Decode(u[1:]))
	return strings.ToLower(first) + rest
}

// addImpPraesSFTMutual ports the known-word IMP:SIN:SFT ↔ 1:SIN:PRÄ:SFT mutual tags.
func (t *GermanTagger) addImpPraesSFTMutual(word string, sentenceTokens []string, idxPos int, readings []*languagetool.AnalyzedToken) []*languagetool.AnalyzedToken {
	if t == nil || word == "" {
		return readings
	}
	low := strings.ToLower(word)
	// skip if starts with separable prefix list or notAVerb
	if startsWithAnyPrefix(low, prefixesSeparableVerbsLongestList) {
		return readings
	}
	if startsWithNotAVerb(low) {
		return readings
	}
	if !isTitleOrLower(word) {
		return readings
	}
	lstPrt := low
	frstPrt := ""
	if startsWithAnyPrefix(low, prefixesNonSeparableVerbs) {
		// remove non-separable prefix
		lstPrt = rePrefixesNonSeparable.ReplaceAllString(low, "")
		if lstPrt == low || lstPrt == "" {
			return readings
		}
		// first part = original prefix casing stripped from end of surface
		frstPrt = strings.TrimSuffix(word, lstPrt)
		// if surface was Title "Verzeih", lstPrt lower "zeih" — TrimSuffix may fail on case
		if frstPrt == word {
			// strip by length of prefix match on lower
			pref := low[:len(low)-len(lstPrt)]
			if len(word) >= len(pref) {
				frstPrt = word[:len(pref)]
			}
		}
	}
	if lstPrt == "gar" || lstPrt == "mal" || lstPrt == "null" || lstPrt == "trotz" {
		return readings
	}
	// Java: sentenceTokens.indexOf(word)==0
	//   || word.equals(word.substring(0,1).toLowerCase()+word.substring(1))
	// First-char lower (rest unchanged), NOT full toLowerCase.
	atStart := idxPos == 0
	if !atStart && !isFirstCharLowerRestUnchanged(word) {
		return readings
	}
	// avoid duplicates
	has := func(sub string) bool {
		for _, r := range readings {
			if r != nil && r.GetPOSTag() != nil && strings.Contains(*r.GetPOSTag(), sub) {
				return true
			}
		}
		return false
	}
	for _, v := range t.TagWordExact(lstPrt) {
		if strings.HasPrefix(v.PosTag, "VER:IMP:SIN:SFT") && !has("VER:1:SIN:PRÄ:SFT") {
			lemma := strings.ToLower(frstPrt) + v.Lemma
			readings = append(readings, toToken(word, tagging.NewTaggedWord(lemma, "VER:1:SIN:PRÄ:SFT")))
		}
		if strings.HasPrefix(v.PosTag, "VER:1:SIN:PRÄ:SFT") && !has("VER:IMP:SIN:SFT") {
			lemma := strings.ToLower(frstPrt) + v.Lemma
			readings = append(readings, toToken(word, tagging.NewTaggedWord(lemma, "VER:IMP:SIN:SFT")))
		}
	}
	return readings
}

// prefixesSeparableVerbsLongest is an alias for the generated longest-first list.
func prefixesSeparableVerbsLongest() []string {
	return prefixesSeparableVerbsLongestList
}

// tagElativeUnknown ports elative intensifier strip for unknown words
// (bitter|dunkel|… + lastPart tagged).
func (t *GermanTagger) tagElativeUnknown(word string) []*languagetool.AnalyzedToken {
	if t == nil || word == "" {
		return nil
	}
	// Java startsWithAny on original then removePattern with grund|ur|un|voll included
	if !reElativeStart.MatchString(word) {
		return nil
	}
	lastPart := reElativePrefix.ReplaceAllString(word, "")
	if tagging.UTF16Len(lastPart) <= 3 {
		return nil
	}
	firstPart := strings.TrimSuffix(word, lastPart)
	var readings []*languagetool.AnalyzedToken
	for _, tw := range t.TagWordExact(lastPart) {
		if tagging.UTF16Len(firstPart) == 2 && strings.HasPrefix(tw.PosTag, "VER") {
			continue
		}
		lemma := firstPart + tw.Lemma
		if tw.Lemma == "" {
			lemma = firstPart + lastPart
		}
		readings = append(readings, toToken(word, tagging.NewTaggedWord(lemma, tw.PosTag)))
	}
	return readings
}

// reElativeStart: word starts with one of the Java start prefixes (before removePattern).
var reElativeStart = regexp.MustCompile(`^(bitter|dunkel|erz|extra|früh|gemein|hyper|lau|mega|minder|stock|super|tod|ultra|un|ur)`)
