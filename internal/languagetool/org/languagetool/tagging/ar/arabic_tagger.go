package ar

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const ArabicDictPath = "/ar/arabic.dict"

// ArabicTagger ports org.languagetool.tagging.ar.ArabicTagger.
// Base dict lookup + prefix/suffix stemming (additionalTags) like Java.
type ArabicTagger struct {
	*tagging.BaseTagger
	TagManager         *ArabicTagManager
	newStylePronounTag bool
}

func NewArabicTagger(wt tagging.WordTagger) *ArabicTagger {
	return &ArabicTagger{
		BaseTagger: tagging.NewBaseTagger(wt, ArabicDictPath, "ar", false),
		TagManager: NewArabicTagManager(),
	}
}

// EnableNewStylePronounTag ports ArabicTagger.enableNewStylePronounTag.
func (t *ArabicTagger) EnableNewStylePronounTag() {
	if t != nil {
		t.newStylePronounTag = true
	}
}

func (t *ArabicTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0
	for _, word := range sentenceTokens {
		readings := t.tagOne(word)
		out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
		pos += utf8.RuneCountInString(word)
	}
	return out
}

// tagOne ports the per-token body of ArabicTagger.tag.
func (t *ArabicTagger) tagOne(word string) []*languagetool.AnalyzedToken {
	striped := tools.RemoveTashkeel(word)
	var readings []*languagetool.AnalyzedToken
	for _, tw := range t.TagWord(striped) {
		readings = append(readings, tagged(word, tw))
	}
	// Java: if not a stop word, additional prefix/suffix stemming
	if !t.IsStopWordReading(readings) {
		readings = append(readings, t.additionalTags(word)...)
	}
	if len(readings) == 0 {
		readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, nil, nil)}
	}
	return readings
}

