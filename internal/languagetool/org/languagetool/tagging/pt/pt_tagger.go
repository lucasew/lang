package pt

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Java PortugueseTagger / BaseTagger resource path.
const PortugueseDictPath = "/pt/portuguese.dict"

// Patterns from PortugueseTagger.java
const (
	ordinalSuffixMasc = "oºᵒ"
	ordinalSuffixFem  = "aªᵃ"
	ordinalSuffixPl   = "sˢ"
)

var (
	// Java ADJ_PART_FS / VERB use Matcher.matches() → full-string match.
	ptAdjPartFS = regexp.MustCompile(`^(?:V.P..SF.|A[QO].[FC][SN].)$`)
	ptVerb      = regexp.MustCompile(`^V.+$`)
	// Java PREFIXES_FOR_VERBS = (soto-)(...+) CASE_INSENSITIVE|UNICODE_CASE
	ptPrefixesForVerbs = regexp.MustCompile(`(?i)^(soto-)(...+)$`)
	// Java lowerWord.replaceAll("^(.+)mente$", "$1")
	ptMenteStem = regexp.MustCompile(`^(.+)mente$`)

	ptOrdinalSuffixes = "[" + ordinalSuffixMasc + ordinalSuffixFem + "][" + ordinalSuffixPl + "]?"
	ptOrdinalPattern  = regexp.MustCompile(`^\d+[\d,.]*\.?` + ptOrdinalSuffixes + `$`)
	ptOrdinalMascSg   = regexp.MustCompile("[" + ordinalSuffixMasc + "]$")
	ptOrdinalFemSg    = regexp.MustCompile("[" + ordinalSuffixFem + "]$")
	ptOrdinalMascPl   = regexp.MustCompile("[" + ordinalSuffixMasc + "][" + ordinalSuffixPl + "]$")
	ptOrdinalFemPl    = regexp.MustCompile("[" + ordinalSuffixFem + "][" + ordinalSuffixPl + "]$")
	ptOrdinalReplace  = regexp.MustCompile(ptOrdinalSuffixes)
	// Java PERCENT_PATTERN / DEGREE_PATTERN (matches on full string)
	ptPercentPattern = regexp.MustCompile(`^−?\d+[\d,.]*%$`)
	ptDegreePattern  = regexp.MustCompile(`^−?\d+[\d,.]*°$`)
)

// PortugueseTagger ports org.languagetool.tagging.pt.PortugueseTagger.
type PortugueseTagger struct {
	*tagging.BaseTagger
	// dictLookup is Java DictionaryLookup(getDictionary()) used by tagMenteAdverbs /
	// tagPrefixedVerbs. When nil, those paths fall back to WordTagger (injected maps / tests).
	dictLookup tagging.WordTagger
}

// NewPortugueseTagger builds a PortugueseTagger over the given WordTagger.
// Java: super("/pt/portuguese.dict", new Locale("pt")); tagLowercaseWithUppercase default true.
func NewPortugueseTagger(wt tagging.WordTagger) *PortugueseTagger {
	return &PortugueseTagger{BaseTagger: tagging.NewBaseTagger(wt, PortugueseDictPath, "pt", true)}
}

// NewPortugueseTaggerWithDictLookup sets the binary-dict stemmer for -mente / soto-
// (Java new DictionaryLookup(getDictionary()) inside tag()).
func NewPortugueseTaggerWithDictLookup(wt, dictLookup tagging.WordTagger) *PortugueseTagger {
	t := NewPortugueseTagger(wt)
	t.dictLookup = dictLookup
	return t
}

// Tag ports PortugueseTagger.tag.
func (t *PortugueseTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0
	for _, word := range sentenceTokens {
		// Java: typewriter apostrophe detect + curly → typewriter when length > 1
		containsTypewriter := false
		w := word
		if tagging.UTF16Len(w) > 1 {
			if strings.Contains(w, "'") {
				containsTypewriter = true
			}
			w = strings.ReplaceAll(w, "’", "'")
		}
		// Java lowerWord = word.toLowerCase(locale) on (possibly) replaced word
		lower := strings.ToLower(w)
		isLower := w == lower
		isMixed := portugueseIsMixedCase(w)

		var readings []*languagetool.AnalyzedToken
		// normal case: asAnalyzedTokenListForTaggedWords(word, getWordTagger().tag(word))
		for _, tw := range t.TagWordExact(w) {
			readings = append(readings, tagged(w, tw))
		}
		// tag non-lowercase (alluppercase or startuppercase), but not mixedcase
		if !isLower && !isMixed {
			for _, tw := range t.TagWordExact(lower) {
				readings = append(readings, tagged(w, tw))
			}
		}
		// tag ordinals and percentages / degrees
		if len(readings) == 0 {
			readings = append(readings, tagNumberExpressionsPT(w)...)
		}
		// additional tagging with adverbs ending in -mente
		if len(readings) == 0 && !isMixed {
			readings = append(readings, t.tagMenteAdverbs(w, lower)...)
		}
		// additional tagging with prefixes for verbs
		ignoreSpelling := false
		if len(readings) == 0 && !isMixed {
			pref := t.tagPrefixedVerbs(w)
			if len(pref) > 0 {
				readings = append(readings, pref...)
				ignoreSpelling = true
			}
		}
		if len(readings) == 0 {
			readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(w, nil, nil)}
		}
		atr := languagetool.NewAnalyzedTokenReadingsList(readings, pos)
		if ignoreSpelling && atr != nil {
			atr.IgnoreSpelling()
		}
		if containsTypewriter && atr != nil {
			atr.SetChunkTags([]string{"containsTypewriterApostrophe"})
		}
		out = append(out, atr)
		// Java: pos += word.length() after reassignment (UTF-16 code units)
		pos += tagging.UTF16Len(w)
	}
	return out
}

