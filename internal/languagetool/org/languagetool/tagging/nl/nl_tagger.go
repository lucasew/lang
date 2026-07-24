package nl

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const DutchDictPath = "/nl/dutch.dict"

// DutchTagger ports org.languagetool.tagging.nl.DutchTagger.
//
// Compound acceptance uses GetCompoundParts (Java Dutch.getCompoundAcceptor().getParts).
// Nil GetCompoundParts → compound branch inactive (fail-closed; no invent).
// GetPostags uses raw WordTagger only (avoids compound re-entry loop).
type DutchTagger struct {
	*tagging.BaseTagger
	// GetCompoundParts ports CompoundAcceptor.getParts; inject from rules/nl to avoid import cycle.
	GetCompoundParts func(word string) []string
}

func NewDutchTagger(wt tagging.WordTagger) *DutchTagger {
	return &DutchTagger{BaseTagger: tagging.NewBaseTagger(wt, DutchDictPath, "nl", false)}
}

// Java alwaysNeedsHet / alwaysNeedsDe / alwaysNeedsMrv for compound ZNW tag overrides.
var (
	alwaysNeedsHet = map[string]struct{}{
		"patroon": {}, "punt": {}, "gemaal": {}, "weer": {}, "kussen": {}, "deel": {},
	}
	alwaysNeedsDe = map[string]struct{}{
		"keten": {}, "boor": {}, "dans": {},
	}
	alwaysNeedsMrv = map[string]struct{}{
		"pies": {}, "koeken": {}, "heden": {},
	}
)

// Accent / hyphen patterns from DutchTagger (static Pattern.compile).
var (
	pattern1A = regexp.MustCompile(`([^aeiouáéíóú])(á)([^aeiouáéíóú])`)
	pattern1E = regexp.MustCompile(`([^aeiouáéíóú])(é)([^aeiouáéíóú])`)
	pattern1I = regexp.MustCompile(`([^aeiouáéíóú])(í)([^aeiouáéíóú])`)
	pattern1O = regexp.MustCompile(`([^aeiouáéíóú])(ó)([^aeiouáéíóú])`)
	pattern1U = regexp.MustCompile(`([^aeiouáéíóú])(ú)([^aeiouáéíóú])`)

	charPatternAA = regexp.MustCompile(`áá`)
	charPatternAE = regexp.MustCompile(`áé`)
	charPatternAI = regexp.MustCompile(`áí`)
	charPatternAU = regexp.MustCompile(`áú`)
	charPatternEE = regexp.MustCompile(`éé`)
	charPatternEI = regexp.MustCompile(`éí`)
	charPatternEU = regexp.MustCompile(`éú`)
	charPatternIE = regexp.MustCompile(`íé`)
	charPatternOE = regexp.MustCompile(`óé`)
	charPatternOI = regexp.MustCompile(`óí`)
	charPatternOO = regexp.MustCompile(`óó`)
	charPatternOU = regexp.MustCompile(`óú`)
	charPatternUI = regexp.MustCompile(`úí`)
	charPatternUU = regexp.MustCompile(`úú`)
	charPatternIJ = regexp.MustCompile(`íj`)

	pattern2A = regexp.MustCompile(`(^|[^aeiou])á([^aeiou]|$)`)
	pattern2E = regexp.MustCompile(`(^|[^aeiou])é([^aeiou]|$)`)
	pattern2I = regexp.MustCompile(`(^|[^aeiou])í([^aeiou]|$)`)
	pattern2O = regexp.MustCompile(`(^|[^aeiou])ó([^aeiou]|$)`)
	pattern2U = regexp.MustCompile(`(^|[^aeiou])ú([^aeiou]|$)`)

	hyphen1Pattern = regexp.MustCompile(`(^.*)-(.*$)`)
	hyphen2Pattern = regexp.MustCompile(`([a-z])-([a-z])`)
)

func (t *DutchTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0
	for _, originalWord := range sentenceTokens {
		atr := t.tagOne(originalWord, pos)
		out = append(out, atr)
		// Java: pos += word.length() after restoring originalWord
		pos += tagging.UTF16Len(originalWord)
	}
	return out
}

