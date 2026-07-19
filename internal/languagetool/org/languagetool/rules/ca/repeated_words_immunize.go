package ca

import (
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

var (
	caRWAntiOnce  sync.Once
	caRWAntiRules []*disambigrules.DisambiguationPatternRule
)

func catalanRepeatedWordsSentenceWithImmunizationAntiRules() []*disambigrules.DisambiguationPatternRule {
	caRWAntiOnce.Do(func() {
		aps := CatalanRepeatedWordsAntiPatterns
		caRWAntiRules = make([]*disambigrules.DisambiguationPatternRule, 0, len(aps))
		for _, toks := range aps {
			if len(toks) == 0 {
				continue
			}
			caRWAntiRules = append(caRWAntiRules, disambigrules.NewDisambiguationPatternRule(
				"INTERNAL_ANTIPATTERN", "(no description)", "ca",
				toks, "", nil, disambigrules.ActionImmunize,
			))
		}
	})
	return caRWAntiRules
}

// catalanRepeatedWordsSentenceWithImmunization ports Rule.getSentenceWithImmunization for repeated-words anti-patterns.
func catalanRepeatedWordsSentenceWithImmunization(sentence *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if sentence == nil {
		return nil
	}
	aps := catalanRepeatedWordsSentenceWithImmunizationAntiRules()
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
