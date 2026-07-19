package uk

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// PosTagHelper ports tagging.uk.PosTagHelper helpers for Ukrainian POS tags.
type PosTagHelper struct{}

// NoVidminokSubstr ports PosTagHelper.NO_VIDMINOK_SUBSTR.
const NoVidminokSubstr = ":nv"

// Java Matcher.matches() patterns (full-string).
var (
	// (noun:(?:[iu]n)?anim|numr|adj|adjp.*):(.):v_.*
	numRegex = regexp.MustCompile(`^(?:noun:(?:[iu]n)?anim|numr|adj|adjp.*):(.):v_.*$`)
	// (noun:(?:[iu]n)?anim|numr|adj|adjp.*):[mfnp]:(v_...).*
	conjRegex = regexp.MustCompile(`^(?:noun:(?:[iu]n)?anim|numr|adj|adjp.*):[mfnp]:(v_...).*$`)
	// gender shares NUM_REGEX
	// (noun:(?:[iu]n)?anim|adj|numr|adjp.*):(.:v_...).*
	genderConjRegex = regexp.MustCompile(`^(?:noun:(?:[iu]n)?anim|adj|numr|adjp.*):(.:v_...).*$`)
	// :(comp.|adjp:.*?(:(im)?perf)+) — Java non-greedy; RE2 has no lookaround, use equivalent
	cleanupPattern = regexp.MustCompile(`:(?:comp.|adjp:.*?(?::(?:im)?perf)+)`)

	// AdjCompRegex ports PosTagHelper.ADJ_COMP_REGEX.
	AdjCompRegex = regexp.MustCompile(`:comp[bcs]`)

	// Common POS patterns (Java public static finals).
	NounVNazPattern  = regexp.MustCompile(`^noun.*:v_naz.*$`)
	AdjVNazPattern   = regexp.MustCompile(`^adj:.:v_naz.*$`)
	VerbInfPattern   = regexp.MustCompile(`^verb.*:inf.*$`)
	AdjVKlyPattern   = regexp.MustCompile(`^adj:.:v_kly.*$`)
	VerbPattern      = regexp.MustCompile(`^verb.*$`)
	VerbAdvpPattern  = regexp.MustCompile(`^(?:verb|advp).*$`)

	wordPattern          = regexp.MustCompile(`(?i)^[а-яіїєґa-z'\-]+$`)
	predictInsertPattern = regexp.MustCompile(`^noninfl:(?:predic|insert).*$`)
	// Java ".*[ую]" Matcher.matches() — surface ends with у/ю
	maleUATokenRE = regexp.MustCompile(`^.*[ую]$`)
)

// VidminkyMap ports PosTagHelper.VIDMINKY_MAP (case code → Ukrainian name).
// Iteration order for messages matches Java LinkedHashMap insertion.
var VidminkyMap = map[string]string{
	"v_naz": "називний",
	"v_rod": "родовий",
	"v_dav": "давальний",
	"v_zna": "знахідний",
	"v_oru": "орудний",
	"v_mis": "місцевий",
	"v_kly": "кличний",
}

// VidminkyIMap ports PosTagHelper.VIDMINKY_I_MAP (includes v_inf for verb gov messages).
var VidminkyIMap = map[string]string{
	"v_naz": "називний",
	"v_rod": "родовий",
	"v_dav": "давальний",
	"v_zna": "знахідний",
	"v_oru": "орудний",
	"v_mis": "місцевий",
	"v_kly": "кличний",
	"v_inf": "інфінітив",
}

// VidminkyOrder is LinkedHashMap insertion order for VIDMINKY_MAP.
var VidminkyOrder = []string{"v_naz", "v_rod", "v_dav", "v_zna", "v_oru", "v_mis", "v_kly"}

// BaseGenders ports PosTagHelper.BASE_GENDERS.
var BaseGenders = []string{"m", "f", "n", "p"}

// VidminokName returns the Ukrainian case name for a v_* code, or the code itself.
func VidminokName(code string) string {
	if n, ok := VidminkyMap[code]; ok {
		return n
	}
	return code
}

// VidminokIName returns VIDMINKY_I_MAP name (incl. інфінітив).
func VidminokIName(code string) string {
	if n, ok := VidminkyIMap[code]; ok {
		return n
	}
	return code
}