func (t *DutchTagger) tagOne(originalWord string, startPos int) *languagetool.AnalyzedTokenReadings {
	// Java: normalize weird apostrophes like tokenizer
	word := originalWord
	word = strings.ReplaceAll(word, "`", "'")
	word = strings.ReplaceAll(word, "’", "'")
	word = strings.ReplaceAll(word, "‘", "'")
	word = strings.ReplaceAll(word, "´", "'")

	lowerWord := strings.ToLower(word)
	isLowercase := word == lowerWord
	isMixedCase := tools.IsMixedCase(word)
	isAllUpper := tools.IsAllUppercase(word)

	var l []*languagetool.AnalyzedToken
	ignoreSpelling := false

	// normal case: tag surface (surface token = originalWord for readings)
	l = append(l, t.asAnalyzed(originalWord, t.tagExact(word))...)
	// non-lowercase, not mixed: also lowercase tags
	if !isLowercase && !isMixedCase {
		l = append(l, t.asAnalyzed(originalWord, t.tagExact(lowerWord))...)
	}
	// all-uppercase proper nouns: first-upper
	if len(l) == 0 && isAllUpper {
		firstUpper := tools.UppercaseFirstChar(lowerWord)
		l = append(l, t.asAnalyzed(originalWord, t.tagExact(firstUpper))...)
	}

	if len(l) == 0 {
		word2 := word
		// remove single accented characters (pattern1)
		word2 = pattern1A.ReplaceAllString(word2, "${1}a${3}")
		word2 = pattern1E.ReplaceAllString(word2, "${1}e${3}")
		word2 = pattern1I.ReplaceAllString(word2, "${1}i${3}")
		word2 = pattern1O.ReplaceAllString(word2, "${1}o${3}")
		word2 = pattern1U.ReplaceAllString(word2, "${1}u${3}")
		// remove allowed accented digraphs
		word2 = charPatternAA.ReplaceAllString(word2, "aa")
		word2 = charPatternAE.ReplaceAllString(word2, "ae")
		word2 = charPatternAI.ReplaceAllString(word2, "ai")
		word2 = charPatternAU.ReplaceAllString(word2, "au")
		word2 = charPatternEE.ReplaceAllString(word2, "ee")
		word2 = charPatternEI.ReplaceAllString(word2, "ei")
		word2 = charPatternEU.ReplaceAllString(word2, "eu")
		word2 = charPatternIE.ReplaceAllString(word2, "ie")
		word2 = charPatternOE.ReplaceAllString(word2, "oe")
		word2 = charPatternOI.ReplaceAllString(word2, "oi")
		word2 = charPatternOO.ReplaceAllString(word2, "oo")
		word2 = charPatternOU.ReplaceAllString(word2, "ou")
		word2 = charPatternUI.ReplaceAllString(word2, "ui")
		word2 = charPatternUU.ReplaceAllString(word2, "uu")
		word2 = charPatternIJ.ReplaceAllString(word2, "ij")
		// pattern2 residual accents
		word2 = pattern2A.ReplaceAllString(word2, "${1}a${2}")
		word2 = pattern2E.ReplaceAllString(word2, "${1}e${2}")
		word2 = pattern2I.ReplaceAllString(word2, "${1}i${2}")
		word2 = pattern2O.ReplaceAllString(word2, "${1}o${2}")
		word2 = pattern2U.ReplaceAllString(word2, "${1}u${2}")

		// hyphen: if part2 tags, drop hyphen between letters
		if strings.Contains(word2, "-") {
			m := hyphen1Pattern.FindStringSubmatch(word2)
			if len(m) == 3 {
				part2 := m[2]
				if len(t.tagExact(part2)) > 0 {
					word2 = hyphen2Pattern.ReplaceAllString(word2, "${1}${2}")
				}
			}
		}

		if word2 != word {
			l2 := t.asAnalyzed(originalWord, t.tagExact(word2))
			if len(l2) > 0 {
				l = append(l, l2...)
				ignoreSpelling = true
			}
		}

		// Tag unknown compound words (Java: word.length() > 5)
		if len(l) == 0 && tagging.UTF16Len(word) > 5 && t.GetCompoundParts != nil {
			parts := t.GetCompoundParts(word)
			if len(parts) == 2 {
				part1, part2 := parts[0], parts[1]
				// recursive tag of part2 only (Java tag(singletonList(part2)))
				part2ATR := t.tagOne(part2, 0)
				part1lc := strings.ToLower(part1)
				if part2ATR != nil {
					for _, part2Reading := range part2ATR.GetReadings() {
						if part2Reading == nil || part2Reading.GetPOSTag() == nil {
							continue
						}
						posTag := *part2Reading.GetPOSTag()
						if strings.HasSuffix(part1, "-") {
							if strings.HasPrefix(posTag, "ENM:LOC") {
								l = append(l, languagetool.NewAnalyzedToken(word, strPtr(posTag), strPtr(part2)))
								break
							}
						}
						if strings.HasPrefix(posTag, "ZNW") {
							tag := posTag
							if _, ok := alwaysNeedsHet[part2]; ok {
								tag = "ZNW:EKV:HET"
							} else if _, ok := alwaysNeedsDe[part2]; ok {
								tag = "ZNW:EKV:DE_"
							} else if _, ok := alwaysNeedsMrv[part2]; ok {
								tag = "ZNW:MRV:DE_"
							}
							lemma := part1lc
							if part2Reading.GetLemma() != nil {
								lemma = part1lc + *part2Reading.GetLemma()
							} else {
								lemma = part1lc + part2
							}
							l = append(l, languagetool.NewAnalyzedToken(word, strPtr(tag), strPtr(lemma)))
							if _, ok := alwaysNeedsHet[part2]; ok {
								break
							}
							if _, ok := alwaysNeedsDe[part2]; ok {
								break
							}
							if _, ok := alwaysNeedsMrv[part2]; ok {
								break
							}
						}
					}
				}
			}
		}
	}

	if len(l) == 0 {
		l = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(originalWord, nil, nil)}
	}

	atr := languagetool.NewAnalyzedTokenReadingsList(l, startPos)
	if ignoreSpelling {
		// Java: if lowercase and first-upper form exists in dict, clear to null reading;
		// else ignoreSpelling. Non-lowercase → ignoreSpelling.
		if isLowercase {
			up := tools.UppercaseFirstChar(originalWord)
			fu := t.tagExact(up)
			if len(fu) == 0 {
				atr.IgnoreSpelling()
			} else {
				// uppercased form exists → this lowercase is probably wrong; null reading
				atr = languagetool.NewAnalyzedTokenReadingsList(
					[]*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(originalWord, nil, nil)},
					startPos,
				)
			}
		} else {
			atr.IgnoreSpelling()
		}
	}
	return atr
}

// GetPostags ports DutchTagger.getPostags: raw word-tagger tags only
// (no compound / accent path — prevents CompoundAcceptor loop).
func (t *DutchTagger) GetPostags(word string) []*languagetool.AnalyzedToken {
	if t == nil {
		return nil
	}
	return t.asAnalyzed(word, t.tagExact(word))
}

// tagExact is Java getWordTagger().tag(word) — no BaseTagger case-merge.
func (t *DutchTagger) tagExact(word string) []tagging.TaggedWord {
	if t == nil || t.BaseTagger == nil {
		return nil
	}
	return t.TagWordExact(word)
}

func (t *DutchTagger) asAnalyzed(surface string, tws []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	if len(tws) == 0 {
		return nil
	}
	out := make([]*languagetool.AnalyzedToken, 0, len(tws))
	for _, tw := range tws {
		out = append(out, tagged(surface, tw))
	}
	return out
}

func tagged(surface string, tw tagging.TaggedWord) *languagetool.AnalyzedToken {
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

func strPtr(s string) *string { return &s }
