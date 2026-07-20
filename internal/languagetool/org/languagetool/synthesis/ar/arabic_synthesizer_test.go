package ar

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

func TestInflectMafoulAndTanwin(t *testing.T) {
	// Java ArabicSynthesizer.inflectMafoulMutlq / inflectAdjectiveTanwinNasb
	m := InflectMafoulMutlq("عمل")
	require.True(t, strings.HasPrefix(m, "عمل"))
	require.Contains(t, m, string(tools.ArabicFathatan))
	require.True(t, strings.HasSuffix(m, string(tools.ArabicAlef)))

	masc := InflectAdjectiveTanwinNasb("قوي", false)
	require.Contains(t, masc, string(tools.ArabicFathatan))
	fem := InflectAdjectiveTanwinNasb("قوي", true)
	require.Contains(t, fem, string(tools.ArabicTehMarbuta))
}

func TestArabicSynthesizer(t *testing.T) {
	man, err := synthesis.NewManualSynthesizer(strings.NewReader("كتب\tكتب\tNxx\n"))
	require.NoError(t, err)
	s := NewArabicSynthesizer(man)
	lemma := "كتب"
	tok := languagetool.NewAnalyzedToken("كتب", nil, &lemma)
	forms, err := s.Synthesize(tok, "Nxx")
	require.NoError(t, err)
	require.Equal(t, []string{"كتب"}, forms)
	require.Equal(t, ArabicSynthDict, s.ResourceFileName)
}

// Twin of ArabicSynthesizer.correctTag / getPosTagCorrection.
func TestArabicSynthesizer_CorrectTagAndPosTagCorrection(t *testing.T) {
	s := NewArabicSynthesizer(nil)
	require.Equal(t, "", s.CorrectTag(""))
	require.Equal(t, "", s.GetPosTagCorrection(""))

	// Noun tag length 12: conj@9 jar@10 pronoun@11 (Go flat layout).
	// Start with conj W, definite L, no jar.
	tag := []rune("Nxx-M1I-xW-L")
	require.Equal(t, 12, len(tag))
	src := string(tag)
	got := s.CorrectTag(src)
	// setConjunction("-") clears CONJ; setDefinite("-") clears definite on unattached;
	// UnifyPronounTag only if attached (H).
	require.Equal(t, '-', s.tm().GetFlag(got, "CONJ"))
	// After setDefinite("-") on unattached definite noun, PRONOUN should be '-'
	require.Equal(t, '-', s.tm().GetFlag(got, "PRONOUN"))
	require.Equal(t, got, s.GetPosTagCorrection(src))
}

// Twin of ArabicSynthesizer.correctStem prefix/suffix adjust.
func TestArabicSynthesizer_CorrectStem(t *testing.T) {
	s := NewArabicSynthesizer(nil)
	// attached pronoun: strip trailing ه
	attached := []rune("Nxx-M1I-x--H")
	require.Equal(t, 12, len(attached))
	require.Equal(t, "كتاب", s.CorrectStem("كتابه", string(attached)))

	// definite: prefix ال
	def := []rune("Nxx-M1I-x--L")
	require.Equal(t, 12, len(def))
	require.Equal(t, "الكتاب", s.CorrectStem("كتاب", string(def)))

	// jar ب
	jar := []rune("Nxx-M1I-x-B-")
	require.Equal(t, 12, len(jar))
	require.Equal(t, "بكتاب", s.CorrectStem("كتاب", string(jar)))

	// conjunction و
	conj := []rune("Nxx-M1I-xW--")
	require.Equal(t, 12, len(conj))
	require.Equal(t, "وكتاب", s.CorrectStem("كتاب", string(conj)))

	// nil/empty postag → unchanged
	require.Equal(t, "كتاب", s.CorrectStem("كتاب", ""))
}

func TestArabicTagManager_DefiniteAndJarPrefixes(t *testing.T) {
	// re-export path: synthesizer uses tagging/ar prefixes
	s := NewArabicSynthesizer(nil)
	tm := s.tm()
	def := string([]rune("Nxx-M1I-x--L"))
	require.Equal(t, "ال", tm.GetDefinitePrefix(def))
	// jar ل + definite L → definite prefix "ل" (Java getDefinitePrefix)
	jarL := string([]rune("Nxx-M1I-x-LL"))
	require.Equal(t, 12, len([]rune(jarL)))
	require.Equal(t, "ل", tm.GetJarPrefix(jarL))
	require.Equal(t, "ل", tm.GetDefinitePrefix(jarL))
	require.Equal(t, "ب", tm.GetJarPrefix(string([]rune("Nxx-M1I-x-B-"))))
	require.Equal(t, "", tm.GetJarPrefix(""))
}
