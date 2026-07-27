package pt

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// PortugueseWordTokenizer ports org.languagetool.tokenizers.pt.PortugueseWordTokenizer.
// Java field: private final PortugueseTagger tagger = new PortugueseTagger().
type PortugueseWordTokenizer struct {
	// tagger is Java PortugueseTagger used in wordsToAdd for hyphen compounds.
	// Nil → fail closed (split hyphens) unless IsTaggedPT package hook is set.
	tagger PortugueseHyphenTagger
}

// PortugueseHyphenTagger is the subset of PortugueseTagger used by wordsToAdd
// (Java: tagger.tag(...).get(0).isTagged()).
// Implementations typically wrap *tagging/pt.PortugueseTagger.Tag.
type PortugueseHyphenTagger interface {
	Tag(tokens []string) []PortugueseHyphenReadings
}

// PortugueseHyphenReadings is the subset of AnalyzedTokenReadings consulted
// for hyphen compound keep/split. *languagetool.AnalyzedTokenReadings implements this.
type PortugueseHyphenReadings interface {
	IsTagged() bool
}

// ATRTagFunc adapts a batch-tag function to PortugueseHyphenTagger.
type ATRTagFunc func(tokens []string) []PortugueseHyphenReadings

// Tag implements PortugueseHyphenTagger.
func (f ATRTagFunc) Tag(tokens []string) []PortugueseHyphenReadings {
	if f == nil {
		return nil
	}
	return f(tokens)
}

// portugueseTaggerLoader is registered by tagging/pt (import-cycle fence) so
// NewPortugueseWordTokenizer can construct like Java new PortugueseTagger().
var portugueseTaggerLoader func() PortugueseHyphenTagger

// RegisterPortugueseTaggerLoader registers the Java-equivalent of
// "new PortugueseTagger()" for PortugueseWordTokenizer construction.
// Called from tagging/pt when official portuguese.dict loading is available.
func RegisterPortugueseTaggerLoader(f func() PortugueseHyphenTagger) {
	portugueseTaggerLoader = f
}

// IsTaggedPT optional package-level PortugueseTagger.tag(...).isTagged() hook.
// Prefer the instance field set by NewPortugueseWordTokenizer (Java).
// Kept as nil-safe fallback for older inject sites (speller / binary POS wire).
// Without tagger field and without this hook, hyphen compounds are split.
var IsTaggedPT func(s string) bool

// NewPortugueseWordTokenizer ports Java PortugueseWordTokenizer() which does
// tagger = new PortugueseTagger(). When tagging/pt has registered a loader and
// official portuguese.dict is available, the real tagger is wired automatically.
func NewPortugueseWordTokenizer() *PortugueseWordTokenizer {
	w := &PortugueseWordTokenizer{}
	if portugueseTaggerLoader != nil {
		w.tagger = portugueseTaggerLoader()
	}
	return w
}

// NewPortugueseWordTokenizerWithTagger builds a tokenizer with an explicit tagger
// (tests / callers that already hold a PortugueseTagger).
func NewPortugueseWordTokenizerWithTagger(t PortugueseHyphenTagger) *PortugueseWordTokenizer {
	return &PortugueseWordTokenizer{tagger: t}
}

// SetTagger sets the hyphen-compound tagger (optional; Java has final field set in ctor).
func (w *PortugueseWordTokenizer) SetTagger(t PortugueseHyphenTagger) {
	if w == nil {
		return
	}
	w.tagger = t
}

const (
	decimalCommaSubst     = '\uE001'
	nonBreakingSpaceSubst = '\uE002'
	nonBreakingDotSubst   = '\uE003'
	nonBreakingColonSubst = '\uE004'
	// Java HYPHEN_SUBST_TEXT; also concatenated into wordChars via Pattern.toString().
	hyphenSubstText = "\u0001\u0001PT_HYPHEN\u0001\u0001"
)

