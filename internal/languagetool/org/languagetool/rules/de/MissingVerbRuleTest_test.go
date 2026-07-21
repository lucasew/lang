package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/MissingVerbRuleTest.java
// Morph/POS inject only (no surface invent). Java is king for counts, spans, special cases.
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

func TestMissingVerbRule_Test(t *testing.T) {
	// Java Test#test: untagged AnalyzePlain has no PKT → not a "real sentence" (isRealSentence).
	rule := NewMissingVerbRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Da ist ein Verb, mal so zum testen."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Dieser Satz kein Verb."))))
}

func TestMissingVerbRule_JavaAssertGood(t *testing.T) {
	rule := NewMissingVerbRule(nil)

	// "Da ist ein Verb, mal so zum testen."
	require.Empty(t, rule.Match(mvSent(
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
	)))

	// Headline-like without sentence-end PKT? Java uses full LT; "Überschrift ohne Verb aber doch nicht zu kurz"
	// has no terminal .?! in the test string — getAnalyzedSentence may still not mark PKT.
	// Without final .?! isRealSentence is false → 0 matches.
	require.Empty(t, rule.Match(mvSent(
		atrWithPOS("Überschrift", "SUB:NOM:SIN:FEM", "Überschrift"),
		atrWithPOS("ohne", "APPR", "ohne"),
		atrWithPOS("Verb", "SUB:AKK:SIN:NEU", "Verb"),
		atrWithPOS("aber", "KON", "aber"),
		atrWithPOS("doch", "ADV", "doch"),
		atrWithPOS("nicht", "ADV", "nicht"),
		atrWithPOS("zu", "ADV", "zu"),
		atrWithPOS("kurz", "ADJ:PRD:GRU", "kurz"),
		// no PKT terminator
	)))

	// "Sprechen Sie vielleicht zufällig Türkisch?" — imperative/capital start; VER tag on first word
	require.Empty(t, rule.Match(mvSent(
		atrWithPOS("Sprechen", "VER:1:PLU:PRÄ:KJ1", "sprechen"),
		atrWithPOS("Sie", "PRO:PER:NOM:PLU:*", "sie"),
		atrWithPOS("vielleicht", "ADV", "vielleicht"),
		atrWithPOS("zufällig", "ADV", "zufällig"),
		atrWithPOS("Türkisch", "SUB:AKK:SIN:NEU", "Türkisch"),
		mvPKT("?"),
	)))

	// "Leg den Tresor in den Koffer im Kofferraum."
	require.Empty(t, rule.Match(mvSent(
		atrWithPOS("Leg", "VER:IMP:SIN:SFT", "legen"),
		atrWithPOS("den", "ART:DEF:AKK:SIN:MAS", "der"),
		atrWithPOS("Tresor", "SUB:AKK:SIN:MAS", "Tresor"),
		atrWithPOS("in", "APPR:AKK", "in"),
		atrWithPOS("den", "ART:DEF:AKK:SIN:MAS", "der"),
		atrWithPOS("Koffer", "SUB:AKK:SIN:MAS", "Koffer"),
		atrWithPOS("im", "APPRART:DAT:MAS", "in"),
		atrWithPOS("Kofferraum", "SUB:DAT:SIN:MAS", "Kofferraum"),
		mvPKT("."),
	)))

	// "Bring doch einfach deine Kinder mit."
	require.Empty(t, rule.Match(mvSent(
		atrWithPOS("Bring", "VER:IMP:SIN:SFT", "bringen"),
		atrWithPOS("doch", "ADV", "doch"),
		atrWithPOS("einfach", "ADV", "einfach"),
		atrWithPOS("deine", "PRO:POS:AKK:PLU:NEU", "dein"),
		atrWithPOS("Kinder", "SUB:AKK:PLU:NEU", "Kind"),
		atrWithPOS("mit", "PTKVZ", "mit"),
		mvPKT("."),
	)))

	// "Gut so." / "Ja!" — short (< MIN_TOKENS_FOR_ERROR=5 tokens without ws)
	// tokens: START + Gut + so + . = 4 < 5 → no error even without VER
	require.Empty(t, rule.Match(mvSent(
		atrWithPOS("Gut", "ADJ:PRD:GRU", "gut"),
		atrWithPOS("so", "ADV", "so"),
		mvPKT("."),
	)))
	require.Empty(t, rule.Match(mvSent(
		atrWithPOS("Ja", "ITJ", "ja"),
		mvPKT("!"),
	)))

	// "Vielen Dank für alles, was Du für mich getan hast." — special case rule1 + also has VER
	require.Empty(t, rule.Match(mvSent(
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
	)))

	// "Herzlichen Glückwunsch zu Deinem zwanzigsten Geburtstag." — special case rule2 (no VER needed)
	require.Empty(t, rule.Match(mvSent(
		atrWithPOS("Herzlichen", "ADJ:AKK:SIN:MAS:GRU:IND", "herzlich"),
		atrWithPOS("Glückwunsch", "SUB:AKK:SIN:MAS", "Glückwunsch"),
		atrWithPOS("zu", "APPR:DAT", "zu"),
		atrWithPOS("Deinem", "PRO:POS:DAT:SIN:MAS", "dein"),
		atrWithPOS("zwanzigsten", "ADJ:DAT:SIN:MAS:GRU:IND", "zwanzigst"),
		atrWithPOS("Geburtstag", "SUB:DAT:SIN:MAS", "Geburtstag"),
		mvPKT("."),
	)))
}

