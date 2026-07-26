package sr

// Twin of languagetool-language-modules/sr/src/test/java/org/languagetool/rules/sr/SimpleGrammarEkavianReplaceRuleTest.java
// (package org.languagetool.rules.sr imports ekavian.SimpleGrammarEkavianReplaceRule).
// Full outcome matrix also lives in ekavian/SimpleGrammarEkavianReplaceRuleTest_test.go;
// this file mirrors Java package placement.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/sr/ekavian"
	"github.com/stretchr/testify/require"
)

func TestSimpleGrammarEkavianReplaceRule_Rule(t *testing.T) {
	// Java: new SimpleGrammarEkavianReplaceRule(TestTools.getEnglishMessages())
	rule := ekavian.NewSimpleGrammarEkavianReplaceRule(nil)
	require.Equal(t, "SR_EKAVIAN_SIMPLE_GRAMMAR_REPLACE_RULE", rule.GetID())
	require.Equal(t, "Провера граматички погрешних речи или израза", rule.GetDescription())
	require.Equal(t, "Граматички погрешна реч тј. израз", rule.GetShort())

	// correct sentences:
	matches := rule.Match(languagetool.AnalyzePlain("Данас је диван дан."))
	require.Equal(t, 0, len(matches))

	// incorrect sentences:
	matches = rule.Match(languagetool.AnalyzePlain("Син је вишљи од оца."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()))
	require.Equal(t, []string{"виши"}, matches[0].GetSuggestedReplacements())
	require.Equal(t, "Не каже се „вишљи“ него „виши“.", matches[0].GetMessage())

	matches = rule.Match(languagetool.AnalyzePlain("У то оправдано сумљам."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()))
	require.Equal(t, []string{"сумњам"}, matches[0].GetSuggestedReplacements())
	require.Equal(t, "Не каже се „сумљам“ него „сумњам“.", matches[0].GetMessage())
}
