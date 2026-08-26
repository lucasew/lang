package gl

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Java GalicianTagger resource path: super("/gl/galician.dict", new Locale("gl")).
const GalicianDictPath = "/gl/galician.dict"

var (
	// Java ADJ_PART_FS / VERB use Matcher.matches() → full-string match.
	glAdjPartFS = regexp.MustCompile(`^(?:V.P..SF.|A[QO].[FC][SN].)$`)
	glVerb      = regexp.MustCompile(`^V.+$`)
	// Java PREFIXES_FOR_VERBS: (auto|re)(...+) CASE_INSENSITIVE|UNICODE_CASE + matches().
	glPrefixesForVerbs = regexp.MustCompile(`(?i)^(auto|re)(...+)$`)
	// Java lowerWord.replaceAll("^(.+)mente$", "$1")
	glMenteStem = regexp.MustCompile(`^(.+)mente$`)
)

// GalicianTagger ports org.languagetool.tagging.gl.GalicianTagger.
type GalicianTagger struct {
	*tagging.BaseTagger
	// dictLookup is Java DictionaryLookup(getDictionary()) used by additionalTags.
	// When nil, additionalTags falls back to WordTagger (injected maps / tests).
	dictLookup tagging.WordTagger
}

// NewGalicianTagger builds a GalicianTagger over the given WordTagger.
// Java: super("/gl/galician.dict", new Locale("gl")); overwriteWithManualTagger() → false.
func NewGalicianTagger(wt tagging.WordTagger) *GalicianTagger {
	return &GalicianTagger{BaseTagger: tagging.NewBaseTagger(wt, GalicianDictPath, "gl", true)}
}

// NewGalicianTaggerWithDictLookup sets the binary-dict stemmer for additionalTags
// (Java new DictionaryLookup(getDictionary()) inside additionalTags).
func NewGalicianTaggerWithDictLookup(wt, dictLookup tagging.WordTagger) *GalicianTagger {
	t := NewGalicianTagger(wt)
	t.dictLookup = dictLookup
	return t
}

// OverwriteWithManualTagger ports GalicianTagger.overwriteWithManualTagger() → false.
func (t *GalicianTagger) OverwriteWithManualTagger() bool { return false }

// Tag ports GalicianTagger.tag: apostrophe normalize + typewriter chunk tag;
// exact+lowercase WordTagger; additionalTags for -mente / auto|re; null fallback;
// pos += word.length() (UTF-16 of the working word after replace).
func (t *GalicianTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
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
		isMixed := tools.IsMixedCase(w)

		var readings []*languagetool.AnalyzedToken
		// normal case: asAnalyzedTokenListForTaggedWords(word, getWordTagger().tag(word))
		for _, tw := range t.TagWordExact(w) {
			readings = append(readings, toTok(w, tw))
		}
		// non-lowercase, not mixed: also lowercase tags with surface = working word
		if !isLower && !isMixed {
			for _, tw := range t.TagWordExact(lower) {
				readings = append(readings, toTok(w, tw))
			}
		}
		// additional tagging with prefixes / -mente (Java: only when empty && !mixed)
		if len(readings) == 0 && !isMixed {
			if extra := t.additionalTags(w); extra != nil {
				readings = append(readings, extra...)
			}
		}
		if len(readings) == 0 {
			readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(w, nil, nil)}
		}
		atr := languagetool.NewAnalyzedTokenReadingsList(readings, pos)
		if containsTypewriter && atr != nil {
			atr.SetChunkTags([]string{"containsTypewriterApostrophe"})
		}
		out = append(out, atr)
		// Java: pos += word.length() after reassignment (UTF-16 code units)
		pos += tagging.UTF16Len(w)
	}
	return out
}

// additionalTags ports GalicianTagger.additionalTags (mente adverbs + auto/re verb prefixes).
// Java uses DictionaryLookup(getDictionary()) — not getWordTagger() — for stem lookups.
func (t *GalicianTagger) additionalTags(word string) []*languagetool.AnalyzedToken {
	if t == nil {
		return nil
	}
	stemmer := t.dictLookup
	if stemmer == nil {
		if t.WordTagger == nil {
			return nil
		}
		stemmer = t.WordTagger
	}

	// Any well-formed adverb with suffix -mente is tagged as an adverb of manner (RM)
	// Java: word.endsWith("mente") — case-sensitive on the working surface
	if strings.HasSuffix(word, "mente") {
		lowerWord := strings.ToLower(word)
		// Java: lowerWord.replaceAll("^(.+)mente$", "$1") — at least one char before mente
		possibleAdj := lowerWord
		if m := glMenteStem.FindStringSubmatch(lowerWord); m != nil {
			possibleAdj = m[1]
		}
		for _, tw := range stemmer.Tag(possibleAdj) {
			if tw.PosTag != "" && glAdjPartFS.MatchString(tw.PosTag) {
				p, lemma := "RM", lowerWord
				return []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, &p, &lemma)}
			}
		}
	}
	// Any well-formed verb with prefixes is tagged as a verb copying the original tags
	if m := glPrefixesForVerbs.FindStringSubmatch(word); m != nil {
		pref := strings.ToLower(m[1])
		// Java: matcher.group(2).toLowerCase() — no locale
		possibleVerb := strings.ToLower(m[2])
		var out []*languagetool.AnalyzedToken
		for _, tw := range stemmer.Tag(possibleVerb) {
			if tw.PosTag != "" && glVerb.MatchString(tw.PosTag) {
				p := tw.PosTag
				lemma := pref + tw.Lemma
				out = append(out, languagetool.NewAnalyzedToken(word, &p, &lemma))
			}
		}
		// Java returns additionalTaggedTokens even when empty (not null) after prefix match
		return out
	}
	return nil
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
