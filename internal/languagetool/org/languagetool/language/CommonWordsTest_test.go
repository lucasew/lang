package language

// Twin of CommonWordsTest — loads inspiration common_words.txt via detector.
import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language/identifier/detector"
	"github.com/stretchr/testify/require"
)

func resourcePath(t *testing.T, lang, file string) string {
	t.Helper()
	_, this, _, ok := runtime.Caller(0)
	require.True(t, ok)
	dir := filepath.Dir(this)
	for i := 0; i < 12; i++ {
		cand := filepath.Join(dir, "inspiration/languagetool/languagetool-language-modules", lang,
			"src/main/resources/org/languagetool/resource", lang, file)
		if _, err := os.Stat(cand); err == nil {
			return cand
		}
		dir = filepath.Dir(dir)
	}
	t.Fatalf("resource not found: %s/%s", lang, file)
	return ""
}

func loadCommonWords(t *testing.T) *detector.CommonWordsDetector {
	t.Helper()
	cw := detector.NewCommonWordsDetector()
	for _, lang := range []string{"de", "en", "es", "pt", "ca"} {
		f, err := os.Open(resourcePath(t, lang, "common_words.txt"))
		require.NoError(t, err)
		require.NoError(t, cw.LoadWords(lang, f))
		_ = f.Close()
	}
	return cw
}

func TestCommonWords_Test(t *testing.T) {
	cw := loadCommonWords(t)

	res1 := cw.GetKnownWordsPerLanguage("Das ist bequem")
	_, hasEN := res1["en"]
	require.False(t, hasEN)
	require.Equal(t, 2, res1["de"])

	res2 := cw.GetKnownWordsPerLanguage("Das ist bequem ")
	_, hasEN = res2["en"]
	require.False(t, hasEN)
	require.Equal(t, 3, res2["de"])

	res3 := cw.GetKnownWordsPerLanguage("bequem")
	_, hasEN = res3["en"]
	require.False(t, hasEN)
	require.Equal(t, 1, res3["de"])

	res4 := cw.GetKnownWordsPerLanguage("this is a test")
	require.Equal(t, 3, res4["en"])

	res5 := cw.GetKnownWordsPerLanguage("Ideábamos una declaracion con el.")
	require.Equal(t, 5, res5["es"])

	res6 := cw.GetKnownWordsPerLanguage("Ideábamos una declaracion con el; desassigna mencions.")
	require.Equal(t, 3, res6["es"])

	res7 := cw.GetKnownWordsPerLanguage("Estagiário de desenvolvedor 'web' ou relacionados a programador.")
	require.Equal(t, 3, res7["es"])
	require.Equal(t, 4, res7["pt"])
	require.Equal(t, 2, res7["ca"])

	res8 := cw.GetKnownWordsPerLanguage("Autohaus-Wirklichkeit")
	_, hasEN = res8["en"]
	require.False(t, hasEN)
	require.Equal(t, 1, res8["de"])

	res9 := cw.GetKnownWordsPerLanguage("Costum de certes cultures que imposa a un pare l’adopció d’un comportament idèntic al de la mare en el període anterior o posterior al part")
	require.Equal(t, 20, res9["ca"])
	require.Equal(t, 10, res9["es"])
}
