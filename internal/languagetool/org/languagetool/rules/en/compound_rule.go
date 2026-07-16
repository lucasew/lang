package en

import (
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

var (
	compoundOnce sync.Once
	compoundData *rules.CompoundRuleData
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
type CompoundRule struct {
	*rules.AbstractCompoundRule
}

// NewCompoundRule constructs EN_COMPOUNDS.
func NewCompoundRule(messages map[string]string) *CompoundRule {
	base := &rules.AbstractCompoundRule{
		Messages:                    messages,
		ID:                          "EN_COMPOUNDS",
		Description:                 "Hyphenated words: $match",
		WithHyphenMessage:           "This word is normally spelled with a hyphen.",
		WithoutHyphenMessage:        "This word is normally spelled as one.",
		WithOrWithoutHyphenMessage:  "This expression is normally spelled as one or with a hyphen.",
		ShortDesc:                   "Compound",
		SentenceStartsWithUpperCase: true,
		Data:                        loadCompoundData(),
		// Without Morfologik, treat suggestions as correctly spelled (Java isMisspelled default false).
	}
	base.UseSubRuleSpecificIDs()
	return &CompoundRule{AbstractCompoundRule: base}
}

// Match applies light anti-patterns then AbstractCompoundRule.
// Full DisambiguationPatternRule anti-patterns are not ported; cover the ones
// required by CompoundRuleTest (contraction 're, &|and co, etc.).
func (r *CompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	immunizeENCompoundAntiPatterns(sentence)
	return r.AbstractCompoundRule.Match(sentence)
}

// immunizeENCompoundAntiPatterns ports a subset of CompoundRule.ANTI_PATTERNS.
func immunizeENCompoundAntiPatterns(sentence *languagetool.AnalyzedSentence) {
	tokens := sentence.GetTokensWithoutWhitespace()
	for i := 0; i < len(tokens); i++ {
		tok := tokens[i].GetToken()
		// ['’`´‘] + re  (they're / we're / …)
		if isApostropheToken(tok) && i+1 < len(tokens) {
			next := tokens[i+1].GetToken()
			if next == "re" || next == "Re" || next == "RE" {
				tokens[i].Immunize(0)
				tokens[i+1].Immunize(0)
			}
		}
		// and|& + co  (Tiffany & Co)
		if (tok == "and" || tok == "And" || tok == "&") && i+1 < len(tokens) {
			next := tokens[i+1].GetToken()
			if next == "co" || next == "Co" || next == "CO" {
				tokens[i].Immunize(0)
				tokens[i+1].Immunize(0)
			}
		}
		// first + ever + green
		if equalFold(tok, "first") && i+2 < len(tokens) &&
			equalFold(tokens[i+1].GetToken(), "ever") &&
			equalFold(tokens[i+2].GetToken(), "green") {
			tokens[i].Immunize(0)
			tokens[i+1].Immunize(0)
			tokens[i+2].Immunize(0)
		}
	}
}

func isApostropheToken(s string) bool {
	switch s {
	case "'", "’", "`", "´", "‘":
		return true
	}
	return false
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		// still use simple ASCII fold for anti-pattern tokens
	}
	if a == b {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}
