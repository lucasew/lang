package bitext

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/bitext"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// SentenceMatcher is the pattern-rule surface used for source/target legs.
type SentenceMatcher interface {
	GetID() string
	GetDescription() string
	GetMessage() string
	Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error)
}

// BitextPatternRule ports org.languagetool.rules.patterns.bitext.BitextPatternRule.
type BitextPatternRule struct {
	bitext.BitextRuleBase
	SrcRule SentenceMatcher
	TrgRule SentenceMatcher
}

func NewBitextPatternRule(src, trg SentenceMatcher) *BitextPatternRule {
	r := &BitextPatternRule{SrcRule: src, TrgRule: trg}
	if src != nil {
		r.ID = src.GetID()
		r.Description = src.GetDescription()
	}
	if trg != nil {
		r.Message = trg.GetMessage()
	}
	return r
}

func (r *BitextPatternRule) GetSrcRule() SentenceMatcher { return r.SrcRule }
func (r *BitextPatternRule) GetTrgRule() SentenceMatcher { return r.TrgRule }

// Match always returns nil (Java returns empty; use MatchBitext).
func (r *BitextPatternRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	return nil, nil
}

// MatchBitext runs target rule only when source rule matches.
func (r *BitextPatternRule) MatchBitext(source, target *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.SrcRule == nil || r.TrgRule == nil {
		return nil
	}
	srcHits, err := r.SrcRule.Match(source)
	if err != nil || len(srcHits) == 0 {
		return nil
	}
	trgHits, err := r.TrgRule.Match(target)
	if err != nil {
		return nil
	}
	return trgHits
}

// Ensure PatternRule can be used as SentenceMatcher.
var _ SentenceMatcher = (*patterns.PatternRule)(nil)
var _ bitext.BitextRule = (*BitextPatternRule)(nil)
