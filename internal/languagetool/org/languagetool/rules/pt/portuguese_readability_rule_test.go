package pt

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// demoLTReadabilityPT ports new JLanguageTool(new Demo()) for multi-sentence Analyze
// (same setup as core readability_rule_test / DE GermanReadabilityRuleTest).
func demoLTReadabilityPT() *languagetool.JLanguageTool {
	lt := languagetool.NewJLanguageTool("xx")
	lt.IgnoredCharacters = languagetool.GermanIgnoredCharactersRegex
	lt.SentenceTokenizer = tokenizers.NewSimpleSentenceTokenizer()
	return lt
}

func TestPortugueseReadabilityRule_IDsAndMetadata(t *testing.T) {
	simple := NewPortugueseReadabilityRule(true, -1)
	diff := NewPortugueseReadabilityRule(false, -1)

	require.Equal(t, "READABILITY_RULE_SIMPLE_PT", simple.GetID())
	require.Equal(t, "READABILITY_RULE_DIFFICULT_PT", diff.GetID())
	require.Equal(t, "READABILITY_RULE_SIMPLE_PT", simple.GetIDFor(true))
	require.Equal(t, "READABILITY_RULE_DIFFICULT_PT", simple.GetIDFor(false))

	require.Equal(t, "Legibilidade: texto demasiado simples", simple.GetDescription())
	require.Equal(t, "Legibilidade: texto demasiado complexo", diff.GetDescription())

	require.Equal(t, 3, simple.Level)
	require.True(t, simple.IsDefaultOff())
	require.Equal(t, rules.ITSStyle, simple.GetLocQualityIssueType())
	require.NotNil(t, simple.GetCategory())
	require.Equal(t, "TEXT_ANALYSIS", simple.GetCategory().GetID().String())
	require.Equal(t, "Análise de Texto", simple.GetCategory().GetName())
	require.Equal(t, -1, simple.MinToCheckParagraph())

	require.Equal(t, "Nível de legibilidade 0 (muito difícil) a 6 (muito fácil):", simple.GetConfigureText())
	opts := simple.GetRuleOptions()
	require.Len(t, opts, 1)
	require.Equal(t, 3, opts[0].DefaultValue)
	require.Equal(t, 0, opts[0].MinConfigurableValue)
	require.Equal(t, 6, opts[0].MaxConfigurableValue)
	require.Equal(t, simple.GetConfigureText(), opts[0].ConfigureText)
}

func TestPortugueseReadabilityRule_CategoryAnáliseDeTexto(t *testing.T) {
	// Java always sets "Análise de Texto" after super (not message-bundle English "Text Analysis").
	r := NewPortugueseReadabilityRule(false, -1)
	require.Equal(t, "TEXT_ANALYSIS", r.GetCategory().GetID().String())
	require.Equal(t, "Análise de Texto", r.GetCategory().GetName())
	// English base would be "Text Analysis" — PT must not leave that.
	require.NotEqual(t, "Text Analysis", r.GetCategory().GetName())
}

func TestPortugueseReadabilityRule_PrintMessageLevel(t *testing.T) {
	r := NewPortugueseReadabilityRule(false, -1)
	require.Equal(t, " {Nível 0: Muito complexo}", r.PrintMessageLevel(0))
	require.Equal(t, " {Nível 1: Complexo}", r.PrintMessageLevel(1))
	require.Equal(t, " {Nível 2: Moderadamente complexo}", r.PrintMessageLevel(2))
	require.Equal(t, " {Nível 3: Meio-termo}", r.PrintMessageLevel(3))
	require.Equal(t, " {Nível 4: Moderadamente simples}", r.PrintMessageLevel(4))
	require.Equal(t, " {Nível 5: Simples}", r.PrintMessageLevel(5))
	require.Equal(t, " {Nível 6: Muito simples}", r.PrintMessageLevel(6))
	require.Equal(t, "", r.PrintMessageLevel(7))
	require.Equal(t, "", r.PrintMessageLevel(-1))
}

func TestPortugueseReadabilityRule_GetMessage(t *testing.T) {
	diff := NewPortugueseReadabilityRule(false, -1)
	msg := diff.GetMessage(1, 40, 20, 2)
	require.Equal(t,
		"Legibilidade {FRE: 40, ASL: 20, ASW: 2}: O texto deste parágrafo é difícil {Nível 1: Complexo}. Tem muitas palavras por frase e muitas sílabas por palavra.",
		msg)

	easy := NewPortugueseReadabilityRule(true, -1)
	msgE := easy.GetMessage(5, 85, 8, 1)
	require.Equal(t,
		"Legibilidade {FRE: 85, ASL: 8, ASW: 1}: O texto deste parágrafo é fácil {Nível 5: Simples}. Tem poucas palavras por frase e poucas sílabas por palavra.",
		msgE)
}

