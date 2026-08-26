package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCore_BritishSpellerID(t *testing.T) {
	lt := languagetool.NewJLanguageTool("en-GB")
	RegisterCoreEnglishLanguageRules(lt)
	require.Contains(t, lt.GetAllRegisteredRuleIDs(), "MORFOLOGIK_RULE_EN_GB")
}

func TestRegisterCore_CanadianSpellerID(t *testing.T) {
	lt := languagetool.NewJLanguageTool("en-CA")
	RegisterCoreEnglishLanguageRules(lt)
	require.Contains(t, lt.GetAllRegisteredRuleIDs(), "MORFOLOGIK_RULE_EN_CA")
}
