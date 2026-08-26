package bigdata

// Twin of ContextBuilderTest
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func javaList(ss []string) string {
	return "[" + strings.Join(ss, ", ") + "]"
}

func TestContextBuilder_GetContext(t *testing.T) {
	cb := NewContextBuilder()
	check := func(input string, pos, contextSize int, expected string) {
		t.Helper()
		sent := languagetool.AnalyzePlain(input)
		var toks []string
		for _, tr := range sent.GetTokensWithoutWhitespace() {
			toks = append(toks, tr.GetToken())
		}
		got := cb.GetContext(toks, pos, contextSize)
		require.Equal(t, expected, javaList(got), "pos=%d size=%d", pos, contextSize)
	}
	check("And this is a test.", 3 /*is*/, 1, "[this, is, a]")
	check("And this is a test.", 3, 2, "[And, this, is, a, test]")
	check("And this is a test.", 3, 3, "[_START_, And, this, is, a, test, .]")
	check("And this is a test.", 3, 4, "[_START_, And, this, is, a, test, ., _END_]")

	check("This", 1, 0, "[This]")
	check("This", 1, 1, "[_START_, This, _END_]")
	check("This", 1, 2, "[_START_, This, _END_]")
}
