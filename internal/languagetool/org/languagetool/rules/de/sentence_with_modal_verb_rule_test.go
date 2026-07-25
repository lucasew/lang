package de

// Twin of org.languagetool.rules.de.SentenceWithModalVerbRule
// (no Java *Test class — behavior matrix + real de-DE analysis when GermanTagger supplies VER:MOD+VER:INF).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	detag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/de"
	"github.com/stretchr/testify/require"
)

func TestSentenceWithModalVerbRule_Meta(t *testing.T) {
	r := NewSentenceWithModalVerbRule(nil)
	require.Equal(t, "SENTENCE_WITH_MODAL_VERB_DE", r.GetID())
	require.Equal(t, sentenceWithModalDefaultMinPercent, r.MinPercent)
	require.Equal(t, 18, r.MinPercent)
	require.Equal(t, float64(100), r.Denominator)
	require.True(t, r.ExcludeDirectSpeech)
	require.True(t, r.IsDefaultOff())
	require.Equal(t, "Statistische Stilanalyse: Sätze mit Modalverben", r.GetDescription())
	require.Equal(t, sentenceWithModalConfigureText, r.GetConfigureText())
	require.Equal(t,
		"Anzeigen wenn mehr als ...% Sätze eines Kapitels Modalverben enthalten:",
		r.GetConfigureText())
	require.NotNil(t, r.GetCategory())
	require.True(t, r.GetCategory().GetID().Equals(rules.NewCategoryId("CREATIVE_WRITING")))
	require.Equal(t, rules.ITSStyle, r.GetLocQualityIssueType())

	opts := r.GetRuleOptions()
	require.Len(t, opts, 1)
	require.Equal(t, 18, opts[0].DefaultValue)
	require.Equal(t, sentenceWithModalConfigureText, opts[0].ConfigureText)
	require.Equal(t, 0, opts[0].MinConfigurableValue)
	require.Equal(t, 100, opts[0].MaxConfigurableValue)
}

func TestSentenceWithModalVerbRule_LimitMessage(t *testing.T) {
	r := NewSentenceWithModalVerbRule(nil)
	require.Equal(t,
		"Modalverb: Modalverben blähen den Text häufig auf und sollten vermieden werden.",
		r.getLimitMessage(0, 0))
	require.Equal(t,
		"Mehr als 18% Sätze mit Modalverben {50%} gefunden. Modalverben blähen den Text häufig auf und sollten vermieden werden.",
		r.getLimitMessage(18, 50))
	// Java: (int)(percent + 0.5d) rounding
	require.Equal(t,
		"Mehr als 18% Sätze mit Modalverben {51%} gefunden. Modalverben blähen den Text häufig auf und sollten vermieden werden.",
		r.getLimitMessage(18, 50.6))
}

func TestSentenceWithModalVerbRule_UserConfigMinPercent(t *testing.T) {
	uc := languagetool.NewUserConfig()
	uc.SetConfigValueByID("SENTENCE_WITH_MODAL_VERB_DE", []any{7})
	r := NewSentenceWithModalVerbRuleWithUserConfig(nil, uc)
	require.Equal(t, 7, r.MinPercent)

	r0 := NewSentenceWithModalVerbRuleWithUserConfig(nil, nil)
	require.Equal(t, 18, r0.MinPercent)

	rMin := NewSentenceWithModalVerbRuleWithMinPercent(nil, 0)
	require.Equal(t, 0, rMin.MinPercent)
}