func TestMissingVerbRule_JavaAssertBad(t *testing.T) {
	rule := NewMissingVerbRule(nil)

	// "Dieser Satz kein Verb."
	// tokensWithoutWhitespace: START, Dieser, Satz, kein, Verb, . → len 6 ≥ 5
	// lastToken after full scan is "."; end = start(".")+1
	dieser := mvSent(
		atrWithPOS("Dieser", "PDAT:NOM:SIN:MAS", "dieser"),
		atrWithPOS("Satz", "SUB:NOM:SIN:MAS", "Satz"),
		atrWithPOS("kein", "PIAT:NOM:SIN:MAS", "kein"),
		atrWithPOS("Verb", "SUB:NOM:SIN:NEU", "Verb"),
		mvPKT("."),
	)
	ms := rule.Match(dieser)
	require.Len(t, ms, 1)
	require.Equal(t, "Dieser Satz scheint kein Verb zu enthalten", ms[0].GetMessage())
	require.Empty(t, ms[0].GetShortMessage())
	require.Equal(t, 0, ms[0].GetFromPos())
	// Java: lastToken is final "." → end = its startPos + length
	toks := dieser.GetTokensWithoutWhitespace()
	last := toks[len(toks)-1]
	require.Equal(t, last.GetStartPos()+utf16LenDE(last.GetToken()), ms[0].GetToPos())

	// "Aus einer Idee sich erste Wortgruppen, aus Wortgruppen einzelne Sätze, aus Sätzen ganze Texte."
	aus := mvSent(
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
	)
	ms = rule.Match(aus)
	require.Len(t, ms, 1)
	require.Equal(t, 0, ms[0].GetFromPos())
	toks = aus.GetTokensWithoutWhitespace()
	last = toks[len(toks)-1]
	require.Equal(t, last.GetStartPos()+utf16LenDE(last.GetToken()), ms[0].GetToPos())

	// "Ich ein neues Rad."
	ich := mvSent(
		atrWithPOS("Ich", "PRO:PER:NOM:SIN:MAS", "ich"),
		atrWithPOS("ein", "ART:IND:AKK:SIN:NEU", "ein"),
		atrWithPOS("neues", "ADJA:AKK:SIN:NEU:GRU:IND", "neu"),
		atrWithPOS("Rad", "SUB:AKK:SIN:NEU", "Rad"),
		mvPKT("."),
	)
	ms = rule.Match(ich)
	require.Len(t, ms, 1)
	require.Equal(t, "Dieser Satz scheint kein Verb zu enthalten", ms[0].GetMessage())
	toks = ich.GetTokensWithoutWhitespace()
	last = toks[len(toks)-1]
	require.Equal(t, last.GetStartPos()+utf16LenDE(last.GetToken()), ms[0].GetToPos())
}

func TestMissingVerbRule_MorphMissing(t *testing.T) {
	// Demo example pair surface: "In diesem Satz kein Wort."
	// All content tokens capitalized/tagged non-VER → missing verb.
	sent := mvSent(
		atrWithPOS("In", "APPR:DAT", "in"),
		atrWithPOS("diesem", "PDAT:DAT:SIN:MAS", "dieser"),
		atrWithPOS("Satz", "SUB:DAT:SIN:MAS", "Satz"),
		atrWithPOS("kein", "PIAT:NOM:SIN:NEU", "kein"),
		atrWithPOS("Wort", "SUB:NOM:SIN:NEU", "Wort"),
		mvPKT("."),
	)
	rule := NewMissingVerbRule(nil)
	ms := rule.Match(sent)
	require.NotEmpty(t, ms)
	require.Equal(t, "Dieser Satz scheint kein Verb zu enthalten", ms[0].GetMessage())
	require.Empty(t, ms[0].GetShortMessage())
	require.Equal(t, "Satz ohne Verb", rule.GetDescription())
	require.Equal(t, "MISSING_VERB", rule.GetID())
	require.True(t, rule.IsDefaultOff())
}

