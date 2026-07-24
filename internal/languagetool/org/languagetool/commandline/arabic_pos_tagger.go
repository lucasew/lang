package commandline

import (
	"os"
	"path/filepath"

	atticmorfo "github.com/lucasew/lang/internal/attic/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ar"
)

// RegisterArabicPOSTagger installs TagWord using Java ArabicTagger semantics:
// CombiningTagger(morfologik, manual added.txt) + tashkeel strip + affix stemming.
func RegisterArabicPOSTagger(lt *languagetool.JLanguageTool, dictPath string) bool {
	if lt == nil || dictPath == "" {
		return false
	}
	d, err := atticmorfo.OpenDictionary(dictPath)
	if err != nil || d == nil {
		return false
	}
	var wordTagger tagging.WordTagger = morfologikWordTagger{d: d}
	// Java BaseTagger: CombiningTagger(morfologik, manual, removal, overwrite=false).
	// Manual additions (resource/ar/added.txt) hold demonstratives (ذلك/هذان/…) as DMS/DMD.
	if manual := loadArabicManualTagger(dictPath); manual != nil {
		wordTagger = tagging.NewCombiningTagger(wordTagger, manual, false)
	}
	tagger := ar.NewArabicTagger(wordTagger)
	lt.TagWord = func(token string) []languagetool.TokenTag {
		if token == "" {
			return nil
		}
		if len([]rune(token)) == 1 {
			r := []rune(token)[0]
			if r == ' ' || r == '\t' || r == '\n' {
				return nil
			}
		}
		tws := tagger.TagTokens(token)
		if len(tws) == 0 {
			return nil
		}
		out := make([]languagetool.TokenTag, 0, len(tws))
		for _, tw := range tws {
			out = append(out, languagetool.TokenTag{POS: tw.PosTag, Lemma: tw.Lemma})
		}
		return out
	}
	return true
}

// loadArabicManualTagger loads resource/ar/added.txt (+ added_custom.txt) beside the dict
// or under the inspiration module tree (Java BaseTagger getManualAdditionsFileNames).
func loadArabicManualTagger(dictPath string) tagging.WordTagger {
	dir := filepath.Dir(dictPath)
	var manuals []tagging.WordTagger
	for _, name := range []string{"added.txt", "added_custom.txt"} {
		p := filepath.Join(dir, name)
		f, err := os.Open(p)
		if err != nil {
			if alt := walkUpArabicResource(name); alt != "" {
				f, err = os.Open(alt)
			}
		}
		if err != nil || f == nil {
			continue
		}
		mt, err := tagging.NewManualTagger(f)
		_ = f.Close()
		if err != nil || mt == nil {
			continue
		}
		manuals = append(manuals, mt)
	}
	if len(manuals) == 0 {
		return nil
	}
	// Fold multiple manuals: later files take precedence via CombiningTagger(tagger1, tagger2)
	// where tagger2 results come first (Java CombiningTagger).
	out := manuals[0]
	for i := 1; i < len(manuals); i++ {
		out = tagging.NewCombiningTagger(out, manuals[i], false)
	}
	return out
}

func walkUpArabicResource(name string) string {
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "ar",
		"src", "main", "resources", "org", "languagetool", "resource", "ar", name)
	wd, _ := os.Getwd()
	dir := wd
	for i := 0; i < 12; i++ {
		p := filepath.Join(dir, rel)
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// morfologikWordTagger adapts attic Morfologik dict to tagging.WordTagger.
type morfologikWordTagger struct {
	d *atticmorfo.Dictionary
}

func (w morfologikWordTagger) Tag(word string) []tagging.TaggedWord {
	if w.d == nil || word == "" {
		return nil
	}
	forms, err := w.d.Lookup(word)
	if err != nil || len(forms) == 0 {
		return nil
	}
	out := make([]tagging.TaggedWord, 0, len(forms))
	for _, f := range forms {
		out = append(out, tagging.NewTaggedWord(f.Stem, f.Tag))
	}
	return out
}
