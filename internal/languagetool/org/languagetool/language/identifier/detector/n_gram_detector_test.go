package detector

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestNGramDetectorScripts(t *testing.T) {
	d := NewNGramDetector(500)
	d.TrainFromText("en", "the cat sat on the mat hello world")
	d.TrainFromText("de", "der hund liegt auf der matte hallo welt")
	require.Equal(t, "en", d.TopLanguage("hello the world the cat"))
	// Chinese characters boost zh
	scores := d.DetectLanguages("你好世界")
	require.Contains(t, scores, "zh")
	require.Greater(t, scores["zh"], 0.0)
}

// Minimal ZIP layout matching Java NGramDetector constructor file names.
func TestNGramDetectorFromZip_LoadAndScore(t *testing.T) {
	// canLanguageBeDetected needs registry (Java Languages always loaded)
	for _, m := range []struct{ n, c string }{{"English", "en"}, {"German", "de"}} {
		meta := languagetool.LanguageMeta{Name: m.n, Code: m.c}
		if !languagetool.GlobalLanguages.IsLanguageSupported(m.c) {
			languagetool.GlobalLanguages.Register(meta)
		}
	}
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "model.zip")
	writeMiniNGramZip(t, zipPath)

	d, err := NewNGramDetectorFromZip(zipPath, 50)
	require.NoError(t, err)
	require.True(t, d.zipLoaded)
	require.NotEmpty(t, d.vocab)
	require.Len(t, d.codes, 2)
	require.Len(t, d.knpBigramProbs, 2)

	// encode + detect should return language scores (not empty)
	scores := d.DetectLanguagesAdditional("hello", nil)
	require.NotEmpty(t, scores)
	// either en/de or zz (threshold path)
	_, hasEN := scores["en"]
	_, hasDE := scores["de"]
	_, hasZZ := scores["zz"]
	require.True(t, hasEN || hasDE || hasZZ)
}

func writeMiniNGramZip(t *testing.T, zipPath string) {
	t.Helper()
	f, err := os.Create(zipPath)
	require.NoError(t, err)
	defer f.Close()
	zw := zip.NewWriter(f)
	write := func(name, body string) {
		w, err := zw.Create(name)
		require.NoError(t, err)
		_, err = w.Write([]byte(body))
		require.NoError(t, err)
	}
	// flag column "1" includes language
	write("iso_codes.tsv", "English\ten\teng\t1\nGerman\tde\tdeu\t1\n")
	// vocab: SOS-like tokens used by encode (▁ + letters)
	// indices 0,1,2,... match Java order
	write("vocab.txt", "<unk>\n<s>\n"+
		"▁\n"+
		"▁h\n"+
		"h\n"+
		"he\n"+
		"hel\n"+
		"ell\n"+
		"llo\n"+
		"lo\n"+
		"o\n")
	// thresholds: first line start index; rows of per-lang thresholds (low so we get real scores)
	write("thresholds.txt", "100\n-999 -999\n")
	// bigram probs for lang 0 and 1 — key "i_j" from join of indices
	// high self-transition from 1 (SOS) to various
	write("00.txt", "1 2 0.5\n2 3 0.5\n3 4 0.5\n4 5 0.5\n5 6 0.5\n6 7 0.5\n7 8 0.5\n8 9 0.5\n")
	write("01.txt", "1 2 0.1\n2 3 0.1\n")
	require.NoError(t, zw.Close())
}
