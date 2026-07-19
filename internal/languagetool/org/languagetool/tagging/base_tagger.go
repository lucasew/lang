package tagging

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// BaseTagger ports the non-Morfologik surface of org.languagetool.tagging.BaseTagger:
// WordTagger-backed lookup with BaseTagger.getAnalyzedTokens case-merge rules.
// (Full AnalyzedTokenReadings assembly lives outside this package to avoid cycles.)
type BaseTagger struct {
	WordTagger                WordTagger
	DictionaryPath            string
	LocaleLanguage            string // e.g. "en"
	TagLowercaseWithUppercase bool
	OverwriteWithManual       bool
	// AdditionalTags optional language hook (Java additionalTags); nil = none.
	AdditionalTags func(word string, wt WordTagger) []TaggedWord
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

// TagWordExact is a raw dictionary lookup (Java getWordTagger().tag(word)).
// GermanTagger and similar overrides that reimplement case logic must use this.
func (t *BaseTagger) TagWordExact(word string) []TaggedWord {
	if t == nil || t.WordTagger == nil {
		return nil
	}
	return t.WordTagger.Tag(word)
}

// TagWord ports BaseTagger.getAnalyzedTokens case-merge at the TaggedWord level:
//  1. tag surface form
//  2. if not lowercase and not mixed case, also add lowercase tags
//  3. if tagLowercaseWithUppercase and still empty and word is lowercase, try UppercaseFirstChar
//  4. if still empty, AdditionalTags hook
// Does not invent a null-POS token (callers that need AnalyzedToken do that).
func (t *BaseTagger) TagWord(word string) []TaggedWord {
	if t == nil || t.WordTagger == nil {
		return nil
	}
	lowerWord := strings.ToLower(word)
	isLowercase := word == lowerWord
	isMixedCase := tools.IsMixedCase(word)

	taggerTokens := t.WordTagger.Tag(word)
	var lowerTaggerTokens []TaggedWord
	if !isLowercase {
		lowerTaggerTokens = t.WordTagger.Tag(lowerWord)
	} else {
		lowerTaggerTokens = taggerTokens
	}

	// normal case (Java addTokens)
	result := append([]TaggedWord(nil), taggerTokens...)
	// non-lowercase (Title or ALLCAPS), not mixed: also lower tags
	if !isLowercase && !isMixedCase {
		result = append(result, lowerTaggerTokens...)
	}
	// lowercase word with start-uppercase tags when both empty
	// Java: only UppercaseFirstChar — not full ToUpper
	if t.TagLowercaseWithUppercase && len(lowerTaggerTokens) == 0 && len(taggerTokens) == 0 && isLowercase {
		up := tools.UppercaseFirstChar(word)
		if up != word {
			result = append(result, t.WordTagger.Tag(up)...)
		}
	}
	// language-dependent additionalTags
	if len(result) == 0 && t.AdditionalTags != nil {
		result = append(result, t.AdditionalTags(word, t.WordTagger)...)
	}
	return result
}

// TagWords tags each token with TagWord (BaseTagger.getAnalyzedTokens per token).
func (t *BaseTagger) TagWords(sentenceTokens []string) [][]TaggedWord {
	out := make([][]TaggedWord, len(sentenceTokens))
	for i, w := range sentenceTokens {
		out[i] = t.TagWord(w)
	}
	return out
}
