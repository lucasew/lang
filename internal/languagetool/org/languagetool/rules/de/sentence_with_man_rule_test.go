package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSentenceWithManRule(t *testing.T) {
	rule := NewSentenceWithManRuleWithMinPercent(nil, 0)
	// Java: hasLemma("man") only
	manSent := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Man", "PRO:IND:NOM:SIN:MAS", "man"),
		atrWithPOS("sollte", "VER:MOD:3:SIN:PRT:SFT", "sollen"),
		atrWithPOS("das", "PRO:DEM:AKK:SIN:NEU", "das"),
		atrWithPOS("vermeiden", "VER:INF:NON", "vermeiden"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 1, len(rule.Match(manSent)))
	erSent := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("sollte", "VER:MOD:3:SIN:PRT:SFT", "sollen"),
		atrWithPOS("das", "PRO:DEM:AKK:SIN:NEU", "das"),
		atrWithPOS("vermeiden", "VER:INF:NON", "vermeiden"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(erSent)))
	// untagged "man" must not invent a hit
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Man sollte das vermeiden."))))
}

func TestSentenceWithModalVerbRule(t *testing.T) {
	rule := NewSentenceWithModalVerbRuleWithMinPercent(nil, 0)
	// Java: VER:MOD + VER:INF
	hit := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("muss", "VER:MOD:3:SIN:PRS:SFT", "müssen"),
		atrWithPOS("das", "PRO:DEM:AKK:SIN:NEU", "das"),
		atrWithPOS("erledigen", "VER:INF:NON", "erledigen"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 1, len(rule.Match(hit)))
	noModal := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("erledigt", "VER:3:SIN:PRS:NON", "erledigen"),
		atrWithPOS("das", "PRO:DEM:AKK:SIN:NEU", "das"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(noModal)))
	// modal alone is not enough
	modalOnly := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("muss", "VER:MOD:3:SIN:PRS:SFT", "müssen"),
		atrWithPOS("das", "PRO:DEM:AKK:SIN:NEU", "das"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(modalOnly)))
	// no surface invent on AnalyzePlain
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Er muss das erledigen."))))
}

func TestPassiveSentenceRule(t *testing.T) {
	rule := NewPassiveSentenceRuleWithMinPercent(nil, 0)
	// Java: lemma werden + VER:PA2
	hit := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "der"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
		atrWithPOS("wird", "VER:AUX:3:SIN:PRS:SFT", "werden"),
		atrWithPOS("gebaut", "VER:PA2:NON", "bauen"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 1, len(rule.Match(hit)))
	active := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("baut", "VER:3:SIN:PRS:NON", "bauen"),
		atrWithPOS("das", "ART:DEF:AKK:SIN:NEU", "der"),
		atrWithPOS("Haus", "SUB:AKK:SIN:NEU", "Haus"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(active)))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Das Haus wird gebaut."))))
}
