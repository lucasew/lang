package gl

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Java GalicianTagger resource path.
const GalicianDictPath = "/gl/galician.dict"

var (
	glAdjPartFS        = regexp.MustCompile(`^V.P..SF.|A[QO].[FC][SN].$`)
	glVerb             = regexp.MustCompile(`^V.+`)
	glPrefixesForVerbs = regexp.MustCompile(`(?i)^(auto|re)(...+)$`)
)

// GalicianTagger ports org.languagetool.tagging.gl.GalicianTagger.
type GalicianTagger struct {
	*tagging.BaseTagger
}

func NewGalicianTagger(wt tagging.WordTagger) *GalicianTagger {
	// Java: tagLowercaseWithUppercase default true; overwriteWithManualTagger false.
	return &GalicianTagger{BaseTagger: tagging.NewBaseTagger(wt, GalicianDictPath, "gl", true)}
}

func (t *GalicianTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0
	for _, word := range sentenceTokens {
		// Java: typewriter apostrophe + chunk tag when original had '
		containsTypewriter := len(word) > 1 && strings.Contains(word, "'")
		w := strings.ReplaceAll(word, "’", "'")
		lower := strings.ToLower(w)
		isLower := w == lower
		isMixed := tools.IsMixedCase(w)

		var readings []*languagetool.AnalyzedToken
		// exact WordTagger lookups (Java getWordTagger().tag)
		for _, tw := range t.TagWordExact(w) {
			readings = append(readings, toTok(word, tw))
		}
		if !isLower && !isMixed {
			for _, tw := range t.TagWordExact(lower) {
				readings = append(readings, toTok(word, tw))
			}
		}
		// additional mente / verb prefixes
		if len(readings) == 0 && !isMixed {
			for _, at := range t.additionalTags(w) {
				readings = append(readings, at)
			}
		}
		if len(readings) == 0 {
			readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, nil, nil)}
		}
		atr := languagetool.NewAnalyzedTokenReadingsList(readings, pos)
		if containsTypewriter && atr != nil {
			atr.SetChunkTags([]string{"containsTypewriterApostrophe"})
		}
		out = append(out, atr)
		pos += tagging.UTF16Len(word)
	}
	return out
}

// additionalTags ports GalicianTagger.additionalTags (mente adverbs + auto/re verb prefixes).
func (t *GalicianTagger) additionalTags(word string) []*languagetool.AnalyzedToken {
	if t == nil || t.WordTagger == nil {
		return nil
	}
	lower := strings.ToLower(word)
	var out []*languagetool.AnalyzedToken
	if strings.HasSuffix(lower, "mente") {
		possibleAdj := strings.TrimSuffix(lower, "mente")
		for _, tw := range t.TagWordExact(possibleAdj) {
			if tw.PosTag != "" && glAdjPartFS.MatchString(tw.PosTag) {
				p, lemma := "RM", lower
				return []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, &p, &lemma)}
			}
		}
	}
	if m := glPrefixesForVerbs.FindStringSubmatch(word); m != nil {
		pref := strings.ToLower(m[1])
		possibleVerb := strings.ToLower(m[2])
		for _, tw := range t.TagWordExact(possibleVerb) {
			if tw.PosTag != "" && glVerb.MatchString(tw.PosTag) {
				p := tw.PosTag
				lemma := pref + tw.Lemma
				out = append(out, languagetool.NewAnalyzedToken(word, &p, &lemma))
			}
		}
	}
	return out
}

func toTok(surface string, tw tagging.TaggedWord) *languagetool.AnalyzedToken {
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
