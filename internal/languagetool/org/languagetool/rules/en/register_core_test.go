package en

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreEnglishLanguageRules_Check(t *testing.T) {
	lt := languagetool.NewJLanguageTool("en-US")
	RegisterCoreEnglishLanguageRules(lt)

	// Java English.createDefaultChunker → pre-disambig Chunker (not post-disambig).
	require.NotNil(t, lt.Chunker)
	require.Nil(t, lt.PostDisambiguationChunker)

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

	// long sentence (40+ words). Java LongSentenceRule is Tag.picky — only active at
	// Level.PICKY (isRuleActiveForLevelAndToneTags). Capitalize first word so
	// UPPERCASE_SENTENCE_START does not compete; DisableCleanOverlapping so STYLE
	// demotion does not hide the match after picky demotion in overlap filter.
	lt.Level = languagetool.LevelPicky
	lt.DisableCleanOverlapping()
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

	// en-US core registers American extras only (not invent GB/NZ replace on all locales).
	ids := map[string]struct{}{}
	for _, id := range lt.GetAllRegisteredRuleIDs() {
		ids[id] = struct{}{}
	}
	require.Contains(t, ids, "METRIC_UNITS_EN_US")
	require.Contains(t, ids, "EN_US_SIMPLE_REPLACE")
	_, hasImp := ids["METRIC_UNITS_EN_IMPERIAL"]
	_, hasGB := ids["EN_GB_SIMPLE_REPLACE"]
	_, hasNZ := ids["EN_NZ_SIMPLE_REPLACE"]
	require.False(t, hasImp, "Imperial not on en-US")
	require.False(t, hasGB, "British replace not on en-US")
	require.False(t, hasNZ, "NZ replace not on en-US")
}

func TestRegisterEnglishVariantExtraRules(t *testing.T) {
	us := languagetool.NewJLanguageTool("en-US")
	RegisterEnglishVariantExtraRules(us)
	usIDs := us.GetAllRegisteredRuleIDs()
	require.Contains(t, usIDs, "METRIC_UNITS_EN_US")
	require.Contains(t, usIDs, "EN_US_SIMPLE_REPLACE")

	gb := languagetool.NewJLanguageTool("en-GB")
	RegisterEnglishVariantExtraRules(gb)
	gbIDs := gb.GetAllRegisteredRuleIDs()
	require.Contains(t, gbIDs, "METRIC_UNITS_EN_IMPERIAL")
	require.Contains(t, gbIDs, "EN_GB_SIMPLE_REPLACE")
	require.NotContains(t, gbIDs, "EN_US_SIMPLE_REPLACE")

	nz := languagetool.NewJLanguageTool("en-NZ")
	RegisterEnglishVariantExtraRules(nz)
	nzIDs := nz.GetAllRegisteredRuleIDs()
	require.Contains(t, nzIDs, "EN_NZ_SIMPLE_REPLACE")
	require.Contains(t, nzIDs, "METRIC_UNITS_EN_IMPERIAL")

	za := languagetool.NewJLanguageTool("en-ZA")
	RegisterEnglishVariantExtraRules(za)
	for _, id := range za.GetAllRegisteredRuleIDs() {
		require.NotEqual(t, "METRIC_UNITS_EN_US", id)
		require.NotEqual(t, "METRIC_UNITS_EN_IMPERIAL", id)
		require.NotEqual(t, "EN_US_SIMPLE_REPLACE", id)
		require.NotEqual(t, "EN_GB_SIMPLE_REPLACE", id)
		require.NotEqual(t, "EN_NZ_SIMPLE_REPLACE", id)
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

// Java English.getRelevantRules base IDs must all be registered (speller/variants are extras).
func TestRegisterCoreEnglishLanguageRules_JavaRelevantBaseIDs(t *testing.T) {
	lt := languagetool.NewJLanguageTool("en-US")
	RegisterCoreEnglishLanguageRules(lt)
	// skip picky path — profanity already in main registration
	ids := lt.GetAllRegisteredRuleIDs()
	for _, id := range language.EnglishRelevantRuleIDs() {
		require.Contains(t, ids, id, "missing Java English.getRelevantRules id %s", id)
	}
	// invent SharedLayout-only extras must not reappear
	require.NotContains(t, ids, "WHITESPACE_PUNCTUATION")
}
