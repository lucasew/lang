package rules_test

// Integration smokes for language RegisterCore* packs (import from lang packages).
// Expectations follow Java getRelevantRules (no invent word-repeat for FR/NL).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ca"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/es"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/fr"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/it"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/nl"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/pl"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/pt"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ru"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/uk"
	"github.com/stretchr/testify/require"
)

func TestLangCoreRegisters(t *testing.T) {
	cases := []struct {
		name string
		reg  func(*languagetool.JLanguageTool)
		lang string
		// bad text that should produce at least one match (often multi-space or word-repeat)
		bad string
		// optional rule id that must appear among matches (empty = any match ok)
		id string
	}{
		// Java French: no WordRepeatRule — multi-space via WHITESPACE_RULE
		{"fr", fr.RegisterCoreFrenchRules, "fr", "bonjour  monde", "WHITESPACE_RULE"},
		{"es", es.RegisterCoreSpanishRules, "es", "hola hola", "SPANISH_WORD_REPEAT_RULE"},
		// Java Dutch: no WordRepeatRule — multi-space via WHITESPACE_RULE
		{"nl", nl.RegisterCoreDutchRules, "nl", "hallo  wereld", "WHITESPACE_RULE"},
		// Java Polish: WordRepeatRule (WORD_REPEAT_RULE) + PolishWordRepeatRule (PL_WORD_REPEAT)
		{"pl", pl.RegisterCorePolishRules, "pl", "test test", "WORD_REPEAT_RULE"},
		// UK word-repeat is POS-gated (Java): untagged doubles are ignored — smoke registration only
		{"uk", uk.RegisterCoreUkrainianRules, "uk", "a  b", ""},
		{"it", it.RegisterCoreItalianRules, "it", "ciao ciao", "ITALIAN_WORD_REPEAT_RULE"},
		{"pt", pt.RegisterCorePortugueseRules, "pt", "teste teste", "PORTUGUESE_WORD_REPEAT_RULE"},
		// Java RussianSimpleWordRepeatRule extends WordRepeatRule → WORD_REPEAT_RULE id
		{"ru", ru.RegisterCoreRussianRules, "ru", "тест тест", "WORD_REPEAT_RULE"},
		{"ca", ca.RegisterCoreCatalanRules, "ca", "hola hola", "CATALAN_WORD_REPEAT_RULE"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lt := languagetool.NewJLanguageTool(tc.lang)
			tc.reg(lt)
			// multi-space when language registers WHITESPACE_RULE
			if tc.name != "uk" {
				// uk still has multi-space via MultipleWhitespace
			}
			require.NotEmpty(t, lt.Check("a  b"), tc.name)
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
			if tc.name == "uk" {
				// Faithful UKRAINIAN_WORD_REPEAT_RULE is registered (POS-gated Match)
				require.Contains(t, m[0].EnabledRules, "UKRAINIAN_WORD_REPEAT_RULE")
			}
		})
	}
}
