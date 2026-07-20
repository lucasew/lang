package patterns

import (
	"sort"
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

// CanBeIgnoredFor ports TokenHint.canBeIgnoredFor via getTokenOffsets / getLemmaOffsets.
func (th TokenHint) CanBeIgnoredFor(sentence *languagetool.AnalyzedSentence) bool {
	if sentence == nil || len(th.LowerCaseValues) == 0 {
		return false
	}
	for _, v := range th.LowerCaseValues {
		if th.getHintIndices(sentence, v) != nil {
			return false
		}
	}
	return true
}

// GetPossibleIndices ports TokenHint.getPossibleIndices — sorted non-blank indices
// where any hint value may appear (for anchor-based match starts).
func (th TokenHint) GetPossibleIndices(sentence *languagetool.AnalyzedSentence) []int {
	if sentence == nil || len(th.LowerCaseValues) == 0 {
		return nil
	}
	var result []int
	seen := map[int]struct{}{}
	for _, v := range th.LowerCaseValues {
		idxs := th.getHintIndices(sentence, v)
		for _, i := range idxs {
			if _, ok := seen[i]; ok {
				continue
			}
			seen[i] = struct{}{}
			result = append(result, i)
		}
	}
	if len(result) == 0 {
		return nil
	}
	sort.Ints(result)
	return result
}

// getHintIndices ports TokenHint.getHintIndices.
func (th TokenHint) getHintIndices(sentence *languagetool.AnalyzedSentence, hint string) []int {
	if sentence == nil {
		return nil
	}
	if th.Inflected {
		return sentence.GetLemmaOffsets(hint)
	}
	return sentence.GetTokenOffsets(hint)
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
	// Java AbstractTokenBasedRule constructor.
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
		// Java: form hints first; if null, lemma hints with inflected=true.
		inflected := false
		vals := token.CalcFormHints()
		if vals == nil {
			inflected = true
			vals = token.CalcLemmaHints()
		}
		if vals != nil {
			h := NewTokenHint(inflected, vals, i)
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
	// Java: sort by fewer values first, then longer min value length desc.
	sort.SliceStable(hints, func(i, j int) bool {
		if len(hints[i].LowerCaseValues) != len(hints[j].LowerCaseValues) {
			return len(hints[i].LowerCaseValues) < len(hints[j].LowerCaseValues)
		}
		return minLen(hints[i].LowerCaseValues) > minLen(hints[j].LowerCaseValues)
	})
	r.TokenHints = hints
	r.AnchorHint = anchor
	if minCount > 127 {
		minCount = 127
	}
	r.MinTokenCount = minCount
}

func minLen(vals []string) int {
	if len(vals) == 0 {
		return 0
	}
	m := len(vals[0])
	for _, v := range vals[1:] {
		if len(v) < m {
			m = len(v)
		}
	}
	return m
}

// hasStringThatMustMatch ports PatternToken.hasStringThatMustMatch.
func hasStringThatMustMatch(token *PatternToken) bool {
	if token == nil {
		return false
	}
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
	// Java: getNonWhitespaceTokenCount() < minTokenCount
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
