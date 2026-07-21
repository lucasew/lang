package de

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/hunspell"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

// GermanSpellerMaxEditDistance ports GermanSpellerRule.MAX_EDIT_DISTANCE (2).
const GermanSpellerMaxEditDistance = 2

// GetSpellingFilePaths ports GermanSpellerRule.getSpellingFilePaths:
// CompoundAwareHunspellRule.getSpellingFilePaths(langCode) plus
// "/" + langCode + "/hunspell/spelling_recommendation.txt".
func GetSpellingFilePaths(langCode string) []string {
	if langCode == "" {
		langCode = "de"
	}
	paths := hunspell.SpellingFilePaths(langCode)
	return append(paths, "/"+langCode+"/hunspell/spelling_recommendation.txt")
}

// expandingPrepareLine ports GermanSpellerRule.ExpandingReader + LineExpander.expandLine
// for MorfologikMultiSpeller plain-text construction.
func expandingPrepareLine(line string) []string {
	return WireLineExpander().ExpandLine(line)
}

// GetSpeller ports GermanSpellerRule.getSpeller(language, userConfig, languageVariantPlainTextDict).
// Binary path: "/de/hunspell/de_" + country + ".dict" (DE/AT/CH).
// Returns nil when the binary dict is missing (Java returns null in LO/OO context).
//
// languageVariantPlainTextDict is e.g. "de/hunspell/spelling-de-AT.txt" (empty = none).
// userWords is UserDictWordsForMulti(accepted, premiumUID); free users pass nil.
//
// Java merges getSpellingFilePaths streams via ExpandingReader and a separate
// variant ExpandingReader; OpenMultiSpellerFromClasspathWithUser loads each
// plain path + variant through prepareLine (LineExpander twin).
func GetSpeller(countryCode, languageVariantPlainTextDict string, userWords []string) *morfologik.MorfologikMultiSpeller {
	country := strings.ToUpper(strings.TrimSpace(countryCode))
	if country == "" {
		country = "DE"
	}
	morfoFile := "/de/hunspell/de_" + country + ".dict"
	if morfologik.DiscoverLanguageDict(morfoFile) == "" {
		return nil
	}
	plain := GetSpellingFilePaths("de")
	variant := strings.TrimSpace(languageVariantPlainTextDict)
	return morfologik.OpenMultiSpellerFromClasspathWithUser(
		morfoFile, plain, variant, GermanSpellerMaxEditDistance, expandingPrepareLine, userWords)
}
