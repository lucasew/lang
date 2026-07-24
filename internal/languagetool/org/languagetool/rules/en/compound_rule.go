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
// Java: CompoundRule(messages, english, userConfig) — shortCode "en" for toId (en-US → en).
func NewCompoundRule(messages map[string]string) *CompoundRule {
	base := &rules.AbstractCompoundRule{
		ID:                          "EN_COMPOUNDS",
		Description:                 "Hyphenated words: $match",
		WithHyphenMessage:           "This word is normally spelled with a hyphen.",
		WithoutHyphenMessage:        "This word is normally spelled as one.",
		WithOrWithoutHyphenMessage:  "This expression is normally spelled as one or with a hyphen.",
		ShortDesc:                   "Compound",
		SentenceStartsWithUpperCase: true,
		// Java Language.getShortCode() for en-US is "en" (toId / makeAntiPatterns language).
		LanguageCode: "en",
		Data:         loadCompoundData(),
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
	// When SpellingIsMisspelled is nil, treat as correctly spelled (cannot invent a speller).
	base.IsMisspelled = func(word string) bool {
		if r.SpellingIsMisspelled == nil {
			return false
		}
		return r.SpellingIsMisspelled(word)
	}
	return r
}

// Match applies ANTI_PATTERNS immunization then AbstractCompoundRule.
// Java AbstractCompoundRule.match: getSentenceWithImmunization(sentence) via getAntiPatterns().
func (r *CompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractCompoundRule.Match(getSentenceWithCompoundImmunization(sentence))
}

// GetAntiPatterns ports CompoundRule.getAntiPatterns.
func (r *CompoundRule) GetAntiPatterns() []*disambigrules.DisambiguationPatternRule {
	return compoundAntiPatterns()
}

// compoundAntiPatterns ports CompoundRule.getAntiPatterns (cached IMMUNIZE rules).
// Java: makeAntiPatterns(..., AMERICAN_ENGLISH) with language shortCode "en".
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
// for CompoundRule.ANTI_PATTERNS (copy sentence, then IMMUNIZE via Replace).
func getSentenceWithCompoundImmunization(sentence *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if sentence == nil {
		return nil
	}
	aps := compoundAntiPatterns()
	if len(aps) == 0 {
		return sentence
	}
	// Java: AnalyzedSentence immunizedSentence = sentence.copy(sentence);
	immunized := sentence.Copy(sentence)
	for _, ap := range aps {
		if ap == nil {
			continue
		}
		if next := ap.Replace(immunized); next != nil {
			immunized = next
		}
	}
	return immunized
}
