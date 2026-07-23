package pl

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const PolishDictPath = "/pl/polish.dict"

// PolishTagger ports org.languagetool.tagging.pl.PolishTagger.
type PolishTagger struct {
	*tagging.BaseTagger
}

// NewPolishTagger builds a PolishTagger over the given WordTagger
// (Java: super("/pl/polish.dict", new Locale("pl"))).
func NewPolishTagger(wt tagging.WordTagger) *PolishTagger {
	// Java BaseTagger ctor sets tagLowercaseWithUppercase true by default, but
	// PolishTagger overrides tag() entirely and does not call getAnalyzedTokens.
	return &PolishTagger{BaseTagger: tagging.NewBaseTagger(wt, PolishDictPath, "pl", false)}
}

// Tag ports PolishTagger.tag — exact WordTagger lookups with Polish-specific
// case retries and POS tags split on '+'.
func (t *PolishTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	tokenReadings := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0

	for _, word := range sentenceTokens {
		var l []*languagetool.AnalyzedToken
		// Java: word.toLowerCase(locale) with Locale("pl") — Polish maps like Unicode.
		lowerWord := strings.ToLower(word)
		taggerTokens := asAnalyzedTokenListForTaggedWords(word, t.TagWordExact(word))
		lowerTaggerTokens := asAnalyzedTokenListForTaggedWords(word, t.TagWordExact(lowerWord))
		isLowercase := word == lowerWord

		// normal case
		l = addTokens(taggerTokens, l)

		if !isLowercase {
			// lowercase (Java: always when surface ≠ lower, including mixed case)
			l = addTokens(lowerTaggerTokens, l)
		}

		// uppercase
		if len(lowerTaggerTokens) == 0 && len(taggerTokens) == 0 {
			if isLowercase {
				upperTaggerTokens := asAnalyzedTokenListForTaggedWords(word,
					t.TagWordExact(tools.UppercaseFirstChar(word)))
				if len(upperTaggerTokens) > 0 {
					l = addTokens(upperTaggerTokens, l)
				} else {
					l = append(l, languagetool.NewAnalyzedToken(word, nil, nil))
				}
			} else {
				l = append(l, languagetool.NewAnalyzedToken(word, nil, nil))
			}
		}
		tokenReadings = append(tokenReadings, languagetool.NewAnalyzedTokenReadingsList(l, pos))
		// Java: pos += word.length() (UTF-16 code units)
		pos += tagging.UTF16Len(word)
	}

	return tokenReadings
}

// asAnalyzedTokenListForTaggedWords ports BaseTagger.asAnalyzedTokenListForTaggedWords
// used by PolishTagger.tag (surface form is the original word).
func asAnalyzedTokenListForTaggedWords(word string, taggedWords []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	if len(taggedWords) == 0 {
		return nil
	}
	out := make([]*languagetool.AnalyzedToken, 0, len(taggedWords))
	for _, tw := range taggedWords {
		var pos, lemma *string
		if tw.PosTag != "" {
			p := tw.PosTag
			pos = &p
		}
		if tw.Lemma != "" {
			l := tw.Lemma
			lemma = &l
		}
		out = append(out, languagetool.NewAnalyzedToken(word, pos, lemma))
	}
	return out
}

// addTokens ports PolishTagger.addTokens: each POS tag is split on '+' into
// separate AnalyzedToken readings sharing token and lemma.
func addTokens(taggedTokens, l []*languagetool.AnalyzedToken) []*languagetool.AnalyzedToken {
	if taggedTokens == nil {
		return l
	}
	for _, at := range taggedTokens {
		if at == nil {
			continue
		}
		// Java: StringTools.asString(at.getPOSTag()).split("\\+")
		// asString(null) → null → NPE; dictionary readings always carry a tag.
		posStr := ""
		if p := at.GetPOSTag(); p != nil {
			posStr = *p
		}
		tagsArr := strings.Split(posStr, "+")
		lemma := at.GetLemma()
		for _, currTag := range tagsArr {
			tag := currTag
			l = append(l, languagetool.NewAnalyzedToken(at.GetToken(), &tag, lemma))
		}
	}
	return l
}
