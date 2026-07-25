package de

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// demoLTReadability ports new JLanguageTool(new Demo()) for multi-sentence Analyze
// (same setup as core readability_rule_test).
func demoLTReadabilityDE() *languagetool.JLanguageTool {
	lt := languagetool.NewJLanguageTool("xx")
	lt.IgnoredCharacters = languagetool.GermanIgnoredCharactersRegex
	lt.SentenceTokenizer = tokenizers.NewSimpleSentenceTokenizer()
	return lt
}

func TestGermanReadabilityRule_IDsAndMetadata(t *testing.T) {
	simple := NewGermanReadabilityRule(nil, true)
	diff := NewGermanReadabilityRule(nil, false)

	require.Equal(t, "READABILITY_RULE_SIMPLE_DE", simple.GetID())
	require.Equal(t, "READABILITY_RULE_DIFFICULT_DE", diff.GetID())
	require.Equal(t, "READABILITY_RULE_SIMPLE_DE", simple.GetIDFor(true))
	require.Equal(t, "READABILITY_RULE_DIFFICULT_DE", simple.GetIDFor(false))

	require.Equal(t, "Lesbarkeit: Zu einfacher Text", simple.GetDescription())
	require.Equal(t, "Lesbarkeit: Zu schwieriger Text", diff.GetDescription())

	require.Equal(t, 3, simple.Level)
	require.True(t, simple.IsDefaultOff())
	require.Equal(t, rules.ITSStyle, simple.GetLocQualityIssueType())
	require.NotNil(t, simple.GetCategory())
	require.Equal(t, "CREATIVE_WRITING", simple.GetCategory().GetID().String())
	require.Equal(t, -1, simple.MinToCheckParagraph())

	require.Equal(t, "Grad der Lesbarkeit 0 (sehr schwierig) bis 6 (sehr einfach):", simple.GetConfigureText())
	opts := simple.GetRuleOptions()
	require.Len(t, opts, 1)
	require.Equal(t, 3, opts[0].DefaultValue)
	require.Equal(t, 0, opts[0].MinConfigurableValue)
	require.Equal(t, 6, opts[0].MaxConfigurableValue)
	require.Equal(t, simple.GetConfigureText(), opts[0].ConfigureText)
}

func TestGermanReadabilityRule_CategoryFromMessages(t *testing.T) {
	msgs := map[string]string{"category_creative_writing": "Kreatives Schreiben"}
	r := NewGermanReadabilityRule(msgs, false)
	require.Equal(t, "CREATIVE_WRITING", r.GetCategory().GetID().String())
	require.Equal(t, "Kreatives Schreiben", r.GetCategory().GetName())
}

func TestGermanReadabilityRule_PrintMessageLevel(t *testing.T) {
	r := NewGermanReadabilityRule(nil, false)
	require.Equal(t, " {Grad 0: Sehr schwer}", r.PrintMessageLevel(0))
	require.Equal(t, " {Grad 1: Schwer}", r.PrintMessageLevel(1))
	require.Equal(t, " {Grad 2: Mittelschwer}", r.PrintMessageLevel(2))
	require.Equal(t, " {Grad 3: Mittel}", r.PrintMessageLevel(3))
	require.Equal(t, " {Grad 4: Mittelleicht}", r.PrintMessageLevel(4))
	require.Equal(t, " {Grad 5: Leicht}", r.PrintMessageLevel(5))
	require.Equal(t, " {Grad 6: Sehr leicht}", r.PrintMessageLevel(6))
	require.Equal(t, "", r.PrintMessageLevel(7))
	require.Equal(t, "", r.PrintMessageLevel(-1))
}

