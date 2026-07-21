package ekavian

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

const (
	MorfologikEkavianSpellerRuleID = "MORFOLOGIK_RULE_SR_EKAVIAN"
	EkavianSpellerDict             = "/sr/dictionary/ekavian/serbian.dict"
	// Java MorfologikEkavianSpellerRule path overrides (not hunspell/).
	EkavianIgnoreFile   = "sr/dictionary/ekavian/ignored.txt"
	EkavianSpellingFile = "sr/dictionary/ekavian/spelling.txt"
	EkavianProhibitFile = "sr/dictionary/ekavian/prohibit.txt"
)

// MorfologikEkavianSpellerRule ports rules.sr.ekavian.MorfologikEkavianSpellerRule.
type MorfologikEkavianSpellerRule struct {
	*morfologik.MorfologikSpellerRule
	// incorrectExamples / correctExamples port Rule.addExamplePair (not on SpellingCheckRule).
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

func NewMorfologikEkavianSpellerRule() *MorfologikEkavianSpellerRule {
	r := &MorfologikEkavianSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikEkavianSpellerRuleID, "sr", EkavianSpellerDict, nil),
	}
	// Java getIgnoreFileName / getSpellingFileName / getProhibitFileName under dictionary/ekavian/.
	if r.SpellingCheckRule != nil {
		r.GetIgnoreFileNameFn = func() string { return "/" + EkavianIgnoreFile }
		r.GetSpellingFileNameFn = func() string { return "/" + EkavianSpellingFile }
		r.GetProhibitFileNameFn = func() string { return "/" + EkavianProhibitFile }
		// clear default additional prohibit/custom that don't apply to SR paths
		r.GetAdditionalProhibitFileNamesFn = func() []string { return nil }
		r.GetAdditionalSpellingFileNamesFn = func() []string { return []string{spelling.GlobalSpellingFile} }
		spelling.ReapplyDefaultSpellingWordLists(r.SpellingCheckRule)
	}
	// Java: бткие → битке
	r.AddExamplePair(
		rules.Wrong("Изгубила све сам <marker>бткие</marker>, ал' још водим рат."),
		rules.Fixed("Изгубила све сам <marker>битке</marker>, ал' још водим рат."),
	)
	return r
}

// AddExamplePair ports Rule.addExamplePair.
func (r *MorfologikEkavianSpellerRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *MorfologikEkavianSpellerRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *MorfologikEkavianSpellerRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// GetIgnoreFileName ports getIgnoreFileName.
func (r *MorfologikEkavianSpellerRule) GetIgnoreFileName() string { return "/" + EkavianIgnoreFile }

// GetSpellingFileName ports getSpellingFileName.
func (r *MorfologikEkavianSpellerRule) GetSpellingFileName() string { return "/" + EkavianSpellingFile }

// GetProhibitFileName ports getProhibitFileName.
func (r *MorfologikEkavianSpellerRule) GetProhibitFileName() string { return "/" + EkavianProhibitFile }
