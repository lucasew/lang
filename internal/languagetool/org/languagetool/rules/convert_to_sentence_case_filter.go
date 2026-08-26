package rules

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// ConvertToSentenceCaseFilter ports org.languagetool.rules.ConvertToSentenceCaseFilter.
//
// Java protected tokenIsException is overridden in language subclasses (e.g. EN "me")
// via TokenIsException; nil means default false.
type ConvertToSentenceCaseFilter struct {
	// TokenIsException ports protected boolean tokenIsException(String s).
	// Default (nil) returns false like the Java base class.
	TokenIsException func(s string) bool
}

func NewConvertToSentenceCaseFilter() *ConvertToSentenceCaseFilter {
	return &ConvertToSentenceCaseFilter{}
}

// AcceptRuleMatch ports ConvertToSentenceCaseFilter.acceptRuleMatch.
// Null match / null patternTokens element → panic (Java NPE).
func (f *ConvertToSentenceCaseFilter) AcceptRuleMatch(match *RuleMatch, _ map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, _ []int) *RuleMatch {
	// Java: RuleMatch ruleMatch = match;
	ruleMatch := match
	firstDone := false
	var replacement strings.Builder
	var originalStr strings.Builder
	// Java: for (int i = patternTokenPos; i < patternTokens.length; i++)
	for i := patternTokenPos; i < len(patternTokens); i++ {
		// Java: patternTokens[i].getStartPos() / getEndPos() — null element → NPE
		tok := patternTokens[i]
		if tok.GetStartPos() < match.GetFromPos() || tok.GetEndPos() > match.GetToPos() {
			continue
		}
		normalizedCase := f.normalizedCase(tok)
		// Java: if (i + 1 < patternTokens.length && patternTokens[i + 1].getToken().equals("."))
		if i+1 < len(patternTokens) && patternTokens[i+1].GetToken() == "." {
			// Java: normalizedCase.length() == 1 (UTF-16 code units)
			if utf16Len(normalizedCase) == 1 {
				normalizedCase = strings.ToUpper(normalizedCase)
			} else if normalizedCase == "corp" {
				normalizedCase = "Corp"
			}
		}
		tokenString := tok.GetToken()
		tokenCapitalized := tools.UppercaseFirstChar(normalizedCase)
		// Java: if (!firstDone & !isPunctuation(tokenString) && !tokenString.isEmpty())
		if !firstDone && !isPunctuationCTS(tokenString) && tokenString != "" {
			firstDone = true
			replacement.WriteString(tokenCapitalized)
			originalStr.WriteString(tokenString)
		} else {
			if tok.IsWhitespaceBefore() {
				replacement.WriteString(" ")
				originalStr.WriteString(" ")
			}
			replacement.WriteString(normalizedCase)
			originalStr.WriteString(tokenString)
		}
	}
	if replacement.String() == originalStr.String() {
		return nil
	}
	ruleMatch.SetSuggestedReplacement(replacement.String())
	return ruleMatch
}

// isPunctuationCTS ports ConvertToSentenceCaseFilter.isPunctuation:
// Pattern.matches("\\p{IsPunctuation}", s) — entire string is one punctuation char.
func isPunctuationCTS(s string) bool {
	if s == "" {
		return false
	}
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError || size != len(s) {
		return false
	}
	return unicode.IsPunct(r)
}

// normalizedCase ports ConvertToSentenceCaseFilter.normalizedCase.
func (f *ConvertToSentenceCaseFilter) normalizedCase(atr *languagetool.AnalyzedTokenReadings) string {
	tokenLowercase := strings.ToLower(atr.GetToken())
	if atr.HasTypographicApostrophe() {
		tokenLowercase = strings.ReplaceAll(tokenLowercase, "'", "’")
	}
	if f.tokenIsException(tokenLowercase) {
		// exception: the lemma is "I"
		return tokenLowercase
	}
	tokenCapitalized := tools.UppercaseFirstChar(tokenLowercase)
	lemmaIsCapitalized := false
	lemmaIsLowercase := false
	// Java: for (AnalyzedToken at : atr)
	for _, at := range atr.GetReadings() {
		// null reading → NPE on hasNoTag (no soft skip)
		if at.HasNoTag() || at.GetLemma() == nil {
			return tokenCapitalized
		}
		// for multi-words lemmas, take the first word
		lemma := strings.Split(*at.GetLemma(), " ")[0]
		lemmaIsCapitalized = lemmaIsCapitalized || tools.IsCapitalizedWord(lemma)
		// Java: lemmaIsLowercase = lemmaIsLowercase || !StringTools.isNotAllLowercase(lemma);
		lemmaIsLowercase = lemmaIsLowercase || !tools.IsNotAllLowercase(lemma)
	}
	if lemmaIsLowercase {
		return tokenLowercase
	}
	if lemmaIsCapitalized {
		return tokenCapitalized
	}
	return atr.GetToken()
}

// tokenIsException ports protected boolean tokenIsException(String s) default false.
func (f *ConvertToSentenceCaseFilter) tokenIsException(s string) bool {
	if f.TokenIsException != nil {
		return f.TokenIsException(s)
	}
	return false
}
