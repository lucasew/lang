package ru

import (
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// Reuses compoundsFS embed from russian_compound_rule.go (same data/compounds.txt).
// Java RussianDashRule.getCompoundsData loads /ru/compounds.txt (resource dir).
const compoundsResourcePath = "/ru/compounds.txt"

var (
	dashOnce     sync.Once
	dashPatterns []string
)

func loadDashPatterns() []string {
	dashOnce.Do(func() {
		// Java: loadCompoundFile("/ru/compounds.txt") once (DCL / sync.Once).
		f, err := compoundsFS.Open("data/compounds.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		p, err := rules.LoadDashCompoundPatterns(f)
		if err != nil {
			panic(err)
		}
		dashPatterns = p
	})
	return dashPatterns
}

// isCyrillicLetter ports Java RussianDashRule isBoundary inverse:
//
//	PATTERN = [\u0400-\u04FF]
//	isBoundary(s) = !PATTERN.matcher(s).matches()
//
// Exactly the Cyrillic block U+0400..U+04FF — no invent extras.
func isCyrillicLetter(r rune) bool {
	return r >= 0x0400 && r <= 0x04FF
}

// RussianDashRule ports org.languagetool.rules.ru.RussianDashRule.
type RussianDashRule struct {
	*rules.AbstractDashRule
}

func NewRussianDashRule(messages map[string]string) *RussianDashRule {
	// Java: super(messages); setTags(Collections.emptyList());
	// No setUrl / addExamplePair / setLocQualityIssueType in Java RU dash rule.
	base := &rules.AbstractDashRule{
		ID:               "RU_DASH_RULE",
		CompoundPatterns: loadDashPatterns(),
		Message:          "Использовано тире вместо дефиса.",
		Description:      "Тире вместо дефиса («из — за» вместо «из-за»).",
		IsLetter:         isCyrillicLetter,
	}
	rules.InitDashRuleMeta(base, messages)
	// Java RussianDashRule ctor: setTags(Collections.emptyList()) clears AbstractDashRule picky.
	base.Tags = nil
	return &RussianDashRule{AbstractDashRule: base}
}

func (r *RussianDashRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractDashRule.Match(sentence)
}
