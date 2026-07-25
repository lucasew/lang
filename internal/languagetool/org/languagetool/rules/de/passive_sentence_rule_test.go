package de

// Twin of org.languagetool.rules.de.PassiveSentenceRule
// (no Java *Test class — behavior matrix + real de-DE analysis when GermanTagger supplies werden+VER:PA2).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	detag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/de"
	"github.com/stretchr/testify/require"
)

func TestPassiveSentenceRule_Meta(t *testing.T) {
	r := NewPassiveSentenceRule(nil)
	require.Equal(t, "PASSIVE_SENTENCE_DE", r.GetID())
	require.Equal(t, passiveSentenceDefaultMinPercent, r.MinPercent)
	require.Equal(t, 8, r.MinPercent)
	require.Equal(t, float64(100), r.Denominator)
	require.True(t, r.ExcludeDirectSpeech)
	require.True(t, r.IsDefaultOff())
	require.Equal(t, "Statistische Stilanalyse: Passivsätze", r.GetDescription())
	require.Equal(t, passiveSentenceConfigureText, r.GetConfigureText())
	require.Equal(t,
		"Anzeigen wenn mehr als ...% Sätze eines Kapitels Passivsätze sind:",
		r.GetConfigureText())
	require.NotNil(t, r.GetCategory())
	require.True(t, r.GetCategory().GetID().Equals(rules.NewCategoryId("CREATIVE_WRITING")))
	require.Equal(t, rules.ITSStyle, r.GetLocQualityIssueType())

	opts := r.GetRuleOptions()
	require.Len(t, opts, 1)
	require.Equal(t, 8, opts[0].DefaultValue)
	require.Equal(t, passiveSentenceConfigureText, opts[0].ConfigureText)
	require.Equal(t, 0, opts[0].MinConfigurableValue)
	require.Equal(t, 100, opts[0].MaxConfigurableValue)
}

func TestPassiveSentenceRule_LimitMessage(t *testing.T) {
	r := NewPassiveSentenceRule(nil)
	require.Equal(t,
		"Passivsatz: Aktiv formulierte Sätze sprechen im Regelfall den Leser stärker an.",
		r.getLimitMessage(0, 0))
	require.Equal(t,
		"Mehr als 8% Passivsätze {50%} gefunden. Aktiv formulierte Sätze sprechen im Regelfall den Leser stärker an.",
		r.getLimitMessage(8, 50))
	// Java: (int)(percent + 0.5d) rounding
	require.Equal(t,
		"Mehr als 8% Passivsätze {51%} gefunden. Aktiv formulierte Sätze sprechen im Regelfall den Leser stärker an.",
		r.getLimitMessage(8, 50.6))
}

func TestPassiveSentenceRule_UserConfigMinPercent(t *testing.T) {
	uc := languagetool.NewUserConfig()
	uc.SetConfigValueByID("PASSIVE_SENTENCE_DE", []any{3})
	r := NewPassiveSentenceRuleWithUserConfig(nil, uc)
	require.Equal(t, 3, r.MinPercent)

	r0 := NewPassiveSentenceRuleWithUserConfig(nil, nil)
	require.Equal(t, 8, r0.MinPercent)

	rMin := NewPassiveSentenceRuleWithMinPercent(nil, 0)
	require.Equal(t, 0, rMin.MinPercent)
}

