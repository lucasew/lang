package patterns_test

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestRegisterTokenSequence_Check(t *testing.T) {
	lt := languagetool.NewJLanguageTool("en")
	patterns.RegisterTokenSequence(lt, "COULD_OF", "en",
		[]string{"could", "of"},
		"Did you mean 'could have'?",
		"could have",
	)
	m := lt.Check("I could of done it.")
	require.NotEmpty(t, m)
	found := false
	for _, x := range m {
		if x.RuleID == "COULD_OF" {
			found = true
			require.Contains(t, x.Suggestions, "could have")
		}
	}
	require.True(t, found, "%+v", m)
}

func TestRegisterTokenSequences(t *testing.T) {
	lt := languagetool.NewJLanguageTool("en")
	patterns.RegisterTokenSequences(lt, "en", []patterns.TokenSequenceSpec{
		{ID: "SHOULD_OF", Tokens: []string{"should", "of"}, Message: "should have?", Suggestion: "should have"},
		{ID: "WOULD_OF", Tokens: []string{"would", "of"}, Message: "would have?", Suggestion: "would have"},
	})
	require.NotEmpty(t, lt.Check("You should of told me."))
	require.NotEmpty(t, lt.Check("I would of gone."))
}
