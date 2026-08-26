package sr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCore_EkavianSpellerID(t *testing.T) {
	lt := languagetool.NewJLanguageTool("sr")
	RegisterCoreSerbianRules(lt)
	require.Contains(t, lt.GetAllRegisteredRuleIDs(), "MORFOLOGIK_RULE_SR_EKAVIAN")
	require.NotContains(t, lt.GetAllRegisteredRuleIDs(), "MORFOLOGIK_RULE_SR_JEKAVIAN")
}

func TestRegisterCore_SerbiaEkavianSpellerID(t *testing.T) {
	lt := languagetool.NewJLanguageTool("sr-RS")
	RegisterCoreSerbianRules(lt)
	require.Contains(t, lt.GetAllRegisteredRuleIDs(), "MORFOLOGIK_RULE_SR_EKAVIAN")
}

func TestRegisterCore_JekavianCountrySpellerID(t *testing.T) {
	for _, code := range []string{"sr-BA", "sr-HR", "sr-ME"} {
		lt := languagetool.NewJLanguageTool(code)
		RegisterCoreSerbianRules(lt)
		require.Contains(t, lt.GetAllRegisteredRuleIDs(), "MORFOLOGIK_RULE_SR_JEKAVIAN", code)
		require.NotContains(t, lt.GetAllRegisteredRuleIDs(), "MORFOLOGIK_RULE_SR_EKAVIAN", code)
	}
}
