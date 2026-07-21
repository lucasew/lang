package jekavian

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

const (
	MorfologikJekavianSpellerRuleID = "MORFOLOGIK_RULE_SR_JEKAVIAN"
	JekavianSpellerDict             = "/sr/dictionary/jekavian/serbian.dict"
	// Java MorfologikJekavianSpellerRule path overrides (not hunspell/).
	JekavianIgnoreFile   = "sr/dictionary/jekavian/ignored.txt"
	JekavianSpellingFile = "sr/dictionary/jekavian/spelling.txt"
	JekavianProhibitFile = "sr/dictionary/jekavian/prohibit.txt"
)

// MorfologikJekavianSpellerRule ports rules.sr.jekavian.MorfologikJekavianSpellerRule.
type MorfologikJekavianSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

func NewMorfologikJekavianSpellerRule() *MorfologikJekavianSpellerRule {
	r := &MorfologikJekavianSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikJekavianSpellerRuleID, "sr", JekavianSpellerDict, nil),
	}
	// Java getIgnoreFileName / getSpellingFileName / getProhibitFileName under dictionary/jekavian/.
	if r.SpellingCheckRule != nil {
		r.GetIgnoreFileNameFn = func() string { return "/" + JekavianIgnoreFile }
		r.GetSpellingFileNameFn = func() string { return "/" + JekavianSpellingFile }
		r.GetProhibitFileNameFn = func() string { return "/" + JekavianProhibitFile }
		r.GetAdditionalProhibitFileNamesFn = func() []string { return nil }
		r.GetAdditionalSpellingFileNamesFn = func() []string { return []string{spelling.GlobalSpellingFile} }
		spelling.ReapplyDefaultSpellingWordLists(r.SpellingCheckRule)
	}
	// Java MorfologikSpellerRule.initSpeller (path getters already set).
	r.InitSpellersFromGetters(nil, nil)
	return r
}

// GetIgnoreFileName ports getIgnoreFileName.
func (r *MorfologikJekavianSpellerRule) GetIgnoreFileName() string { return "/" + JekavianIgnoreFile }

// GetSpellingFileName ports getSpellingFileName.
func (r *MorfologikJekavianSpellerRule) GetSpellingFileName() string {
	return "/" + JekavianSpellingFile
}

// GetProhibitFileName ports getProhibitFileName.
func (r *MorfologikJekavianSpellerRule) GetProhibitFileName() string {
	return "/" + JekavianProhibitFile
}
