package multitoken

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// MultitokenSpellerFilter ports
// org.languagetool.rules.spelling.multitoken.MultitokenSpellerFilter.
//
// Java obtains Language from ((PatternRule) match.getRule()).getLanguage() and then
// uses Language APIs (getShortCode, getDefaultLanguageVariant, getMultitokenSpeller,
// getDefaultSpellingRule, getWordTokenizer). Go wires those Language APIs as fields/hooks
// on this leaf (no full Language composite required for the filter sector).
type MultitokenSpellerFilter struct {
	// Speller ports Language.getMultitokenSpeller() (after optional default-variant resolve).
	Speller *MultitokenSpeller
	// ShortCode ports Language.getShortCode() for the en/de/pt/nl areTokensAcceptedBySpeller gate.
	ShortCode string
	// ShortCodeWithCountryAndVariant ports Language.getShortCodeWithCountryAndVariant().
	// When UTF-16 length == 2 inside the en/de/pt/nl gate, Java reassigns
	// lang = lang.getDefaultLanguageVariant() before isMisspelled + getMultitokenSpeller.
	ShortCodeWithCountryAndVariant string
	// GetDefaultLanguageVariant ports Language.getDefaultLanguageVariant().
	// Called only when short code is en/de/pt/nl and ShortCodeWithCountryAndVariant length is 2.
	// Returned hooks replace Speller/IsMisspelled/Tokenize for isMisspelled + getSuggestions
	// (Java mutates the local lang reference).
	GetDefaultLanguageVariant func() *MultitokenSpellerLanguageHooks
	// IsMisspelled ports SpellingCheckRule.isMisspelled (Java language.getDefaultSpellingRule()).
	// Nil → null SpellingCheckRule → isMisspelled returns false.
	IsMisspelled func(token string) bool
	// Tokenize ports language.getWordTokenizer().tokenize; nil → WordTokenizer.
	Tokenize func(s string) []string
}

// MultitokenSpellerLanguageHooks is the Language surface re-bound when
// getDefaultLanguageVariant() is used (bare short codes in testing).
type MultitokenSpellerLanguageHooks struct {
	Speller      *MultitokenSpeller
	IsMisspelled func(token string) bool
	Tokenize     func(s string) []string
}

// AcceptRuleMatch ports MultitokenSpellerFilter.acceptRuleMatch.
func (f *MultitokenSpellerFilter) AcceptRuleMatch(
	match *rules.RuleMatch,
	_ map[string]string,
	patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings,
	_ []int,
) *rules.RuleMatch {
	if f == nil || match == nil {
		return match
	}
	// Java: if (Arrays.stream(patternTokens).allMatch(x -> x.isIgnoredBySpeller())) return null;
	// Stream.allMatch on empty stream is true.
	// nil patternTokens is a Go convenience (tests without tokens) — skip the gate
	// (Java always passes the matched array from the pattern engine; null would NPE).
	if patternTokens != nil && allIgnoredBySpeller(patternTokens) {
		return nil
	}

	// Java: String underlinedError = match.getOriginalErrorStr();
	underlinedError := match.GetOriginalErrorStr()

	// Language local (may be reassigned to default variant inside en/de/pt/nl gate).
	sp := f.Speller
	isMiss := f.IsMisspelled
	tokenize := f.Tokenize

	// Java:
	//   boolean areTokensAcceptedBySpeller = false;
	//   if (lang.getShortCode().equals("en") || ... "de" || "pt" || "nl") {
	//     if (lang.getShortCodeWithCountryAndVariant().length()==2) {
	//       lang = lang.getDefaultLanguageVariant();
	//     }
	//     areTokensAcceptedBySpeller = !isMisspelled(underlinedError, lang);
	//   }
	//   List<String> replacements = lang.getMultitokenSpeller().getSuggestions(...);
	acceptedBySpeller := false
	if isSpellerAcceptedShortCode(f.ShortCode) {
		scv := f.ShortCodeWithCountryAndVariant
		if scv == "" {
			scv = f.ShortCode
		}
		if utf16Len(scv) == 2 && f.GetDefaultLanguageVariant != nil {
			if v := f.GetDefaultLanguageVariant(); v != nil {
				sp = v.Speller
				isMiss = v.IsMisspelled
				tokenize = v.Tokenize
			}
		}
		acceptedBySpeller = !isMisspelledWith(underlinedError, isMiss, tokenize)
	}
	if sp == nil {
		// Java: lang.getMultitokenSpeller() would NPE; fail-closed like empty suggestions.
		return nil
	}
	replacements := sp.GetSuggestionsOpts(underlinedError, acceptedBySpeller)
	if len(replacements) == 0 {
		return nil
	}

	// Java: underlinedError.length()>4 && StringTools.isAllUppercase — UTF-16 length
	if utf16Len(underlinedError) > 4 && tools.IsAllUppercase(underlinedError) {
		allUpper := make([]string, 0, len(replacements))
		for _, replacement := range replacements {
			// Java: replacement.toUpperCase() (default locale)
			newReplacement := strings.ToUpper(replacement)
			if !stringListContains(allUpper, newReplacement) && underlinedError != newReplacement {
				allUpper = append(allUpper, newReplacement)
			}
		}
		replacements = allUpper
	} else {
		// Capitalize suggestion at sentence start
		// Java: wordsStartPos = 1; skip punctuation / not-word; patternTokenPos==wordsStartPos
		if match.Sentence != nil && isPatternAtSentenceStart(match.Sentence, patternTokenPos) {
			capitalized := make([]string, 0, len(replacements))
			for _, replacement := range replacements {
				newReplacement := replacement
				// do not capitalize iPad (mixed case) — only when all-lower
				if replacement == strings.ToLower(replacement) {
					newReplacement = tools.UppercaseFirstChar(replacement)
				}
				if !stringListContains(capitalized, newReplacement) && underlinedError != newReplacement {
					capitalized = append(capitalized, newReplacement)
				}
			}
			replacements = capitalized
		}
	}
	if len(replacements) == 0 {
		return nil
	}
	match.SetSuggestedReplacements(replacements)
	return match
}

