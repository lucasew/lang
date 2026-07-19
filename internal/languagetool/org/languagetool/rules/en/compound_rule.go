package en

import (
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

var (
	compoundOnce sync.Once
	compoundData *rules.CompoundRuleData

	compoundAntiOnce  sync.Once
	compoundAntiRules []*disambigrules.DisambiguationPatternRule
)

func loadCompoundData() *rules.CompoundRuleData {
	compoundOnce.Do(func() {
		f, err := compoundsFS.Open("data/compounds.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		d, err := rules.NewCompoundRuleData(f, "/en/compounds.txt")
		if err != nil {
			panic(err)
		}
		compoundData = d
	})
	return compoundData
}

// CompoundRule ports org.languagetool.rules.en.CompoundRule.
// isMisspelled uses SpellingIsMisspelled (Java english.getDefaultSpellingRule().isMisspelled).
// Without SpellingIsMisspelled, isMisspelled is false (AbstractCompoundRule default).
type CompoundRule struct {
	*rules.AbstractCompoundRule
	// SpellingIsMisspelled ports Morfologik speller isMisspelled; nil → misspelled=false.
	SpellingIsMisspelled func(word string) bool
}

// NewCompoundRule constructs EN_COMPOUNDS.
func NewCompoundRule(messages map[string]string) *CompoundRule {
	base := &rules.AbstractCompoundRule{
		ID:                          "EN_COMPOUNDS",
		Description:                 "Hyphenated words: $match",
		WithHyphenMessage:           "This word is normally spelled with a hyphen.",
		WithoutHyphenMessage:        "This word is normally spelled as one.",
		WithOrWithoutHyphenMessage:  "This expression is normally spelled as one or with a hyphen.",
		ShortDesc:                   "Compound",
		SentenceStartsWithUpperCase: true,
		Data:                        loadCompoundData(),
	}
	base.UseSubRuleSpecificIDs()
	rules.InitCompoundRuleMeta(base, messages)
	// Java: setUrl hyphen insights; addExamplePair part time → part-time
	base.URL = "https://languagetool.org/insights/post/hyphen/"
	base.AddExamplePair(
		rules.Wrong("I now have a <marker>part time</marker> job."),
		rules.Fixed("I now have a <marker>part-time</marker> job."),
	)
	r := &CompoundRule{AbstractCompoundRule: base}
	// Java CompoundRule.isMisspelled: english.getDefaultSpellingRule().isMisspelled(word)
	base.IsMisspelled = func(word string) bool {
		if r.SpellingIsMisspelled == nil {
			return false
		}
		return r.SpellingIsMisspelled(word)
	}
	return r
}

// Match applies ANTI_PATTERNS immunization then AbstractCompoundRule.
// Java: getSentenceWithImmunization(sentence) via getAntiPatterns().
func (r *CompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractCompoundRule.Match(getSentenceWithCompoundImmunization(sentence))
}

// compoundAntiPatterns ports CompoundRule.getAntiPatterns (cached IMMUNIZE rules).
func compoundAntiPatterns() []*disambigrules.DisambiguationPatternRule {
	compoundAntiOnce.Do(func() {
		aps := CompoundRuleAntiPatterns
		compoundAntiRules = make([]*disambigrules.DisambiguationPatternRule, 0, len(aps))
		for _, toks := range aps {
			if len(toks) == 0 {
				continue
			}
			// Java makeAntiPatterns: INTERNAL_ANTIPATTERN + IMMUNIZE
			rule := disambigrules.NewDisambiguationPatternRule(
				"INTERNAL_ANTIPATTERN", "(no description)", "en",
				toks, "", nil, disambigrules.ActionImmunize,
			)
			compoundAntiRules = append(compoundAntiRules, rule)
		}
	})
	return compoundAntiRules
}

// getSentenceWithCompoundImmunization ports Rule.getSentenceWithImmunization
// for CompoundRule.ANTI_PATTERNS.
func getSentenceWithCompoundImmunization(sentence *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if sentence == nil {
		return nil
	}
	aps := compoundAntiPatterns()
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
		if ap == nil {
			continue
		}
		immunized = ap.Replace(immunized)
	}
	return immunized
}
