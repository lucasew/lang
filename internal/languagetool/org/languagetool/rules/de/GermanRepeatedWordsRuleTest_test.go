package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/GermanRepeatedWordsRuleTest.java
//
// Java setup:
//   rule = new GermanRepeatedWordsRule(TestTools.getMessages("de"));
//   lt = new JLanguageTool(Languages.getLanguageForShortCode("de"));
//   assertCorrectText / getRuleMatches → rule.match(lt.analyzeText(s), aText)
//
// Go analysis path (Java analyzeText):
//   NewJLanguageTool("de") + WireGermanTagWord (createDefaultTagger) + lt.Analyze
// Synonyms: official /de/synonyms.txt (data/synonyms.txt embed, byte-identical to Java resource).
// Synthesizer: GermanSynthesizer.INSTANCE via openDiscoveredGermanSynthesizer.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// germanRepeatedWordsLT ports Java @Before: de language + real tagger resources.
func germanRepeatedWordsLT(t *testing.T) (*GermanRepeatedWordsRule, *languagetool.JLanguageTool) {
	t.Helper()
	// Java: Languages.getLanguageForShortCode("de"); new JLanguageTool(lang)
	lt := languagetool.NewJLanguageTool("de")
	// Java createDefaultTagger → GermanTagger.INSTANCE (required for lemma/POS lookup)
	require.True(t, WireGermanTagWord(lt), "GermanTagger resources required for GermanRepeatedWordsRule twin")
	require.NotNil(t, DiscoveredGermanTagger(), "GermanTagger.INSTANCE resources required")
	// Java: new GermanRepeatedWordsRule(TestTools.getMessages("de"))
	rule := NewGermanRepeatedWordsRule(nil)
	return rule, lt
}

// analyzeDERepeatedWordsText ports JLanguageTool("de").analyzeText (SRX + getAnalyzedSentence).
func analyzeDERepeatedWordsText(t *testing.T, lt *languagetool.JLanguageTool, text string) []*languagetool.AnalyzedSentence {
	t.Helper()
	sents := lt.Analyze(text)
	require.NotEmpty(t, sents, "Analyze(%q)", text)
	return sents
}

// TestGermanRepeatedWordsRule_Rule twins Java GermanRepeatedWordsRuleTest.testRule.
func TestGermanRepeatedWordsRule_Rule(t *testing.T) {
	rule, lt := germanRepeatedWordsLT(t)

	assertCorrectText := func(s string) {
		t.Helper()
		require.Equal(t, 0, len(rule.MatchList(analyzeDERepeatedWordsText(t, lt, s))), "correct: %q", s)
	}
	getRuleMatches := func(s string) []*rules.RuleMatch {
		t.Helper()
		return rule.MatchList(analyzeDERepeatedWordsText(t, lt, s))
	}

	// Java: assertCorrectText("Lauf so schnell wie möglich. Das ist MÖGLICH!");
	// MÖGLICH is all-uppercase (isException) and möglich is commented-out in synonyms.txt —
	// either path yields 0 matches; real analysis must not invent a hit.
	assertCorrectText("Lauf so schnell wie möglich. Das ist MÖGLICH!")

	// Java: "Das ist hervorragende Arbeit. Die Ergebnisse sind hervorragend."
	// → 1 match; suggestions exactly [brillant, exzellent, perfekt, fantastisch, traumhaft]
	matches := getRuleMatches("Das ist hervorragende Arbeit. Die Ergebnisse sind hervorragend.")
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"brillant", "exzellent", "perfekt", "fantastisch", "traumhaft"},
		matches[0].GetSuggestedReplacements())

	// Java-visible rule surface (specific id from toId(lemma, de), short message, description)
	require.Equal(t, "DE_REPEATEDWORDS_HERVORRAGEND", matches[0].GetSpecificRuleId())
	require.Equal(t, "Stil: Wortwiederholung", matches[0].GetShortMessage())
	require.Equal(t, "Synonyme für wiederholte Wörter.", rule.GetDescription())
	require.Equal(t, "DE_REPEATEDWORDS", rule.GetID())
	require.Equal(t, 1, rule.MinToCheckParagraph())

	// Java JSON id = getSpecificRuleId(); ToLocalMatches must not drop lemma suffix
	lm := rules.ToLocalMatches(matches)
	require.Len(t, lm, 1)
	require.Equal(t, "DE_REPEATEDWORDS_HERVORRAGEND", lm[0].RuleID)
	require.Equal(t, "Stil: Wortwiederholung", lm[0].ShortMessage)
}

