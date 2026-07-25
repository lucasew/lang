package de

import (
	"embed"
	"io"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

//go:embed data/compounds.txt data/compound-cities.txt
var compoundsFS embed.FS

var (
	deCompoundOnce sync.Once
	deCompoundData *rules.CompoundRuleData

	deCompoundAntiOnce  sync.Once
	deCompoundAntiRules []*disambigrules.DisambiguationPatternRule
)

func mustLoadCompoundData(expander rules.LineExpander) *rules.CompoundRuleData {
	f1, err := compoundsFS.Open("data/compounds.txt")
	if err != nil {
		panic(err)
	}
	defer f1.Close()
	f2, err := compoundsFS.Open("data/compound-cities.txt")
	if err != nil {
		panic(err)
	}
	defer f2.Close()
	d, err := rules.NewCompoundRuleDataMulti(expander, []io.Reader{f1, f2}, []string{"/de/compounds.txt", "/de/compound-cities.txt"})
	if err != nil {
		panic(err)
	}
	return d
}

func loadDECompoundData() *rules.CompoundRuleData {
	deCompoundOnce.Do(func() {
		deCompoundData = mustLoadCompoundData(nil)
	})
	return deCompoundData
}

// GermanCompoundRule ports org.languagetool.rules.de.GermanCompoundRule.
//
// Java isMisspelled:
//
//	return GermanyGerman.getInstance().getDefaultSpellingRule().isMisspelled(word);
//
// SpellingIsMisspelled is that hook. When nil, keep AbstractCompoundRule default
// (isMisspelled=false → keep suggestions). Do not substitute raw FilterDict alone:
// it rejects Java-accepted compound suggestions (e.g. "Stand-by-Betrieb") that the
// full GermanSpellerRule/hunspell path accepts.
type GermanCompoundRule struct {
	*rules.AbstractCompoundRule
	// SpellingIsMisspelled ports getDefaultSpellingRule().isMisspelled; nil → misspelled=false.
	SpellingIsMisspelled func(word string) bool
}

func NewGermanCompoundRule(messages map[string]string) *GermanCompoundRule {
	base := &rules.AbstractCompoundRule{
		ID:                          "DE_COMPOUNDS",
		Description:                 "Zusammenschreibung von Wörtern, z. B. 'CD-ROM' statt 'CD ROM'",
		WithHyphenMessage:           "Dieses Wort wird mit Bindestrich geschrieben.",
		WithoutHyphenMessage:        "Dieses Wort wird zusammengeschrieben.",
		WithOrWithoutHyphenMessage:  "Diese Wörter werden zusammengeschrieben oder mit Bindestrich getrennt.",
		ShortDesc:                   "Zusammenschreibung von Wörtern",
		SentenceStartsWithUpperCase: true,
		// Java makeAntiPatterns(..., GermanyGerman) / Language.getShortCode() → "de"
		LanguageCode: "de",
		Data:         loadDECompoundData(),
	}
	rules.InitCompoundRuleMeta(base, messages)
	// Java GermanCompoundRule: HNO Arzt → HNO-Arzt
	base.AddExamplePair(
		rules.Wrong("Wenn es schlimmer wird, solltest Du zum <marker>HNO Arzt</marker> gehen."),
		rules.Fixed("Wenn es schlimmer wird, solltest Du zum <marker>HNO-Arzt</marker> gehen."),
	)
	r := &GermanCompoundRule{AbstractCompoundRule: base}
	base.IsMisspelled = func(word string) bool {
		if r.SpellingIsMisspelled != nil {
			return r.SpellingIsMisspelled(word)
		}
		// AbstractCompoundRule default: isMisspelled → false
		return false
	}
	return r
}

// Match applies ANTI_PATTERNS immunization then AbstractCompoundRule.
// Java AbstractCompoundRule.match: getSentenceWithImmunization(sentence) via getAntiPatterns().
func (r *GermanCompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractCompoundRule.Match(getSentenceWithDECompoundImmunization(sentence))
}

// GetAntiPatterns ports GermanCompoundRule.getAntiPatterns.
func (r *GermanCompoundRule) GetAntiPatterns() []*disambigrules.DisambiguationPatternRule {
	return deCompoundAntiPatterns()
}

// deCompoundAntiPatterns ports GermanCompoundRule.getAntiPatterns (cached IMMUNIZE rules).
func deCompoundAntiPatterns() []*disambigrules.DisambiguationPatternRule {
	deCompoundAntiOnce.Do(func() {
		aps := GermanCompoundRuleAntiPatterns
		deCompoundAntiRules = make([]*disambigrules.DisambiguationPatternRule, 0, len(aps))
		for _, toks := range aps {
			if len(toks) == 0 {
				continue
			}
			// Java makeAntiPatterns: INTERNAL_ANTIPATTERN + IMMUNIZE, GermanyGerman shortCode "de"
			rule := disambigrules.NewDisambiguationPatternRule(
				"INTERNAL_ANTIPATTERN", "(no description)", "de",
				toks, "", nil, disambigrules.ActionImmunize,
			)
			deCompoundAntiRules = append(deCompoundAntiRules, rule)
		}
	})
	return deCompoundAntiRules
}

// getSentenceWithDECompoundImmunization ports Rule.getSentenceWithImmunization
// for GermanCompoundRule.ANTI_PATTERNS (also used by SwissCompoundRule).
func getSentenceWithDECompoundImmunization(sentence *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if sentence == nil {
		return nil
	}
	aps := deCompoundAntiPatterns()
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

// isDigits reports whether s is non-empty and consists only of ASCII digits.
// Used by agreement and compound helpers in this package.
func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
