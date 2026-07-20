package de

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreGermanRules_Check(t *testing.T) {
	lt := languagetool.NewJLanguageTool("de-DE")
	RegisterCoreGermanRules(lt)

	require.Empty(t, lt.Check("Ein Test, der keine Fehler geben sollte."))
	// word repeat (German rule id)
	m := lt.Check("Ein Test Test, der Fehler geben sollte.")
	require.NotEmpty(t, m)
	var hasRepeat bool
	for _, x := range m {
		if x.RuleID == "GERMAN_WORD_REPEAT_RULE" || x.RuleID == "WORD_REPEAT_RULE" {
			hasRepeat = true
		}
	}
	require.True(t, hasRepeat)

	// multi whitespace
	require.NotEmpty(t, lt.Check("Hallo  Welt"))

	// double punct
	require.NotEmpty(t, lt.Check("Warte.. jetzt"))
}

func TestRegisterCoreGermanRules_TextLevel(t *testing.T) {
	lt := languagetool.NewJLanguageTool("de")
	RegisterCoreGermanRules(lt)
	// three successive "Auch" starts
	m := lt.Check("Auch heute. Auch morgen. Auch übermorgen.")
	found := false
	for _, x := range m {
		if x.RuleID == "GERMAN_WORD_REPEAT_BEGINNING_RULE" {
			found = true
		}
	}
	require.True(t, found, "%+v", m)

	// long sentence (41+ words). Java LongSentenceRule is Tag.picky — only active at
	// Level.PICKY (isRuleActiveForLevelAndToneTags). Disable overlapping cleanup so
	// STYLE/shorter matches do not hide TOO_LONG_SENTENCE_DE after demotion.
	lt.Level = languagetool.LevelPicky
	lt.DisableCleanOverlapping()
	cycle := []string{"Eins", "zwei", "drei", "vier", "fünf", "sechs", "sieben", "acht", "neun", "zehn"}
	var b strings.Builder
	for i := 0; i < 45; i++ {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(cycle[i%len(cycle)])
	}
	b.WriteByte('.')
	m2 := lt.Check(b.String())
	foundLS := false
	for _, x := range m2 {
		if x.RuleID == "TOO_LONG_SENTENCE_DE" {
			foundLS = true
		}
	}
	require.True(t, foundLS, "%+v", m2)
}

func TestRegisterCoreGermanRules_NoSoftWegenInvent(t *testing.T) {
	// Soft DE_WEGEN_DEM invent removed; grammar.xml rules are incomplete without XML load.
	lt := languagetool.NewJLanguageTool("de")
	RegisterCoreGermanRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.NotContains(t, ids, "DE_WEGEN_DEM")
	require.NotContains(t, ids, "DE_TROTZ_DEM")
	// Must not invent a soft hit for wegen dem (official rule needs grammar.xml).
	for _, x := range lt.Check("Das war wegen dem Wetter.") {
		require.NotEqual(t, "DE_WEGEN_DEM", x.RuleID)
	}
}

