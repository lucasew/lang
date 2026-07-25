package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/NonSignificantVerbsRuleTest.java
//
// Java:
//   Language lang = Languages.getLanguageForShortCode("de-DE");
//   JLanguageTool lt = new JLanguageTool(lang);
//   setUpRule(lt): disable all; add NonSignificantVerbsRule; enable
//   assertEquals(3, lt.check("Wenn man das machen kann, sollte man das tun. Das ist so.").size());
//   assertEquals(0, lt.check("Der Vorgang war abgeschlossen. Das hatte er nicht bedacht.").size());
//
// Go: de-DE analysis via GermanWordTokenizer + GermanTagger.Tag (Java createDefaultTagger().tag
// with ignoreCase=true). WireGermanTagWord is Lookup-only and leaves sentence-start "Das" as
// isPosTagUnknown forever (flag is construction-time final in Java). Building ATR from Tag
// sets isPosTagUnknown correctly — required for isException isUnknownWord. No POS invent.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	detok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/de"
	detag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/de"
	"github.com/stretchr/testify/require"
)

// setUpNonSignificantVerbsRule ports NonSignificantVerbsRuleTest.setUpRule (+ tagger ready).
func setUpNonSignificantVerbsRule(t *testing.T) (*NonSignificantVerbsRule, *languagetool.JLanguageTool, *detag.GermanTagger) {
	t.Helper()
	// Java: Languages.getLanguageForShortCode("de-DE")
	lt := languagetool.NewJLanguageTool("de-DE")
	// Ensure resources open (same as WireGermanTagWord path).
	require.True(t, WireGermanTagWord(lt), "GermanTagger resources required for NonSignificantVerbsRule twin")
	tagger := DiscoveredGermanTagger()
	require.NotNil(t, tagger, "GermanTagger.INSTANCE resources required")

	// Java: for (Rule rule : lt.getAllRules()) { lt.disableRule(rule.getId()); }
	for _, id := range lt.GetAllRegisteredRuleIDs() {
		lt.DisableRule(id)
	}
	// Java: new NonSignificantVerbsRule(TestTools.getMessages(lang.getShortCode()), lang, new UserConfig())
	rule := NewNonSignificantVerbsRule(nil)
	// Text-level registration for Check path when analysis is re-built externally.
	lt.AddTextLevelRuleChecker(rule.GetID(), rules.AsTextLevelChecker(rule.MatchList))
	lt.EnableRule(rule.GetID())
	return rule, lt, tagger
}

// analyzeWithGermanTagger ports getAnalyzedSentence for one sentence string:
// GermanWordTokenizer.tokenize + GermanTagger.tag (ignoreCase=true) + SENT_START/SENT_END.
// Same shape as Java JLanguageTool.getRawAnalyzedSentence for de (no invent POS).
func analyzeWithGermanTagger(tagger *detag.GermanTagger, sentence string) *languagetool.AnalyzedSentence {
	if tagger == nil {
		return languagetool.AnalyzePlain(sentence)
	}
	wt := detok.NewGermanWordTokenizer()
	surfaces := wt.Tokenize(sentence)
	// Java: language.getTagger().tag(tokens) — includes whitespace tokens.
	tagged := tagger.Tag(surfaces)

	ss := languagetool.SentenceStartTagName
	startTok := languagetool.NewAnalyzedToken("", &ss, nil)
	startR := languagetool.NewAnalyzedTokenReadings(startTok)
	startR.SetStartPos(0)

	readings := make([]*languagetool.AnalyzedTokenReadings, 0, len(tagged)+1)
	readings = append(readings, startR)
	prev := ""
	// positions from surfaces (UTF-16), same as tokenizers.BuildPositions
	positions := tokenizers.BuildPositions(surfaces)
	for i, ar := range tagged {
		if ar == nil {
			continue
		}
		// Rebuild ATR at correct startPos so isPosTagUnknown is set from Tag readings
		// (Tag already created ATR; re-wrap if startPos differs).
		if i < len(positions) && ar.GetStartPos() != positions[i] {
			ar.SetStartPos(positions[i])
		}
		if prev != "" {
			ar.SetWhitespaceBeforeToken(prev)
		}
		readings = append(readings, ar)
		if i < len(surfaces) {
			prev = surfaces[i]
		}
	}
	// SENT_END on last content token (Java getRawAnalyzedSentence).
	if len(readings) > 1 {
		last := readings[len(readings)-1]
		if last != nil && !last.IsSentenceEnd() {
			last.SetSentEnd()
		}
	}
	return languagetool.NewAnalyzedSentence(readings)
}