// GenderMap ports PosTagHelper.GENDER_MAP.
var GenderMap = map[string]string{
	"m": "ч.р.",
	"f": "ж.р.",
	"n": "с.р.",
	"p": "мн.",
	"s": "одн.",
	"i": "інф.",
	"o": "безос. форма",
}

// GenderName returns the Ukrainian gender label, or the code itself.
func GenderName(code string) string {
	if n, ok := GenderMap[code]; ok {
		return n
	}
	return code
}

// PersonMap ports PosTagHelper.PERSON_MAP (includes s/p from Java).
var PersonMap = map[string]string{
	"1": "1-а особа",
	"2": "2-а особа",
	"3": "3-я особа",
	"s": "одн.",
	"p": "мн.",
}

// PersonName returns PERSON_MAP label or the code itself.
func PersonName(code string) string {
	if n, ok := PersonMap[code]; ok {
		return n
	}
	return code
}

// HasPos reports whether pos contains the given tag fragment (colon-separated).
func HasPos(pos, fragment string) bool {
	if pos == "" || fragment == "" {
		return false
	}
	for _, p := range strings.Split(pos, ":") {
		if p == fragment {
			return true
		}
	}
	return false
}

// IsNoun reports noun tags (noun:...).
func IsNoun(pos string) bool {
	return strings.HasPrefix(pos, "noun")
}

// IsVerb reports verb tags.
func IsVerb(pos string) bool {
	return strings.HasPrefix(pos, "verb")
}

// IsAdj reports adjective tags.
func IsAdj(pos string) bool {
	return strings.HasPrefix(pos, "adj")
}

// Gender returns m/f/n/s/p from POS if present (colon fragment scan; not Java getGender).
func Gender(pos string) string {
	for _, g := range []string{"m", "f", "n", "s", "p"} {
		if HasPos(pos, g) {
			return g
		}
	}
	return ""
}

// Case returns nom/gen/dat/acc/ins/loc/voc if present.
func Case(pos string) string {
	for _, c := range []string{"v_naz", "v_rod", "v_dav", "v_zna", "v_oru", "v_mis", "v_kly", "nom", "gen", "dat", "acc", "ins", "loc", "voc"} {
		if HasPos(pos, c) {
			return c
		}
	}
	return ""
}

// GetGender ports PosTagHelper.getGender (NUM/GENDER_REGEX Matcher.matches).
func GetGender(posTag string) string {
	m := numRegex.FindStringSubmatch(posTag)
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}

// GetNum ports PosTagHelper.getNum ("p" stays plural; other gender → "s").
func GetNum(posTag string) string {
	m := numRegex.FindStringSubmatch(posTag)
	if len(m) < 2 {
		return ""
	}
	if m[1] != "p" {
		return "s"
	}
	return "p"
}

// GetConj ports PosTagHelper.getConj (v_… case code).
func GetConj(posTag string) string {
	m := conjRegex.FindStringSubmatch(posTag)
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}

// GetGenderConj ports PosTagHelper.getGenderConj ("m:v_naz" style).
func GetGenderConj(posTag string) string {
	m := genderConjRegex.FindStringSubmatch(posTag)
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}

// posMatchesFull ports Java Pattern.matcher(s).matches() / String.matches.
func posMatchesFull(re *regexp.Regexp, posTag string) bool {
	if re == nil || posTag == "" {
		return false
	}
	loc := re.FindStringIndex(posTag)
	return loc != nil && loc[0] == 0 && loc[1] == len(posTag)
}

// HasPosTagToken ports hasPosTag(AnalyzedToken, Pattern) — Matcher.matches().
func HasPosTagToken(token *languagetool.AnalyzedToken, re *regexp.Regexp) bool {
	if token == nil || re == nil {
		return false
	}
	pos := token.GetPOSTag()
	if pos == nil {
		return false
	}
	return posMatchesFull(re, *pos)
}

// HasPosTagTokenString ports hasPosTag(AnalyzedToken, String) — String.matches.
func HasPosTagTokenString(token *languagetool.AnalyzedToken, posTagRegex string) bool {
	if token == nil || posTagRegex == "" {
		return false
	}
	pos := token.GetPOSTag()
	if pos == nil {
		return false
	}
	re, err := regexp.Compile("^(?:" + posTagRegex + ")$")
	if err != nil {
		return false
	}
	return re.MatchString(*pos)
}