// TestGermanRepeatedWordsRule_IsExceptionCapitalizedMidSentence ports
// GermanRepeatedWordsRule.isException: isCapitalized && !sentStart → true.
// Hand-built tokens only for the isException predicate (same pattern as EN NNP unit test).
func TestGermanRepeatedWordsRule_IsExceptionCapitalizedMidSentence(t *testing.T) {
	ss := languagetool.SentenceStartTagName
	s1 := languagetool.NewAnalyzedSentence(withPositions(
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Leider", "ADV", "leider"),
		atrWithPOS("regnet", "VER:3:SIN:PRÄ:SFT", "regnen"),
		atrWithPOS("es", "PRO:PER:NOM:SIN:NEU", "es"),
		atrWithPOS(".", "PKT", "."),
	))
	s2cap := languagetool.NewAnalyzedSentence(withPositions(
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Es", "PRO:PER:NOM:SIN:NEU", "es"),
		atrWithPOS("war", "VER:3:SIN:PRT:NON", "sein"),
		atrWithPOS("Leider", "ADV", "leider"), // mid-sentence capitalized
		atrWithPOS("kalt", "ADJ:PRD:GRU", "kalt"),
		atrWithPOS(".", "PKT", "."),
	))
	rule := NewGermanRepeatedWordsRule(nil)
	// isCapitalized(Leider) mid-sentence → exception → no match
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2cap})))

	s2low := languagetool.NewAnalyzedSentence(withPositions(
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Es", "PRO:PER:NOM:SIN:NEU", "es"),
		atrWithPOS("war", "VER:3:SIN:PRT:NON", "sein"),
		atrWithPOS("leider", "ADV", "leider"),
		atrWithPOS("kalt", "ADJ:PRD:GRU", "kalt"),
		atrWithPOS(".", "PKT", "."),
	))
	// First sentence "Leider" at start is kept; second "leider" not exception → match
	require.Equal(t, 1, len(rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2low})))
}

// TestGermanRepeatedWordsIsException_EIG ports isException hasPosTagStartingWith("EIG:").
func TestGermanRepeatedWordsIsException_EIG(t *testing.T) {
	eig := atrWithPOS("Berlin", "EIG:NOM:SIN:NEU", "Berlin")
	toks := []*languagetool.AnalyzedTokenReadings{eig}
	require.True(t, germanRepeatedWordsIsException(toks, 0, true, true, false))
	// all-uppercase exception
	require.True(t, germanRepeatedWordsIsException(toks, 0, true, false, true))
	// mid-sentence capitalized exception
	require.True(t, germanRepeatedWordsIsException(toks, 0, false, true, false))
	// sentence-start non-EIG lower-like (not capitalized, not all-upper) is not exception via those gates;
	// EIG: still returns true via POS — use non-EIG token for false case.
	adv := atrWithPOS("leider", "ADV", "leider")
	require.False(t, germanRepeatedWordsIsException([]*languagetool.AnalyzedTokenReadings{adv}, 0, true, false, false))
}

// TestGermanRepeatedWords_CategoryAndIssueType ports AbstractRepeatedWordsRule ctor meta (no picky on DE).
func TestGermanRepeatedWords_CategoryAndIssueType(t *testing.T) {
	r := NewGermanRepeatedWordsRule(nil)
	require.NotNil(t, r.GetCategory())
	require.Equal(t, rules.NewCategoryId("REPETITIONS_STYLE"), r.GetCategory().GetID())
	require.Equal(t, rules.ITSStyle, r.GetLocQualityIssueType())
	require.False(t, r.HasTag(rules.TagPicky), "DE has no picky tag in Java")
	// Java getMessage
	require.Equal(t, "Dieses Wort kommt in einem nahe gelegenen vorherigen Satz bereits vor. "+
		"Verwenden Sie ein Synonym, um Ihren Text abwechslungsreicher zu gestalten, außer die Wiederholung ist beabsichtigt.",
		r.Message)
	require.Equal(t, "Stil: Wortwiederholung", r.ShortMsg)
}

// TestGermanRepeatedWords_SynonymsFromOfficialResource checks wordsToCheck keys from official synonyms.txt
// (Java loadWords("/de/synonyms.txt")) — no invent list.
func TestGermanRepeatedWords_SynonymsFromOfficialResource(t *testing.T) {
	r := NewGermanRepeatedWordsRule(nil)
	require.NotNil(t, r.WordsToCheck)
	// Active entries in official de/synonyms.txt (commented lines are not loaded).
	for _, key := range []string{"meist", "nun", "insbesondere", "teilweise", "außerdem", "leider", "hervorragend", "beeindruckend"} {
		_, ok := r.WordsToCheck[key]
		require.True(t, ok, "missing official synonyms key %q", key)
	}
	// Commented-out möglich must not appear
	_, ok := r.WordsToCheck["möglich"]
	require.False(t, ok, "commented #möglich must not load")
	// hervorragend synonyms order as file
	h := r.WordsToCheck["hervorragend"]
	require.NotNil(t, h)
	require.Equal(t, []string{"brillant", "exzellent", "perfekt", "fantastisch", "traumhaft"}, h.Synonyms)
	require.Equal(t, "ADJ.*", h.Postag)
}
