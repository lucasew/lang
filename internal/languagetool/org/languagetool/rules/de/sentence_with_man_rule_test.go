package de

// Twin of org.languagetool.rules.de.SentenceWithManRule
// (no Java *Test class — behavior matrix + real de-DE analysis when GermanTagger supplies lemma "man").

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	detag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/de"
	"github.com/stretchr/testify/require"
)

func TestSentenceWithManRule_Meta(t *testing.T) {
	r := NewSentenceWithManRule(nil)
	require.Equal(t, "SENTENCE_WITH_MAN_DE", r.GetID())
	require.Equal(t, sentenceWithManDefaultMinPercent, r.MinPercent)
	require.Equal(t, 15, r.MinPercent)
	require.Equal(t, float64(1000), r.Denominator)
	require.True(t, r.ExcludeDirectSpeech)
	require.True(t, r.IsDefaultOff())
	require.Equal(t, "Statistische Stilanalyse: Sätze mit indirekter Leseransprache 'man'", r.GetDescription())
	require.Equal(t, sentenceWithManConfigureText, r.GetConfigureText())
	require.Equal(t,
		"Anzeigen wenn mehr als ...‰ Sätze eines Kapitels 'man' enthalten:",
		r.GetConfigureText())
	require.NotNil(t, r.GetCategory())
	require.True(t, r.GetCategory().GetID().Equals(rules.NewCategoryId("CREATIVE_WRITING")))
	require.Equal(t, rules.ITSStyle, r.GetLocQualityIssueType())

	opts := r.GetRuleOptions()
	require.Len(t, opts, 1)
	require.Equal(t, 15, opts[0].DefaultValue)
	require.Equal(t, sentenceWithManConfigureText, opts[0].ConfigureText)
	require.Equal(t, 0, opts[0].MinConfigurableValue)
	require.Equal(t, 100, opts[0].MaxConfigurableValue)
}

func TestSentenceWithManRule_LimitMessage(t *testing.T) {
	r := NewSentenceWithManRule(nil)
	require.Equal(t,
		"Sätze mit der indirekten Leseransprache 'man' sind stilistisch wenig elegant formuliert. Lässt sich das Wort vermeiden?",
		r.getLimitMessage(0, 0))
	require.Equal(t,
		"Mehr als 15‰ Sätze mit der indirekten Leseransprache 'man' {50‰} gefunden. Lässt sich das Wort vermeiden?",
		r.getLimitMessage(15, 50))
	// Java: (int)(percent + 0.5d) rounding
	require.Equal(t,
		"Mehr als 15‰ Sätze mit der indirekten Leseransprache 'man' {51‰} gefunden. Lässt sich das Wort vermeiden?",
		r.getLimitMessage(15, 50.6))
}

func TestSentenceWithManRule_UserConfigMinPercent(t *testing.T) {
	uc := languagetool.NewUserConfig()
	uc.SetConfigValueByID("SENTENCE_WITH_MAN_DE", []any{7})
	r := NewSentenceWithManRuleWithUserConfig(nil, uc)
	require.Equal(t, 7, r.MinPercent)

	r0 := NewSentenceWithManRuleWithUserConfig(nil, nil)
	require.Equal(t, 15, r0.MinPercent)

	rMin := NewSentenceWithManRuleWithMinPercent(nil, 0)
	require.Equal(t, 0, rMin.MinPercent)
}

