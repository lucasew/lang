package commandline

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/corepack"
	"github.com/stretchr/testify/require"
)

// Soft Analyze attaches SENT_END on the last content word (Java AddReading drops
// null POS). PL AdvancedWordRepeat + UK WordRepeat must still flag adjacent
// doubles without a trailing period (server multi-lang smoke).
func TestGolden_WordRepeat_NoTrailingPeriod(t *testing.T) {
	for _, tc := range []struct{ lang, text, want string }{
		{"pl", "test test", "PL_WORD_REPEAT"},
		{"uk", "без без", "UKRAINIAN_WORD_REPEAT_RULE"},
		{"fr", "bonjour bonjour", "FR_WORD_REPEAT_RULE"},
	} {
		t.Run(tc.lang, func(t *testing.T) {
			lt := languagetool.NewJLanguageTool(tc.lang)
			corepack.Register(lt, tc.lang)
			found := false
			for _, m := range lt.Check(tc.text) {
				if m.RuleID == tc.want {
					found = true
				}
			}
			require.True(t, found, "want %s", tc.want)
		})
	}
}
