package en

import (
	"strings"

	atticmorfo "github.com/lucasew/lang/internal/attic/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// CommonDemoSpellerSuggestions is a soft map of frequent EN typos → fixes.
// Used with binary and map spellers for suggestion UX until full Morfologik
// suggestion generation is wired.
var CommonDemoSpellerSuggestions = map[string][]string{
	"teh":      {"the"},
	"recieve":  {"receive"},
	"seperate": {"separate"},
	"occured":  {"occurred"},
	"definately": {"definitely"},
	"accomodate": {"accommodate"},
	"untill":   {"until"},
	"wich":     {"which"},
	"thier":    {"their"},
}

// RegisterBinaryEnglishSpeller installs MORFOLOGIK_RULE_EN_US backed by a CFSA2
// en_US.dict (attic morfologik loader). Returns false if the dictionary cannot be opened.
// nearestKnown is an optional small word set for edit-distance suggestions (not the full dict).
// suggestions may be nil (uses CommonDemoSpellerSuggestions).
func RegisterBinaryEnglishSpeller(lt *languagetool.JLanguageTool, dictPath string, nearestKnown map[string]struct{}, suggestions map[string][]string) bool {
	if lt == nil || dictPath == "" {
		return false
	}
	d, err := atticmorfo.OpenDictionary(dictPath)
	if err != nil || d == nil {
		return false
	}
	isKnown := func(w string) bool {
		if d.Contains(w) {
			return true
		}
		low := strings.ToLower(w)
		if low != w && d.Contains(low) {
			return true
		}
		return false
	}
	if suggestions == nil {
		suggestions = CommonDemoSpellerSuggestions
	}
	// nearestKnown only (full CFSA2 dict is too large for edit-distance scan)
	lt.AddRuleChecker("MORFOLOGIK_RULE_EN_US", languagetool.SimplePredicateSpellerChecker(
		"MORFOLOGIK_RULE_EN_US", isKnown, suggestions, nearestKnown,
	))
	return true
}