var (
	decimalCommaPattern  = regexp.MustCompile(`(?i)([\d]),([\d])`)
	dottedNumbersPattern = regexp.MustCompile(`(?i)([\d])\.([\d])`)
	colonNumbersPattern  = regexp.MustCompile(`(?i)([\d]):([\d])`)
	// Java DATE_PATTERN — replacement uses only $1$2$3 (alts 2–3 empty groups; bug-for-bug).
	datePattern    = regexp.MustCompile(`(?i)([\d]{2})\.([\d]{2})\.([\d]{4})|([\d]{4})\.([\d]{2})\.([\d]{2})|([\d]{4})-([\d]{2})-([\d]{2})`)
	dottedOrdinals = regexp.MustCompile(`(?i)([\d])\.([aoªºᵃᵒ][sˢ]?)`)
	hyphenPattern  = regexp.MustCompile(`(?i)([\p{L}])-([\p{L}\d])`)
	nearbyHyphens  = regexp.MustCompile(`(?i)([\p{L}])-([\p{L}])-([\p{L}])`)
	// Java DECIMAL_SPACE_PATTERN body (lookaround applied in protectPTSpacedNumbers).
	decimalSpaceBody = regexp.MustCompile(`\d{1,3}(?: \d{3})+(?:[` + string(decimalCommaSubst) + string(nonBreakingDotSubst) + `]\d+)?`)

	// Java wordChars includes HYPHEN_SUBST (Pattern → HYPHEN_SUBST_TEXT chars in the class).
	wordChars          = "°\\^\\-\\p{L}\\d\\x{0300}-\\x{036F}\\x{00A8}\\x{2070}-\\x{209F}" + string(decimalCommaSubst) + string(nonBreakingDotSubst) + string(nonBreakingColonSubst) + string(nonBreakingSpaceSubst) + hyphenSubstText
	wordCharsLeftEdge  = `−@€£\$¢¥¤`
	wordCharsRightEdge = `€£\$%‰‱ºªᵃᵒˢ`
	wordPattern        = regexp.MustCompile(`(?i)[` + wordCharsLeftEdge + `]?[` + wordChars + `]+[` + wordCharsRightEdge + `]?|[^` + wordChars + `]`)

	// Java wordsToAdd camel-case hyphen exceptions only (not invent soft compounds).
	javaHyphenExceptions = map[string]bool{
		"mers-cov": true, "mcgraw-hill": true, "sars-cov-2": true, "sars-cov": true,
		"ph-metre": true, "ph-metres": true, "anti-ivg": true, "anti-uv": true,
		"anti-vih": true, "al-qaïda": true,
	}
)

func (w *PortugueseWordTokenizer) Tokenize(text string) []string {
	tokenisedText := text
	if strings.Contains(tokenisedText, ",") {
		tokenisedText = decimalCommaPattern.ReplaceAllString(tokenisedText, "$1"+string(decimalCommaSubst)+"$2")
	}
	// if period is not the last character in the sentence
	dotIndex := strings.IndexByte(tokenisedText, '.')
	if dotIndex >= 0 && dotIndex < len(tokenisedText)-1 {
		// Java DATE_PATTERN_REPL = "$1"+DOT+"$2"+DOT+"$3" (only first alt groups).
		tokenisedText = datePattern.ReplaceAllString(tokenisedText, "$1"+string(nonBreakingDotSubst)+"$2"+string(nonBreakingDotSubst)+"$3")
		tokenisedText = dottedNumbersPattern.ReplaceAllString(tokenisedText, "$1"+string(nonBreakingDotSubst)+"$2")
		tokenisedText = dottedOrdinals.ReplaceAllString(tokenisedText, "$1"+string(nonBreakingDotSubst)+"$2")
	}
	tokenisedText = protectPTSpacedNumbers(tokenisedText)
	if strings.Contains(tokenisedText, ":") {
		tokenisedText = colonNumbersPattern.ReplaceAllString(tokenisedText, "$1"+string(nonBreakingColonSubst)+"$2")
	}
	if strings.Contains(tokenisedText, "-") {
		tokenisedText = nearbyHyphens.ReplaceAllString(tokenisedText, "$1"+hyphenSubstText+"$2"+hyphenSubstText+"$3")
		tokenisedText = hyphenPattern.ReplaceAllString(tokenisedText, "$1"+hyphenSubstText+"$2")
	}

	var tokenList []string
	for _, loc := range wordPattern.FindAllStringIndex(tokenisedText, -1) {
		token := tokenisedText[loc[0]:loc[1]]
		// 0xFE00-0xFE0F are non-spacing marks (Java token.length()==1 → one UTF-16 code unit / BMP).
		if len(tokenList) > 0 {
			r, size := utf8.DecodeRuneInString(token)
			if size == len(token) && r >= 0xFE00 && r <= 0xFE0F {
				tokenList[len(tokenList)-1] = tokenList[len(tokenList)-1] + token
				continue
			}
		}
		token = strings.ReplaceAll(token, string(decimalCommaSubst), ",")
		token = strings.ReplaceAll(token, string(nonBreakingColonSubst), ":")
		token = strings.ReplaceAll(token, string(nonBreakingSpaceSubst), " ")
		// outside of if as we also replace back sentence-ending abbreviations
		token = strings.ReplaceAll(token, string(nonBreakingDotSubst), ".")
		token = strings.ReplaceAll(token, hyphenSubstText, "-")
		tokenList = append(tokenList, w.wordsToAdd(token)...)
	}
	return tokenizers.JoinEMailsAndUrls(tokenList)
}

