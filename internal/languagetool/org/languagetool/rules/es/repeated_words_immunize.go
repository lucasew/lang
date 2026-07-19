package es

import (
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

var (
	esRWAntiOnce  sync.Once
	esRWAntiRules []*disambigrules.DisambiguationPatternRule
)

func spanishRepeatedWordsSentenceWithImmunizationAntiRules() []*disambigrules.DisambiguationPatternRule {
	esRWAntiOnce.Do(func() {
		aps := SpanishRepeatedWordsAntiPatterns
		esRWAntiRules = make([]*disambigrules.DisambiguationPatternRule, 0, len(aps))
		for _, toks := range aps {
			if len(toks) == 0 {
				continue
			}
			esRWAntiRules = append(esRWAntiRules, disambigrules.NewDisambiguationPatternRule(
				"INTERNAL_ANTIPATTERN", "(no description)", "es",
				toks, "", nil, disambigrules.ActionImmunize,
			))
		}
	})
	return esRWAntiRules
}

// spanishRepeatedWordsSentenceWithImmunization ports Rule.getSentenceWithImmunization for repeated-words anti-patterns.
func spanishRepeatedWordsSentenceWithImmunization(sentence *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if sentence == nil {
		return nil
	}
	aps := spanishRepeatedWordsSentenceWithImmunizationAntiRules()
	if len(aps) == 0 {
		return sentence
	}
	src := sentence.GetTokens()
	cloned := make([]*languagetool.AnalyzedTokenReadings, len(src))
	for i, t := range src {
		if t == nil {
			continue
		}
		cloned[i] = languagetool.NewAnalyzedTokenReadingsFromOld(t, t.GetReadings(), "")
	}
	immunized := languagetool.NewAnalyzedSentence(cloned)
	for _, ap := range aps {
		if ap != nil {
			immunized = ap.Replace(immunized)
		}
	}
	return immunized
}
