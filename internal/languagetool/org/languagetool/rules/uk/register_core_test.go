package uk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreRules_NoSoftInventSequences(t *testing.T) {
	// Soft invent token sequences removed; official grammar.xml is incomplete until loaded.
	lt := languagetool.NewJLanguageTool("uk")
	RegisterCoreUkrainianRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.NotContains(t, ids, "UK_В_В")
}

func TestRegisterCoreRules_JavaRelevantRules(t *testing.T) {
	lt := languagetool.NewJLanguageTool("uk")
	RegisterCoreUkrainianRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	// Ports of Ukrainian.getRelevantRules (non-XML)
	for _, id := range []string{
		TokenAgreementVerbNounRuleID,
		TokenAgreementNounVerbRuleID,
		TokenAgreementAdjNounRuleID,
		TokenAgreementPrepNounRuleID,
		TokenAgreementNumrNounRuleID,
		"UK_SIMPLE_REPLACE",
		"UK_SIMPLE_REPLACE_SOFT",
		"UK_SIMPLE_REPLACE_RENAMED",
		"UK_HIDDEN_CHARS",
		"DASH", // TypographyRule
		"UK_MIXED_ALPHABETS",
		"UK_MISSING_HYPHEN",
		"UKRAINIAN_WORD_REPEAT_RULE",
		"MORFOLOGIK_RULE_UK_UA",
		"UPPERCASE_SENTENCE_START",
		// Java CommaWhitespaceRule / UkrainianCommaWhitespaceRule getId()
		"COMMA_PARENTHESIS_WHITESPACE",
	} {
		require.Contains(t, ids, id, "missing rule %s", id)
	}
	// Java intentionally omits DoublePunctuationRule
	require.NotContains(t, ids, "DOUBLE_PUNCTUATION")
}
