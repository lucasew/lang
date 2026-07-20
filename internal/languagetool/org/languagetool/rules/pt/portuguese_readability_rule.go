package pt

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PortugueseReadabilityRule ports org.languagetool.rules.pt.PortugueseReadabilityRule metadata.
type PortugueseReadabilityRule struct {
	*rules.ReadabilityRule
	TooEasyTest bool
}

func NewPortugueseReadabilityRule(tooEasy bool, level int) *PortugueseReadabilityRule {
	return &PortugueseReadabilityRule{
		ReadabilityRule: rules.NewReadabilityRule(tooEasy, level),
		TooEasyTest:     tooEasy,
	}
}

func (r *PortugueseReadabilityRule) GetID() string {
	if r.TooEasyTest {
		return "READABILITY_RULE_SIMPLE_PT"
	}
	return "READABILITY_RULE_DIFFICULT_PT"
}

func (r *PortugueseReadabilityRule) GetDescription() string {
	if r.TooEasyTest {
		return "Legibilidade: texto demasiado simples"
	}
	return "Legibilidade: texto demasiado complexo"
}

// MatchList ports ReadabilityRule.match (paragraph FRE check; default-off).
func (r *PortugueseReadabilityRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.ReadabilityRule == nil || len(sentences) == 0 {
		return nil
	}
	var words []string
	nSent := 0
	var startSent *languagetool.AnalyzedSentence
	startPos, endPos := -1, -1
	pos := 0
	for _, s := range sentences {
		if s == nil {
			continue
		}
		nSent++
		toks := s.GetTokensWithoutWhitespace()
		if startSent == nil && len(toks) > 1 {
			startSent = s
			startPos = pos + toks[1].GetStartPos()
			if len(toks) > 3 {
				endPos = pos + toks[3].GetEndPos()
			} else {
				endPos = pos + toks[len(toks)-1].GetEndPos()
			}
		}
		for _, t := range toks {
			if t == nil || t.IsSentenceStart() || t.IsSentenceEnd() || t.IsNonWord() {
				continue
			}
			words = append(words, t.GetToken())
		}
		pos += s.GetCorrectedTextLength()
	}
	_, _, too := r.EvaluateParagraph(nSent, words)
	if !too || startSent == nil || startPos < 0 || endPos < 0 {
		return nil
	}
	msg := "Legibilidade: texto demasiado complexo"
	if r.TooEasyTest {
		msg = "Legibilidade: texto demasiado simples"
	}
	return []*rules.RuleMatch{rules.NewRuleMatch(r, startSent, startPos, endPos, msg)}
}
