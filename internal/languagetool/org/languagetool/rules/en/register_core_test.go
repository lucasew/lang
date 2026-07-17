package en

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreEnglishLanguageRules_Check(t *testing.T) {
	lt := languagetool.NewJLanguageTool("en-US")
	RegisterCoreEnglishLanguageRules(lt)

	require.NotEmpty(t, lt.Check("This is an test."))
	require.NotEmpty(t, lt.Check("hello  world"))
	// English word-repeat id
	m := lt.Check("this this")
	require.NotEmpty(t, m)
	var hasEN bool
	for _, x := range m {
		if x.RuleID == "ENGLISH_WORD_REPEAT_RULE" {
			hasEN = true
		}
	}
	require.True(t, hasEN)
	// phrase
	m = lt.Check("Guide tot he Galaxy")
	require.NotEmpty(t, m)

	// long sentence (40+ words)
	var b strings.Builder
	for i := 0; i < 45; i++ {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString("word")
	}
	b.WriteByte('.')
	m = lt.Check(b.String())
	var hasLong bool
	for _, x := range m {
		if x.RuleID == "TOO_LONG_SENTENCE" {
			hasLong = true
		}
	}
	require.True(t, hasLong, "%+v", m)
}
