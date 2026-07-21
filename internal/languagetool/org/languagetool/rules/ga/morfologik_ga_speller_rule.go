package ga

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	taggingga "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ga"
)

// irishTokenizingPattern ports MorfologikIrishSpellerRule.tokenizingPattern = Pattern.compile("-").
var irishTokenizingPattern = regexp.MustCompile(`-`)

const (
	// MorfologikIrishSpellerRuleID ports MorfologikIrishSpellerRule.getId().
	// Java: "MORFOLOGIK_RULE_GA_IE" (not MORFOLOGIK_RULE_GA).
	MorfologikIrishSpellerRuleID = "MORFOLOGIK_RULE_GA_IE"
	// IrishSpellerDict ports MorfologikIrishSpellerRule.getFileName() → RESOURCE_FILENAME.
	// Java: "/ga/hunspell/ga_IE.dict"
	IrishSpellerDict = "/ga/hunspell/ga_IE.dict"
)

// MorfologikIrishSpellerRule ports rules.ga.MorfologikIrishSpellerRule.
// tokenizingPattern("-"); isMisspelled normalizes maths / halfwidth Latin via tagging/ga.Utils.
type MorfologikIrishSpellerRule struct {
	*morfologik.MorfologikSpellerRule
	// incorrectExamples / correctExamples port Rule.addExamplePair (not on SpellingCheckRule).
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

func NewMorfologikIrishSpellerRule() *MorfologikIrishSpellerRule {
	base := morfologik.NewMorfologikSpellerRule(
		MorfologikIrishSpellerRuleID, "ga", IrishSpellerDict, nil)
	// Java MorfologikIrishSpellerRule: super.ignoreWordsWithLength = 1
	if base.SpellingCheckRule != nil {
		base.IgnoreWordsWithLength = 1
	}
	r := &MorfologikIrishSpellerRule{MorfologikSpellerRule: base}
	// Java tokenizingPattern(): Pattern.compile("-") — base Match splits per segment.
	r.TokenizingPattern = irishTokenizingPattern
	// Wrap IsMisspelled for maths/halfwidth normalization (Java isMisspelled override).
	inner := r.IsMisspelled
	r.IsMisspelled = func(word string) bool {
		return r.irishIsMisspelled(word, inner)
	}
	// Java: botun → botún
	r.AddExamplePair(
		rules.Wrong("Tá <marker>botun</marker> san abairt seo."),
		rules.Fixed("Tá <marker>botún</marker> san abairt seo."),
	)
	return r
}

// AddExamplePair ports Rule.addExamplePair.
func (r *MorfologikIrishSpellerRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *MorfologikIrishSpellerRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *MorfologikIrishSpellerRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// irishIsMisspelled ports isMisspelled: simplify mathematical / halfwidth before dict check.
func (r *MorfologikIrishSpellerRule) irishIsMisspelled(word string, inner func(string) bool) bool {
	check := word
	if taggingga.IsAllMathsChars(word) {
		check = taggingga.SimplifyMathematical(word)
	} else if taggingga.IsAllHalfWidthChars(word) {
		check = taggingga.HalfwidthLatinToLatin(word)
	}
	if inner != nil {
		return inner(check)
	}
	return false
}