// HasPosTagReadings ports hasPosTag(AnalyzedTokenReadings, Pattern).
func HasPosTagReadings(atr *languagetool.AnalyzedTokenReadings, re *regexp.Regexp) bool {
	if atr == nil {
		return false
	}
	return HasPosTagTokens(atr.GetReadings(), re)
}

// HasPosTagTokens ports hasPosTag(Collection<AnalyzedToken>, Pattern).
func HasPosTagTokens(tokens []*languagetool.AnalyzedToken, re *regexp.Regexp) bool {
	for _, t := range tokens {
		if HasPosTagToken(t, re) {
			return true
		}
	}
	return false
}

// HasPosTagPartToken ports hasPosTagPart(AnalyzedToken, String).
func HasPosTagPartToken(token *languagetool.AnalyzedToken, part string) bool {
	if token == nil || part == "" {
		return false
	}
	pos := token.GetPOSTag()
	return pos != nil && strings.Contains(*pos, part)
}

// HasPosTagPartReadings ports hasPosTagPart(AnalyzedTokenReadings, String).
func HasPosTagPartReadings(atr *languagetool.AnalyzedTokenReadings, part string) bool {
	if atr == nil {
		return false
	}
	return HasPosTagPartTokens(atr.GetReadings(), part)
}

// HasPosTagPartTokens ports hasPosTagPart(List<AnalyzedToken>, String).
func HasPosTagPartTokens(tokens []*languagetool.AnalyzedToken, part string) bool {
	for _, t := range tokens {
		if HasPosTagPartToken(t, part) {
			return true
		}
	}
	return false
}

// HasPosTagPartAllReadings ports hasPosTagPartAll(AnalyzedTokenReadings, String).
func HasPosTagPartAllReadings(atr *languagetool.AnalyzedTokenReadings, part string) bool {
	if atr == nil {
		return false
	}
	return HasPosTagPartAllTokens(atr.GetReadings(), part)
}

// HasPosTagPartAllTokens ports hasPosTagPartAll(List, String) — skips SENT_END/PARAGRAPH_END.
func HasPosTagPartAllTokens(tokens []*languagetool.AnalyzedToken, part string) bool {
	foundTag := false
	for _, t := range tokens {
		if t == nil {
			continue
		}
		pos := t.GetPOSTag()
		if pos == nil {
			continue
		}
		if *pos == languagetool.SentenceEndTagName || *pos == languagetool.ParagraphEndTagName {
			continue
		}
		if !strings.Contains(*pos, part) {
			return false
		}
		foundTag = true
	}
	return foundTag
}

// HasPosTagAllTokens ports hasPosTagAll(List, Pattern).
func HasPosTagAllTokens(tokens []*languagetool.AnalyzedToken, re *regexp.Regexp) bool {
	foundTag := false
	for _, t := range tokens {
		if t == nil {
			continue
		}
		pos := t.GetPOSTag()
		if pos == nil {
			continue
		}
		if *pos == languagetool.SentenceEndTagName || *pos == languagetool.ParagraphEndTagName {
			continue
		}
		if !posMatchesFull(re, *pos) {
			return false
		}
		foundTag = true
	}
	return foundTag
}

// HasPosTagStartToken ports hasPosTagStart(AnalyzedToken, String).
func HasPosTagStartToken(token *languagetool.AnalyzedToken, prefix string) bool {
	if token == nil || prefix == "" {
		return false
	}
	pos := token.GetPOSTag()
	return pos != nil && strings.HasPrefix(*pos, prefix)
}

// HasPosTagStartReadings ports hasPosTagStart(AnalyzedTokenReadings, String).
func HasPosTagStartReadings(atr *languagetool.AnalyzedTokenReadings, prefix string) bool {
	if atr == nil {
		return false
	}
	return HasPosTagStartTokens(atr.GetReadings(), prefix)
}

// HasPosTagStartTokens ports hasPosTagStart(List, String).
func HasPosTagStartTokens(tokens []*languagetool.AnalyzedToken, prefix string) bool {
	for _, t := range tokens {
		if HasPosTagStartToken(t, prefix) {
			return true
		}
	}
	return false
}

// HasPosTagPart2 ports hasPosTagPart2(List<TaggedWord>, String).
func HasPosTagPart2(words []tagging.TaggedWord, part string) bool {
	for _, w := range words {
		if w.PosTag != "" && strings.Contains(w.PosTag, part) {
			return true
		}
	}
	return false
}