// analyzeTextWithGermanTagger ports multi-sentence Analyze (SRX split + getAnalyzedSentence).
func analyzeTextWithGermanTagger(lt *languagetool.JLanguageTool, tagger *detag.GermanTagger, text string) []*languagetool.AnalyzedSentence {
	parts := lt.SentenceTokenize(text)
	if len(parts) == 0 {
		return nil
	}
	out := make([]*languagetool.AnalyzedSentence, 0, len(parts))
	for _, p := range parts {
		out = append(out, analyzeWithGermanTagger(tagger, p))
	}
	return out
}

// nsvCheck ports lt.check(text).size() with only NonSignificantVerbsRule enabled.
// Analysis uses createDefaultTagger().tag equivalent (see analyzeWithGermanTagger).
func nsvCheck(t *testing.T, rule *NonSignificantVerbsRule, lt *languagetool.JLanguageTool, tagger *detag.GermanTagger, text string) int {
	t.Helper()
	sents := analyzeTextWithGermanTagger(lt, tagger, text)
	return len(rule.MatchList(sents))
}

// TestNonSignificantVerbsRule_Rule ports NonSignificantVerbsRuleTest.testRule.
func TestNonSignificantVerbsRule_Rule(t *testing.T) {
	rule, lt, tagger := setUpNonSignificantVerbsRule(t)

	// Java subclass surface (id, default limit, description, default-off).
	require.Equal(t, "NON_SIGNIFICANT_VERB_DE", rule.GetID())
	require.Equal(t, nonSignificantDefaultMinPerMill, rule.MinPercent)
	require.Equal(t, 8, rule.MinPercent)
	require.Equal(t, float64(1000), rule.Denominator)
	require.True(t, rule.ExcludeDirectSpeech)
	require.True(t, rule.IsDefaultOff())
	require.Equal(t, "Statistische Stilanalyse: Verben mit wenig Aussagekraft", rule.GetDescription())
	require.NotNil(t, rule.GetCategory())
	require.True(t, rule.GetCategory().GetID().Equals(rules.NewCategoryId("CREATIVE_WRITING")))
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())

	// Java assertEquals(3, lt.check(...).size()) — machen, tun, ist (over 8‰)
	require.Equal(t, 3, nsvCheck(t, rule, lt, tagger, "Wenn man das machen kann, sollte man das tun. Das ist so."))
	// Java assertEquals(0, ...) — sein/haben + PA2 exceptions
	require.Equal(t, 0, nsvCheck(t, rule, lt, tagger, "Der Vorgang war abgeschlossen. Das hatte er nicht bedacht."))
}

// TestNonSignificantVerbsRule_RealAnalysisLemmas documents GermanTagger lemmas/POS
// for hasAnyLemma + PA2 + isUnknownWord — not hand-built ATR.
func TestNonSignificantVerbsRule_RealAnalysisLemmas(t *testing.T) {
	_, lt, tagger := setUpNonSignificantVerbsRule(t)

	sents := analyzeTextWithGermanTagger(lt, tagger, "Wenn man das machen kann, sollte man das tun. Das ist so.")
	require.Len(t, sents, 2)

	var sawMachen, sawTun bool
	for _, tok := range sents[0].GetTokensWithoutWhitespace() {
		if tok == nil {
			continue
		}
		if tok.GetToken() == "machen" {
			require.True(t, tok.HasAnyLemma("machen"), "machen must have lemma machen")
			sawMachen = true
		}
		if tok.GetToken() == "tun" {
			require.True(t, tok.HasAnyLemma("tun"), "tun must have lemma tun")
			sawTun = true
		}
	}
	require.True(t, sawMachen && sawTun)

	var sawIst bool
	for _, tok := range sents[1].GetTokensWithoutWhitespace() {
		if tok == nil {
			continue
		}
		if tok.GetToken() == "Das" {
			require.False(t, tok.IsPosTagUnknown(), "Das after Tag must not be isPosTagUnknown")
			require.False(t, isUnknownWordNS(tok), "Das must not be isUnknownWord")
		}
		if tok.GetToken() == "ist" {
			require.True(t, tok.HasAnyLemma("sein"), "ist must have lemma sein")
			sawIst = true
		}
	}
	require.True(t, sawIst)

	exc := analyzeTextWithGermanTagger(lt, tagger, "Der Vorgang war abgeschlossen. Das hatte er nicht bedacht.")
	require.Len(t, exc, 2)
	for _, tok := range exc[0].GetTokensWithoutWhitespace() {
		if tok != nil && tok.GetToken() == "abgeschlossen" {
			require.True(t,
				tok.HasPosTagStartingWith("PA2") || tok.HasPosTagStartingWith("VER:PA2"),
				"abgeschlossen needs PA2/VER:PA2 for isException")
		}
	}
	for _, tok := range exc[1].GetTokensWithoutWhitespace() {
		if tok != nil && tok.GetToken() == "bedacht" {
			require.True(t,
				tok.HasPosTagStartingWith("PA2") || tok.HasPosTagStartingWith("VER:PA2"),
				"bedacht needs PA2/VER:PA2 for isException")
		}
	}
}
