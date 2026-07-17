package patterns

// PatternRule wired into JLanguageTool.Check via rules.AsSentenceChecker.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestPatternRule_CheckAdapter(t *testing.T) {
	pr := NewPatternRule("FOO_BAR", "en",
		[]*PatternToken{
			NewPatternTokenBuilder().Token("foo").Build(),
			NewPatternTokenBuilder().Token("bar").Build(),
		},
		"foo bar", "Avoid foo bar", "foo bar")
	lt := languagetool.NewJLanguageTool("en")
	lt.AddRuleChecker(pr.GetID(), rules.AsSentenceChecker(pr.Match))
	require.NotEmpty(t, lt.Check("say foo bar now"))
	require.Empty(t, lt.Check("say foo baz now"))
}
