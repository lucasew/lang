package uk

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// ukrLettersOnly approximates Java UKR_LETTERS_PATTERN [А-ЯІЇЄҐа-яіїєґ'-]+
func ukrLettersOnly(word string) bool {
	if word == "" {
		return false
	}
	for _, r := range word {
		switch {
		case r == '\'' || r == '-' || r == '’':
			continue
		case unicode.In(r, unicode.Cyrillic):
			continue
		default:
			return false
		}
	}
	return true
}

// GuessOtherTagsReadings ports CompoundTagger.guessOtherTagsInternal ending
// paradigms for capitalized proper names (no invent dictionary of names):
//
//	*штрассе / *штрасе → noun:inanim:f:…:nv:prop[:alt]
//	*дзе / *швілі / *іані → noun:inanim:m|f:…:nv:prop:lname
//
// No-dash prefix compounds stay in TryNoDashPrefixTags (already dict-gated).
func GuessOtherTagsReadings(token string) []*languagetool.AnalyzedToken {
	if token == "" || tagging.UTF16Len(token) <= 7 {
		return nil
	}
	if !ukrLettersOnly(token) {
		return nil
	}
	// Java: StringTools.isCapitalizedWord(word)
	if !tools.IsCapitalizedWord(token) {
		return nil
	}
	low := strings.ToLower(token)
	if strings.HasSuffix(low, "штрассе") {
		return generateTokensForNv(token, "f", ":prop:alt")
	}
	if strings.HasSuffix(low, "штрасе") {
		return generateTokensForNv(token, "f", ":prop")
	}
	if strings.HasSuffix(low, "дзе") || strings.HasSuffix(low, "швілі") || strings.HasSuffix(low, "іані") {
		return generateTokensForNv(token, "mf", ":prop:lname")
	}
	return nil
}
