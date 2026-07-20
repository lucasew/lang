package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestEnglishForL2SpeakersFalseFriendRules(t *testing.T) {
	de := NewEnglishForGermansFalseFriendRule()
	require.Equal(t, "EN_FOR_DE_SPEAKERS_FALSE_FRIENDS", de.GetID())
	require.Equal(t, []string{"confusion_sets_l2_de.txt"}, de.GetFilenames())
	require.Equal(t, "de", de.MotherTongue)
	// Java addExamplePair: handy → phone
	require.Equal(t, []string{"phone"}, de.GetIncorrectExamples()[0].GetCorrections())
	require.Contains(t, de.ExampleWrong, "<marker>handy</marker>")

	fr := NewEnglishForFrenchFalseFriendRule()
	require.Equal(t, "EN_FOR_FR_SPEAKERS_FALSE_FRIENDS", fr.GetID())
	require.Equal(t, []string{"complete"}, fr.GetIncorrectExamples()[0].GetCorrections())

	es := NewEnglishForSpaniardsFalseFriendRule()
	require.Equal(t, "EN_FOR_ES_SPEAKERS_FALSE_FRIENDS", es.GetID())
	require.Equal(t, []string{"produce"}, es.GetIncorrectExamples()[0].GetCorrections())

	nl := NewEnglishForDutchmenFalseFriendRule()
	require.Equal(t, "EN_FOR_NL_SPEAKERS_FALSE_FRIENDS", nl.GetID())
	require.Equal(t, []string{"wall"}, nl.GetIncorrectExamples()[0].GetCorrections())
}

func TestEnglishForL2_MatchNilLM(t *testing.T) {
	r := NewEnglishForGermansFalseFriendRule()
	require.Empty(t, r.Match(languagetool.AnalyzePlain("My handy is broken.")))
}

func TestEnglishForL2_MessageFromFalseFriendRules(t *testing.T) {
	ClearL2FalseFriendRuleCache()
	r := NewEnglishForGermansFalseFriendRule()
	// Inject a minimal false-friend rule: token "handy"
	pt := patterns.Token("handy")
	ff := patterns.NewFalseFriendPatternRule("FF", "en", []*patterns.PatternToken{pt},
		"desc", `"handy" (English) means "handy" (German).`, "short")
	l2FFMu.Lock()
	l2FFRules["de"] = []*patterns.FalseFriendPatternRule{ff}
	l2FFMu.Unlock()
	defer ClearL2FalseFriendRuleCache()

	text := rules.NewConfusionString("handy", nil)
	better := rules.NewConfusionString("phone", nil)
	msg := r.l2Message(text, better)
	require.Contains(t, msg, "handy")
	require.Contains(t, msg, "German")
}

func TestEnglishForL2_BaseformMatch(t *testing.T) {
	r := NewEnglishForGermansFalseFriendRule()
	r.TagWord = func(token string) []languagetool.TokenTag {
		if token == "went" {
			return []languagetool.TokenTag{{POS: "VBD", Lemma: "go"}}
		}
		return nil
	}
	pt := patterns.NewPatternTokenBuilder().Token("go").MatchInflectedForms().Build()
	cs := rules.NewConfusionString("went", nil)
	require.True(t, r.isBaseformMatch(cs, pt))
	require.False(t, r.isBaseformMatch(cs, patterns.Token("go"))) // not inflected
}

func TestEnglishForL2_WithLMAndPair(t *testing.T) {
	lm := ngrams.FuncLanguageModel(func(tokens []string) ngrams.Probability {
		for _, tok := range tokens {
			if tok == "phone" {
				return ngrams.NewProbabilitySimple(0.9, 1.0)
			}
			if tok == "handy" {
				return ngrams.NewProbabilitySimple(0.001, 1.0)
			}
		}
		return ngrams.NewProbabilitySimple(0.1, 1.0)
	})
	r := NewEnglishForGermansFalseFriendRuleWithLM(lm)
	pair := rules.NewConfusionPairTokens("handy", "phone", 10, true)
	r.SetConfusionPair(pair)
	matches := r.Match(languagetool.AnalyzePlain("My handy is broken."))
	require.NotEmpty(t, matches)
	require.Contains(t, matches[0].GetSuggestedReplacements(), "phone")
}
