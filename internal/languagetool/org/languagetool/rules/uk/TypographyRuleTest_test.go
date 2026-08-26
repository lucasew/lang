package uk

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/TypographyRuleTest.java
//
// Java setup:
//   TypographyRule rule = new TypographyRule(TestTools.getMessages("uk"));
//   JLanguageTool lt = new JLanguageTool(Ukrainian.DEFAULT_VARIANT);
//   rule.match(lt.getAnalyzedSentence(...));
//
// Go: real UkrainianWordTokenizer + UkrainianTagger (official ukrainian.dict) via wireUKWordRepeatLT.

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestTypographyRule_Rule twins TypographyRuleTest.testRule.
func TestTypographyRule_Rule(t *testing.T) {
	// Java: new TypographyRule(TestTools.getMessages("uk"))
	rule := NewTypographyRule(nil)
	// Java: new JLanguageTool(Ukrainian.DEFAULT_VARIANT)
	lt := wireUKWordRepeatLT(t)

	// correct sentences:
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("як-небудь"))))

	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("А\u2013Т"))))

	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("ХХ\u2013ХХІ"))))

	// incorrect sentences:
	// TODO (Java): does not work when word is the last in the sentence: "яскраво\u2013рожевий"
	matches := rule.Match(lt.GetAnalyzedSentence("яскраво\u2013рожевий."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"яскраво-рожевий", "яскраво \u2014 рожевий"}, matches[0].GetSuggestedReplacements())

	// test unknown word
	matches = rule.Match(lt.GetAnalyzedSentence("яскраво\u2013шуруровий."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"яскраво-шуруровий", "яскраво \u2014 шуруровий"}, matches[0].GetSuggestedReplacements())

	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("ХХ\u2014ХХІ"))))

	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("Вовка,— волкова"))))

	matches = rule.Match(lt.GetAnalyzedSentence("Вовка\u2014Волкова."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"Вовка-Волкова", "Вовка \u2014 Волкова"}, matches[0].GetSuggestedReplacements())

	matches = rule.Match(lt.GetAnalyzedSentence("цукерок —знову низька"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"цукерок-знову", "цукерок \u2014 знову"}, matches[0].GetSuggestedReplacements())

	matches = rule.Match(lt.GetAnalyzedSentence("—знову низька"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"\u2014 знову"}, matches[0].GetSuggestedReplacements())

	matches = rule.Match(lt.GetAnalyzedSentence("знову—"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"знову \u2014"}, matches[0].GetSuggestedReplacements())

	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("\u2014 Київ, 1994"))))

	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("\u2013 Київ, 1994"))))

	matches = rule.Match(lt.GetAnalyzedSentence("важливіше \u2013потенційні"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"важливіше-потенційні", "важливіше \u2014 потенційні"}, matches[0].GetSuggestedReplacements())

	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("Рахунки 1 класу –"))))

	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("\u2013"))))

	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence(" \u2013"))))

	matches = rule.Match(lt.GetAnalyzedSentence("любили ,—люби"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()))
}

// TestTypographyRule_Metadata checks id / description / short / category (Java surfaces).
func TestTypographyRule_Metadata(t *testing.T) {
	r := NewTypographyRule(nil)
	require.Equal(t, "DASH", r.GetID())
	require.Equal(t, "Коротка риска замість дефісу", r.GetDescription())
	require.Equal(t, "Коротка риска", r.GetShort())
	require.NotNil(t, r.GetCategory())
}
