package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/MissingVerbRuleTest.java
// Morph/POS inject only where full de-DE analysis does not yet attach PKT/VER like Java LT.
// Java is king for counts, spans, special cases, and the assertGood/assertBad matrix.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func mvPKT(tok string) *languagetool.AnalyzedTokenReadings {
	r := atrWithPOS(tok, "PKT", tok)
	r.SetSentEnd()
	return r
}

func mvSent(toks ...*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedSentence {
	all := make([]*languagetool.AnalyzedTokenReadings, 0, len(toks)+1)
	all = append(all, sentStartATR())
	all = append(all, toks...)
	return languagetool.NewAnalyzedSentence(withPositions(all...))
}

// assertGoodMV ports MissingVerbRuleTest.assertGood (0 matches).
func assertGoodMV(t *testing.T, rule *MissingVerbRule, sent *languagetool.AnalyzedSentence, label string) {
	t.Helper()
	require.Equal(t, 0, len(rule.Match(sent)), "Found unexpected error in: %q", label)
}

// assertBadMV ports MissingVerbRuleTest.assertBad (exactly 1 match).
func assertBadMV(t *testing.T, rule *MissingVerbRule, sent *languagetool.AnalyzedSentence, label string) {
	t.Helper()
	ms := rule.Match(sent)
	require.Equal(t, 1, len(ms), "Did not find expected error in: %q", label)
	require.Equal(t, "Dieser Satz scheint kein Verb zu enthalten", ms[0].GetMessage())
	require.Equal(t, 0, ms[0].GetFromPos())
	toks := sent.GetTokensWithoutWhitespace()
	last := toks[len(toks)-1]
	require.Equal(t, last.GetStartPos()+utf16LenDE(last.GetToken()), ms[0].GetToPos())
}

// TestMissingVerbRule_Test ports MissingVerbRuleTest.test full matrix.
func TestMissingVerbRule_Test(t *testing.T) {
	rule := NewMissingVerbRule(nil)

	// --- assertGood (Java MissingVerbRuleTest) ---

	// "Da ist ein Verb, mal so zum testen."
	assertGoodMV(t, rule, mvSent(
		atrWithPOS("Da", "ADV", "da"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("ein", "ART:IND:NOM:SIN:NEU", "ein"),
		atrWithPOS("Verb", "SUB:NOM:SIN:NEU", "Verb"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("mal", "ADV", "mal"),
		atrWithPOS("so", "ADV", "so"),
		atrWithPOS("zum", "APPRART:DAT:NEU", "zu"),
		atrWithPOS("testen", "VER:INF:NON", "testen"),
		mvPKT("."),
	), "Da ist ein Verb, mal so zum testen.")

	// "Überschrift ohne Verb aber doch nicht zu kurz" — no terminal .?! → not real sentence
	assertGoodMV(t, rule, mvSent(
		atrWithPOS("Überschrift", "SUB:NOM:SIN:FEM", "Überschrift"),
		atrWithPOS("ohne", "APPR", "ohne"),
		atrWithPOS("Verb", "SUB:AKK:SIN:NEU", "Verb"),
		atrWithPOS("aber", "KON", "aber"),
		atrWithPOS("doch", "ADV", "doch"),
		atrWithPOS("nicht", "ADV", "nicht"),
		atrWithPOS("zu", "ADV", "zu"),
		atrWithPOS("kurz", "ADJ:PRD:GRU", "kurz"),
	), "Überschrift ohne Verb aber doch nicht zu kurz")

	// "Sprechen Sie vielleicht zufällig Türkisch?" — VER on first content word
	assertGoodMV(t, rule, mvSent(
		atrWithPOS("Sprechen", "VER:1:PLU:PRÄ:KJ1", "sprechen"),
		atrWithPOS("Sie", "PRO:PER:NOM:PLU:*", "sie"),
		atrWithPOS("vielleicht", "ADV", "vielleicht"),
		atrWithPOS("zufällig", "ADV", "zufällig"),
		atrWithPOS("Türkisch", "SUB:AKK:SIN:NEU", "Türkisch"),
		mvPKT("?"),
	), "Sprechen Sie vielleicht zufällig Türkisch?")

	// "Leg den Tresor in den Koffer im Kofferraum."
	assertGoodMV(t, rule, mvSent(
		atrWithPOS("Leg", "VER:IMP:SIN:SFT", "legen"),
		atrWithPOS("den", "ART:DEF:AKK:SIN:MAS", "der"),
		atrWithPOS("Tresor", "SUB:AKK:SIN:MAS", "Tresor"),
		atrWithPOS("in", "APPR:AKK", "in"),
		atrWithPOS("den", "ART:DEF:AKK:SIN:MAS", "der"),
		atrWithPOS("Koffer", "SUB:AKK:SIN:MAS", "Koffer"),
		atrWithPOS("im", "APPRART:DAT:MAS", "in"),
		atrWithPOS("Kofferraum", "SUB:DAT:SIN:MAS", "Kofferraum"),
		mvPKT("."),
	), "Leg den Tresor in den Koffer im Kofferraum.")

	// "Bring doch einfach deine Kinder mit."
	assertGoodMV(t, rule, mvSent(
		atrWithPOS("Bring", "VER:IMP:SIN:SFT", "bringen"),
		atrWithPOS("doch", "ADV", "doch"),
		atrWithPOS("einfach", "ADV", "einfach"),
		atrWithPOS("deine", "PRO:POS:AKK:PLU:NEU", "dein"),
		atrWithPOS("Kinder", "SUB:AKK:PLU:NEU", "Kind"),
		atrWithPOS("mit", "PTKVZ", "mit"),
		mvPKT("."),
	), "Bring doch einfach deine Kinder mit.")

	// "Gut so." / "Ja!" — short (< MIN_TOKENS_FOR_ERROR=5 tokens without ws incl. SENT_START)
	assertGoodMV(t, rule, mvSent(
		atrWithPOS("Gut", "ADJ:PRD:GRU", "gut"),
		atrWithPOS("so", "ADV", "so"),
		mvPKT("."),
	), "Gut so.")
	assertGoodMV(t, rule, mvSent(
		atrWithPOS("Ja", "ITJ", "ja"),
		mvPKT("!"),
	), "Ja!")

	// "Vielen Dank für alles, was Du für mich getan hast." — special case rule1 + VER
	assertGoodMV(t, rule, mvSent(
		atrWithPOS("Vielen", "PIAT:DAT:SIN:MAS", "viel"),
		atrWithPOS("Dank", "SUB:DAT:SIN:MAS", "Dank"),
		atrWithPOS("für", "APPR:AKK", "für"),
		atrWithPOS("alles", "PIS:AKK:SIN:NEU", "alles"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("was", "PRELS:AKK:SIN:NEU", "was"),
		atrWithPOS("Du", "PRO:PER:NOM:SIN:MAS", "du"),
		atrWithPOS("für", "APPR:AKK", "für"),
		atrWithPOS("mich", "PRO:PER:AKK:SIN:MAS", "ich"),
		atrWithPOS("getan", "VER:PA2:SFT", "tun"),
		atrWithPOS("hast", "VER:2:SIN:PRÄ:NON", "haben"),
		mvPKT("."),
	), "Vielen Dank für alles, was Du für mich getan hast.")

	// "Herzlichen Glückwunsch zu Deinem zwanzigsten Geburtstag." — special case rule2
	assertGoodMV(t, rule, mvSent(
		atrWithPOS("Herzlichen", "ADJ:AKK:SIN:MAS:GRU:IND", "herzlich"),
		atrWithPOS("Glückwunsch", "SUB:AKK:SIN:MAS", "Glückwunsch"),
		atrWithPOS("zu", "APPR:DAT", "zu"),
		atrWithPOS("Deinem", "PRO:POS:DAT:SIN:MAS", "dein"),
		atrWithPOS("zwanzigsten", "ADJ:DAT:SIN:MAS:GRU:IND", "zwanzigst"),
		atrWithPOS("Geburtstag", "SUB:DAT:SIN:MAS", "Geburtstag"),
		mvPKT("."),
	), "Herzlichen Glückwunsch zu Deinem zwanzigsten Geburtstag.")

	// --- assertBad (Java MissingVerbRuleTest) ---

	// "Dieser Satz kein Verb."
	assertBadMV(t, rule, mvSent(
		atrWithPOS("Dieser", "PDAT:NOM:SIN:MAS", "dieser"),
		atrWithPOS("Satz", "SUB:NOM:SIN:MAS", "Satz"),
		atrWithPOS("kein", "PIAT:NOM:SIN:MAS", "kein"),
		atrWithPOS("Verb", "SUB:NOM:SIN:NEU", "Verb"),
		mvPKT("."),
	), "Dieser Satz kein Verb.")

	// "Aus einer Idee sich erste Wortgruppen, aus Wortgruppen einzelne Sätze, aus Sätzen ganze Texte."
	assertBadMV(t, rule, mvSent(
		atrWithPOS("Aus", "APPR:DAT", "aus"),
		atrWithPOS("einer", "ART:IND:DAT:SIN:FEM", "ein"),
		atrWithPOS("Idee", "SUB:DAT:SIN:FEM", "Idee"),
		atrWithPOS("sich", "PRF:AKK:SIN", "sich"),
		atrWithPOS("erste", "ADJA:NOM:PLU:FEM:GRU:IND", "erst"),
		atrWithPOS("Wortgruppen", "SUB:NOM:PLU:FEM", "Wortgruppe"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("aus", "APPR:DAT", "aus"),
		atrWithPOS("Wortgruppen", "SUB:DAT:PLU:FEM", "Wortgruppe"),
		atrWithPOS("einzelne", "ADJA:NOM:PLU:MAS:GRU:IND", "einzeln"),
		atrWithPOS("Sätze", "SUB:NOM:PLU:MAS", "Satz"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("aus", "APPR:DAT", "aus"),
		atrWithPOS("Sätzen", "SUB:DAT:PLU:MAS", "Satz"),
		atrWithPOS("ganze", "ADJA:NOM:PLU:MAS:GRU:IND", "ganz"),
		atrWithPOS("Texte", "SUB:NOM:PLU:MAS", "Text"),
		mvPKT("."),
	), "Aus einer Idee sich erste Wortgruppen, aus Wortgruppen einzelne Sätze, aus Sätzen ganze Texte.")

	// "Ich ein neues Rad."
	assertBadMV(t, rule, mvSent(
		atrWithPOS("Ich", "PRO:PER:NOM:SIN:MAS", "ich"),
		atrWithPOS("ein", "ART:IND:AKK:SIN:NEU", "ein"),
		atrWithPOS("neues", "ADJA:AKK:SIN:NEU:GRU:IND", "neu"),
		atrWithPOS("Rad", "SUB:AKK:SIN:NEU", "Rad"),
		mvPKT("."),
	), "Ich ein neues Rad.")

	// Java: //assertBad("Ich einen neuen Fehler gefunden.", lt);  // see issue #42 — not asserted
}

func TestMissingVerbRule_MetaAndExamples(t *testing.T) {
	r := NewMissingVerbRule(nil)
	require.Equal(t, "MISSING_VERB", r.GetID())
	require.Equal(t, "Satz ohne Verb", r.GetDescription())
	require.True(t, r.IsDefaultOff())
	require.NotNil(t, r.GetCategory())
	require.NotEmpty(t, r.GetIncorrectExamples())
	require.NotEmpty(t, r.GetCorrectExamples())
	// Demo surface from Java addExamplePair
	require.Equal(t, []string{"fehlt"}, r.GetIncorrectExamples()[0].GetCorrections())
}

func TestMissingVerbRule_SpecialCasePatternRules(t *testing.T) {
	// isSpecialCase uses PatternRule match (case-sensitive PatternToken), not invent string-at-start.
	rule := NewMissingVerbRule(nil)

	// Surface path: PatternRule matches plain tokens
	require.Empty(t, rule.Match(languagetool.AnalyzePlain("Vielen Dank."))) // not real (no PKT) → 0
	// Morph: special case with PKT, no VER, ≥5 tokens → still 0 via rule1
	assertGoodMV(t, rule, mvSent(
		atrWithPOS("Vielen", "PIAT", "viel"),
		atrWithPOS("Dank", "SUB:DAT:SIN:MAS", "Dank"),
		atrWithPOS("für", "APPR", "für"),
		atrWithPOS("alles", "PIS", "alles"),
		mvPKT("."),
	), "Vielen Dank für alles.")

	assertGoodMV(t, rule, mvSent(
		atrWithPOS("Herzlichen", "ADJA", "herzlich"),
		atrWithPOS("Glückwunsch", "SUB", "Glückwunsch"),
		atrWithPOS("zu", "APPR", "zu"),
		atrWithPOS("allem", "PIS", "alles"),
		mvPKT("."),
	), "Herzlichen Glückwunsch zu allem.")

	// PatternRule matches mid-sentence too (Java rule1.match anywhere)
	assertGoodMV(t, rule, mvSent(
		atrWithPOS("Also", "ADV", "also"),
		atrWithPOS("nochmal", "ADV", "nochmal"),
		atrWithPOS("Vielen", "PIAT", "viel"),
		atrWithPOS("Dank", "SUB", "Dank"),
		atrWithPOS("dafür", "ADV", "dafür"),
		mvPKT("."),
	), "Also nochmal Vielen Dank dafür.")
}

func TestMissingVerbRule_ShortSentence(t *testing.T) {
	// fewer than MIN_TOKENS_FOR_ERROR (5 tokens without whitespace including START)
	// START + Hallo + Welt + . = 4 < 5
	sent := mvSent(
		atrWithPOS("Hallo", "ITJ", "hallo"),
		atrWithPOS("Welt", "SUB:NOM:SIN:FEM", "Welt"),
		mvPKT("."),
	)
	require.Empty(t, NewMissingVerbRule(nil).Match(sent))
}

func TestMissingVerbRule_IsRealSentence(t *testing.T) {
	rule := NewMissingVerbRule(nil)
	// no terminal punct → not real
	require.Empty(t, rule.Match(mvSent(
		atrWithPOS("Dieser", "PDAT", "dieser"),
		atrWithPOS("Satz", "SUB", "Satz"),
		atrWithPOS("kein", "PIAT", "kein"),
		atrWithPOS("Verb", "SUB", "Verb"),
	)))
	// terminal ";" is not .?! → not real even with PKT tag
	semi := atrWithPOS(";", "PKT", ";")
	semi.SetSentEnd()
	require.Empty(t, rule.Match(mvSent(
		atrWithPOS("Dieser", "PDAT", "dieser"),
		atrWithPOS("Satz", "SUB", "Satz"),
		atrWithPOS("kein", "PIAT", "kein"),
		atrWithPOS("Verb", "SUB", "Verb"),
		semi,
	)))
	// plain analysis: no PKT → not real (both good and bad strings)
	require.Empty(t, rule.Match(languagetool.AnalyzePlain("Da ist ein Verb, mal so zum testen.")))
	require.Empty(t, rule.Match(languagetool.AnalyzePlain("Dieser Satz kein Verb.")))
}

func TestMissingVerbRule_VerbAtSentenceStartHook(t *testing.T) {
	// Java verbAtSentenceStart: first content token capital, re-tag lowercased as VER.
	// Without hook: capitalized non-VER tagged word does not count → missing verb if rest non-VER.
	// With hook returning true for "sprechen": counts as verbFound.
	noHook := NewMissingVerbRule(nil)
	sent := mvSent(
		atrWithPOS("Sprechen", "EIG:NOM:SIN:NEU", "Sprechen"), // mis-tagged capital start
		atrWithPOS("Sie", "PRO:PER:NOM:PLU:*", "sie"),
		atrWithPOS("vielleicht", "ADV", "vielleicht"),
		atrWithPOS("Türkisch", "SUB:AKK:SIN:NEU", "Türkisch"),
		mvPKT("?"),
	)
	require.Len(t, noHook.Match(sent), 1)

	withHook := NewMissingVerbRule(nil).WithTagFirstLowercased(func(lower string) bool {
		return lower == "sprechen"
	})
	sent2 := mvSent(
		atrWithPOS("Sprechen", "EIG:NOM:SIN:NEU", "Sprechen"),
		atrWithPOS("Sie", "PRO:PER:NOM:PLU:*", "sie"),
		atrWithPOS("vielleicht", "ADV", "vielleicht"),
		atrWithPOS("Türkisch", "SUB:AKK:SIN:NEU", "Türkisch"),
		mvPKT("?"),
	)
	require.Empty(t, withHook.Match(sent2))
}

func TestMissingVerbRule_VerbAtSentenceStartWiredTagger(t *testing.T) {
	// Prefer real GermanTagger when resources available (Java language.getTagger().tag).
	rule := WireMissingVerbRule(NewMissingVerbRule(nil))
	if rule.TagFirstLowercased == nil {
		t.Skip("GermanTagger resources unavailable")
	}
	// "Sprechen" mis-tagged as non-VER; lowercased re-tag should find VER.
	sent := mvSent(
		atrWithPOS("Sprechen", "EIG:NOM:SIN:NEU", "Sprechen"),
		atrWithPOS("Sie", "PRO:PER:NOM:PLU:*", "sie"),
		atrWithPOS("vielleicht", "ADV", "vielleicht"),
		atrWithPOS("Türkisch", "SUB:AKK:SIN:NEU", "Türkisch"),
		mvPKT("?"),
	)
	require.Empty(t, rule.Match(sent), "wired tagger should accept sentence-start Sprechen as VER")
}

func TestMissingVerbRule_UntaggedLowercaseCountsAsVerb(t *testing.T) {
	// Java: !isTagged && !isCapitalizedWord → treat as possible verb (avoid false alarms)
	sent := mvSent(
		atrWithPOS("In", "APPR:DAT", "in"),
		atrWithPOS("diesem", "PDAT:DAT:SIN:MAS", "dieser"),
		atrWithPOS("Satz", "SUB:DAT:SIN:MAS", "Satz"),
		// untagged lowercase token — empty POS → nil pos → not tagged
		atrWithPOS("fehlt", "", ""),
		atrWithPOS("nichts", "PIS", "nichts"),
		mvPKT("."),
	)
	require.Empty(t, NewMissingVerbRule(nil).Match(sent))
}

func TestMissingVerbRule_OperatorPrecedence_UntaggedCapitalNotVerb(t *testing.T) {
	// Capitalized untagged word is NOT treated as verb by (!tagged && !capitalized).
	// Without VER and without sentence-start hook → missing verb if long enough.
	sent := mvSent(
		atrWithPOS("Foo", "", ""), // untagged + capitalized
		atrWithPOS("Bar", "SUB", "Bar"),
		atrWithPOS("Baz", "SUB", "Baz"),
		atrWithPOS("Qux", "SUB", "Qux"),
		mvPKT("."),
	)
	// i==1 is Foo; verbAtSentenceStart nil → no; !tagged && !cap = true && false = false
	require.Len(t, NewMissingVerbRule(nil).Match(sent), 1)
}

func TestMissingVerbRule_DemoExampleMorph(t *testing.T) {
	// Java addExamplePair surface: "In diesem Satz kein Wort."
	sent := mvSent(
		atrWithPOS("In", "APPR:DAT", "in"),
		atrWithPOS("diesem", "PDAT:DAT:SIN:MAS", "dieser"),
		atrWithPOS("Satz", "SUB:DAT:SIN:MAS", "Satz"),
		atrWithPOS("kein", "PIAT:NOM:SIN:NEU", "kein"),
		atrWithPOS("Wort", "SUB:NOM:SIN:NEU", "Wort"),
		mvPKT("."),
	)
	ms := NewMissingVerbRule(nil).Match(sent)
	require.Len(t, ms, 1)
	require.Equal(t, "Dieser Satz scheint kein Verb zu enthalten", ms[0].GetMessage())
	require.Empty(t, ms[0].GetShortMessage())
}
