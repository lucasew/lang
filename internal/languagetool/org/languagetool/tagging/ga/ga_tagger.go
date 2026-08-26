package ga

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// IrishDictPath is the Java BaseTagger resource path (super("/ga/irish.dict", Locale("ga"))).
const IrishDictPath = "/ga/irish.dict"

// IrishTagger ports org.languagetool.tagging.ga.IrishTagger.
type IrishTagger struct {
	*tagging.BaseTagger
}

// NewIrishTagger builds an IrishTagger over the given WordTagger
// (Java: super("/ga/irish.dict", new Locale("ga"))).
func NewIrishTagger(wt tagging.WordTagger) *IrishTagger {
	// Java BaseTagger default tagLowercaseWithUppercase is true, but IrishTagger
	// overrides tag() entirely and never calls getAnalyzedTokens.
	return &IrishTagger{BaseTagger: tagging.NewBaseTagger(wt, IrishDictPath, "ga", false)}
}

// Tag ports IrishTagger.tag — exact WordTagger lookups with Irish case cascade
// (Utils.toLowerCaseIrish) and filterMorph demutation guesses.
func (t *IrishTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	tokenReadings := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0

	for _, word := range sentenceTokens {
		var l []*languagetool.AnalyzedToken
		// Java: Utils.toLowerCaseIrish(word) — not Locale/string lower only.
		lowerWord := ToLowerCaseIrish(word)

		taggerTokens := asAnalyzedTokenListForTaggedWords(word, t.TagWordExact(word))
		lowerTaggerTokens := asAnalyzedTokenListForTaggedWords(word, t.TagWordExact(lowerWord))
		isLowercase := word == lowerWord

		// normal case
		l = addTokens(taggerTokens, l)

		if !isLowercase {
			// lowercase
			l = addTokens(lowerTaggerTokens, l)
		}

		// uppercase / morph guess branch when both exact lookups empty
		if len(lowerTaggerTokens) == 0 && len(taggerTokens) == 0 {
			guessedTokens := asAnalyzedTokenListForTaggedWords(word, t.filterMorph(word))
			lowerGuessedTokens := asAnalyzedTokenListForTaggedWords(word, t.filterMorph(ToLowerCaseIrish(word)))
			if len(guessedTokens) > 0 {
				l = addTokens(guessedTokens, l)
			}
			if len(guessedTokens) == 0 && len(lowerGuessedTokens) > 0 {
				l = addTokens(lowerGuessedTokens, l)
			}
			if isLowercase {
				upperTaggerTokens := asAnalyzedTokenListForTaggedWords(word,
					t.TagWordExact(tools.UppercaseFirstChar(word)))
				if len(upperTaggerTokens) > 0 {
					l = addTokens(upperTaggerTokens, l)
				} else {
					upperGuessedTokens := asAnalyzedTokenListForTaggedWords(word,
						t.filterMorph(tools.UppercaseFirstChar(word)))
					if len(upperGuessedTokens) > 0 {
						l = addTokens(upperGuessedTokens, l)
					} else {
						l = append(l, languagetool.NewAnalyzedToken(word, nil, nil))
					}
				}
			} else {
				// Java: always null token for non-lowercase when dict empty,
				// even if filterMorph already added readings.
				l = append(l, languagetool.NewAnalyzedToken(word, nil, nil))
			}
		}
		tokenReadings = append(tokenReadings, languagetool.NewAnalyzedTokenReadingsList(l, pos))
		// Java: pos += word.length() (UTF-16 code units)
		pos += tagging.UTF16Len(word)
	}

	return tokenReadings
}

// filterMorph ports IrishTagger.filterMorph.
func (t *IrishTagger) filterMorph(in string) []tagging.TaggedWord {
	var tagged []tagging.TaggedWord
	tocheck := MorphWord(in)
	if len(tocheck) == 0 {
		return tagged
	}
	for _, rt := range tocheck {
		if rt == nil {
			continue
		}
		pfx := false
		cur := t.TagWordExact(rt.GetWord())
		if rt.GetPrefix() != "" {
			pfx = true
			tryword := rt.GetPrefix() + Lenite(rt.GetWord())
			joined := t.TagWordExact(tryword)
			hyphword := rt.GetPrefix() + "-" + Lenite(rt.GetWord())
			hyphen := t.TagWordExact(hyphword)

			if len(joined) > 0 {
				cur = joined
				pfx = false
			} else if len(hyphen) > 0 {
				pfx = false
				cur = hyphen
			} else {
				pfx = true
			}
		}

		if len(cur) == 0 {
			continue
		}
		for _, tw := range cur {
			appendTag := rt.GetAppendTag()
			if pfx {
				appendTag = rt.GetAppendTag() + ":NonStdCmpd"
			}
			// Java: tw.getPosTag().matches(rt.getRestrictToPos()) — full-string match.
			if javaStringMatches(tw.PosTag, rt.GetRestrictToPos()) {
				lemma := tw.Lemma
				if pfx {
					lemma = rt.GetPrefix() + Lenite(tw.Lemma)
				}
				tagged = append(tagged, tagging.NewTaggedWord(lemma, tw.PosTag+appendTag))
			}
		}
	}
	return tagged
}

// javaStringMatches ports Java String.matches(regex): whole-string match.
// Empty regex only matches the empty string (same as Java Pattern).
func javaStringMatches(s, regex string) bool {
	// Compile as full match: ^(regex)$ — group so | alternation stays correct.
	re, err := regexp.Compile("^(?:" + regex + ")$")
	if err != nil {
		return false
	}
	return re.MatchString(s)
}

// asAnalyzedTokenListForTaggedWords ports BaseTagger.asAnalyzedTokenListForTaggedWords
// (surface form is the original word).
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

// addTokens ports IrishTagger.addTokens (simple addAll; no '+' split).
func addTokens(taggedTokens, l []*languagetool.AnalyzedToken) []*languagetool.AnalyzedToken {
	if taggedTokens == nil {
		return l
	}
	return append(l, taggedTokens...)
}
