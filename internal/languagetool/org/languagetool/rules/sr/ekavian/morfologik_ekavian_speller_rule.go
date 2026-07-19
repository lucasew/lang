package ekavian

import (
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
}

func NewMorfologikEkavianSpellerRule() *MorfologikEkavianSpellerRule {
	r := &MorfologikEkavianSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikEkavianSpellerRuleID, "sr", EkavianSpellerDict, nil),
	}
	// Java getIgnoreFileName / getSpellingFileName / getProhibitFileName under dictionary/ekavian/.
	if r.SpellingCheckRule != nil {
		spelling.ApplySpellingResourcePaths(r.SpellingCheckRule, EkavianIgnoreFile, EkavianSpellingFile, EkavianProhibitFile)
	}
	return r
}

// GetIgnoreFileName ports getIgnoreFileName.
func (r *MorfologikEkavianSpellerRule) GetIgnoreFileName() string { return "/" + EkavianIgnoreFile }

// GetSpellingFileName ports getSpellingFileName.
func (r *MorfologikEkavianSpellerRule) GetSpellingFileName() string { return "/" + EkavianSpellingFile }

// GetProhibitFileName ports getProhibitFileName.
func (r *MorfologikEkavianSpellerRule) GetProhibitFileName() string { return "/" + EkavianProhibitFile }
