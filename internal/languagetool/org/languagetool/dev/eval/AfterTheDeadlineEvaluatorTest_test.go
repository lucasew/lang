package eval

// Twin of AfterTheDeadlineEvaluatorTest
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Port of AfterTheDeadlineEvaluatorTest.testIsExpectedErrorFound
func TestAfterTheDeadlineEvaluator_IsExpectedErrorFound(t *testing.T) {
	evaluator := NewAfterTheDeadlineEvaluator("fake")
	example := rules.NewIncorrectExample("This <marker>is is</marker> a test")
	ok, err := evaluator.IsExpectedErrorFound(example, `<results><error><string>is is</string></error></results>`)
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = evaluator.IsExpectedErrorFound(example, `<results><error><string>This is</string></error></results>`)
	require.NoError(t, err)
	require.False(t, ok)

	ok, err = evaluator.IsExpectedErrorFound(example, `<results></results>`)
	require.NoError(t, err)
	require.False(t, ok)

	ok, err = evaluator.IsExpectedErrorFound(example,
		`<results>`+
			`<error><string>foo bar</string></error>`+
			`<error><string>is is</string></error>`+
			`</results>`)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestAfterTheDeadlineEvaluator_QueryExample(t *testing.T) {
	evaluator := NewAfterTheDeadlineEvaluator("http://fake/")
	evaluator.Query = func(sentence string) (string, error) {
		require.Equal(t, "This is is a test", sentence)
		return `<results><error><string>is is</string></error></results>`, nil
	}
	ok, err := evaluator.QueryExample(rules.NewIncorrectExample("This <marker>is is</marker> a test"))
	require.NoError(t, err)
	require.True(t, ok)
}