// TagTokens returns POS/lemma for one surface (TagWord inject helper).
func (t *ArabicTagger) TagTokens(word string) []tagging.TaggedWord {
	readings := t.tagOne(word)
	out := make([]tagging.TaggedWord, 0, len(readings))
	for _, r := range readings {
		if r == nil {
			continue
		}
		pos, lemma := "", ""
		if r.GetPOSTag() != nil {
			pos = *r.GetPOSTag()
		}
		if r.GetLemma() != nil {
			lemma = *r.GetLemma()
		}
		if pos == "" && lemma == "" {
			continue
		}
		out = append(out, tagging.NewTaggedWord(lemma, pos))
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

// IsStopWordReading reports if any reading is a particle (P…).
func (t *ArabicTagger) IsStopWordReading(readings []*languagetool.AnalyzedToken) bool {
	if t == nil || t.TagManager == nil {
		return false
	}
	for _, r := range readings {
		if r != nil && r.GetPOSTag() != nil && t.TagManager.IsStopWord(*r.GetPOSTag()) {
			return true
		}
	}
	return false
}

// additionalTags ports ArabicTagger.additionalTags.
func (t *ArabicTagger) additionalTags(word string) []*languagetool.AnalyzedToken {
	if t == nil {
		return nil
	}
	striped := tools.RemoveTashkeel(word)
	var out []*languagetool.AnalyzedToken
	prefixes := getPrefixIndexList(striped)
	suffixes := getSuffixIndexList(striped)
	for _, i := range prefixes {
		for _, j := range suffixes {
			if i == 0 && j == utf8.RuneCountInString(striped) {
				continue
			}
			stems := getStem(striped, i, j)
			tags := t.getTags(striped, i, j)
			for _, stem := range stems {
				for _, tw := range t.TagWord(stem) {
					posTag := tw.PosTag
					posTag = t.TagManager.ModifyPosTag(posTag, tags)
					if posTag == "" {
						continue
					}
					lemma := tw.Lemma
					out = append(out, languagetool.NewAnalyzedToken(word, &posTag, strPtr(lemma)))
				}
			}
		}
	}
	return out
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func getSuffixIndexList(possibleWord string) []int {
	// Work in runes for Arabic.
	rs := []rune(possibleWord)
	n := len(rs)
	suffixIndexes := []int{n}
	suffixPos := n
	if strings.HasSuffix(possibleWord, "ك") ||
		strings.HasSuffix(possibleWord, "ها") ||
		strings.HasSuffix(possibleWord, "هما") ||
		strings.HasSuffix(possibleWord, "كما") ||
		strings.HasSuffix(possibleWord, "هم") ||
		strings.HasSuffix(possibleWord, "هن") ||
		strings.HasSuffix(possibleWord, "كم") ||
		strings.HasSuffix(possibleWord, "كن") ||
		strings.HasSuffix(possibleWord, "نا") {
		if strings.HasSuffix(possibleWord, "ك") {
			suffixPos -= 1
		} else if strings.HasSuffix(possibleWord, "هما") || strings.HasSuffix(possibleWord, "كما") {
			suffixPos -= 3
		} else {
			suffixPos -= 2
		}
		suffixIndexes = append(suffixIndexes, suffixPos)
	}
	return suffixIndexes
}

func getPrefixIndexList(possibleWord string) []int {
	prefixIndexes := []int{0}
	// four letters
	if strings.HasPrefix(possibleWord, "وكال") ||
		strings.HasPrefix(possibleWord, "وبال") ||
		strings.HasPrefix(possibleWord, "فكال") ||
		strings.HasPrefix(possibleWord, "فبال") {
		prefixIndexes = append(prefixIndexes, 4)
	}
	// three letters
	if strings.HasPrefix(possibleWord, "ولل") ||
		strings.HasPrefix(possibleWord, "فلل") ||
		strings.HasPrefix(possibleWord, "فال") ||
		strings.HasPrefix(possibleWord, "وال") ||
		strings.HasPrefix(possibleWord, "بال") ||
		strings.HasPrefix(possibleWord, "كال") {
		prefixIndexes = append(prefixIndexes, 3)
	}
	// two letters
	if strings.HasPrefix(possibleWord, "لل") ||
		strings.HasPrefix(possibleWord, "وك") ||
		strings.HasPrefix(possibleWord, "ول") ||
		strings.HasPrefix(possibleWord, "وب") ||
		strings.HasPrefix(possibleWord, "فك") ||
		strings.HasPrefix(possibleWord, "فل") ||
		strings.HasPrefix(possibleWord, "فب") ||
		strings.HasPrefix(possibleWord, "ال") ||
		strings.HasPrefix(possibleWord, "فسأ") ||
		strings.HasPrefix(possibleWord, "فسن") ||
		strings.HasPrefix(possibleWord, "فسي") ||
		strings.HasPrefix(possibleWord, "فست") ||
		strings.HasPrefix(possibleWord, "وسأ") ||
		strings.HasPrefix(possibleWord, "وسن") ||
		strings.HasPrefix(possibleWord, "وسي") ||
		strings.HasPrefix(possibleWord, "وست") {
		prefixIndexes = append(prefixIndexes, 2)
	}
	// one letter
	if strings.HasPrefix(possibleWord, "ك") ||
		strings.HasPrefix(possibleWord, "ل") ||
		strings.HasPrefix(possibleWord, "ب") ||
		strings.HasPrefix(possibleWord, "و") ||
		strings.HasPrefix(possibleWord, "ف") ||
		strings.HasPrefix(possibleWord, "سأ") ||
		strings.HasPrefix(possibleWord, "سن") ||
		strings.HasPrefix(possibleWord, "سي") ||
		strings.HasPrefix(possibleWord, "ست") {
		prefixIndexes = append(prefixIndexes, 1)
	}
	return prefixIndexes
}

func (t *ArabicTagger) getTags(word string, posStart, posEnd int) []string {
	var tags []string
	prefix := getPrefix(word, posStart)
	suffix := getSuffix(word, posEnd)
	// prefixes
	if strings.HasPrefix(prefix, "و") || strings.HasPrefix(prefix, "ف") {
		tags = append(tags, "CONJ;W")
		prefix = regexp.MustCompile(`^[وف]`).ReplaceAllString(prefix, "")
	}
	if strings.HasPrefix(prefix, "ك") {
		tags = append(tags, "JAR;K")
	} else if strings.HasPrefix(prefix, "ل") {
		tags = append(tags, "JAR;L")
	} else if strings.HasPrefix(prefix, "ب") {
		tags = append(tags, "JAR;B")
	} else if strings.HasPrefix(prefix, "س") {
		tags = append(tags, "ISTIQBAL;S")
	}
	if strings.HasSuffix(prefix, "ال") || strings.HasSuffix(prefix, "لل") {
		tags = append(tags, "PRONOUN;D")
	}
	// suffixes
	if t != nil && t.newStylePronounTag {
		switch suffix {
		case "ني":
			tags = append(tags, "PRONOUN;b")
		case "نا":
			tags = append(tags, "PRONOUN;c")
		case "ك":
			tags = append(tags, "PRONOUN;d")
		case "كما":
			tags = append(tags, "PRONOUN;e")
		case "كم":
			tags = append(tags, "PRONOUN;f")
		case "كن":
			tags = append(tags, "PRONOUN;g")
		case "ه":
			tags = append(tags, "PRONOUN;H")
		case "ها":
			tags = append(tags, "PRONOUN;i")
		case "هما":
			tags = append(tags, "PRONOUN;j")
		case "هم":
			tags = append(tags, "PRONOUN;k")
		case "هن":
			tags = append(tags, "PRONOUN;n")
		}
	} else {
		switch suffix {
		case "ني", "نا", "ك", "كما", "كم", "كن", "ه", "ها", "هما", "هم", "هن":
			tags = append(tags, "PRONOUN;H")
		}
	}
	return tags
}

func getPrefix(word string, pos int) string {
	rs := []rune(word)
	if pos < 0 {
		pos = 0
	}
	if pos > len(rs) {
		pos = len(rs)
	}
	return string(rs[:pos])
}

func getSuffix(word string, pos int) string {
	rs := []rune(word)
	if pos < 0 {
		pos = 0
	}
	if pos > len(rs) {
		pos = len(rs)
	}
	return string(rs[pos:])
}

var attachedPronounRE = regexp.MustCompile(`(ك|ها|هما|هم|هن|كما|كم|كن|نا|ي)$`)

func getStem(word string, posStart, posEnd int) []string {
	rs := []rune(word)
	n := len(rs)
	if posStart < 0 {
		posStart = 0
	}
	if posStart > n {
		posStart = n
	}
	if posEnd < posStart {
		posEnd = posStart
	}
	if posEnd > n {
		posEnd = n
	}
	// Java: stem = word.substring(posStart); then maybe replace suffix with ه
	stem := string(rs[posStart:])
	if posEnd != n {
		stem = attachedPronounRE.ReplaceAllString(stem, "ه")
	}
	var stemList []string
	prefix := getPrefix(word, posStart)
	if strings.HasSuffix(prefix, "لل") {
		stemList = append(stemList, "ل"+stem)
	}
	stemList = append(stemList, stem)
	return stemList
}
