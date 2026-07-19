package en

import (
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

var (
	enRWAntiOnce  sync.Once
	enRWAntiRules []*disambigrules.DisambiguationPatternRule
)

func englishRepeatedWordsSentenceWithImmunizationAntiRules() []*disambigrules.DisambiguationPatternRule {
	enRWAntiOnce.Do(func() {
		aps := EnglishRepeatedWordsAntiPatterns
		enRWAntiRules = make([]*disambigrules.DisambiguationPatternRule, 0, len(aps))
		for _, toks := range aps {
			if len(toks) == 0 {
				continue
			}
			enRWAntiRules = append(enRWAntiRules, disambigrules.NewDisambiguationPatternRule(
				"INTERNAL_ANTIPATTERN", "(no description)", "en",
				toks, "", nil, disambigrules.ActionImmunize,
			))
		}
	})
	return enRWAntiRules
}

// englishRepeatedWordsSentenceWithImmunization ports Rule.getSentenceWithImmunization for repeated-words anti-patterns.
func englishRepeatedWordsSentenceWithImmunization(sentence *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if sentence == nil {
		return nil
	}
	aps := englishRepeatedWordsSentenceWithImmunizationAntiRules()
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
