package br

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Java BretonTagger resource path: super("/br/breton.dict", new Locale("br")).
const BretonDictPath = "/br/breton.dict"

// BretonTaggerDictPath is the Java resource path constant (alias for checklist/tests).
const BretonTaggerDictPath = BretonDictPath

// patternSuffix ports Java Pattern.compile("(?iu)(..+)-(mañ|se|hont)$").
// Java (?iu) = CASE_INSENSITIVE | UNICODE_CASE; Go (?i) is Unicode-aware for letters.
var patternSuffix = regexp.MustCompile(`(?i)(..+)-(mañ|se|hont)$`)

// BretonTagger ports org.languagetool.tagging.br.BretonTagger.
// Not a pure thin BaseTagger: overrides Tag with demonstrative suffix strip retry.
type BretonTagger struct {
	*tagging.BaseTagger
}

// NewBretonTagger builds a BretonTagger over the given WordTagger.
// Java: super("/br/breton.dict", new Locale("br")) → tagLowercaseWithUppercase true.
func NewBretonTagger(wt tagging.WordTagger) *BretonTagger {
	return &BretonTagger{BaseTagger: tagging.NewBaseTagger(wt, BretonDictPath, "br", true)}
}

// Tag ports BretonTagger.tag:
//   - length>50 (Java String.length = UTF-16) → null POS
//   - exact + lower (NO isMixedCase skip — differs from BaseTagger)
//   - if both empty and lowercase: UppercaseFirstChar probe
//   - if still empty: strip -mañ/-se/-hont suffix and retry probeWord
//   - asAnalyzedTokenListForTaggedWords always uses original surface word
//   - pos += word.length() (UTF-16)
//
// conversionLocale is Java Locale.getDefault(); Go uses strings.ToLower (Unicode).
func (t *BretonTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0
	for _, word := range sentenceTokens {
		probeWord := word
		// Java: if (probeWord.length() > 50)
		if tagging.UTF16Len(probeWord) > 50 {
			l := []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, nil, nil)}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(l, pos))
			pos += tagging.UTF16Len(word)
			continue
		}

		// Retry loop when stripping -mañ/-se/-hont.
		for {
			var l []*languagetool.AnalyzedToken
			// Java: lowerWord = probeWord.toLowerCase(conversionLocale) // Locale.getDefault()
			lowerWord := strings.ToLower(probeWord)
			// asAnalyzedTokenListForTaggedWords(word, getWordTagger().tag(...)) — surface = original word
			taggerTokens := asAnalyzedTokens(word, t.TagWordExact(probeWord))
			lowerTaggerTokens := asAnalyzedTokens(word, t.TagWordExact(lowerWord))
			isLowercase := probeWord == lowerWord

			// Normal case.
			addTokens(taggerTokens, &l)

			if !isLowercase {
				// Lowercase — NO isMixedCase skip (differs from BaseTagger).
				addTokens(lowerTaggerTokens, &l)
			}

			// Uppercase / suffix / null.
			if len(lowerTaggerTokens) == 0 && len(taggerTokens) == 0 {
				if isLowercase {
					upper := tools.UppercaseFirstChar(probeWord)
					upperTaggerTokens := asAnalyzedTokens(word, t.TagWordExact(upper))
					if len(upperTaggerTokens) > 0 {
						addTokens(upperTaggerTokens, &l)
					}
				}
				if len(l) == 0 {
					if m := patternSuffix.FindStringSubmatch(probeWord); m != nil {
						// Remove the suffix and probe dictionary again.
						// So given "xxx-mañ", probe again with "xxx".
						probeWord = m[1]
						continue
					}
					l = append(l, languagetool.NewAnalyzedToken(word, nil, nil))
				}
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(l, pos))
			pos += tagging.UTF16Len(word)
			break
		}
	}
	return out
}

// asAnalyzedTokens ports BaseTagger.asAnalyzedTokenListForTaggedWords.
func asAnalyzedTokens(word string, tagged []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	if len(tagged) == 0 {
		return nil
	}
	out := make([]*languagetool.AnalyzedToken, 0, len(tagged))
	for _, tw := range tagged {
		out = append(out, taggedToToken(word, tw))
	}
	return out
}

func taggedToToken(surface string, tw tagging.TaggedWord) *languagetool.AnalyzedToken {
	var pos, lemma *string
	if tw.PosTag != "" {
		p := tw.PosTag
		pos = &p
	}
	if tw.Lemma != "" {
		l := tw.Lemma
		lemma = &l
	}
	return languagetool.NewAnalyzedToken(surface, pos, lemma)
}

// addTokens ports BretonTagger.addTokens.
func addTokens(taggedTokens []*languagetool.AnalyzedToken, l *[]*languagetool.AnalyzedToken) {
	if taggedTokens == nil {
		return
	}
	*l = append(*l, taggedTokens...)
}
