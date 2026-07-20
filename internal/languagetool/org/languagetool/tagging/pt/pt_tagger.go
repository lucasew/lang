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
	// ADJ_PART_FS = V.P..SF.|A[QO].[FC][SN].
	ptAdjPartFS = regexp.MustCompile(`^V.P..SF.|A[QO].[FC][SN].$`)
	ptVerb      = regexp.MustCompile(`^V.+`)
	// Java PREFIXES_FOR_VERBS = (soto-)(...+)
	ptPrefixesForVerbs = regexp.MustCompile(`(?i)^(soto-)(...+)$`)

	ptOrdinalSuffixes = "[" + ordinalSuffixMasc + ordinalSuffixFem + "][" + ordinalSuffixPl + "]?"
	ptOrdinalPattern  = regexp.MustCompile(`^\d+[\d,.]*\.?` + ptOrdinalSuffixes + `$`)
	ptOrdinalMascSg   = regexp.MustCompile("[" + ordinalSuffixMasc + "]$")
	ptOrdinalFemSg    = regexp.MustCompile("[" + ordinalSuffixFem + "]$")
	ptOrdinalMascPl   = regexp.MustCompile("[" + ordinalSuffixMasc + "][" + ordinalSuffixPl + "]$")
	ptOrdinalFemPl    = regexp.MustCompile("[" + ordinalSuffixFem + "][" + ordinalSuffixPl + "]$")
	ptOrdinalReplace  = regexp.MustCompile(ptOrdinalSuffixes)
	ptPercentPattern  = regexp.MustCompile(`^−?\d+[\d,.]*%$`)
	ptDegreePattern   = regexp.MustCompile(`^−?\d+[\d,.]*°$`)
)

// PortugueseTagger ports org.languagetool.tagging.pt.PortugueseTagger.
type PortugueseTagger struct {
	*tagging.BaseTagger
}

func NewPortugueseTagger(wt tagging.WordTagger) *PortugueseTagger {
	// Java BaseTagger default tagLowercaseWithUppercase=true; Tag() reimplements case.
	return &PortugueseTagger{BaseTagger: tagging.NewBaseTagger(wt, PortugueseDictPath, "pt", true)}
}

func (t *PortugueseTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0
	for _, word := range sentenceTokens {
		containsTypewriter := len(word) > 1 && strings.Contains(word, "'")
		w := strings.ReplaceAll(word, "’", "'")
		lower := strings.ToLower(w)
		isLower := w == lower
		isMixed := portugueseIsMixedCase(w)

		var readings []*languagetool.AnalyzedToken
		// Java getWordTagger().tag — exact
		for _, tw := range t.TagWordExact(w) {
			readings = append(readings, tagged(word, tw))
		}
		if !isLower && !isMixed {
			for _, tw := range t.TagWordExact(lower) {
				readings = append(readings, tagged(word, tw))
			}
		}
		// ordinals / percent / degree
		if len(readings) == 0 {
			readings = append(readings, tagNumberExpressionsPT(w)...)
		}
		// -mente adverbs (RG)
		if len(readings) == 0 && !isMixed {
			readings = append(readings, t.tagMenteAdverbs(w, lower)...)
		}
		// soto- verb prefixes
		ignoreSpelling := false
		if len(readings) == 0 && !isMixed {
			pref := t.tagPrefixedVerbs(w)
			if len(pref) > 0 {
				readings = append(readings, pref...)
				ignoreSpelling = true
			}
		}
		if len(readings) == 0 {
			readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, nil, nil)}
		}
		atr := languagetool.NewAnalyzedTokenReadingsList(readings, pos)
		if ignoreSpelling && atr != nil {
			atr.IgnoreSpelling()
		}
		if containsTypewriter && atr != nil {
			atr.SetChunkTags([]string{"containsTypewriterApostrophe"})
		}
		out = append(out, atr)
		pos += len([]rune(word))
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
	numberGender := ""
	// Check pl before sg (pl patterns are longer suffixes)
	switch {
	case ptOrdinalMascPl.MatchString(word):
		numberGender = "MP"
	case ptOrdinalFemPl.MatchString(word):
		numberGender = "FP"
	case ptOrdinalMascSg.MatchString(word):
		numberGender = "MS"
	case ptOrdinalFemSg.MatchString(word):
		numberGender = "FS"
	}
	if numberGender == "" {
		return nil
	}
	nounTag := "NC" + numberGender + "000"
	adjTag := "AO0" + numberGender + "0"
	return []*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken(word, &nounTag, &lemma),
		languagetool.NewAnalyzedToken(word, &adjTag, &lemma),
	}
}

func (t *PortugueseTagger) tagMenteAdverbs(word, lowerWord string) []*languagetool.AnalyzedToken {
	if !strings.HasSuffix(strings.ToLower(word), "mente") {
		return nil
	}
	possibleAdj := strings.TrimSuffix(lowerWord, "mente")
	for _, tw := range t.TagWordExact(possibleAdj) {
		if tw.PosTag != "" && ptAdjPartFS.MatchString(tw.PosTag) {
			p, lemma := "RG", lowerWord
			return []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, &p, &lemma)}
		}
	}
	return nil
}

func (t *PortugueseTagger) tagPrefixedVerbs(word string) []*languagetool.AnalyzedToken {
	m := ptPrefixesForVerbs.FindStringSubmatch(word)
	if m == nil {
		return nil
	}
	prefix := strings.ToLower(m[1])
	possibleVerb := strings.ToLower(m[2])
	var out []*languagetool.AnalyzedToken
	for _, tw := range t.TagWordExact(possibleVerb) {
		if tw.PosTag == "" || !ptVerb.MatchString(tw.PosTag) {
			continue
		}
		lemma := prefix + tw.Lemma
		// Java: only if combined lemma not in dict
		if len(t.TagWordExact(lemma)) > 0 {
			continue
		}
		p, l := tw.PosTag, lemma
		out = append(out, languagetool.NewAnalyzedToken(word, &p, &l))
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