// HasPosTag2 ports hasPosTag2(List<TaggedWord>, Pattern).
func HasPosTag2(words []tagging.TaggedWord, re *regexp.Regexp) bool {
	for _, w := range words {
		if w.PosTag != "" && posMatchesFull(re, w.PosTag) {
			return true
		}
	}
	return false
}

// HasPosTagStart2 ports hasPosTagStart2(List<TaggedWord>, String).
func HasPosTagStart2(words []tagging.TaggedWord, prefix string) bool {
	for _, w := range words {
		if w.PosTag != "" && strings.HasPrefix(w.PosTag, prefix) {
			return true
		}
	}
	return false
}

// GetGenders ports PosTagHelper.getGenders (concat unique gender chars for matching tags).
func GetGenders(atr *languagetool.AnalyzedTokenReadings, re *regexp.Regexp) string {
	if atr == nil || re == nil {
		return ""
	}
	var sb strings.Builder
	for _, t := range atr.GetReadings() {
		if t == nil {
			continue
		}
		pos := t.GetPOSTag()
		if pos == nil || !posMatchesFull(re, *pos) {
			continue
		}
		g := GetGender(*pos)
		if g != "" && !strings.Contains(sb.String(), g) {
			sb.WriteString(g)
		}
	}
	return sb.String()
}

// GetGendersString ports getGenders(…, String) compiling the pattern as full-match.
func GetGendersString(atr *languagetool.AnalyzedTokenReadings, posTagRegex string) string {
	if posTagRegex == "" {
		return ""
	}
	re, err := regexp.Compile("^(?:" + posTagRegex + ")$")
	if err != nil {
		return ""
	}
	return GetGenders(atr, re)
}

// GenerateTokensForNv ports PosTagHelper.generateTokensForNv.
// Lemma = surface; skips v_kly; appends :nv + optional extraTags.
func GenerateTokensForNv(word, genders, extraTags string) []*languagetool.AnalyzedToken {
	var out []*languagetool.AnalyzedToken
	for _, gen := range genders {
		for _, vidm := range VidminkyOrder {
			if vidm == "v_kly" {
				continue
			}
			pos := "noun:inanim:" + string(gen) + ":" + vidm + NoVidminokSubstr
			if extraTags != "" {
				pos += extraTags
			}
			p, l := pos, word
			out = append(out, languagetool.NewAnalyzedToken(word, &p, &l))
		}
	}
	return out
}

// AddIfNotContains ports addIfNotContains(String, String).
func AddIfNotContains(tag, addTag string) string {
	if addTag != "" && !strings.Contains(tag, addTag) {
		return tag + addTag
	}
	return tag
}

// AddIfNotContainsMany ports addIfNotContains(String, String...).
func AddIfNotContainsMany(tag string, addTags ...string) string {
	for _, a := range addTags {
		if a != "" && !strings.Contains(tag, a) {
			tag += a
		}
	}
	return tag
}

// AddIfNotContainsWords ports addIfNotContains(List<TaggedWord>, addTag, lemma?).
// lemma "" keeps each word's lemma.
func AddIfNotContainsWords(words []tagging.TaggedWord, addTag, lemma string) []tagging.TaggedWord {
	out := make([]tagging.TaggedWord, 0, len(words))
	for _, w := range words {
		l := w.Lemma
		if lemma != "" {
			l = lemma
		}
		out = append(out, tagging.NewTaggedWord(l, AddIfNotContains(w.PosTag, addTag)))
	}
	return out
}

// Adjust ports PosTagHelper.adjust (lemma prefix/suffix + cleanExtraTags + addTags).
func Adjust(words []tagging.TaggedWord, lemmaPrefix, lemmaSuffix string, addTags ...string) []tagging.TaggedWord {
	out := make([]tagging.TaggedWord, 0, len(words))
	for _, w := range words {
		lemma := adjustLemma(w.Lemma, lemmaPrefix, lemmaSuffix)
		tag := cleanExtraTags(w.PosTag)
		tag = AddIfNotContainsMany(tag, addTags...)
		out = append(out, tagging.NewTaggedWord(lemma, tag))
	}
	return out
}

func adjustLemma(lemma, prefix, suffix string) string {
	if prefix != "" {
		lemma = prefix + lemma
	}
	if suffix != "" {
		lemma += suffix
	}
	return lemma
}

