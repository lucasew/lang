package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/CompoundInfinitivRuleTest.java
//
// Java:
//   Language german = Languages.getLanguageForShortCode("de");
//   CompoundInfinitivRule rule = new CompoundInfinitivRule(TestTools.getMessages("de"), german, null);
//   JLanguageTool lt = new JLanguageTool(german);
//   assertThat(rule.match(lt.getAnalyzedSentence(...)).length, is(N));
//
// Go: de JLanguageTool + WireGermanRuntimeResourcesFor (createDefaultTagger + disambiguator
// + de_DE.dict speller). Match uses real ZUS/VER:INF tags and Morfologik de_DE.dict —
// same path as Java (no surface invent, no fake known-word map).
//
// Sentence-start imperative short forms ("Hör") need GermanTagger.Tag sentence context
// (Java createDefaultTagger().tag → getImperativeForm). WireGermanTagWord is Lookup-only,
// so the twin re-applies GermanTagger.Tag for untagged first content tokens when needed.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// newCompoundInfinitivFixture mirrors Java CompoundInfinitivRuleTest fields.
func newCompoundInfinitivFixture(t *testing.T) (*CompoundInfinitivRule, *languagetool.JLanguageTool) {
	t.Helper()
	// Java: new CompoundInfinitivRule(messages, german, null) — speller lazy on match
	rule := NewCompoundInfinitivRule(nil)
	lt := languagetool.NewJLanguageTool("de")
	WireGermanRuntimeResourcesFor(lt, "DE")
	require.True(t, FilterDictAvailable() || DiscoverGermanHunspellDict("DE") != "",
		"de_DE.dict required for CompoundInfinitivRule speller twin")
	require.True(t, WireGermanTagWord(lt) || lt.TagWord != nil,
		"GermanTagger resources required for CompoundInfinitivRule twin")
	return rule, lt
}

// ciMatchLen ports rule.match(lt.getAnalyzedSentence(text)).length with Java-visible
// sentence-start imperative enrichment when TagWord Lookup left the first word untagged.
func ciMatchLen(t *testing.T, rule *CompoundInfinitivRule, lt *languagetool.JLanguageTool, text string) int {
	t.Helper()
	sent := lt.GetAnalyzedSentence(text)
	require.NotNil(t, sent)
	enrichSentenceStartImperativeFromGermanTagger(sent)
	return len(rule.Match(sent))
}

// enrichSentenceStartImperativeFromGermanTagger ports the Java tagger arm that
// Lookup-based TagWord cannot: GermanTagger.getImperativeForm at sentence start
// (e.g. "Hör" → VER:IMP:SIN:SFT). Only fills tokens that still lack any POS tag.
func enrichSentenceStartImperativeFromGermanTagger(sent *languagetool.AnalyzedSentence) {
	if sent == nil {
		return
	}
	tagger := DiscoveredGermanTagger()
	if tagger == nil {
		return
	}
	all := sent.GetTokens()
	if len(all) < 2 {
		return
	}
	// First content token (index 1 after SENT_START) — only enrich when untagged.
	first := all[1]
	if first == nil || first.IsWhitespace() {
		return
	}
	if hasAnyPOSTagCI(first) {
		return
	}
	// Full token list including whitespace (Java GermanTagger.tag sentenceTokens).
	surfaces := make([]string, 0, len(all)-1)
	for i := 1; i < len(all); i++ {
		if all[i] == nil {
			surfaces = append(surfaces, "")
			continue
		}
		surfaces = append(surfaces, all[i].GetToken())
	}
	tagged := tagger.Tag(surfaces)
	if len(tagged) == 0 || tagged[0] == nil {
		return
	}
	for _, r := range tagged[0].GetReadings() {
		if r == nil {
			continue
		}
		if p := r.GetPOSTag(); p != nil && *p != "" {
			first.AddReading(r, "GermanTagger.Tag")
		}
	}
}

func hasAnyPOSTagCI(tok *languagetool.AnalyzedTokenReadings) bool {
	if tok == nil {
		return false
	}
	for _, r := range tok.GetReadings() {
		if r == nil {
			continue
		}
		if p := r.GetPOSTag(); p != nil && *p != "" && *p != languagetool.SentenceStartTagName &&
			*p != languagetool.SentenceEndTagName {
			return true
		}
	}
	return false
}