func TestGermanReadabilityRule_GetMessage(t *testing.T) {
	diff := NewGermanReadabilityRule(nil, false)
	msg := diff.GetMessage(1, 40, 20, 2)
	require.Equal(t,
		"Lesbarkeit: Der Text dieses Absatzes ist zu schwierig {Grad 1: Schwer}. Zu viele Wörter pro Satz und zu viele Silben pro Wort.",
		msg)

	easy := NewGermanReadabilityRule(nil, true)
	msgE := easy.GetMessage(5, 85, 8, 1)
	require.Equal(t,
		"Lesbarkeit: Der Text dieses Absatzes ist zu einfach {Grad 5: Leicht}. Zu wenige Wörter pro Satz und zu wenige Silben pro Wort.",
		msgE)
}

func TestGermanReadabilityRule_GetFleschReadingEase_Amstad(t *testing.T) {
	r := NewGermanReadabilityRule(nil, false)
	// FRE = 180 - asl - 58.5 * asw
	require.InDelta(t, 180-10-(58.5*1.5), r.GetFleschReadingEase(10, 1.5), 1e-9)
	require.InDelta(t, 180.0, r.GetFleschReadingEase(0, 0), 1e-9)
	require.InDelta(t, 180-20-(58.5*2), r.GetFleschReadingEase(20, 2), 1e-9)
	// hook path used by base MatchList (same numbers)
	require.InDelta(t, 180-10-(58.5*1.5), r.ReadabilityRule.GetFleschReadingEase(10, 1.5), 1e-9)
}

func TestGermanReadabilityRule_SimpleSyllablesCount(t *testing.T) {
	r := NewGermanReadabilityRule(nil, false)
	// empty
	require.Equal(t, 0, r.SimpleSyllablesCount(""))
	// single vowel / consonant
	require.Equal(t, 1, r.SimpleSyllablesCount("a"))
	require.Equal(t, 1, r.SimpleSyllablesCount("x"))
	// "Haus": H,a,u,s — a vowel; u with cl=a → diphthong lastDouble; n=1
	// i=0 H not; start: a is vowel → 1; i=1 a already counted in start; wait:
	// Java: if vowel(0) n++; for i=1..len-1:
	// i=1 'a' vowel cl='H' → n++ → 2? Let's trace carefully:
	// word "Haus" length 4
	// char0 'H' not vowel → n=0
	// i=1 'a' vowel cl='H' → not lastDouble, not diphthong (diphthongs check c against cl vowels) → n=1
	// i=2 'u' vowel cl='a' → c=='u' && cl=='a' → lastDouble=true
	// i=3 's' not → lastDouble=false
	// n=1 → 1
	require.Equal(t, 1, r.SimpleSyllablesCount("Haus"))
	// "Idee": I,d,e,e
	// I vowel → n=1; i=1 d; i=2 e cl=d → n=2; i=3 e cl=e → lastDouble (e after e)
	// n=2 → 2
	require.Equal(t, 2, r.SimpleSyllablesCount("Idee"))
	// German umlaut vowels via IsVowel
	require.Equal(t, 1, r.SimpleSyllablesCount("Öl"))
	// "schön": s,c,h,ö,n — ö vowel → 1
	require.Equal(t, 1, r.SimpleSyllablesCount("schön"))
	// multi-syllable German word
	// "Verantwortlichkeiten" should have several syllables
	require.GreaterOrEqual(t, r.SimpleSyllablesCount("Verantwortlichkeiten"), 4)
	// hook path
	require.Equal(t, r.SimpleSyllablesCount("Haus"), r.ReadabilityRule.SimpleSyllablesCount("Haus"))
	// diphthong ai
	// "Mai": M,a,i — a → n=1; i with cl=a → lastDouble; n=1
	require.Equal(t, 1, r.SimpleSyllablesCount("Mai"))
	// "Baum": B,a,u,m — a →1; u cl=a lastDouble; n=1
	require.Equal(t, 1, r.SimpleSyllablesCount("Baum"))
	// "See": S,e,e — e →1; e cl=e lastDouble; n=1
	require.Equal(t, 1, r.SimpleSyllablesCount("See"))
	// "Haar": H,a,a,r — a →1; a cl=a lastDouble; n=1
	require.Equal(t, 1, r.SimpleSyllablesCount("Haar"))
	// two separate vowels: "Ida" I,d,a → I n=1; a cl=d n=2
	require.Equal(t, 2, r.SimpleSyllablesCount("Ida"))
	// after lastDouble the next vowel counts: "Idee" already; "Aaa" A,a,a
	// A n=1; a cl=A lastDouble; a lastDouble → n=2 lastDouble=false
	require.Equal(t, 2, r.SimpleSyllablesCount("Aaa"))
}