func passiveMorphSent(toks ...*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedSentence {
	return languagetool.NewAnalyzedSentence(withPositions(toks...))
}

func TestPassiveSentenceRule_ConditionMatrix(t *testing.T) {
	rule := NewPassiveSentenceRuleWithMinPercent(nil, 0)

	// werden then VER:PA2 → hit (returns werden token)
	werdenThenPA2 := passiveMorphSent(
		sentStartATR(),
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "der"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
		atrWithPOS("wird", "VER:AUX:3:SIN:PRS:SFT", "werden"),
		atrWithPOS("gebaut", "VER:PA2:NON", "bauen"),
		atrWithPOS(".", "PKT", "."),
	)
	matches := rule.Match(werdenThenPA2)
	require.Equal(t, 1, len(matches))
	require.Contains(t, matches[0].GetMessage(), "Passivsatz:")

	// VER:PA2 then werden → hit (returns werden token)
	pa2ThenWerden := passiveMorphSent(
		sentStartATR(),
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "der"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
		atrWithPOS("gebaut", "VER:PA2:NON", "bauen"),
		atrWithPOS("wird", "VER:AUX:3:SIN:PRS:SFT", "werden"),
		atrWithPOS(".", "PKT", "."),
	)
	require.Equal(t, 1, len(rule.Match(pa2ThenWerden)))

	// mark between werden and PA2 aborts
	werdenMarkPA2 := passiveMorphSent(
		sentStartATR(),
		atrWithPOS("wird", "VER:AUX:3:SIN:PRS:SFT", "werden"),
		atrWithPOS(",", ",", ","),
		atrWithPOS("gebaut", "VER:PA2:NON", "bauen"),
		atrWithPOS(".", "PKT", "."),
	)
	require.Equal(t, 0, len(rule.Match(werdenMarkPA2)))

	// mark between PA2 and werden aborts
	pa2MarkWerden := passiveMorphSent(
		sentStartATR(),
		atrWithPOS("gebaut", "VER:PA2:NON", "bauen"),
		atrWithPOS(",", ",", ","),
		atrWithPOS("wird", "VER:AUX:3:SIN:PRS:SFT", "werden"),
		atrWithPOS(".", "PKT", "."),
	)
	require.Equal(t, 0, len(rule.Match(pa2MarkWerden)))

	// werden alone is not enough (period is mark — no PA2)
	werdenOnly := passiveMorphSent(
		sentStartATR(),
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "der"),
		atrWithPOS("wird", "VER:AUX:3:SIN:PRS:SFT", "werden"),
		atrWithPOS(".", "PKT", "."),
	)
	require.Equal(t, 0, len(rule.Match(werdenOnly)))

	// PA2 alone is not enough
	pa2Only := passiveMorphSent(
		sentStartATR(),
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "der"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
		atrWithPOS("gebaut", "VER:PA2:NON", "bauen"),
		atrWithPOS(".", "PKT", "."),
	)
	require.Equal(t, 0, len(rule.Match(pa2Only)))

	// active sentence (no werden lemma / no PA2 pair)
	active := passiveMorphSent(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("baut", "VER:3:SIN:PRS:NON", "bauen"),
		atrWithPOS("das", "ART:DEF:AKK:SIN:NEU", "der"),
		atrWithPOS("Haus", "SUB:AKK:SIN:NEU", "Haus"),
		atrWithPOS(".", "PKT", "."),
	)
	require.Equal(t, 0, len(rule.Match(active)))

	// surface "wird" with wrong lemma must not invent a hit
	wrongLemma := passiveMorphSent(
		sentStartATR(),
		atrWithPOS("wird", "VER:3:SIN:PRS:NON", "werdenX"), // not lemma "werden"
		atrWithPOS("gebaut", "VER:PA2:NON", "bauen"),
		atrWithPOS(".", "PKT", "."),
	)
	require.Equal(t, 0, len(rule.Match(wrongLemma)))

	// untagged AnalyzePlain must not invent hits
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Das Haus wird gebaut."))))
}

func TestPassiveSentenceRule_PercentThreshold(t *testing.T) {
	// Two sentences, one passive → 50%. Java: percent > minPercent.
	passive := passiveMorphSent(
		sentStartATR(),
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "der"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
		atrWithPOS("wird", "VER:AUX:3:SIN:PRS:SFT", "werden"),
		atrWithPOS("gebaut", "VER:PA2:NON", "bauen"),
		atrWithPOS(".", "PKT", "."),
	)
	other := passiveMorphSent(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("geht", "VER:3:SIN:PRS:NON", "gehen"),
		atrWithPOS(".", "PKT", "."),
	)
	sents := []*languagetool.AnalyzedSentence{passive, other}

	// 50% > 8 default → hit with threshold message
	rDef := NewPassiveSentenceRule(nil)
	ms := rDef.MatchList(sents)
	require.Equal(t, 1, len(ms))
	require.Contains(t, ms[0].GetMessage(), "Mehr als 8%")
	require.Contains(t, ms[0].GetMessage(), "50%")

	// 50% > 49 → hit
	r49 := NewPassiveSentenceRuleWithMinPercent(nil, 49)
	require.Equal(t, 1, len(r49.MatchList(sents)))

	// 50% > 50 is false → no hit
	r50 := NewPassiveSentenceRuleWithMinPercent(nil, 50)
	require.Equal(t, 0, len(r50.MatchList(sents)))

	// minPercent 0 shows all with limit==0 message
	r0 := NewPassiveSentenceRuleWithMinPercent(nil, 0)
	ms0 := r0.MatchList(sents)
	require.Equal(t, 1, len(ms0))
	require.Equal(t,
		"Passivsatz: Aktiv formulierte Sätze sprechen im Regelfall den Leser stärker an.",
		ms0[0].GetMessage())
}

