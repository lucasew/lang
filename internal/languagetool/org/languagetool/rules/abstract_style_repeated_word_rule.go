package rules

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// AbstractStyleRepeatedWordRule ports
// org.languagetool.rules.AbstractStyleRepeatedWordRule surface:
// flags repeated content words within a sentence or across nearby sentences.
type AbstractStyleRepeatedWordRule struct {
	ID                     string
	Description            string
	MaxDistanceOfSentences int // default 1
	// IsTokenToCheck filters which tokens participate (nil = all non-trivial words).
	IsTokenToCheck func(tok *languagetool.AnalyzedTokenReadings) bool
	// TokenKey maps token to comparison key (default: lowercased surface).
	TokenKey func(tok *languagetool.AnalyzedTokenReadings) string
	// MessageSameSentence / MessageConsecutive optional message builders.
	MessageSameSentence func(word string) string
	MessageConsecutive  func(word string) string
}

func NewAbstractStyleRepeatedWordRule() *AbstractStyleRepeatedWordRule {
	return &AbstractStyleRepeatedWordRule{
		ID:                     "STYLE_REPEATED_WORD_RULE",
		Description:            "Repeated words in consecutive sentences",
		MaxDistanceOfSentences: 1,
	}
}

func (r *AbstractStyleRepeatedWordRule) GetID() string {
	if r.ID != "" {
		return r.ID
	}
	return "STYLE_REPEATED_WORD_RULE"
}

func (r *AbstractStyleRepeatedWordRule) GetDescription() string {
	if r.Description != "" {
		return r.Description
	}
	return "Repeated words in consecutive sentences"
}

func (r *AbstractStyleRepeatedWordRule) key(tok *languagetool.AnalyzedTokenReadings) string {
	if r.TokenKey != nil {
		return r.TokenKey(tok)
	}
	return strings.ToLower(tok.GetToken())
}

func (r *AbstractStyleRepeatedWordRule) check(tok *languagetool.AnalyzedTokenReadings) bool {
	if r.IsTokenToCheck != nil {
		return r.IsTokenToCheck(tok)
	}
	t := tok.GetToken()
	if len(t) < 2 {
		return false
	}
	// skip pure punctuation
	for _, ch := range t {
		if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || ch > 127 {
			return true
		}
	}
	return false
}

// MatchList finds repeated checked tokens within a sentence and across adjacent sentences.
func (r *AbstractStyleRepeatedWordRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	if r == nil {
		return nil
	}
	maxDist := r.MaxDistanceOfSentences
	if maxDist < 0 {
		maxDist = 0
	}
	var out []*RuleMatch
	pos := 0
	// per-sentence keys with positions
	type occ struct {
		key  string
		from int
		to   int
		sent *languagetool.AnalyzedSentence
	}
	var sentOccs [][]occ
	for _, s := range sentences {
		if s == nil {
			sentOccs = append(sentOccs, nil)
			continue
		}
		var ocs []occ
		seenInSent := map[string]occ{}
		for _, tok := range s.GetTokensWithoutWhitespace() {
			if !r.check(tok) {
				continue
			}
			k := r.key(tok)
			if k == "" {
				continue
			}
			o := occ{key: k, from: pos + tok.GetStartPos(), to: pos + tok.GetEndPos(), sent: s}
			if prev, ok := seenInSent[k]; ok {
				msg := "Repeated word in the same sentence"
				if r.MessageSameSentence != nil {
					msg = r.MessageSameSentence(k)
				}
				out = append(out, NewRuleMatch(r, s, o.from, o.to, msg))
				_ = prev
			} else {
				seenInSent[k] = o
			}
			ocs = append(ocs, o)
		}
		sentOccs = append(sentOccs, ocs)
		pos += s.GetCorrectedTextLength()
	}
	// consecutive sentences
	for i := 0; i < len(sentOccs); i++ {
		keys := map[string]occ{}
		for _, o := range sentOccs[i] {
			keys[o.key] = o
		}
		for d := 1; d <= maxDist && i+d < len(sentOccs); d++ {
			for _, o := range sentOccs[i+d] {
				if prev, ok := keys[o.key]; ok {
					msg := "Repeated word in consecutive sentences"
					if r.MessageConsecutive != nil {
						msg = r.MessageConsecutive(o.key)
					}
					out = append(out, NewRuleMatch(r, o.sent, o.from, o.to, msg))
					_ = prev
				}
			}
		}
	}
	return out
}