func TestPortugueseReadabilityRule_GetFleschReadingEase_Portuguese(t *testing.T) {
	r := NewPortugueseReadabilityRule(false, -1)
	// FRE = 206.84 - 1.02*asl - 60.0*asw
	require.InDelta(t, 206.84-(1.02*10)-(60.0*1.5), r.GetFleschReadingEase(10, 1.5), 1e-9)
	require.InDelta(t, 206.84, r.GetFleschReadingEase(0, 0), 1e-9)
	require.InDelta(t, 206.84-(1.02*20)-(60.0*2), r.GetFleschReadingEase(20, 2), 1e-9)
	// hook path used by base MatchList (same numbers) — not English FRE
	require.InDelta(t, 206.84-(1.02*10)-(60.0*1.5), r.ReadabilityRule.GetFleschReadingEase(10, 1.5), 1e-9)
	// must differ from English 206.835 - 1.015*asl - 84.6*asw
	eng := 206.835 - (1.015 * 10) - (84.6 * 1.5)
	require.NotEqual(t, eng, r.GetFleschReadingEase(10, 1.5))
}

func TestPortugueseReadabilityRule_IsVowel(t *testing.T) {
	// Accented vowels from Java isVowel
	for _, c := range []rune{'a', 'e', 'i', 'o', 'u', 'y', 'A', 'E', 'I', 'O', 'U', 'Y',
		'á', 'é', 'í', 'ó', 'ú', 'à', 'Á', 'É', 'Í', 'Ó', 'Ú', 'À',
		'â', 'ê', 'ô', 'Â', 'Ê', 'Ô', 'Ü'} {
		require.True(t, isVowelPT(c), "expected vowel: %c", c)
	}
	// Java does not list ü, ã, õ as vowels
	for _, c := range []rune{'ü', 'ã', 'õ', 'Ã', 'Õ', 'b', 'x'} {
		require.False(t, isVowelPT(c), "expected non-vowel: %c", c)
	}
}

func TestPortugueseReadabilityRule_SimpleSyllablesCount(t *testing.T) {
	r := NewPortugueseReadabilityRule(false, -1)
	// empty
	require.Equal(t, 0, r.SimpleSyllablesCount(""))
	// single vowel / consonant
	require.Equal(t, 1, r.SimpleSyllablesCount("a"))
	require.Equal(t, 1, r.SimpleSyllablesCount("x"))
	// accented first vowel
	require.Equal(t, 1, r.SimpleSyllablesCount("á"))
	require.Equal(t, 1, r.SimpleSyllablesCount("ô"))
	// "casa": c,a,s,a — a →1; a cl=s → n=2
	require.Equal(t, 2, r.SimpleSyllablesCount("casa"))
	// "pai": p,a,i — a →1; i cl=a → diphthong lastDouble; n=1
	require.Equal(t, 1, r.SimpleSyllablesCount("pai"))
	// "ideia": i,d,e,i,a
	// i →1; e cl=d → n=2; i cl=e → diphthong lastDouble; a lastDouble → n=3
	require.Equal(t, 3, r.SimpleSyllablesCount("ideia"))
	// "see" style: "lee" l,e,e — e→1; e cl=e lastDouble; n=1
	require.Equal(t, 1, r.SimpleSyllablesCount("lee"))
	// "ia": i →1; a cl=i → c=='a' && cl=='i' lastDouble; n=1
	require.Equal(t, 1, r.SimpleSyllablesCount("ia"))
	// "Aaa": A→1; a cl=A → not PT diphthong (c=='a' needs cl e/i/í/o/u/ú) → n=2; a cl=a → same → n=3
	require.Equal(t, 3, r.SimpleSyllablesCount("Aaa"))
	// after lastDouble next vowel counts: "iaa" i→1; a cl=i lastDouble; a lastDouble→n=2
	require.Equal(t, 2, r.SimpleSyllablesCount("iaa"))
	// multi-syllable Portuguese word
	require.GreaterOrEqual(t, r.SimpleSyllablesCount("internacionalização"), 4)
	// hook path
	require.Equal(t, r.SimpleSyllablesCount("casa"), r.ReadabilityRule.SimpleSyllablesCount("casa"))
	// ü is not a vowel in PT isVowel — "Über" U is vowel (Ü is), wait "Über": Ü→1; e cl=b → n=2
	require.Equal(t, 1, r.SimpleSyllablesCount("Ü"))
	// "über": ü not vowel; e cl=b → n=1 (e counts); n=1
	require.Equal(t, 1, r.SimpleSyllablesCount("über"))
}

