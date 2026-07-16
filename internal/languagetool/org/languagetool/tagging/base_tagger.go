package tagging

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// BaseTagger ports the non-Morfologik surface of org.languagetool.tagging.BaseTagger:
// WordTagger-backed lookup with optional lowercase→uppercase retry.
// (Full AnalyzedTokenReadings assembly lives outside this package to avoid cycles.)
type BaseTagger struct {
	WordTagger                WordTagger
	DictionaryPath            string
	LocaleLanguage            string // e.g. "en"
	TagLowercaseWithUppercase bool
	OverwriteWithManual       bool
}

func NewBaseTagger(wordTagger WordTagger, dictionaryPath, localeLanguage string, tagLowercaseWithUppercase bool) *BaseTagger {
	return &BaseTagger{
		WordTagger:                wordTagger,
		DictionaryPath:            dictionaryPath,
		LocaleLanguage:            localeLanguage,
		TagLowercaseWithUppercase: tagLowercaseWithUppercase,
	}
}

func (t *BaseTagger) GetDictionaryPath() string { return t.DictionaryPath }

func (t *BaseTagger) GetManualAdditionsFileNames() []string {
	lang := t.LocaleLanguage
	return []string{lang + "/added.txt", lang + "/added_custom.txt"}
}

func (t *BaseTagger) GetManualRemovalsFileNames() []string {
	lang := t.LocaleLanguage
	return []string{lang + "/removed.txt", lang + "/removed_custom.txt"}
}

func (t *BaseTagger) OverwriteWithManualTagger() bool { return t.OverwriteWithManual }

// GetWordTagger returns the underlying WordTagger.
func (t *BaseTagger) GetWordTagger() WordTagger { return t.WordTagger }

// TagWord looks up a single surface form (with optional case retry).
func (t *BaseTagger) TagWord(word string) []TaggedWord {
	if t == nil || t.WordTagger == nil {
		return nil
	}
	res := t.WordTagger.Tag(word)
	if len(res) == 0 && t.TagLowercaseWithUppercase && word != "" && !tools.StartsWithUppercase(word) {
		up := tools.UppercaseFirstChar(word)
		if up != word {
			res = t.WordTagger.Tag(up)
		}
		if len(res) == 0 {
			res = t.WordTagger.Tag(strings.ToUpper(word))
		}
	}
	return res
}

// TagWords tags each token independently.
func (t *BaseTagger) TagWords(sentenceTokens []string) [][]TaggedWord {
	out := make([][]TaggedWord, len(sentenceTokens))
	for i, w := range sentenceTokens {
		out[i] = t.TagWord(w)
	}
	return out
}