func modalMorphSent(toks ...*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedSentence {
	return languagetool.NewAnalyzedSentence(withPositions(toks...))
}

func TestSentenceWithModalVerbRule_ConditionMatrix(t *testing.T) {
	rule := NewSentenceWithModalVerbRuleWithMinPercent(nil, 0)

	// MOD then INF → hit (returns MOD token)
	modThenInf := modalMorphSent(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("muss", "VER:MOD:3:SIN:PRS:SFT", "müssen"),
		atrWithPOS("das", "PRO:DEM:AKK:SIN:NEU", "das"),
		atrWithPOS("erledigen", "VER:INF:NON", "erledigen"),
		atrWithPOS(".", "PKT", "."),
	)
	matches := rule.Match(modThenInf)
	require.Equal(t, 1, len(matches))
	require.Contains(t, matches[0].GetMessage(), "Modalverb:")

	// INF then MOD → hit (returns MOD token)
	infThenMod := modalMorphSent(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("wird", "VER:AUX:3:SIN:PRS:NON", "werden"),
		atrWithPOS("das", "PRO:DEM:AKK:SIN:NEU", "das"),
		atrWithPOS("erledigen", "VER:INF:NON", "erledigen"),
		atrWithPOS("müssen", "VER:MOD:INF", "müssen"),
		atrWithPOS(".", "PKT", "."),
	)
	require.Equal(t, 1, len(rule.Match(infThenMod)))

	// mark between MOD and INF aborts
	modMarkInf := modalMorphSent(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("muss", "VER:MOD:3:SIN:PRS:SFT", "müssen"),
		atrWithPOS(",", ",", ","),
		atrWithPOS("erledigen", "VER:INF:NON", "erledigen"),
		atrWithPOS(".", "PKT", "."),
	)
	require.Equal(t, 0, len(rule.Match(modMarkInf)))

	// mark between INF and MOD aborts
	infMarkMod := modalMorphSent(
		sentStartATR(),
		atrWithPOS("erledigen", "VER:INF:NON", "erledigen"),
		atrWithPOS(",", ",", ","),
		atrWithPOS("müssen", "VER:MOD:INF", "müssen"),
		atrWithPOS(".", "PKT", "."),
	)
	require.Equal(t, 0, len(rule.Match(infMarkMod)))

	// modal alone is not enough (period is mark — no INF)
	modalOnly := modalMorphSent(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("muss", "VER:MOD:3:SIN:PRS:SFT", "müssen"),
		atrWithPOS("das", "PRO:DEM:AKK:SIN:NEU", "das"),
		atrWithPOS(".", "PKT", "."),
	)
	require.Equal(t, 0, len(rule.Match(modalOnly)))

	// INF alone is not enough
	infOnly := modalMorphSent(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("will", "VER:3:SIN:PRS:NON", "wollen"), // not VER:MOD in this fixture
		atrWithPOS("das", "PRO:DEM:AKK:SIN:NEU", "das"),
		atrWithPOS("erledigen", "VER:INF:NON", "erledigen"),
		atrWithPOS(".", "PKT", "."),
	)
	require.Equal(t, 0, len(rule.Match(infOnly)))

	// no modal / no inf
	plain := modalMorphSent(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("erledigt", "VER:3:SIN:PRS:NON", "erledigen"),
		atrWithPOS("das", "PRO:DEM:AKK:SIN:NEU", "das"),
		atrWithPOS(".", "PKT", "."),
	)
	require.Equal(t, 0, len(rule.Match(plain)))

	// untagged AnalyzePlain must not invent hits
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Er muss das erledigen."))))
}

func TestSentenceWithModalVerbRule_PercentThreshold(t *testing.T) {
	// Two sentences, one modal+inf → 50%. Java: percent > minPercent.
	modal := modalMorphSent(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("muss", "VER:MOD:3:SIN:PRS:SFT", "müssen"),
		atrWithPOS("das", "PRO:DEM:AKK:SIN:NEU", "das"),
		atrWithPOS("erledigen", "VER:INF:NON", "erledigen"),
		atrWithPOS(".", "PKT", "."),
	)
	other := modalMorphSent(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("geht", "VER:3:SIN:PRS:NON", "gehen"),
		atrWithPOS(".", "PKT", "."),
	)
	sents := []*languagetool.AnalyzedSentence{modal, other}

	// 50% > 18 default → hit with threshold message
	rDef := NewSentenceWithModalVerbRule(nil)
	ms := rDef.MatchList(sents)
	require.Equal(t, 1, len(ms))
	require.Contains(t, ms[0].GetMessage(), "Mehr als 18%")
	require.Contains(t, ms[0].GetMessage(), "50%")

	// 50% > 49 → hit
	r49 := NewSentenceWithModalVerbRuleWithMinPercent(nil, 49)
	require.Equal(t, 1, len(r49.MatchList(sents)))

	// 50% > 50 is false → no hit
	r50 := NewSentenceWithModalVerbRuleWithMinPercent(nil, 50)
	require.Equal(t, 0, len(r50.MatchList(sents)))

	// minPercent 0 shows all with limit==0 message
	r0 := NewSentenceWithModalVerbRuleWithMinPercent(nil, 0)
	ms0 := r0.MatchList(sents)
	require.Equal(t, 1, len(ms0))
	require.Equal(t,
		"Modalverb: Modalverben blähen den Text häufig auf und sollten vermieden werden.",
		ms0[0].GetMessage())
}

func TestSentenceWithModalVerbRule_ConditionFulfilledDirect(t *testing.T) {
	// Unit: conditionFulfilled return token / mark abort / i++ reuse exhaust
	r := NewSentenceWithModalVerbRuleWithMinPercent(nil, 0)
	// MOD … INF → MOD token
	mod := atrWithPOS("muss", "VER:MOD:3:SIN:PRS:SFT", "müssen")
	inf := atrWithPOS("erledigen", "VER:INF:NON", "erledigen")
	got := r.conditionFulfilled([]*languagetool.AnalyzedTokenReadings{
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		mod,
		atrWithPOS("das", "PRO:DEM:AKK:SIN:NEU", "das"),
		inf,
		atrWithPOS(".", "PKT", "."),
	})
	require.Same(t, mod, got)

	// INF … MOD → MOD token
	mod2 := atrWithPOS("müssen", "VER:MOD:INF", "müssen")
	inf2 := atrWithPOS("erledigen", "VER:INF:NON", "erledigen")
	got2 := r.conditionFulfilled([]*languagetool.AnalyzedTokenReadings{
		inf2,
		mod2,
	})
	require.Same(t, mod2, got2)

	// mark aborts to null
	require.Nil(t, r.conditionFulfilled([]*languagetool.AnalyzedTokenReadings{
		atrWithPOS("muss", "VER:MOD:3:SIN:PRS:SFT", "müssen"),
		atrWithPOS(",", ",", ","),
		atrWithPOS("erledigen", "VER:INF:NON", "erledigen"),
	}))
}

func setUpSentenceWithModalVerbRule(t *testing.T) (*SentenceWithModalVerbRule, *languagetool.JLanguageTool, *detag.GermanTagger) {
	t.Helper()
	lt := languagetool.NewJLanguageTool("de-DE")
	require.True(t, WireGermanTagWord(lt), "GermanTagger resources required for SentenceWithModalVerbRule twin")
	tagger := DiscoveredGermanTagger()
	require.NotNil(t, tagger, "GermanTagger.INSTANCE resources required")
	// show-all for real-analysis twin (minPercent=0)
	rule := NewSentenceWithModalVerbRuleWithMinPercent(nil, 0)
	return rule, lt, tagger
}

func TestSentenceWithModalVerbRule_RealGermanTagger(t *testing.T) {
	rule, lt, tagger := setUpSentenceWithModalVerbRule(t)

	// MOD + INF (muss … erledigen)
	modInf := analyzeTextWithGermanTagger(lt, tagger, "Er muss das erledigen.")
	require.Equal(t, 1, len(rule.MatchList(modInf)))

	// INF + MOD (erledigen müssen)
	infMod := analyzeTextWithGermanTagger(lt, tagger, "Er wird das erledigen müssen.")
	require.Equal(t, 1, len(rule.MatchList(infMod)))

	// modal only — no INF
	modalOnly := analyzeTextWithGermanTagger(lt, tagger, "Er muss das.")
	require.Equal(t, 0, len(rule.MatchList(modalOnly)))

	// no modal
	none := analyzeTextWithGermanTagger(lt, tagger, "Er erledigt das.")
	require.Equal(t, 0, len(rule.MatchList(none)))
}