func TestRegisterCoreGermanRules_AgreementIDsRegistered(t *testing.T) {
	lt := languagetool.NewJLanguageTool("de")
	RegisterCoreGermanRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	want := []string{
		"DE_AGREEMENT", "DE_AGREEMENT2", "DE_SUBJECT_VERB_AGREEMENT",
		"DE_VERBAGREEMENT", "MISSING_VERB",
		"COMMA_IN_FRONT_RELATIVE_CLAUSE", "COMMA_BEHIND_RELATIVE_CLAUSE",
		"COMPOUND_INFINITIV_RULE",
	}
	for _, id := range want {
		require.Contains(t, ids, id)
	}
	// MISSING_VERB + style statistic rules are default-off in Java
	off := lt.GetDefaultOffRuleIDs()
	require.Contains(t, off, "MISSING_VERB")
	require.Contains(t, off, "PASSIVE_SENTENCE_DE")
	require.Contains(t, off, "NON_SIGNIFICANT_VERB_DE")
	require.Contains(t, off, "REDUNDANT_MODAL_VERB")
	require.Contains(t, off, "SENTENCE_WITH_MAN_DE")
	require.Contains(t, off, "SENTENCE_WITH_MODAL_VERB_DE")
	require.Contains(t, off, "SENTENCE_BEGINNING_WITH_CONJUNCTION_DE")
	require.Contains(t, off, "STYLE_REPEATED_WORD_RULE_DE")
	require.Contains(t, off, "TOO_OFTEN_USED_NOUN_DE")
	require.Contains(t, off, "TOO_OFTEN_USED_VERB_DE")
	require.Contains(t, off, "TOO_OFTEN_USED_ADJECTIVE_DE")
	require.Contains(t, off, "STYLE_REPEATED_SENTENCE_BEGINNING")
	require.Contains(t, off, "READABILITY_RULE_SIMPLE_DE")
	require.Contains(t, off, "READABILITY_RULE_DIFFICULT_DE")
	require.Contains(t, off, "DE_UPPER_CASE_NGRAM")
	require.Contains(t, ids, "DE_DU_UPPER_LOWER")
	require.Contains(t, ids, "DE_WIEDER_VS_WIDER")
	require.Contains(t, ids, "GERMAN_PARAGRAPH_REPEAT_BEGINNING_RULE")
	require.Contains(t, off, "DE_SIMILAR_NAMES")
	require.Contains(t, off, "UNNECESSARY_PHRASES_DE")
	require.Contains(t, ids, "UNPAIRED_BRACKETS")
	require.Contains(t, ids, "DE_UNPAIRED_QUOTES")
	require.Contains(t, ids, "OLD_SPELLING_RULE")
	require.Contains(t, ids, "REDUNDANT_MODAL_VERB")
	require.Contains(t, ids, "DE_PROHIBITED_COMPOUNDS")
	require.Contains(t, ids, "DE_SENTENCE_WHITESPACE")
	require.Contains(t, ids, "DE_DOUBLE_PUNCTUATION")
	require.Contains(t, ids, "DE_CONFUSION_RULE")
	require.Contains(t, off, "FILLER_WORDS_DE")
	require.Contains(t, off, "STYLE_REPEATED_SHORT_SENTENCES")
	// de (Germany): German compounds + GERMAN_SPELLER_RULE
	require.Contains(t, ids, "DE_COMPOUNDS")
	require.Contains(t, ids, "GERMAN_SPELLER_RULE")
	require.NotContains(t, ids, "DE_CH_COMPOUNDS")
	// Java German.getRelevantRules: DE_* layout only — not core twins / not WHITESPACE_PUNCTUATION.
	require.NotContains(t, ids, "SENTENCE_WHITESPACE")
	require.NotContains(t, ids, "DOUBLE_PUNCTUATION")
	require.NotContains(t, ids, "WHITESPACE_PUNCTUATION")
	require.NotContains(t, ids, "PARAGRAPH_REPEAT_BEGINNING_RULE")
	require.Contains(t, ids, "UPPERCASE_SENTENCE_START")
	// Java CommaWhitespaceRule id is COMMA_PARENTHESIS_WHITESPACE.
	require.Contains(t, ids, "COMMA_PARENTHESIS_WHITESPACE")
	// Java default-off layout (setDefaultOff) must stay off until EnableRule.
	require.Contains(t, off, "EMPTY_LINE")
	require.Contains(t, off, "TOO_LONG_PARAGRAPH")
	require.Contains(t, off, "WHITESPACE_PARAGRAPH")
	require.Contains(t, off, "WHITESPACE_PARAGRAPH_BEGIN")
	require.Contains(t, off, "PUNCTUATION_PARAGRAPH_END")
	require.Contains(t, off, "GERMAN_PARAGRAPH_REPEAT_BEGINNING_RULE")
}

func TestRegisterCoreGermanRules_VariantAT(t *testing.T) {
	lt := languagetool.NewJLanguageTool("de-AT")
	RegisterCoreGermanRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.Contains(t, ids, "AUSTRIAN_GERMAN_SPELLER_RULE")
	require.Contains(t, ids, "DE_COMPOUNDS")
	require.NotContains(t, ids, "DE_CH_COMPOUNDS")
	require.NotContains(t, ids, "SWISS_GERMAN_SPELLER_RULE")
}

func TestRegisterCoreGermanRules_VariantCH(t *testing.T) {
	lt := languagetool.NewJLanguageTool("de-CH")
	RegisterCoreGermanRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.Contains(t, ids, "SWISS_GERMAN_SPELLER_RULE")
	require.Contains(t, ids, "DE_CH_COMPOUNDS")
	require.NotContains(t, ids, "DE_COMPOUNDS")
	require.NotContains(t, ids, "AUSTRIAN_GERMAN_SPELLER_RULE")
}

func TestGermanVariant(t *testing.T) {
	require.Equal(t, "AT", germanVariant("de-AT"))
	require.Equal(t, "CH", germanVariant("de-CH"))
	require.Equal(t, "DE", germanVariant("de"))
	require.Equal(t, "DE", germanVariant("de-DE"))
}