func TestPortugueseReadabilityRule_Match_Difficult_SingleParagraph(t *testing.T) {
	lt := demoLTReadabilityPT()
	// Long multisyllabic Portuguese words → low PT FRE → rLevel < 3.
	// Final quirk: (!tooEasy && rLevel < level) returns matches without nParagraph>1.
	words := []string{
		"internacionalização", "constitucionalismo", "interdisciplinaridade",
		"caracterizações", "individualizações", "diferenciações",
		"responsabilidades", "superintelectualização", "contrarrevolucionário",
		"eletroencefalografia", "complexificações",
	}
	text := strings.Join(words, " ") + "."
	r := NewPortugueseReadabilityRule(false, -1)
	matches := r.MatchList(lt.Analyze(text))
	require.NotEmpty(t, matches, "expected difficult match for long multisyllabic Portuguese words")
	require.Contains(t, matches[0].Message, "difícil")
	require.Contains(t, matches[0].Message, "muitas palavras por frase")
	require.Contains(t, matches[0].Message, "{Nível ")
	require.Contains(t, matches[0].Message, "FRE:")
	// Span: tokens[1].start .. tokens[MARK_WORDS=3].end
	sents := lt.Analyze(text)
	require.NotEmpty(t, sents)
	toks := sents[0].GetTokensWithoutWhitespace()
	require.Greater(t, len(toks), 3)
	require.Equal(t, toks[1].GetStartPos(), matches[0].GetFromPos())
	require.Equal(t, toks[3].GetEndPos(), matches[0].GetToPos())
}

func TestPortugueseReadabilityRule_Match_TooEasy_NeedsMultipleParagraphs(t *testing.T) {
	lt := demoLTReadabilityPT()
	// Easy Portuguese: short words / short sentences → high FRE.
	// Single paragraph: Java final condition needs nParagraph > 1 for tooEasy → empty.
	easyPara := "O cão sentou no tapete. O gato correu no parque. Um homem viu um carro grande e vermelho na rua."
	r := NewPortugueseReadabilityRule(true, -1)
	require.Empty(t, r.MatchList(lt.Analyze(easyPara)), "single paragraph tooEasy must not return matches (Java quirk)")

	easyLots := strings.Repeat("O cão sentou no tapete. ", 8) + "\n\n" + strings.Repeat("O gato correu até mim. ", 8)
	matches := r.MatchList(lt.Analyze(easyLots))
	require.NotEmpty(t, matches, "very easy multi-paragraph Portuguese text should match tooEasy rule")
	require.Contains(t, matches[0].Message, "fácil")
	require.Contains(t, matches[0].Message, "poucas palavras por frase")
}

func TestPortugueseReadabilityRule_Match_BelowMinWords(t *testing.T) {
	lt := demoLTReadabilityPT()
	r := NewPortugueseReadabilityRule(false, -1)
	require.Empty(t, r.MatchList(lt.Analyze("Olá lá.")))
	rEasy := NewPortugueseReadabilityRule(true, -1)
	require.Empty(t, rEasy.MatchList(lt.Analyze("Olá lá amigo.")))
}

func TestPortugueseReadabilityRule_CtorLevel(t *testing.T) {
	r := NewPortugueseReadabilityRuleFull(nil, nil, false, 5, true)
	require.Equal(t, 5, r.Level)
	require.False(t, r.IsDefaultOff())
	require.Equal(t, "READABILITY_RULE_DIFFICULT_PT", r.GetID())
}

// Keep the original short ID test name path green.
func TestPortugueseReadabilityRule_IDs(t *testing.T) {
	easy := NewPortugueseReadabilityRule(true, 2)
	require.Equal(t, "READABILITY_RULE_SIMPLE_PT", easy.GetID())
	require.Contains(t, easy.GetDescription(), "simples")
	hard := NewPortugueseReadabilityRule(false, 4)
	require.Equal(t, "READABILITY_RULE_DIFFICULT_PT", hard.GetID())
	require.Equal(t, 2, easy.Level)
	require.Equal(t, 4, hard.Level)
}
