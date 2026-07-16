package km

// Twin of languagetool-language-modules/km/src/test/java/org/languagetool/rules/km/KhmerWordRepeatRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestKhmerWordRepeatRule_WordRepeatRule(t *testing.T) {
	rule := NewKhmerWordRepeatRule(map[string]string{
		"repetition":            "Word repetition",
		"desc_repetition_short": "Repetition",
	})
	// correct: ៗ is not the same token as word; space-separated repeat ignored
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("នេះ​ហើយៗ​នោះ។"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("គាត់​ហើយ ហើយ​ខ្ញុំ។"))))
	// incorrect: ZWSP-separated same word
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("នេះ​ហើយ​ហើយ​នោះ។"))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("ខ្ញុំ​និង​និង​គាត់។"))))
}
