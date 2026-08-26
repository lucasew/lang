package de

import (
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

// GermanConfusionAntiPatterns ports GermanConfusionProbabilityRule.ANTI_PATTERNS (11/11).
var GermanConfusionAntiPatterns = [][]*patterns.PatternToken{
	{
		// "Im nur wenige Meter entfernten Schergenturm"
		patterns.NewPatternTokenBuilder().Token("im").SetSkip(8).Build(),
		patterns.PosRegex("PA[12].*"),
	},
	{
		// "Du forderst viel …" / "Schneit es viel …"
		patterns.PosRegex("VER.*"),
		patterns.NewPatternTokenBuilder().Token("es").Min(0).Build(),
		patterns.Token("viel"),
	},
	{
		// "Geht Tom viel fort?"
		patterns.PosRegex("VER.*"),
		patterns.PosRegex("(EIG|SUB).*"),
		patterns.Token("viel"),
	},
	{
		// "… viel … investiert"
		patterns.NewPatternTokenBuilder().Token("viel").SetSkip(8).Build(),
		patterns.PosRegex("PA2.*"),
	},
	{
		// "Warum viel graue Energie …"
		patterns.Token("viel"),
		patterns.NewPatternTokenBuilder().PosRegex("ADJ.*").Min(0).Build(),
		patterns.PosRegex("SUB.*"),
	},
	{
		// "Wie haben ihr …"
		patterns.CsToken("Wie"),
		patterns.PosRegex("VER.*"),
	},
	{
		// "Weist du uns den Weg?"
		patterns.NewPatternTokenBuilder().Token("weist").SetSkip(8).Build(),
		patterns.Token("den"),
		patterns.CsToken("Weg"),
	},
	{
		// "… Tank fasst …"
		patterns.Regex(".*tank|.*bus|.*zug|.*flieger|.*flugzeug|.*container|.*behälter|.*schüssel|.*festplatte|Platte|SSD|.*speicher|.*glas|.*tasse|.*batterie"),
		patterns.Token("fasst"),
	},
	{
		// "… fasst … zusammen"
		patterns.NewPatternTokenBuilder().Token("fasst").SetSkip(-1).Build(),
		patterns.Token("zusammen"),
	},
	{
		// "Wer’s glaubt …"
		patterns.Token("wer"),
		patterns.Regex("['’`´‘]"),
		patterns.Token("s"),
		patterns.Regex("glaubt|will|mag"),
	},
	{
		// "… sieht … gedeckt|bestätigt|aus"
		patterns.NewPatternTokenBuilder().Token("sieht").SetSkip(-1).Build(),
		patterns.Regex("gedeckt|bestätigt|aus"),
	},
}

var (
	deConfusionAPOnce  sync.Once
	deConfusionAPRules []*disambigrules.DisambiguationPatternRule
)

func germanConfusionAntiPatternRules() []*disambigrules.DisambiguationPatternRule {
	deConfusionAPOnce.Do(func() {
		aps := GermanConfusionAntiPatterns
		deConfusionAPRules = make([]*disambigrules.DisambiguationPatternRule, 0, len(aps))
		for _, toks := range aps {
			if len(toks) == 0 {
				continue
			}
			rule := disambigrules.NewDisambiguationPatternRule(
				"INTERNAL_ANTIPATTERN", "(no description)", "de",
				toks, "", nil, disambigrules.ActionImmunize,
			)
			deConfusionAPRules = append(deConfusionAPRules, rule)
		}
	})
	return deConfusionAPRules
}

func germanConfusionSentenceWithImmunization(sentence *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if sentence == nil {
		return nil
	}
	aps := germanConfusionAntiPatternRules()
	if len(aps) == 0 {
		return sentence
	}
	src := sentence.GetTokens()
	cloned := make([]*languagetool.AnalyzedTokenReadings, len(src))
	for i, t := range src {
		if t == nil {
			continue
		}
		cloned[i] = languagetool.NewAnalyzedTokenReadingsFromOld(t, t.GetReadings(), "")
	}
	immunized := languagetool.NewAnalyzedSentence(cloned)
	for _, ap := range aps {
		if ap != nil {
			immunized = ap.Replace(immunized)
		}
	}
	return immunized
}

// deConfusionIsCoveredByAntiPattern ports ConfusionProbabilityRule.isCoveredByAntiPattern
// via DisambiguationPatternRule immunization (Java getSentenceWithImmunization).
func deConfusionIsCoveredByAntiPattern(sentence *languagetool.AnalyzedSentence, startPos, endPos int) bool {
	imm := germanConfusionSentenceWithImmunization(sentence)
	if imm == nil {
		return false
	}
	for _, t := range imm.GetTokensWithoutWhitespace() {
		if t == nil || !t.IsImmunized() {
			continue
		}
		// covers(tmpStart, tmpEnd, googleStart, googleEnd)
		if t.GetStartPos() <= startPos && t.GetEndPos() >= endPos {
			return true
		}
	}
	return false
}
