package gl

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikGalicianSpellerRuleID = "MORFOLOGIK_RULE_GL_ES"
	GalicianSpellerDict = "/gl/hunspell/gl_ES.dict"
)

type MorfologikGalicianSpellerRule struct { *morfologik.MorfologikSpellerRule }

func NewMorfologikGalicianSpellerRule() *MorfologikGalicianSpellerRule {
	return &MorfologikGalicianSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(MorfologikGalicianSpellerRuleID, "gl", GalicianSpellerDict, nil),
	}
}
