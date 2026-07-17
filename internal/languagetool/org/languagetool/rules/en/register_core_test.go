package en

import (
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
}