func TestPassiveSentenceRule_ConditionFulfilledDirect(t *testing.T) {
	// Unit: conditionFulfilled return token / mark abort / i++ reuse
	r := NewPassiveSentenceRuleWithMinPercent(nil, 0)

	// werden … PA2 → werden token
	werden := atrWithPOS("wird", "VER:AUX:3:SIN:PRS:SFT", "werden")
	pa2 := atrWithPOS("gebaut", "VER:PA2:NON", "bauen")
	got := r.conditionFulfilled([]*languagetool.AnalyzedTokenReadings{
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "der"),
		werden,
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
		pa2,
		atrWithPOS(".", "PKT", "."),
	})
	require.Same(t, werden, got)

	// PA2 … werden → werden token
	werden2 := atrWithPOS("wird", "VER:AUX:3:SIN:PRS:SFT", "werden")
	pa22 := atrWithPOS("gebaut", "VER:PA2:NON", "bauen")
	got2 := r.conditionFulfilled([]*languagetool.AnalyzedTokenReadings{
		pa22,
		werden2,
	})
	require.Same(t, werden2, got2)

	// mark aborts werden→PA2 to null
	require.Nil(t, r.conditionFulfilled([]*languagetool.AnalyzedTokenReadings{
		atrWithPOS("wird", "VER:AUX:3:SIN:PRS:SFT", "werden"),
		atrWithPOS(",", ",", ","),
		atrWithPOS("gebaut", "VER:PA2:NON", "bauen"),
	}))

	// mark aborts PA2→werden to null
	require.Nil(t, r.conditionFulfilled([]*languagetool.AnalyzedTokenReadings{
		atrWithPOS("gebaut", "VER:PA2:NON", "bauen"),
		atrWithPOS(",", ",", ","),
		atrWithPOS("wird", "VER:AUX:3:SIN:PRS:SFT", "werden"),
	}))

	// i++ reuse exhaust: after inner loop ends without match, outer continues with advanced i
	// (no second scan of tokens already consumed by inner for)
	require.Nil(t, r.conditionFulfilled([]*languagetool.AnalyzedTokenReadings{
		atrWithPOS("wird", "VER:AUX:3:SIN:PRS:SFT", "werden"),
		atrWithPOS("heute", "ADV", "heute"),
		// no PA2; i advanced past "heute"; outer loop ends
	}))
}

func setUpPassiveSentenceRule(t *testing.T) (*PassiveSentenceRule, *languagetool.JLanguageTool, *detag.GermanTagger) {
	t.Helper()
	lt := languagetool.NewJLanguageTool("de-DE")
	require.True(t, WireGermanTagWord(lt), "GermanTagger resources required for PassiveSentenceRule twin")
	tagger := DiscoveredGermanTagger()
	require.NotNil(t, tagger, "GermanTagger.INSTANCE resources required")
	// show-all for real-analysis twin (minPercent=0)
	rule := NewPassiveSentenceRuleWithMinPercent(nil, 0)
	return rule, lt, tagger
}

func TestPassiveSentenceRule_RealGermanTagger(t *testing.T) {
	rule, lt, tagger := setUpPassiveSentenceRule(t)

	// werden + PA2 (wird … gebaut)
	werdenPA2 := analyzeTextWithGermanTagger(lt, tagger, "Das Haus wird gebaut.")
	require.Equal(t, 1, len(rule.MatchList(werdenPA2)))

	// past participle fronted + werden (gebaut wird)
	pa2Werden := analyzeTextWithGermanTagger(lt, tagger, "Gebaut wird das Haus.")
	require.Equal(t, 1, len(rule.MatchList(pa2Werden)))

	// active — no passive pair
	active := analyzeTextWithGermanTagger(lt, tagger, "Er baut das Haus.")
	require.Equal(t, 0, len(rule.MatchList(active)))

	// werden without PA2
	werdenOnly := analyzeTextWithGermanTagger(lt, tagger, "Er wird Lehrer.")
	require.Equal(t, 0, len(rule.MatchList(werdenOnly)))
}

// TestPassiveSentenceRule keeps the historical smoke entrypoint name used by package scans.
func TestPassiveSentenceRule(t *testing.T) {
	rule := NewPassiveSentenceRuleWithMinPercent(nil, 0)
	hit := passiveMorphSent(
		sentStartATR(),
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "der"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
		atrWithPOS("wird", "VER:AUX:3:SIN:PRS:SFT", "werden"),
		atrWithPOS("gebaut", "VER:PA2:NON", "bauen"),
		atrWithPOS(".", "PKT", "."),
	)
	require.Equal(t, 1, len(rule.Match(hit)))
	active := passiveMorphSent(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("baut", "VER:3:SIN:PRS:NON", "bauen"),
		atrWithPOS("das", "ART:DEF:AKK:SIN:NEU", "der"),
		atrWithPOS("Haus", "SUB:AKK:SIN:NEU", "Haus"),
		atrWithPOS(".", "PKT", "."),
	)
	require.Equal(t, 0, len(rule.Match(active)))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Das Haus wird gebaut."))))
}