func manMorphSent(toks ...*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedSentence {
	return languagetool.NewAnalyzedSentence(withPositions(toks...))
}

func TestSentenceWithManRule_ConditionMatrix(t *testing.T) {
	rule := NewSentenceWithManRuleWithMinPercent(nil, 0)

	// hasLemma("man") → hit (returns man token)
	manSent := manMorphSent(
		sentStartATR(),
		atrWithPOS("Man", "PRO:IND:NOM:SIN:MAS", "man"),
		atrWithPOS("sollte", "VER:MOD:3:SIN:PRT:SFT", "sollen"),
		atrWithPOS("das", "PRO:DEM:AKK:SIN:NEU", "das"),
		atrWithPOS("vermeiden", "VER:INF:NON", "vermeiden"),
		atrWithPOS(".", "PKT", "."),
	)
	matches := rule.Match(manSent)
	require.Equal(t, 1, len(matches))
	require.Contains(t, matches[0].GetMessage(), "indirekten Leseransprache")

	// lower-case surface with lemma man
	manLower := manMorphSent(
		sentStartATR(),
		atrWithPOS("man", "PRO:IND:NOM:SIN:MAS", "man"),
		atrWithPOS("sagt", "VER:3:SIN:PRS:NON", "sagen"),
		atrWithPOS(".", "PKT", "."),
	)
	require.Equal(t, 1, len(rule.Match(manLower)))

	// no man lemma
	erSent := manMorphSent(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("sollte", "VER:MOD:3:SIN:PRT:SFT", "sollen"),
		atrWithPOS("das", "PRO:DEM:AKK:SIN:NEU", "das"),
		atrWithPOS("vermeiden", "VER:INF:NON", "vermeiden"),
		atrWithPOS(".", "PKT", "."),
	)
	require.Equal(t, 0, len(rule.Match(erSent)))

	// surface "man" with wrong lemma must not invent a hit
	wrongLemma := manMorphSent(
		sentStartATR(),
		atrWithPOS("Man", "SUB:NOM:SIN:NEU", "Man"), // proper-noun-ish lemma, not "man"
		atrWithPOS("ist", "VER:3:SIN:PRS:SFT", "sein"),
		atrWithPOS(".", "PKT", "."),
	)
	require.Equal(t, 0, len(rule.Match(wrongLemma)))

	// untagged AnalyzePlain must not invent hits
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Man sollte das vermeiden."))))
}

func TestSentenceWithManRule_ConditionFulfilledDirect(t *testing.T) {
	r := NewSentenceWithManRuleWithMinPercent(nil, 0)
	manTok := atrWithPOS("Man", "PRO:IND:NOM:SIN:MAS", "man")
	got := r.conditionFulfilled([]*languagetool.AnalyzedTokenReadings{
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		manTok,
		atrWithPOS(".", "PKT", "."),
	})
	require.Same(t, manTok, got)

	// first man wins
	man1 := atrWithPOS("man", "PRO:IND:NOM:SIN:MAS", "man")
	man2 := atrWithPOS("man", "PRO:IND:NOM:SIN:MAS", "man")
	got2 := r.conditionFulfilled([]*languagetool.AnalyzedTokenReadings{man1, man2})
	require.Same(t, man1, got2)

	require.Nil(t, r.conditionFulfilled([]*languagetool.AnalyzedTokenReadings{
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("geht", "VER:3:SIN:PRS:NON", "gehen"),
	}))
}

func TestSentenceWithManRule_PercentThreshold(t *testing.T) {
	// Two sentences, one with man → 1*1000/2 = 500‰. Java: percent > minPercent.
	man := manMorphSent(
		sentStartATR(),
		atrWithPOS("Man", "PRO:IND:NOM:SIN:MAS", "man"),
		atrWithPOS("sagt", "VER:3:SIN:PRS:NON", "sagen"),
		atrWithPOS(".", "PKT", "."),
	)
	other := manMorphSent(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("geht", "VER:3:SIN:PRS:NON", "gehen"),
		atrWithPOS(".", "PKT", "."),
	)
	sents := []*languagetool.AnalyzedSentence{man, other}

	// 500‰ > 15 default → hit with threshold message (‰ not %)
	rDef := NewSentenceWithManRule(nil)
	ms := rDef.MatchList(sents)
	require.Equal(t, 1, len(ms))
	require.Contains(t, ms[0].GetMessage(), "Mehr als 15‰")
	require.Contains(t, ms[0].GetMessage(), "500‰")

	// 500‰ > 499 → hit
	r499 := NewSentenceWithManRuleWithMinPercent(nil, 499)
	require.Equal(t, 1, len(r499.MatchList(sents)))

	// 500‰ > 500 is false → no hit
	r500 := NewSentenceWithManRuleWithMinPercent(nil, 500)
	require.Equal(t, 0, len(r500.MatchList(sents)))

	// minPercent 0 shows all with limit==0 message
	r0 := NewSentenceWithManRuleWithMinPercent(nil, 0)
	ms0 := r0.MatchList(sents)
	require.Equal(t, 1, len(ms0))
	require.Equal(t,
		"Sätze mit der indirekten Leseransprache 'man' sind stilistisch wenig elegant formuliert. Lässt sich das Wort vermeiden?",
		ms0[0].GetMessage())
}

func setUpSentenceWithManRule(t *testing.T) (*SentenceWithManRule, *languagetool.JLanguageTool, *detag.GermanTagger) {
	t.Helper()
	lt := languagetool.NewJLanguageTool("de-DE")
	require.True(t, WireGermanTagWord(lt), "GermanTagger resources required for SentenceWithManRule twin")
	tagger := DiscoveredGermanTagger()
	require.NotNil(t, tagger, "GermanTagger.INSTANCE resources required")
	// show-all for real-analysis twin (minPercent=0)
	rule := NewSentenceWithManRuleWithMinPercent(nil, 0)
	return rule, lt, tagger
}

func TestSentenceWithManRule_RealGermanTagger(t *testing.T) {
	rule, lt, tagger := setUpSentenceWithManRule(t)

	// "Man" should get lemma "man" from GermanTagger
	manHit := analyzeTextWithGermanTagger(lt, tagger, "Man sollte das vermeiden.")
	require.Equal(t, 1, len(rule.MatchList(manHit)))

	// no man
	none := analyzeTextWithGermanTagger(lt, tagger, "Er sollte das vermeiden.")
	require.Equal(t, 0, len(rule.MatchList(none)))

	// mid-sentence man
	mid := analyzeTextWithGermanTagger(lt, tagger, "Dann sagt man das.")
	require.Equal(t, 1, len(rule.MatchList(mid)))
}
