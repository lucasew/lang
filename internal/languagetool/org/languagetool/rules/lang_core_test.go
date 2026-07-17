package rules_test

// Integration smokes for language RegisterCore* packs (import from lang packages).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/es"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/fr"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/nl"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/pl"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/uk"
	"github.com/stretchr/testify/require"
)

func TestLangCoreRegisters(t *testing.T) {
	cases := []struct {
		name string
		reg  func(*languagetool.JLanguageTool)
		lang string
		bad  string
		id   string
	}{
		{"fr", fr.RegisterCoreFrenchRules, "fr", "bonjour bonjour", "FR_WORD_REPEAT_RULE"},
		{"es", es.RegisterCoreSpanishRules, "es", "hola hola", "SPANISH_WORD_REPEAT_RULE"},
		{"nl", nl.RegisterCoreDutchRules, "nl", "hallo hallo", "NL_WORD_REPEAT_RULE"},
		{"pl", pl.RegisterCorePolishRules, "pl", "test test", "PL_WORD_REPEAT"},
		// UK word-repeat ignores untagged content words; function doubles still fire
		{"uk", uk.RegisterCoreUkrainianRules, "uk", "без без", "UKRAINIAN_WORD_REPEAT_RULE"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lt := languagetool.NewJLanguageTool(tc.lang)
			tc.reg(lt)
			require.NotEmpty(t, lt.Check("a  b")) // multi-space
			m := lt.Check(tc.bad)
			require.NotEmpty(t, m, tc.name)
			if tc.id != "" {
				found := false
				for _, x := range m {
					if x.RuleID == tc.id {
						found = true
						break
					}
				}
				require.True(t, found, "want id %s in %+v", tc.id, m)
			}
		})
	}
}