// isSpellerAcceptedShortCode ports the Java short-code gate for areTokensAcceptedBySpeller.
func isSpellerAcceptedShortCode(short string) bool {
	switch short {
	case "en", "de", "pt", "nl":
		return true
	default:
		return false
	}
}

// allIgnoredBySpeller ports Arrays.stream(patternTokens).allMatch(isIgnoredBySpeller).
// Empty stream → true (Java Stream.allMatch vacuous truth).
func allIgnoredBySpeller(tokens []*languagetool.AnalyzedTokenReadings) bool {
	for _, t := range tokens {
		if t == nil || !t.IsIgnoredBySpeller() {
			return false
		}
	}
	return true
}

// isPatternAtSentenceStart ports MultitokenSpellerFilter sentence-start walk:
// wordsStartPos skips leading punct/not-word after SENT_START; true when patternTokenPos == wordsStartPos.
func isPatternAtSentenceStart(sentence *languagetool.AnalyzedSentence, patternTokenPos int) bool {
	if sentence == nil {
		return false
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	// Java: int wordsStartPos = 1;
	wordsStartPos := 1
	for wordsStartPos < len(tokens) {
		tok := tokens[wordsStartPos]
		// Java does not null-check; defensive skip only
		if tok == nil {
			wordsStartPos++
			continue
		}
		t := tok.GetToken()
		if tools.IsPunctuationMark(t) || tools.IsNotWordString(t) {
			wordsStartPos++
			continue
		}
		break
	}
	return patternTokenPos == wordsStartPos
}

// isMisspelled ports MultitokenSpellerFilter.isMisspelled(String, Language) using
// this filter's spelling-rule and tokenizer hooks (Language DI).
//
//	if spellerRule == null → false
//	tokens = wordTokenizer.tokenize(s)
//	any token misspelled → true
func (f *MultitokenSpellerFilter) isMisspelled(s string) bool {
	if f == nil {
		return false
	}
	return isMisspelledWith(s, f.IsMisspelled, f.Tokenize)
}

func isMisspelledWith(s string, isMiss func(string) bool, tokenize func(string) []string) bool {
	// Java: if (spellerRule == null) return false;
	if isMiss == nil {
		return false
	}
	var tokens []string
	if tokenize != nil {
		tokens = tokenize(s)
	} else {
		tokens = tokenizers.NewWordTokenizer().Tokenize(s)
	}
	for _, token := range tokens {
		if isMiss(token) {
			return true
		}
	}
	return false
}

// stringListContains ports List.contains for ordered-dedupe of suggestion lists.
func stringListContains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