// TestCompoundInfinitivRule_Rule ports CompoundInfinitivRuleTest.testRule full matrix.
func TestCompoundInfinitivRule_Rule(t *testing.T) {
	rule, lt := newCompoundInfinitivFixture(t)

	assertLen := func(text string, want int) {
		t.Helper()
		got := ciMatchLen(t, rule, lt, text)
		require.Equal(t, want, got, "match length for %q", text)
	}

	// Java assertThat(...length, is(1)) — incorrect (should compound):
	assertLen("Ich brachte ihn dazu, mein Zimmer sauber zu machen.", 1)
	assertLen("Du brauchst nicht bei mir vorbei zu kommen.", 1)
	assertLen("Ich ging zur Seite, um die alte Dame vorbei zu lassen.", 1)

	// Java assertThat(...length, is(0)) — correct / exception / anti-pattern:
	assertLen("Seine Frau gab vor zu schlafen.", 0)
	assertLen("Mein Herz hörte auf zu schlagen.", 0)
	assertLen("Den Sonnenaufgang von einem Berggipfel aus zu sehen, ist eine Wonne.", 0)
	assertLen("Hör auf zu schreien", 0)
	assertLen("Sie riss sich zusammen und fing wieder an zu reden.", 0)
	assertLen("Fang an zu zählen.", 0)
	assertLen("Er hatte nichts weiter zu sagen", 0)
	assertLen("Sie strengte sich an zu schwimmen.", 0)
	assertLen("Er erhob sich von dem Stuhl, auf dem er gesessen hatte, und begann im Zimmer auf und ab zu schreiten.", 0)
	assertLen("Tom stand auf und fing an, auf und ab zu gehen.", 0)
	assertLen("Aber um auf Nummer sicher zu gehen, schrieb er es auf.", 0)
	assertLen("Sie scheint aus Spanien heraus zu kaufen.", 0)
	assertLen("Hab keine Lust, mir Gedanken darüber zu machen.", 0)
	assertLen("Fang an zu lesen, wo du aufgehört hast.", 0)
	assertLen("Fang dort an zu lesen, wo du aufgehört hast.", 0)
	assertLen("Der diensthabende Kollege hatte ganz schön zu tun.", 0)
}

// TestCompoundInfinitivRule_MatchSurface ports Java-visible match fields for a hit
// (message, suggestion particle+zu+infinitive, UTF-16 span).
func TestCompoundInfinitivRule_MatchSurface(t *testing.T) {
	rule, lt := newCompoundInfinitivFixture(t)
	text := "Ich brachte ihn dazu, mein Zimmer sauber zu machen."
	sent := lt.GetAnalyzedSentence(text)
	ms := rule.Match(sent)
	require.Equal(t, 1, len(ms))
	require.Contains(t, ms[0].GetMessage(), "saubermachen")
	// Java: suggestions.add(tokens[i-1] + tokens[i] + tokens[i+1]) → sauberzumachen
	require.Equal(t, []string{"sauberzumachen"}, ms[0].GetSuggestedReplacements())
	toks := sent.GetTokensWithoutWhitespace()
	var sauberTok, machenTok *languagetool.AnalyzedTokenReadings
	for _, tok := range toks {
		if tok == nil {
			continue
		}
		switch tok.GetToken() {
		case "sauber":
			sauberTok = tok
		case "machen":
			machenTok = tok
		}
	}
	require.NotNil(t, sauberTok)
	require.NotNil(t, machenTok)
	require.Equal(t, sauberTok.GetStartPos(), ms[0].GetFromPos())
	require.Equal(t, machenTok.GetEndPos(), ms[0].GetToPos())
}

// TestCompoundInfinitivRule_IsPunctuationUTF16 ports isPunctuation (Java length==1 + equalsAny).
func TestCompoundInfinitivRule_IsPunctuationUTF16(t *testing.T) {
	require.True(t, isPunctuationCI("…"))
	require.True(t, isPunctuationCI("."))
	require.False(t, isPunctuationCI(".."))
	require.False(t, isPunctuationCI(""))
}

// TestCompoundInfinitivRule_Meta ports constructor surface (id, description, url, examples).
func TestCompoundInfinitivRule_Meta(t *testing.T) {
	r := NewCompoundInfinitivRule(nil)
	require.Equal(t, "COMPOUND_INFINITIV_RULE", r.GetID())
	require.Equal(t, "Erweiterter Infinitiv mit zu (Zusammenschreibung)", r.GetDescription())
	require.Equal(t, "https://languagetool.org/insights/de/beitrag/zu-zusammen-oder-getrennt/", r.GetURL())
	require.NotNil(t, r.GetCategory())
	require.Equal(t, "COMPOUNDING", r.GetCategory().GetID().String())
	require.NotEmpty(t, r.GetIncorrectExamples())
	require.NotEmpty(t, r.GetCorrectExamples())
	require.Contains(t, r.GetIncorrectExamples()[0].GetExample(), "sicher zu gehen")
	require.Contains(t, r.GetCorrectExamples()[0].GetExample(), "sicherzugehen")
	require.NotEmpty(t, r.GetAntiPatterns())
	require.Equal(t, len(CompoundInfinitivAntiPatterns), len(r.GetAntiPatterns()))
}
