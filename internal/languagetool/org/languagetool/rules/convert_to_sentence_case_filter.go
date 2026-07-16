package rules

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// ConvertToSentenceCaseFilter ports org.languagetool.rules.ConvertToSentenceCaseFilter
// without a full tagger: tokens supply surface forms and optional lemma-case hints.
type ConvertToSentenceCaseFilter struct {
	// TokenIsException returns true for tokens that stay lower (e.g. pronoun "i").
	TokenIsException func(s string) bool
}

func NewConvertToSentenceCaseFilter() *ConvertToSentenceCaseFilter {
	return &ConvertToSentenceCaseFilter{}
}

// SentenceCaseToken is one token inside the match span.
type SentenceCaseToken struct {
	Token            string
	WhitespaceBefore bool
	// LemmaCase: "lower", "capitalized", or "" (unknown → capitalize by default).
	LemmaCase string
	// HasTypographicApostrophe maps ' → ’ in normalized form.
	HasTypographicApostrophe bool
}

// Suggest builds a sentence-case replacement for tokens fully inside the match.
// Returns "" when the suggestion equals the original (match should be suppressed).
func (f *ConvertToSentenceCaseFilter) Suggest(tokens []SentenceCaseToken) string {
	firstDone := false
	var replacement, original strings.Builder
	for i, tok := range tokens {
		normalized := f.normalizedCase(tok)
		// single-letter before "." → upper; "corp." → "Corp"
		if i+1 < len(tokens) && tokens[i+1].Token == "." {
			if len([]rune(normalized)) == 1 {
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
	case "unknown", "":
		return tokenCap
	default:
		return atr.Token
	}
}

func isPunctuationToken(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsPunct(r) {
			return false
		}
	}
	return true
}
