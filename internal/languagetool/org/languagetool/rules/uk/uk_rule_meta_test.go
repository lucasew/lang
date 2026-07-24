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

func TestUK_SimpleReplaceMeta(t *testing.T) {
	r := NewSimpleReplaceRule(nil)
	require.Equal(t, "UK_SIMPLE_REPLACE", r.GetID())
	require.Equal(t, "Пошук помилкових слів", r.GetDescription())
	require.NotNil(t, r.GetCategory())
	require.Equal(t, rules.NewCategoryId("MISC"), r.GetCategory().GetID())
}

func TestUK_PunctuationCheckMeta(t *testing.T) {
	r := NewPunctuationCheckRule(nil)
	require.Equal(t, "PUNCTUATION_GENERIC_CHECK", r.GetID())
	require.Equal(t, "Use of unusual combination of punctuation characters", r.GetDescription())
	require.Equal(t, rules.NewCategoryId("PUNCTUATION"), r.GetCategory().GetID())
}

func TestUK_FractionalNumrPatterns(t *testing.T) {
	// Java String.matches surfaces
	require.True(t, IsFractionalNumrException(atr("півтора", "numr"), nil))
	require.True(t, IsFractionalNumrException(atr("один-півтора", "numr"), nil))
	require.True(t, IsFractionalNumrException(atr("півтори", "numr"), nil))
	require.True(t, IsFractionalNumrException(atr("пів", "numr"), nil))
	require.True(t, IsFractionalNumrException(atr("1,5", "number"), nil))
	require.False(t, IsFractionalNumrException(atr("три", "numr"), nil))
}

func TestUK_TokenAgreementMetadata(t *testing.T) {
	adj := NewTokenAgreementAdjNounRule()
	require.Equal(t, "UK_ADJ_NOUN_INFLECTION_AGREEMENT", adj.GetID())
	require.Equal(t, "Узгодження відмінків, роду і числа прикметника та іменника", adj.GetDescription())
	require.Equal(t, "Узгодження прикметника та іменника", adj.GetShort())
	require.Equal(t, rules.NewCategoryId("MISC"), adj.GetCategory().GetID())

	nv := NewTokenAgreementNounVerbRule()
	require.Equal(t, "Узгодження іменника та дієслова за родом, числом та особою", nv.GetDescription())
	require.Equal(t, "Узгодження іменника з дієсловом", nv.GetShort())
	require.Equal(t, rules.NewCategoryId("MISC"), nv.GetCategory().GetID())

	numr := NewTokenAgreementNumrNounRule()
	require.Equal(t, "Узгодження відмінків, роду і числа числівника та іменника", numr.GetDescription())
	require.Equal(t, "Узгодження числівника та іменника", numr.GetShort())

	prep := NewTokenAgreementPrepNounRule()
	require.Equal(t, "UK_PREP_NOUN_INFLECTION_AGREEMENT", prep.GetID())
	require.Equal(t, "Узгодження прийменника та іменника у реченні", prep.GetDescription())
	require.Equal(t, "Узгодження прийменника та іменника", prep.GetShort())
	require.Equal(t, rules.NewCategoryId("MISC"), prep.GetCategory().GetID())

	vn := NewTokenAgreementVerbNounRule()
	require.Equal(t, "Узгодження дієслова з іменником", vn.GetDescription())
	require.Equal(t, "Узгодження дієслова з іменником", vn.GetShort())
	require.Equal(t, rules.NewCategoryId("MISC"), vn.GetCategory().GetID())
}
