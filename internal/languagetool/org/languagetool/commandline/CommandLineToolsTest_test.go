package commandline

// Twin of languagetool-commandline/src/test/java/org/languagetool/commandline/CommandLineToolsTest.java
import (
	"bytes"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Port of CommandLineToolsTest.testCheck
func TestCommandLineTools_Check(t *testing.T) {
	sent := languagetool.AnalyzePlain("This is an test.")
	m := rules.NewRuleMatch(rules.NewFakeRule("EN_A_VS_AN"), sent, 8, 10, "Use 'a' instead of 'an' if the following word doesn't start with a vowel sound, e.g. 'a sentence', 'a university'.")
	m.SetSuggestedReplacements([]string{"a"})

	var buf bytes.Buffer
	n, err := CheckText(&buf, "This is an test.", fakeChecker{ms: []*rules.RuleMatch{m}})
	require.NoError(t, err)
	require.Equal(t, 1, n)
	out := buf.String()
	require.Contains(t, out, "Rule ID: EN_A_VS_AN")
	require.Contains(t, out, "premium: false")
	require.NotContains(t, out, "Type:") // Java printMatches has no Type/Severity soft lines
	require.Contains(t, out, "Message:")
	require.Contains(t, out, "Suggestion: a")
	require.Contains(t, out, "Time:")
	require.Contains(t, out, "sentences")

	// clean text → zero matches, still prints timing
	buf.Reset()
	n, err = CheckText(&buf, "All good.", fakeChecker{ms: nil})
	require.NoError(t, err)
	require.Equal(t, 0, n)
	require.Contains(t, buf.String(), "Time:")
}
