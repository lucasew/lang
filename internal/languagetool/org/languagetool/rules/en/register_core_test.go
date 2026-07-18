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
	// Soft invent PHRASE_REPLACE ("tot he") pack removed.

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

	// Soft invent EN_COULD_OF pack removed; official grammar.xml load is the path for that rule.
}

func TestRegisterDemoEnglishSpeller(t *testing.T) {
	lt := languagetool.NewJLanguageTool("en")
	RegisterCoreEnglishLanguageRules(lt)
	RegisterDemoEnglishSpeller(lt, DemoEnglishKnownWords(), map[string][]string{
		"teh": {"the"},
	})
	m := lt.Check("teh cat")
	// "teh" unknown; "cat" may also be unknown — at least one spelling hit with teh suggestion path
	found := false
	for _, x := range m {
		if x.RuleID == "MORFOLOGIK_RULE_EN_US" {
			found = true
			if strings.Contains(strings.ToLower(x.Message), "teh") || len(x.Suggestions) > 0 {
				// ok
			}
		}
	}
	require.True(t, found, "%+v", m)

	// known words not flagged solely for spelling
	m2 := lt.Check("hello world")
	for _, x := range m2 {
		require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", x.RuleID)
	}
}

func TestRegisterDemoEnglishTagger(t *testing.T) {
	lt := languagetool.NewJLanguageTool("en")
	RegisterDemoEnglishTagger(lt)
	sents := lt.Analyze("The cat is here")
	require.NotEmpty(t, sents)
	foundDT := false
	for _, tok := range sents[0].GetTokensWithoutWhitespace() {
		if tok.GetToken() == "The" || tok.GetToken() == "the" {
			// TagWord lowercases The → the
		}
		if strings.EqualFold(tok.GetToken(), "the") {
			rd := tok.GetReadings()
			if len(rd) > 0 && rd[0].GetPOSTag() != nil && *rd[0].GetPOSTag() == "DT" {
				foundDT = true
			}
		}
		if strings.EqualFold(tok.GetToken(), "is") {
			rd := tok.GetReadings()
			require.NotEmpty(t, rd)
			require.NotNil(t, rd[0].GetPOSTag())
			require.Equal(t, "VBZ", *rd[0].GetPOSTag())
		}
	}
	require.True(t, foundDT)
}
