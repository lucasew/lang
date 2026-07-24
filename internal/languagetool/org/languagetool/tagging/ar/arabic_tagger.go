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
	TagManager *ArabicTagManager
	// dictLookup is Java DictionaryLookup(getDictionary()) used by additionalTags.
	// When nil, additionalTags falls back to WordTagger (injected maps / tests).
	dictLookup         tagging.WordTagger
	newStylePronounTag bool
}

// NewArabicTagger builds an ArabicTagger over the given WordTagger.
// Java: super("/ar/arabic.dict", new Locale("ar")) → tagLowercaseWithUppercase true.
func NewArabicTagger(wt tagging.WordTagger) *ArabicTagger {
	return &ArabicTagger{
		BaseTagger: tagging.NewBaseTagger(wt, ArabicDictPath, "ar", true),
		TagManager: NewArabicTagManager(),
	}
}

// NewArabicTaggerWithDictLookup sets the binary-dict stemmer for additionalTags
// (Java new DictionaryLookup(getDictionary()) inside tag()).
func NewArabicTaggerWithDictLookup(wt, dictLookup tagging.WordTagger) *ArabicTagger {
	t := NewArabicTagger(wt)
	t.dictLookup = dictLookup
	return t
}

// EnableNewStylePronounTag ports ArabicTagger.enableNewStylePronounTag.
func (t *ArabicTagger) EnableNewStylePronounTag() {
	if t != nil {
		t.newStylePronounTag = true
	}
}

// Tag ports ArabicTagger.tag: strip tashkeel for WordTagger; if not stop word,
// additionalTags via DictionaryLookup; null fallback; pos += word.length() (UTF-16).
func (t *ArabicTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0
	for _, word := range sentenceTokens {
		readings := t.tagOne(word)
		out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
		// Java: pos += word.length() (UTF-16 code units)
		pos += tagging.UTF16Len(word)
	}
	return out
}

