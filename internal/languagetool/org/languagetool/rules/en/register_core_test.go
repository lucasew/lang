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

	// long sentence (40+ words). Capitalize the first word so UPPERCASE_SENTENCE_START
	// does not fire: Java LongSentenceRule is Tag.picky, so CleanOverlappingFilter demotes
	// it below non-picky layout rules that share the span (faithful, not invent).
	var b strings.Builder
	for i := 0; i < 45; i++ {
		if i > 0 {
			b.WriteByte(' ')
		}
		if i == 0 {
			b.WriteString("Word")
		} else {
			b.WriteString("word")
		}
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

	// en-US core registers US unit conversion (not picky invent of both US+Imperial).
	ids := map[string]struct{}{}
	for _, id := range lt.GetAllRegisteredRuleIDs() {
		ids[id] = struct{}{}
	}
	_, hasUS := ids["METRIC_UNITS_EN_US"]
	_, hasImp := ids["METRIC_UNITS_EN_IMPERIAL"]
	require.True(t, hasUS, "AmericanEnglish unit conversion")
	require.False(t, hasImp, "Imperial not on en-US")
}

func TestRegisterEnglishVariantExtraRules(t *testing.T) {
	us := languagetool.NewJLanguageTool("en-US")
	RegisterEnglishVariantExtraRules(us)
	require.Contains(t, us.GetAllRegisteredRuleIDs(), "METRIC_UNITS_EN_US")

	gb := languagetool.NewJLanguageTool("en-GB")
	RegisterEnglishVariantExtraRules(gb)
	require.Contains(t, gb.GetAllRegisteredRuleIDs(), "METRIC_UNITS_EN_IMPERIAL")

	za := languagetool.NewJLanguageTool("en-ZA")
	RegisterEnglishVariantExtraRules(za)
	for _, id := range za.GetAllRegisteredRuleIDs() {
		require.NotEqual(t, "METRIC_UNITS_EN_US", id)
		require.NotEqual(t, "METRIC_UNITS_EN_IMPERIAL", id)
	}
}

func TestRegisterPickyEnglishRules_OnlyProfanity(t *testing.T) {
	lt := languagetool.NewJLanguageTool("en-US")
	RegisterPickyEnglishRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.Equal(t, []string{"PROFANITY"}, ids)
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
		// Exact surface "The" / "the" both listed in DemoEnglishTagWord (no lowercase invent).
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
