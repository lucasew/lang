package rules

import (
	"regexp"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// AddCommasFilter ports org.languagetool.rules.AddCommasFilter suggestion logic.
type AddCommasFilter struct{}

func NewAddCommasFilter() *AddCommasFilter {
	return &AddCommasFilter{}
}

var openingQuotesRE = regexp.MustCompile(`[«“"‘'„¿¡]`)

// CommaContext holds precomputed neighborhood flags for Accept/Suggest.
type CommaContext struct {
	// MatchedText is sentence[fromPos:toPos].
	MatchedText string
	// FirstToken / LastToken of the match span (without whitespace).
	FirstToken, LastToken string
	// TokenBefore is the token immediately before the match ("" if none / sentence start).
	TokenBefore string
	// TokenAfter is the token immediately after the match ("" if none).
	TokenAfter string
	// AfterHasWhitespaceBefore: next token has whitespace before it (for quote check).
	AfterHasWhitespaceBefore bool
	// SuggestSemicolon enables "; phrase," and ", phrase," suggestions when prev is ",".
	SuggestSemicolon bool
	// MatchAtSentenceStart: pattern starts at first content token.
	MatchAtSentenceStart bool
}

// Accept returns false (suppress match) when both sides already look punctuated/capitalized.
func (f *AddCommasFilter) Accept(ctx CommaContext) bool {
	beforeOK, afterOK := f.sidesOK(ctx)
	return !(beforeOK && afterOK)
}

// Suggest returns replacement strings. Empty means suppress the match.
func (f *AddCommasFilter) Suggest(ctx CommaContext) []string {
	beforeOK, afterOK := f.sidesOK(ctx)
	if beforeOK && afterOK {
		return nil
	}
	if ctx.SuggestSemicolon && ctx.TokenBefore == "," && !afterOK {
		return []string{
			"; " + ctx.MatchedText + ",",
			", " + ctx.MatchedText + ",",
		}
	}
	if beforeOK && !afterOK {
		return []string{ctx.LastToken + ","}
	}
	if !beforeOK && afterOK {
		return []string{", " + ctx.FirstToken}
	}
	return []string{", " + ctx.MatchedText + ","}
}

func (f *AddCommasFilter) sidesOK(ctx CommaContext) (beforeOK, afterOK bool) {
	beforeOK = ctx.MatchAtSentenceStart || ctx.TokenBefore == "" ||
		isPunctuationOrSymbol(ctx.TokenBefore) || tools.IsCapitalizedWord(ctx.FirstToken)
	afterOK = ctx.TokenAfter != "" && isPunctuationOrSymbol(ctx.TokenAfter) &&
		!(ctx.AfterHasWhitespaceBefore && openingQuotesRE.MatchString(ctx.TokenAfter))
	return beforeOK, afterOK
}

// isPunctuationOrSymbol approximates StringTools.isPunctuationOrSymbol (\p{P}|\p{S}).
func isPunctuationOrSymbol(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsPunct(r) && !unicode.IsSymbol(r) {
			return false
		}
	}
	return true
}