func cleanExtraTags(tag string) string {
	if tag == "" {
		return tag
	}
	return cleanupPattern.ReplaceAllString(tag, "")
}

// FilterTokens ports filter(List<AnalyzedToken>, Pattern).
func FilterTokens(tokens []*languagetool.AnalyzedToken, re *regexp.Regexp) []*languagetool.AnalyzedToken {
	var out []*languagetool.AnalyzedToken
	for _, t := range tokens {
		if HasPosTagToken(t, re) {
			out = append(out, t)
		}
	}
	return out
}

// Filter2 ports filter2(List<TaggedWord>, Pattern).
func Filter2(words []tagging.TaggedWord, re *regexp.Regexp) []tagging.TaggedWord {
	var out []tagging.TaggedWord
	for _, w := range words {
		if w.PosTag != "" && posMatchesFull(re, w.PosTag) {
			out = append(out, w)
		}
	}
	return out
}

// Filter2Negative ports filter2Negative.
func Filter2Negative(words []tagging.TaggedWord, re *regexp.Regexp) []tagging.TaggedWord {
	var out []tagging.TaggedWord
	for _, w := range words {
		if w.PosTag == "" || !posMatchesFull(re, w.PosTag) {
			out = append(out, w)
		}
	}
	return out
}

// IsUnknownWord ports PosTagHelper.isUnknownWord.
func IsUnknownWord(atr *languagetool.AnalyzedTokenReadings) bool {
	if atr == nil {
		return false
	}
	// getAnalyzedToken(0).hasNoTag()
	readings := atr.GetReadings()
	if len(readings) == 0 || readings[0] == nil || !readings[0].HasNoTag() {
		return false
	}
	return wordPattern.MatchString(atr.GetToken())
}

// IsPredictOrInsert ports PosTagHelper.isPredictOrInsert.
func IsPredictOrInsert(token *languagetool.AnalyzedToken) bool {
	if token == nil {
		return false
	}
	pos := token.GetPOSTag()
	if pos == nil {
		return false
	}
	return posMatchesFull(predictInsertPattern, *pos)
}

// FilterByPosAndToken ports filter(AnalyzedTokenReadings, postag, token) with Matcher.matches().
func FilterByPosAndToken(atr *languagetool.AnalyzedTokenReadings, postag, tokenRE *regexp.Regexp) []*languagetool.AnalyzedToken {
	if atr == nil || postag == nil || tokenRE == nil {
		return nil
	}
	var out []*languagetool.AnalyzedToken
	for _, t := range atr.GetReadings() {
		if t == nil {
			continue
		}
		pos := t.GetPOSTag()
		if pos == nil || !posMatchesFull(postag, *pos) {
			continue
		}
		tok := t.GetToken()
		if tok == "" || !posMatchesFull(tokenRE, tok) {
			continue
		}
		out = append(out, t)
	}
	return out
}

// HasPosTagAndToken ports hasPosTagAndToken.
func HasPosTagAndToken(atr *languagetool.AnalyzedTokenReadings, postag, tokenRE *regexp.Regexp) bool {
	return len(FilterByPosAndToken(atr, postag, tokenRE)) > 0
}

// HasMaleUA ports PosTagHelper.hasMaleUA.
// Java: hasPosTagAndToken(…, "noun:inanim:m:v_dav(?!:nv).*", ".*[ую]")
// RE2 has no lookaround: accept v_dav tags that do not contain :nv after the case slot.
func HasMaleUA(atr *languagetool.AnalyzedTokenReadings) bool {
	if atr == nil {
		return false
	}
	for _, t := range atr.GetReadings() {
		if t == nil {
			continue
		}
		pos := t.GetPOSTag()
		if pos == nil {
			continue
		}
		// noun:inanim:m:v_dav… without :nv (Java (?!:nv) after v_dav)
		if !strings.HasPrefix(*pos, "noun:inanim:m:v_dav") {
			continue
		}
		rest := (*pos)[len("noun:inanim:m:v_dav"):]
		if strings.HasPrefix(rest, ":nv") {
			continue
		}
		tok := t.GetToken()
		if tok == "" {
			tok = atr.GetToken()
		}
		if posMatchesFull(maleUATokenRE, tok) {
			return true
		}
	}
	return false
}
