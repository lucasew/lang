package rules

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// ConvertToSentenceCaseFilter ports org.languagetool.rules.ConvertToSentenceCaseFilter.
type ConvertToSentenceCaseFilter struct {
	// TokenIsException returns true for tokens that stay lower (e.g. EN "me").
	TokenIsException func(s string) bool
}

func NewConvertToSentenceCaseFilter() *ConvertToSentenceCaseFilter {
	return &ConvertToSentenceCaseFilter{}
}

// SentenceCaseToken is one token inside the match span (unit-test / Suggest path).
type SentenceCaseToken struct {
	Token            string
	WhitespaceBefore bool
	// LemmaCase: "lower", "capitalized", "upper", or "" (no lemma/tag → capitalize).
	LemmaCase string
	// HasTypographicApostrophe maps ' → ’ in normalized form.
	HasTypographicApostrophe bool
}

// AcceptRuleMatch ports ConvertToSentenceCaseFilter.acceptRuleMatch.
// patternTokens is the match slice (PatternRuleMatcher copyOfRange); only tokens
// fully inside the match span contribute. Returns nil when suggestion equals original.
func (f *ConvertToSentenceCaseFilter) AcceptRuleMatch(match *RuleMatch, _ map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, _ []int) *RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	var sc []SentenceCaseToken
	for _, atr := range patternTokens {
		if atr == nil {
			continue
		}
		// Java: skip tokens outside match span
		if atr.GetStartPos() < match.GetFromPos() || atr.GetEndPos() > match.GetToPos() {
			continue
		}
		sc = append(sc, sentenceCaseTokenFromATR(atr))
	}
	sug := f.Suggest(sc)
	if sug == "" {
		return nil
	}
	match.SetSuggestedReplacement(sug)
	return match
}

func sentenceCaseTokenFromATR(atr *languagetool.AnalyzedTokenReadings) SentenceCaseToken {
	return SentenceCaseToken{
		Token:                    atr.GetToken(),
		WhitespaceBefore:         atr.IsWhitespaceBefore(),
		LemmaCase:                lemmaCaseFromReadings(atr),
		HasTypographicApostrophe: atr.HasTypographicApostrophe(),
	}
}

// lemmaCaseFromReadings ports ConvertToSentenceCaseFilter.normalizedCase lemma flags.
// Returns "lower", "capitalized", "upper" (keep surface), or "" (no tag → capitalize).
func lemmaCaseFromReadings(atr *languagetool.AnalyzedTokenReadings) string {
	lemmaIsCapitalized := false
	lemmaIsLowercase := false
	for _, at := range atr.GetReadings() {
		if at == nil {
			continue
		}
		// Java: hasNoTag || lemma == null → treat as capitalize
		if at.HasNoTag() || at.GetLemma() == nil {
			return ""
		}
		// multi-word lemmas: first word only
		lemma := strings.Split(*at.GetLemma(), " ")[0]
		lemmaIsCapitalized = lemmaIsCapitalized || tools.IsCapitalizedWord(lemma)
		// Java: lemmaIsLowercase ||= !isNotAllLowercase(lemma)  ⇒ all lowercase
		lemmaIsLowercase = lemmaIsLowercase || !tools.IsNotAllLowercase(lemma)
	}
	if lemmaIsLowercase {
		return "lower"
	}
	if lemmaIsCapitalized {
		return "capitalized"
	}
	// neither → keep original surface in normalizedCase default branch
	return "upper"
}

// Suggest builds a sentence-case replacement for tokens fully inside the match.
// Returns "" when the suggestion equals the original (match should be suppressed).
func (f *ConvertToSentenceCaseFilter) Suggest(tokens []SentenceCaseToken) string {
	firstDone := false
	var replacement, original strings.Builder
	for i, tok := range tokens {
		normalized := f.normalizedCase(tok)
		// single-letter before "." → upper; "corp." → "Corp"
		// Java: normalizedCase.length() == 1 (UTF-16 code units)
		if i+1 < len(tokens) && tokens[i+1].Token == "." {
			if utf16Len(normalized) == 1 {
				normalized = strings.ToUpper(normalized)
			} else if normalized == "corp" {
				normalized = "Corp"
			}
		}
		tokenString := tok.Token
		if !firstDone && !isPunctuationToken(tokenString) && tokenString != "" {
			firstDone = true
			replacement.WriteString(tools.UppercaseFirstChar(normalized))
			original.WriteString(tokenString)
		} else {
			if tok.WhitespaceBefore {
				replacement.WriteByte(' ')
				original.WriteByte(' ')
			}
			replacement.WriteString(normalized)
			original.WriteString(tokenString)
		}
	}
	if replacement.String() == original.String() {
		return ""
	}
	return replacement.String()
}

func (f *ConvertToSentenceCaseFilter) normalizedCase(atr SentenceCaseToken) string {
	tokenLower := strings.ToLower(atr.Token)
	if atr.HasTypographicApostrophe {
		tokenLower = strings.ReplaceAll(tokenLower, "'", "’")
	}
	if f.TokenIsException != nil && f.TokenIsException(tokenLower) {
		return tokenLower
	}
	tokenCap := tools.UppercaseFirstChar(tokenLower)
	switch atr.LemmaCase {
	case "lower":
		return tokenLower
	case "capitalized":
		return tokenCap
	case "upper":
		return atr.Token
	case "unknown", "":
		// Java: no tag / null lemma → capitalized
		return tokenCap
	default:
		return atr.Token
	}
}

// isPunctuationToken ports Java Pattern.matches("\\p{IsPunctuation}", s)
// — entire string is a single punctuation character.
func isPunctuationToken(s string) bool {
	if s == "" {
		return false
	}
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError || size != len(s) {
		return false
	}
	return unicode.IsPunct(r)
}
