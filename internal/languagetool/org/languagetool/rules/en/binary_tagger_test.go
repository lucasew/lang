package en

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	entok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/en"
	"github.com/stretchr/testify/require"
)

func findEnglishPOSDict(t *testing.T) string {
	t.Helper()
	wd, _ := os.Getwd()
	dir := wd
	for i := 0; i < 12; i++ {
		cand := filepath.Join(dir, "third_party", "english-pos-dict", "org", "languagetool", "resource", "en", "english.dict")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Skip("english.dict not found")
	return ""
}

func TestRegisterBinaryEnglishTagger(t *testing.T) {
	p := findEnglishPOSDict(t)
	lt := languagetool.NewJLanguageTool("en")
	require.True(t, RegisterBinaryEnglishTagger(lt, p))
	require.NotNil(t, lt.TagWord)
	tags := lt.TagWord("houses")
	require.NotEmpty(t, tags)
	var hasNNS bool
	for _, tg := range tags {
		if tg.POS == "NNS" && tg.Lemma == "house" {
			hasNNS = true
		}
	}
	require.True(t, hasNNS, "%+v", tags)
	// case fold
	tags = lt.TagWord("This")
	require.NotEmpty(t, tags)
	// Java EnglishTagger: sentence-start "How" keeps lowercase WRB readings
	tags = lt.TagWord("How")
	var hasWRB, hasNNP bool
	for _, tg := range tags {
		if tg.POS == "WRB" {
			hasWRB = true
		}
		if tg.POS == "NNP" {
			hasNNP = true
		}
	}
	require.True(t, hasWRB, "How should include WRB from lowercase how: %+v", tags)
	require.True(t, hasNNP, "How may still include NNP proper-name: %+v", tags)
}

func TestEnglishIsMixedCase(t *testing.T) {
	require.False(t, englishIsMixedCase("How"))
	require.False(t, englishIsMixedCase("HOW"))
	require.False(t, englishIsMixedCase("how"))
	require.True(t, englishIsMixedCase("iPhone"))
	require.True(t, englishIsMixedCase("McDonald"))
}

func TestBinaryEnglishTagWord_PCT(t *testing.T) {
	// Java UNKNOWN_PCT disambiguation: comma gets PCT so ALL_OF_SUDDEN matches.
	p := findEnglishPOSDict(t)
	lt := languagetool.NewJLanguageTool("en")
	require.True(t, RegisterBinaryEnglishTagger(lt, p))
	tags := lt.TagWord(",")
	var hasPCT bool
	for _, tg := range tags {
		if tg.POS == "PCT" {
			hasPCT = true
		}
	}
	require.True(t, hasPCT, "comma should have PCT: %+v", tags)
}

func TestRegisterBinaryEnglishTagger_WiresTokenizerIsTaggedEN(t *testing.T) {
	p := findEnglishPOSDict(t)
	prev := entok.IsTaggedEN
	t.Cleanup(func() { entok.IsTaggedEN = prev })

	lt := languagetool.NewJLanguageTool("en")
	require.True(t, RegisterBinaryEnglishTagger(lt, p))
	require.NotNil(t, entok.IsTaggedEN)

	// Known dictionary surfaces used by EnglishWordTokenizerTest
	require.True(t, entok.IsTaggedEN("doin'"), "doin' should be tagged in english.dict")
	require.True(t, entok.IsTaggedEN("'m") || entok.IsTaggedEN("I'm"), "clitic or full form")
	// Nonsense should not invent
	require.False(t, entok.IsTaggedEN("xyzzy-not-a-word-qq"))

	// Tokenizer should keep doin' whole when IsTaggedEN is wired
	toks := entok.NewEnglishWordTokenizer().Tokenize("doin' that")
	require.Equal(t, []string{"doin'", " ", "that"}, toks)
}