// tagOne ports the per-token body of ArabicTagger.tag.
func (t *ArabicTagger) tagOne(word string) []*languagetool.AnalyzedToken {
	striped := tools.RemoveTashkeel(word)
	var readings []*languagetool.AnalyzedToken
	// Java: getWordTagger().tag(striped) — exact WordTagger, no BaseTagger case-merge
	for _, tw := range t.TagWordExact(striped) {
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
// Java uses DictionaryLookup(getDictionary()) — not getWordTagger() — for stem lookups.
func (t *ArabicTagger) additionalTags(word string) []*languagetool.AnalyzedToken {
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
	striped := tools.RemoveTashkeel(word)
	var out []*languagetool.AnalyzedToken
	prefixes := getPrefixIndexList(striped)
	suffixes := getSuffixIndexList(striped)
	for _, i := range prefixes {
		for _, j := range suffixes {
			// Java: (i == 0) && (j == striped.length()) — BMP Arabic: length == rune count
			if i == 0 && j == utf8.RuneCountInString(striped) {
				continue
			}
			stems := getStem(striped, i, j)
			tags := t.getTags(striped, i, j)
			for _, stem := range stems {
				// Java: asAnalyzedTokenList(stem, stemmer.lookup(stem))
				for _, tw := range stemmer.Tag(stem) {
					posTag := tw.PosTag
					// modify tags in postag, return null if not compatible
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

// TagSingle ports ArabicTagger.tag(String) — single surface form.
// Named TagSingle to avoid shadowing BaseTagger.TagWord (WordTagger case-merge).
func (t *ArabicTagger) TagSingle(word string) *languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	atr := t.Tag([]string{word})
	if len(atr) == 0 {
		return nil
	}
	return atr[0]
}

// GetProclitic ports ArabicTagger.getProclitic.
func (t *ArabicTagger) GetProclitic(token *languagetool.AnalyzedToken) string {
	if t == nil || token == nil || t.TagManager == nil {
		return ""
	}
	postagPtr := token.GetPOSTag()
	if postagPtr == nil || *postagPtr == "" {
		return ""
	}
	postag := *postagPtr
	word := token.GetToken()
	if t.TagManager.IsVerb(postag) {
		conjflag := t.TagManager.GetFlag(postag, "CONJ")
		istqbalflag := t.TagManager.GetFlag(postag, "ISTIQBAL")
		prefixLength := 0
		if conjflag == 'W' {
			prefixLength++
		}
		if istqbalflag == 'S' {
			prefixLength++
		}
		return getPrefix(word, prefixLength)
	}
	if t.TagManager.IsNoun(postag) {
		conjflag := t.TagManager.GetFlag(postag, "CONJ")
		jarflag := t.TagManager.GetFlag(postag, "JAR")
		prefixLength := 0
		if conjflag != '-' {
			prefixLength++
		}
		if jarflag != '-' {
			prefixLength++
		}
		if t.TagManager.IsDefinite(postag) {
			if jarflag == 'L' {
				prefixLength++
			} else {
				prefixLength += 2
			}
		}
		return getPrefix(word, prefixLength)
	}
	return ""
}

// GetEnclitic ports ArabicTagger.getEnclitic.
func (t *ArabicTagger) GetEnclitic(token *languagetool.AnalyzedToken) string {
	if t == nil || token == nil || t.TagManager == nil {
		return ""
	}
	postagPtr := token.GetPOSTag()
	if postagPtr == nil || *postagPtr == "" {
		return ""
	}
	postag := *postagPtr
	word := token.GetToken()
	flag := t.TagManager.GetFlag(postag, "PRONOUN")
	if flag == '-' {
		return t.TagManager.GetPronounSuffix(postag)
	}
	suffix := ""
	switch {
	case strings.HasSuffix(word, "ه"):
		suffix = "ه"
	case strings.HasSuffix(word, "ها"):
		suffix = "ها"
	case strings.HasSuffix(word, "هما"):
		suffix = "هما"
	case strings.HasSuffix(word, "هم"):
		suffix = "هم"
	case strings.HasSuffix(word, "هن"):
		suffix = "هن"
	case strings.HasSuffix(word, "ك"):
		suffix = "ك"
	case strings.HasSuffix(word, "كما"):
		suffix = "كما"
	case strings.HasSuffix(word, "كم"):
		suffix = "كم"
	case strings.HasSuffix(word, "كن"):
		suffix = "كن"
	case strings.HasSuffix(word, "ني"):
		suffix = "ني"
	case strings.HasSuffix(word, "نا"):
		suffix = "نا"
	// Java unreachable-ish branches for مني/عني (already matched above)
	case (word == "عني" || word == "مني") && strings.HasSuffix(word, "ني"):
		suffix = "ني"
	case (word == "عنا" || word == "منا") && strings.HasSuffix(word, "نا"):
		suffix = "نا"
	default:
		suffix = ""
	}
	return suffix
}

// GetJarProclitic ports ArabicTagger.getJarProclitic.
func (t *ArabicTagger) GetJarProclitic(token *languagetool.AnalyzedToken) string {
	if t == nil || token == nil || t.TagManager == nil {
		return ""
	}
	postagPtr := token.GetPOSTag()
	if postagPtr == nil || *postagPtr == "" {
		return ""
	}
	postag := *postagPtr
	word := token.GetToken()
	if !t.TagManager.IsNoun(postag) {
		return ""
	}
	conjflag := t.TagManager.GetFlag(postag, "CONJ")
	jarflag := t.TagManager.GetFlag(postag, "JAR")
	prefixLength := 0
	if conjflag != '-' {
		prefixLength++
	}
	if jarflag != '-' {
		prefixLength++
	}
	if prefixLength > 0 {
		// Java: word.substring(prefixLength - 1, prefixLength) — one letter
		rs := []rune(word)
		idx := prefixLength - 1
		if idx >= 0 && idx < len(rs) {
			return string(rs[idx : idx+1])
		}
	}
	return ""
}

// GetLemmas ports ArabicTagger.getLemmas — unique lemmas by type (verb/adj/masdar).
func (t *ArabicTagger) GetLemmas(patternTokens *languagetool.AnalyzedTokenReadings, typ string) []string {
	if t == nil || patternTokens == nil || t.TagManager == nil {
		return nil
	}
	var lemmaList []string
	seen := map[string]bool{}
	for _, tok := range patternTokens.GetReadings() {
		if tok == nil || tok.GetPOSTag() == nil {
			continue
		}
		pos := *tok.GetPOSTag()
		ok := (t.TagManager.IsVerb(pos) && typ == "verb") ||
			(t.TagManager.IsAdj(pos) && typ == "adj") ||
			(t.TagManager.IsMasdar(pos) && typ == "masdar")
		if !ok {
			continue
		}
		lemma := ""
		if tok.GetLemma() != nil {
			lemma = *tok.GetLemma()
		}
		if !seen[lemma] {
			seen[lemma] = true
			lemmaList = append(lemmaList, lemma)
		}
	}
	return lemmaList
}
