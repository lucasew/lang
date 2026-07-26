package sr

// Twin of languagetool-language-modules/sr/src/test/java/org/languagetool/rules/sr/SimpleStyleEkavianReplaceRuleTest.java
// (package org.languagetool.rules.sr imports ekavian.SimpleStyleEkavianReplaceRule).
// Full outcome matrix also lives in ekavian/SimpleStyleEkavianReplaceRuleTest_test.go;
// this file mirrors Java package placement.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/sr/ekavian"
	"github.com/stretchr/testify/require"
)

func TestSimpleStyleEkavianReplaceRule_Rule(t *testing.T) {
	// Java: new SimpleStyleEkavianReplaceRule(TestTools.getEnglishMessages())
	rule := ekavian.NewSimpleStyleEkavianReplaceRule(nil)
	require.Equal(t, "SR_EKAVIAN_SIMPLE_STYLE_REPLACE_RULE", rule.GetID())
	require.Equal(t, "Провера стилски лоших речи или израза", rule.GetDescription())
	require.Equal(t, "Стилски лоша реч тј. израз", rule.GetShort())

	// correct sentences:
	matches := rule.Match(languagetool.AnalyzePlain("Он је добар."))
	require.Equal(t, 0, len(matches))

	// incorrect sentences:
	matches = rule.Match(languagetool.AnalyzePlain("Она је дебела."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()))
	require.Equal(t, []string{"елегантно попуњена"}, matches[0].GetSuggestedReplacements())
	require.Equal(t, "Уместо израза „дебела“ било би боље да користите: елегантно попуњена.", matches[0].GetMessage())
}