func TestGermanReadabilityRule_Match_Difficult_SingleParagraph(t *testing.T) {
	lt := demoLTReadabilityDE()
	// Long multisyllabic German words → low Amstad FRE → rLevel < 3.
	// Final quirk: (!tooEasy && rLevel < level) returns matches without nParagraph>1.
	words := []string{
		"Verantwortlichkeiten", "Komplexitätsanalysen", "Internationalisierung",
		"Charakterisierungen", "Individualisierungen", "Differenzierungen",
		"Konstitutionalismus", "Überintellectualisierung", "Gegenrevolutionäre",
		"Elektroenzephalographie", "Interdisziplinarität",
	}
	text := strings.Join(words, " ") + "."
	r := NewGermanReadabilityRule(nil, false)
	matches := r.MatchList(lt.Analyze(text))
	require.NotEmpty(t, matches, "expected difficult match for long multisyllabic German words")
	require.Contains(t, matches[0].Message, "zu schwierig")
	require.Contains(t, matches[0].Message, "Zu viele Wörter pro Satz")
	require.Contains(t, matches[0].Message, "{Grad ")
	// Span: tokens[1].start .. tokens[MARK_WORDS=3].end
	sents := lt.Analyze(text)
	require.NotEmpty(t, sents)
	toks := sents[0].GetTokensWithoutWhitespace()
	require.Greater(t, len(toks), 3)
	require.Equal(t, toks[1].GetStartPos(), matches[0].GetFromPos())
	require.Equal(t, toks[3].GetEndPos(), matches[0].GetToPos())
}

func TestGermanReadabilityRule_Match_TooEasy_NeedsMultipleParagraphs(t *testing.T) {
	lt := demoLTReadabilityDE()
	// Easy German: short words / short sentences → high FRE.
	// Single paragraph: Java final condition needs nParagraph > 1 for tooEasy → empty.
	easyPara := "Der Hund saß auf der Matte. Die Katze lief in den Park. Ein Mann sah ein großes rotes Auto vor dem Haus."
	r := NewGermanReadabilityRule(nil, true)
	require.Empty(t, r.MatchList(lt.Analyze(easyPara)), "single paragraph tooEasy must not return matches (Java quirk)")

	easyLots := strings.Repeat("Der Hund saß auf der Matte. ", 8) + "\n\n" + strings.Repeat("Die Katze lief zu mir. ", 8)
	matches := r.MatchList(lt.Analyze(easyLots))
	require.NotEmpty(t, matches, "very easy multi-paragraph German text should match tooEasy rule")
	require.Contains(t, matches[0].Message, "zu einfach")
	require.Contains(t, matches[0].Message, "Zu wenige Wörter pro Satz")
}

func TestGermanReadabilityRule_Match_BelowMinWords(t *testing.T) {
	lt := demoLTReadabilityDE()
	r := NewGermanReadabilityRule(nil, false)
	require.Empty(t, r.MatchList(lt.Analyze("Hallo dort.")))
	rEasy := NewGermanReadabilityRule(nil, true)
	require.Empty(t, rEasy.MatchList(lt.Analyze("Hallo dort Freund.")))
}

func TestGermanReadabilityRule_CtorLevel(t *testing.T) {
	r := NewGermanReadabilityRuleFull(nil, nil, false, 5, true)
	require.Equal(t, 5, r.Level)
	require.False(t, r.IsDefaultOff())
	require.Equal(t, "READABILITY_RULE_DIFFICULT_DE", r.GetID())
}