// portugueseIsMixedCase ports PortugueseTagger hyphen-aware mixed-case check.
func portugueseIsMixedCase(word string) bool {
	if strings.Contains(word, "-") {
		for _, part := range strings.Split(word, "-") {
			if tools.IsMixedCase(part) {
				return true
			}
		}
		return false
	}
	return tools.IsMixedCase(word)
}

func tagNumberExpressionsPT(word string) []*languagetool.AnalyzedToken {
	if ptOrdinalPattern.MatchString(word) {
		return buildOrdinalTokensPT(word)
	}
	if ptDegreePattern.MatchString(word) || ptPercentPattern.MatchString(word) {
		p, l := "NCMP000", word
		return []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, &p, &l)}
	}
	return nil
}

func buildOrdinalTokensPT(word string) []*languagetool.AnalyzedToken {
	// Java: lemma = word.replaceAll(ORDINAL_SUFFIXES, "º")
	lemma := ptOrdinalReplace.ReplaceAllString(word, "º")
	// Java sequential ifs overwrite numberGenderTags (later match wins: pl after sg).
	numberGender := ""
	if ptOrdinalMascSg.MatchString(word) {
		numberGender = "MS"
	}
	if ptOrdinalFemSg.MatchString(word) {
		numberGender = "FS"
	}
	if ptOrdinalMascPl.MatchString(word) {
		numberGender = "MP"
	}
	if ptOrdinalFemPl.MatchString(word) {
		numberGender = "FP"
	}
	nounTag := "NC" + numberGender + "000"
	adjTag := "AO0" + numberGender + "0"
	return []*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken(word, &nounTag, &lemma),
		languagetool.NewAnalyzedToken(word, &adjTag, &lemma),
	}
}

// tagMenteAdverbs ports PortugueseTagger.tagMenteAdverbs.
// Java uses DictionaryLookup(getDictionary()) — not getWordTagger() — for adj stem lookups.
func (t *PortugueseTagger) tagMenteAdverbs(word, lowerWord string) []*languagetool.AnalyzedToken {
	// Java: word.endsWith("mente") — case-sensitive on the working surface
	if !strings.HasSuffix(word, "mente") {
		return nil
	}
	stemmer := t.stemmer()
	if stemmer == nil {
		return nil
	}
	// Java: lowerWord.replaceAll("^(.+)mente$", "$1") — at least one char before mente
	possibleAdj := lowerWord
	if m := ptMenteStem.FindStringSubmatch(lowerWord); m != nil {
		possibleAdj = m[1]
	}
	for _, tw := range stemmer.Tag(possibleAdj) {
		if tw.PosTag != "" && ptAdjPartFS.MatchString(tw.PosTag) {
			p, lemma := "RG", lowerWord
			return []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, &p, &lemma)}
		}
	}
	return nil
}

// tagPrefixedVerbs ports PortugueseTagger.tagPrefixedVerbs (soto- only as of current Java).
func (t *PortugueseTagger) tagPrefixedVerbs(word string) []*languagetool.AnalyzedToken {
	m := ptPrefixesForVerbs.FindStringSubmatch(word)
	if m == nil {
		return nil
	}
	stemmer := t.stemmer()
	if stemmer == nil {
		return nil
	}
	prefix := strings.ToLower(m[1])
	possibleVerb := strings.ToLower(m[2])
	var out []*languagetool.AnalyzedToken
	for _, tw := range stemmer.Tag(possibleVerb) {
		if tw.PosTag == "" || !ptVerb.MatchString(tw.PosTag) {
			continue
		}
		lemma := prefix + tw.Lemma
		// Java: only if combined lemma not in pure binary dict
		if len(stemmer.Tag(lemma)) > 0 {
			continue
		}
		p, l := tw.PosTag, lemma
		out = append(out, languagetool.NewAnalyzedToken(word, &p, &l))
	}
	return out
}

func (t *PortugueseTagger) stemmer() tagging.WordTagger {
	if t == nil {
		return nil
	}
	if t.dictLookup != nil {
		return t.dictLookup
	}
	return t.WordTagger
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
