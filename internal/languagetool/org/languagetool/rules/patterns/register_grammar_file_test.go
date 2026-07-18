package patterns_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestRegisterGrammarFile_SoftEN(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	// patterns → rules → languagetool → org → languagetool → internal → module root (6)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../.."))
	path := filepath.Join(root, "testdata/grammar/en-soft.xml")
	lt := languagetool.NewJLanguageTool("en")
	n, err := patterns.RegisterGrammarFile(lt, path, "en")
	require.NoError(t, err)
	require.GreaterOrEqual(t, n, 1)
	m := lt.Check("Well, your welcome to try.")
	found := false
	for _, x := range m {
		if x.RuleID == "EN_SOFT_YOUR_YOU_RE" {
			found = true
			require.Contains(t, x.Message, "you're welcome")
			if len(x.Suggestions) > 0 {
				require.Contains(t, x.Suggestions, "you're welcome")
			}
		}
	}
	require.True(t, found, "%+v", m)
}

func TestRegisterGrammarXML_Inline(t *testing.T) {
	lt := languagetool.NewJLanguageTool("en")
	xml := `<rules lang="en"><category id="G"><rule id="X"><pattern><token>foo</token><token>bar</token></pattern><message>bad <suggestion>baz</suggestion></message></rule></category></rules>`
	n, err := patterns.RegisterGrammarXML(lt, xml, "inline", "en")
	require.NoError(t, err)
	require.Equal(t, 1, n)
	m := lt.Check("say foo bar now")
	require.NotEmpty(t, m)
}

func TestRegisterGrammarFile_SoftDE(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../.."))
	path := filepath.Join(root, "testdata/grammar/de-soft.xml")
	lt := languagetool.NewJLanguageTool("de")
	n, err := patterns.RegisterGrammarFile(lt, path, "de")
	require.NoError(t, err)
	require.GreaterOrEqual(t, n, 1)
	m := lt.Check("Ich denke das es stimmt.")
	found := false
	for _, x := range m {
		if x.RuleID == "DE_SOFT_DAS_DASS" {
			found = true
		}
	}
	require.True(t, found, "%+v", m)
}

func TestRegisterGrammarXML_DefaultOff(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules lang="en">
  <category id="STYLE" name="Style">
    <rule id="EN_SOFT_OPT_TEST" name="opt" default="off">
      <pattern><token>prior</token><token>to</token></pattern>
      <message>Did you mean "before"?</message>
    </rule>
  </category>
</rules>`
	lt := languagetool.NewJLanguageTool("en")
	n, err := patterns.RegisterGrammarXML(lt, xml, "test.xml", "en")
	require.NoError(t, err)
	require.Equal(t, 1, n)
	// disabled by default
	ms := lt.Check("Prior to leaving, call.")
	for _, m := range ms {
		require.NotEqual(t, "EN_SOFT_OPT_TEST", m.RuleID)
	}
	lt.EnableRule("EN_SOFT_OPT_TEST")
	ms = lt.Check("Prior to leaving, call.")
	found := false
	for _, m := range ms {
		if m.RuleID == "EN_SOFT_OPT_TEST" {
			found = true
		}
	}
	require.True(t, found, "%+v", ms)
}
