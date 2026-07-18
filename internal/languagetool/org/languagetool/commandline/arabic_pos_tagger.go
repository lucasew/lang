package commandline

import (
	atticmorfo "github.com/lucasew/lang/internal/attic/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ar"
)

// RegisterArabicPOSTagger installs TagWord using Java ArabicTagger semantics
// (tashkeel strip + prefix/suffix stemming against arabic.dict).
func RegisterArabicPOSTagger(lt *languagetool.JLanguageTool, dictPath string) bool {
	if lt == nil || dictPath == "" {
		return false
	}
	d, err := atticmorfo.OpenDictionary(dictPath)
	if err != nil || d == nil {
		return false
	}
	wt := morfologikWordTagger{d: d}
	tagger := ar.NewArabicTagger(wt)
	lt.TagWord = func(token string) []languagetool.TokenTag {
		if token == "" {
			return nil
		}
		// Skip pure whitespace tokens
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
