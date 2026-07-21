package hunspell

import (
	"bufio"
	"io"
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// whitespaceArray ports HunspellRule.WHITESPACE_ARRAY (lengths 0..19).
var whitespaceArray [20]string

func init() {
	for i := 0; i < 20; i++ {
		whitespaceArray[i] = strings.Repeat(" ", i)
	}
}

// NonWordSplitter ports HunspellRule.nonWordPattern.split behavior.
// Java uses Pattern.compile("(?![WORDCHARS])[^\\p{L}]") — RE2 has no lookahead,
// so we implement equivalent: split on runes that are not letters and not in WordChars.
type NonWordSplitter struct {
	// WordChars are extra characters treated as part of words (from .aff WORDCHARS).
	// Empty means only Unicode letters are word characters (NON_ALPHABETIC default).
	WordChars map[rune]struct{}
}

// DefaultNonWordSplitter is the NON_ALPHABETIC default (letters only).
var DefaultNonWordSplitter = NonWordSplitter{}

// isWordChar reports whether r is kept as part of a token (not a split point).
func (s NonWordSplitter) isWordChar(r rune) bool {
	if unicode.IsLetter(r) {
		return true
	}
	if len(s.WordChars) == 0 {
		return false
	}
	_, ok := s.WordChars[r]
	return ok
}

// Split ports Pattern.split on the non-word pattern: consecutive non-word-chars
// are separators. Trailing empty segments are dropped (Java Pattern.split limit 0).
func (s NonWordSplitter) Split(text string) []string {
	if text == "" {
		return nil
	}
	var out []string
	var b strings.Builder
	flush := func() {
		// Always append (including empty for consecutive separators), then trim trailing.
		out = append(out, b.String())
		b.Reset()
	}
	for _, r := range text {
		if s.isWordChar(r) {
			b.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	// Drop trailing empties (Java default split).
	for len(out) > 0 && out[len(out)-1] == "" {
		out = out[:len(out)-1]
	}
	return out
}

// ComputeNonWordSplitter ports HunspellRule.computeNonWordPattern from an .aff reader.
// Reads WORDCHARS line; if absent, returns DefaultNonWordSplitter.
func ComputeNonWordSplitter(r io.Reader) NonWordSplitter {
	if r == nil {
		return DefaultNonWordSplitter
	}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "WORDCHARS ") {
			wordCharsFromAff := line[len("WORDCHARS "):]
			m := make(map[rune]struct{}, len([]rune(wordCharsFromAff)))
			for _, ch := range wordCharsFromAff {
				// Java char class includes each code unit; Go uses runes.
				m[ch] = struct{}{}
			}
			return NonWordSplitter{WordChars: m}
		}
	}
	return DefaultNonWordSplitter
}

// ComputeNonWordSplitterFromString is a convenience for tests / loaded aff text.
func ComputeNonWordSplitterFromString(affContent string) NonWordSplitter {
	return ComputeNonWordSplitter(strings.NewReader(affContent))
}

// ComputeNonWordPattern is a deprecated name retained for callers; returns a
// NonWordSplitter (Java returned Pattern — behavior twin is Split, not regexp).
// Prefer ComputeNonWordSplitter.
func ComputeNonWordPattern(r io.Reader) NonWordSplitter {
	return ComputeNonWordSplitter(r)
}

// ComputeNonWordPatternFromString prefers ComputeNonWordSplitterFromString.
func ComputeNonWordPatternFromString(affContent string) NonWordSplitter {
	return ComputeNonWordSplitterFromString(affContent)
}

// GetDictFilenameInResources ports HunspellRule.getDictFilenameInResources:
// "/" + language.shortCode + "/hunspell/" + langCountry + ".dic"
// langCountry is typically shortCode or shortCode_COUNTRY (e.g. de_DE, da_DK).
func GetDictFilenameInResources(shortCode, langCountry string) string {
	if shortCode == "" {
		shortCode = langCountry
	}
	if langCountry == "" {
		langCountry = shortCode
	}
	return "/" + shortCode + "/hunspell/" + langCountry + FileExtension
}

// GetDictFilenameInResourcesFromLangCode ports getDictFilenameInResources when
// only a language code (possibly with region) is known.
// "de-DE" → shortCode de, langCountry de_DE; "da" → da / da.
func GetDictFilenameInResourcesFromLangCode(langCode string) string {
	c := strings.TrimSpace(langCode)
	if c == "" {
		return ""
	}
	// Normalize en-US / en_US
	c = strings.ReplaceAll(c, "-", "_")
	parts := strings.SplitN(c, "_", 2)
	short := strings.ToLower(parts[0])
	langCountry := short
	if len(parts) == 2 && parts[1] != "" {
		langCountry = short + "_" + strings.ToUpper(parts[1])
	}
	return GetDictFilenameInResources(short, langCountry)
}

// IsQuotedCompound ports HunspellRule.isQuotedCompound — base always false;
// German overrides for quoted compounds. Override via IsQuotedCompoundFn.
func (r *HunspellRule) IsQuotedCompound(sentence *languagetool.AnalyzedSentence, idx int, token string) bool {
	if r != nil && r.IsQuotedCompoundFn != nil {
		return r.IsQuotedCompoundFn(sentence, idx, token)
	}
	return false
}

// TokenizeText ports HunspellRule.tokenizeText — nonWordPattern.split(sentence).
func (r *HunspellRule) TokenizeText(sentence string) []string {
	spl := DefaultNonWordSplitter
	if r != nil {
		spl = r.NonWordSplitter
	}
	return spl.Split(sentence)
}

// GetSentenceTextWithoutUrlsAndImmunizedTokens ports
// HunspellRule.getSentenceTextWithoutUrlsAndImmunizedTokens.
// Builds a string where immunized / URL / email / _english_ignore_ tokens are
// replaced with spaces (UTF-16 length), and other tokens use stringForSpeller.
// Quoted compounds (isQuotedCompound): space + token without first char.
func (r *HunspellRule) GetSentenceTextWithoutUrlsAndImmunizedTokens(sentence *languagetool.AnalyzedSentence) string {
	if sentence == nil {
		return ""
	}
	work := sentence
	if r != nil && r.SpellingCheckRule != nil {
		work = r.SpellingCheckRule.SentenceWithImmunization(sentence)
	}
	tokens := work.GetTokens() // full tokens including whitespace (Java getTokens)
	if len(tokens) <= 1 {
		return ""
	}
	var sb strings.Builder
	// Java: for (int i = 1; i < sentenceTokens.length; i++) — skip SENT_START
	for i := 1; i < len(tokens); i++ {
		tok := tokens[i]
		if tok == nil {
			continue
		}
		token := tok.GetToken()
		quoted := r.IsQuotedCompound(sentence, i, token)
		if tok.IsImmunized() || tok.IsIgnoredBySpeller() ||
			spelling.IsUrl(token) || spelling.IsEMail(token) ||
			quoted ||
			tok.HasPosTag("_english_ignore_") {
			if quoted {
				// sb.append(' ').append(token.substring(1));
				rest := token
				if u := utf16.Encode([]rune(token)); len(u) > 0 {
					rest = string(utf16.Decode(u[1:]))
				}
				sb.WriteByte(' ')
				sb.WriteString(rest)
			} else {
				// replace with spaces of UTF-16 length
				n := utf16LenHun(token)
				if n < 20 {
					sb.WriteString(whitespaceArray[n])
				} else {
					sb.WriteString(strings.Repeat(" ", n))
				}
			}
		} else {
			sb.WriteString(tools.StringForSpeller(token))
		}
	}
	return sb.String()
}

// SetNonWordSplitterFromAff parses .aff content and sets NonWordSplitter (Java init).
func (r *HunspellRule) SetNonWordSplitterFromAff(affContent string) {
	if r == nil {
		return
	}
	r.NonWordSplitter = ComputeNonWordSplitterFromString(affContent)
}
