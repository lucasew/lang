package uk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Java category / ITSIssueType metadata on UK rules.
func TestUK_RuleMetadata(t *testing.T) {
	soft := NewSimpleReplaceSoftRule(nil)
	require.Equal(t, "UK_SIMPLE_REPLACE_SOFT", soft.GetID())
	require.Equal(t, "Пошук нерекомендованих слів", soft.GetDescription())
	require.Equal(t, rules.ITSStyle, soft.GetLocQualityIssueType())
	require.NotNil(t, soft.GetCategory())
	require.Equal(t, rules.NewCategoryId("MISC"), soft.GetCategory().GetID())

	renamed := NewSimpleReplaceRenamedRule(nil)
	require.Equal(t, rules.ITSStyle, renamed.GetLocQualityIssueType())

	hidden := NewHiddenCharacterRule(nil)
	require.Equal(t, "Приховані символи: знак м’якого перенесення", hidden.GetDescription())
	require.NotNil(t, hidden.GetCategory())
	require.Equal(t, rules.NewCategoryId("MISC"), hidden.GetCategory().GetID())

	hyphen := NewMissingHyphenRule(nil)
	require.Equal(t, rules.ITSMisspelling, hyphen.GetLocQualityIssueType())

	mixed := NewMixedAlphabetsRule(nil)
	require.Equal(t, "Змішування кирилиці й латиниці", mixed.GetDescription())
	require.Equal(t, rules.NewCategoryId("MISC"), mixed.GetCategory().GetID())

	typo := NewTypographyRule(nil)
	require.Equal(t, "Коротка риска замість дефісу", typo.GetDescription())
	require.Equal(t, rules.NewCategoryId("TYPOGRAPHY"), typo.GetCategory().GetID())
}