func TestMissingVerbRule_MorphWithVerb(t *testing.T) {
	sent := mvSent(
		atrWithPOS("Da", "ADV", "da"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("ein", "ART:IND:NOM:SIN:NEU", "ein"),
		atrWithPOS("Verb", "SUB:NOM:SIN:NEU", "Verb"),
		mvPKT("."),
	)
	require.Empty(t, NewMissingVerbRule(nil).Match(sent))
}

func TestMissingVerbRule_SpecialCaseVielenDank(t *testing.T) {
	// Surface special-case path (even without verb tags / without full morph)
	require.Empty(t, NewMissingVerbRule(nil).Match(languagetool.AnalyzePlain("Vielen Dank.")))
	// Morph path: Vielen Dank… with PKT, no VER
	require.Empty(t, NewMissingVerbRule(nil).Match(mvSent(
		atrWithPOS("Vielen", "PIAT", "viel"),
		atrWithPOS("Dank", "SUB:DAT:SIN:MAS", "Dank"),
		atrWithPOS("für", "APPR", "für"),
		atrWithPOS("alles", "PIS", "alles"),
		mvPKT("."),
	)))
	require.Empty(t, NewMissingVerbRule(nil).Match(mvSent(
		atrWithPOS("Herzlichen", "ADJA", "herzlich"),
		atrWithPOS("Glückwunsch", "SUB", "Glückwunsch"),
		atrWithPOS("zu", "APPR", "zu"),
		atrWithPOS("allem", "PIS", "alles"),
		mvPKT("."),
	)))
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

func TestMissingVerbRule_VerbAtSentenceStartHook(t *testing.T) {
	// Java verbAtSentenceStart: first content token capital, re-tag lowercased as VER.
	// Without hook: capitalized non-VER tagged word does not count → missing verb if rest non-VER.
	// With hook returning true for "sprechen": counts as verbFound.
	noHook := NewMissingVerbRule(nil)
	// Sprechen tagged as EIG/SUB (mis-tag), rest non-VER, ≥5 tokens
	sent := mvSent(
		atrWithPOS("Sprechen", "EIG:NOM:SIN:NEU", "Sprechen"), // mis-tagged capital start
		atrWithPOS("Sie", "PRO:PER:NOM:PLU:*", "sie"),
		atrWithPOS("vielleicht", "ADV", "vielleicht"),
		atrWithPOS("Türkisch", "SUB:AKK:SIN:NEU", "Türkisch"),
		mvPKT("?"),
	)
	// without hook: i==1 "Sprechen" not VER, not (!tagged&&!cap), hook nil → no verb → error
	require.Len(t, noHook.Match(sent), 1)

	withHook := NewMissingVerbRule(nil).WithTagFirstLowercased(func(lower string) bool {
		return lower == "sprechen"
	})
	// Rebuild sentence (Match may not mutate, but tokens already used — rebuild for safety)
	sent2 := mvSent(
		atrWithPOS("Sprechen", "EIG:NOM:SIN:NEU", "Sprechen"),
		atrWithPOS("Sie", "PRO:PER:NOM:PLU:*", "sie"),
		atrWithPOS("vielleicht", "ADV", "vielleicht"),
		atrWithPOS("Türkisch", "SUB:AKK:SIN:NEU", "Türkisch"),
		mvPKT("?"),
	)
	require.Empty(t, withHook.Match(sent2))
}

func TestMissingVerbRule_UntaggedLowercaseCountsAsVerb(t *testing.T) {
	// Java: !isTagged && !isCapitalizedWord → treat as possible verb (avoid false alarms)
	// untagged "fehlt" lowercase among tagged non-VER words
	sent := mvSent(
		atrWithPOS("In", "APPR:DAT", "in"),
		atrWithPOS("diesem", "PDAT:DAT:SIN:MAS", "dieser"),
		atrWithPOS("Satz", "SUB:DAT:SIN:MAS", "Satz"),
		// untagged lowercase token — no POS
		atrWithPOS("fehlt", "", ""),
		atrWithPOS("nichts", "PIS", "nichts"),
		mvPKT("."),
	)
	// Ensure untagged: empty POS should mean not tagged
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
}

func TestMissingVerbRule_Examples(t *testing.T) {
	r := NewMissingVerbRule(nil)
	require.NotEmpty(t, r.GetIncorrectExamples())
	require.NotEmpty(t, r.GetCorrectExamples())
}