// protectPTSpacedNumbers ports Java DECIMAL_SPACE_PATTERN + appendReplacement loop.
// Java: (?<=^|[\s(])\d{1,3}( \d{3})+(?:[DECIMAL_COMMA_SUBST NON_BREAKING_DOT_SUBST]\d+)?(?=\D|$)
// RE2 has no lookaround; boundaries checked manually. Java \s/\D without UNICODE_CHARACTER_CLASS = ASCII.
func protectPTSpacedNumbers(text string) string {
	if !decimalSpaceBody.MatchString(text) {
		return text
	}
	var b strings.Builder
	i := 0
	for i < len(text) {
		loc := decimalSpaceBody.FindStringIndex(text[i:])
		if loc == nil {
			b.WriteString(text[i:])
			break
		}
		start := i + loc[0]
		end := i + loc[1]
		okLeft := start == 0
		if !okLeft {
			// Java (?<=^|[\s(]) — ASCII \s = [ \t\n\x0B\f\r]
			prev, _ := utf8.DecodeLastRuneInString(text[:start])
			okLeft = prev == '(' || prev == ' ' || prev == '\t' || prev == '\n' ||
				prev == '\v' || prev == '\f' || prev == '\r'
		}
		okRight := end == len(text)
		if !okRight {
			// Java (?=\D|$) — \D = not ASCII digit
			next := text[end]
			okRight = next < '0' || next > '9'
		}
		if okLeft && okRight {
			b.WriteString(text[i:start])
			splitNumber := text[start:end]
			splitNumber = strings.ReplaceAll(splitNumber, " ", string(nonBreakingSpaceSubst))
			splitNumber = strings.ReplaceAll(splitNumber, "\u00A0", string(nonBreakingSpaceSubst))
			b.WriteString(splitNumber)
			i = end
			continue
		}
		// Not a valid boundary match: emit up to start and resume one byte later
		// so a later offset can still match (same as Java find-from-progress).
		if start == i {
			b.WriteByte(text[i])
			i++
		} else {
			b.WriteString(text[i:start])
			i = start
		}
	}
	return b.String()
}

// wordsToAdd ports PortugueseWordTokenizer.wordsToAdd.
func (w *PortugueseWordTokenizer) wordsToAdd(s string) []string {
	var l []string
	if s == "" {
		return l
	}
	if tokenizers.IsCurrencyExpression(s) {
		return tokenizers.SplitCurrencyExpression(s)
	}
	if !strings.Contains(s, "-") {
		l = append(l, s)
		return l
	}
	// Java: tagger.tag(Arrays.asList(s.replace("’", "'"))).get(0).isTagged()
	// In the current POS tag, most apostrophes are curly: to be fixed
	normalized := strings.ReplaceAll(s, "’", "'")
	if w.isTagged(normalized) {
		l = append(l, s)
		return l
	}
	// some camel-case words containing hyphen (is there any better fix?)
	if javaHyphenException(s) {
		l = append(l, s)
		return l
	}
	// if not found, the word is split on hyphens (keep separators)
	// Java: StringTokenizer(s, "-", true)
	var cur strings.Builder
	for _, r := range s {
		if r == '-' {
			if cur.Len() > 0 {
				l = append(l, cur.String())
				cur.Reset()
			}
			l = append(l, "-")
		} else {
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		l = append(l, cur.String())
	}
	return l
}

func javaHyphenException(s string) bool {
	return javaHyphenExceptions[strings.ToLower(s)]
}

// isTagged ports Java tagger.tag(...).get(0).isTagged().
// Prefer instance tagger (Java field); package IsTaggedPT is nil-safe fallback only.
func (w *PortugueseWordTokenizer) isTagged(s string) bool {
	if w != nil && w.tagger != nil {
		atrs := w.tagger.Tag([]string{s})
		if len(atrs) == 0 || atrs[0] == nil {
			return false
		}
		return atrs[0].IsTagged()
	}
	if IsTaggedPT != nil {
		return IsTaggedPT(s)
	}
	return false
}
