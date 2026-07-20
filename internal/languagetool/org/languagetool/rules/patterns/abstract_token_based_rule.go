package patterns

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// TokenHint ports AbstractTokenBasedRule.TokenHint — possible form/lemma values for fast skip.
type TokenHint struct {
	Inflected       bool
	LowerCaseValues []string
	TokenIndex      int
}

func NewTokenHint(inflected bool, possibleValues []string, tokenIndex int) TokenHint {
	seen := map[string]struct{}{}
	var vals []string
	for _, s := range possibleValues {
		low := tools.Intern(strings.ToLower(s))
		if _, ok := seen[low]; ok {
			continue
		}
		seen[low] = struct{}{}
		vals = append(vals, low)
	}
	return TokenHint{Inflected: inflected, LowerCaseValues: vals, TokenIndex: tokenIndex}
}

// CanBeIgnoredFor returns true if none of the hint values appear in the sentence tokens/lemmas.
// Java uses AnalyzedSentence.getLemmaOffsets for inflected hints (tagger lemmas only).
func (th TokenHint) CanBeIgnoredFor(sentence *languagetool.AnalyzedSentence) bool {
	if sentence == nil || len(th.LowerCaseValues) == 0 {
		return false
	}
	want := map[string]struct{}{}
	for _, v := range th.LowerCaseValues {
		want[v] = struct{}{}
	}
	for _, tok := range sentence.GetTokensWithoutWhitespace() {
		if th.Inflected {
			for _, r := range tok.GetReadings() {
				if lem := r.GetLemma(); lem != nil {
					if _, ok := want[strings.ToLower(*lem)]; ok {
						return false
					}
				}
			}
			// Surface equals a listed lemma form (no soft morphological invent).
			if _, ok := want[strings.ToLower(tok.GetToken())]; ok {
				return false
			}
		} else {
			if _, ok := want[strings.ToLower(tok.GetToken())]; ok {
				return false
			}
		}
	}
	return true
}

// AbstractTokenBasedRule ports performance-hint fields of AbstractTokenBasedRule.
type AbstractTokenBasedRule struct {
	*PatternRule
	TokenHints    []TokenHint
	AnchorHint    *TokenHint
	MinTokenCount int
}

func NewAbstractTokenBasedRule(id, description, languageCode string, patternTokens []*PatternToken) *AbstractTokenBasedRule {
	pr := NewPatternRule(id, languageCode, patternTokens, description, "", "")
	r := &AbstractTokenBasedRule{PatternRule: pr}
	r.computeHints(patternTokens)
	return r
}

func (r *AbstractTokenBasedRule) computeHints(patternTokens []*PatternToken) {
	// Java AbstractTokenBasedRule constructor: minTokenCount + tokenHints from calcFormHints/calcLemmaHints.
	minCount := 0
	if len(patternTokens) > 0 && !canMatchSentenceStart(patternTokens[0]) {
		minCount = 1
	}
	var hints []TokenHint
	fixedOffset := true
	var anchor *TokenHint
	for i, token := range patternTokens {
		if token == nil {
			continue
		}
		if token.MinOccurrence > 0 {
			minCount++
		}
		// Java PatternToken.calcFormHints: null when negation || !hasStringThatMustMatch
		// hasStringThatMustMatch: !ref && !mayBeOmitted (min=0) && non-empty string.
		if hasStringThatMustMatch(token) && !token.Regexp && !token.Negation {
			h := NewTokenHint(token.MatchInflected, []string{token.Token}, i)
			hints = append(hints, h)
			if fixedOffset && anchor == nil {
				hh := h
				anchor = &hh
			}
		}
		if fixedOffset && (token.MinOccurrence != 1 || token.SkipNext != 0 || token.MaxOccurrence != 1) {
			fixedOffset = false
		}
	}
	r.TokenHints = hints
	r.AnchorHint = anchor
	if minCount > 127 {
		minCount = 127
	}
	r.MinTokenCount = minCount
}

// hasStringThatMustMatch ports PatternToken.hasStringThatMustMatch.
func hasStringThatMustMatch(token *PatternToken) bool {
	if token == nil {
		return false
	}
	// !isReferenceElement && !MAY_BE_OMITTED && !getString().isEmpty()
	if token.IsReferenceElement() {
		return false
	}
	if token.MinOccurrence == 0 {
		return false
	}
	return token.Token != ""
}

func canMatchSentenceStart(token *PatternToken) bool {
	if token == nil {
		return true
	}
	// Java: isSentenceStart() || getNegation() || !hasStringThatMustMatch()
	if token.Negation || !hasStringThatMustMatch(token) {
		return true
	}
	if token.Pos != nil && token.Pos.PosTag == languagetool.SentenceStartTagName && !token.Pos.Negate {
		return true
	}
	return false
}

// CanBeIgnoredFor ports AbstractTokenBasedRule.canBeIgnoredFor.
func (r *AbstractTokenBasedRule) CanBeIgnoredFor(sentence *languagetool.AnalyzedSentence) bool {
	if sentence == nil {
		return true
	}
	// Java compares getTokensWithoutWhitespace().length (includes SENT_START).
	if len(sentence.GetTokensWithoutWhitespace()) < r.MinTokenCount {
		return true
	}
	for _, th := range r.TokenHints {
		if th.CanBeIgnoredFor(sentence) {
			return true
		}
	}
	return false
}
