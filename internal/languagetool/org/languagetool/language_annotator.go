package languagetool

import (
	"regexp"
	"strings"
	"unicode"
)

const languageAnnotatorMinTokens = 4

var (
	langAnnotatorBoundaryRE = regexp.MustCompile(`^[.?!;:"„“»«()\[\]\n]$`)
	langAnnotatorQuoteRE    = regexp.MustCompile(`^["„“”»«]$`)
	langAnnotatorWordRE     = regexp.MustCompile(`^\w+$`)
)

// LanguageAnnotator ports org.languagetool.LanguageAnnotator with pluggable
// tokenization and word-validity (VagueSpellChecker deferred).
type LanguageAnnotator struct {
	// Tokenize splits input into tokens (defaults to whitespace+punctuation-ish split).
	Tokenize func(input string) []string
	// IsValidWord reports whether token is valid in languageCode.
	IsValidWord func(token, languageCode string) bool
}

func NewLanguageAnnotator() *LanguageAnnotator {
	return &LanguageAnnotator{}
}

func (a *LanguageAnnotator) tokenize(input string) []string {
	if a != nil && a.Tokenize != nil {
		return a.Tokenize(input)
	}
	return defaultWordTokenize(input)
}

func (a *LanguageAnnotator) isValid(token, lang string) bool {
	if a != nil && a.IsValidWord != nil {
		return a.IsValidWord(token, lang)
	}
	return false
}

// DetectLanguages ports LanguageAnnotator.detectLanguages.
// Language codes replace Language objects (short code with optional variant).
func (a *LanguageAnnotator) DetectLanguages(input, mainLang string, secondLangs []string) []FragmentWithLanguage {
	tokens := a.getTokensWithPotentialLanguages(input, mainLang, secondLangs)
	ranges := a.getTokenRanges(tokens)
	withLang := a.getTokenRangesWithLang(ranges, mainLang, secondLangs)

	// Java: walk token ranges, splice on language change using input substrings.
	curPos := 0
	fromPos := 0
	prevLang := mainLang
	var result []FragmentWithLanguage
	for _, tr := range withLang {
		curLang := tr.lang
		if curLang == "" {
			curLang = mainLang
		}
		// single quote inherits previous language
		if len(tr.tokens) == 1 && isAnnotatorQuote(tr.tokens[0]) {
			curLang = prevLang
		} else if curLang != prevLang {
			// FragmentWithLanguage panics on empty fragment — skip empty slices
			if fromPos < curPos && fromPos >= 0 && curPos <= len(input) {
				frag := input[fromPos:curPos]
				if frag != "" {
					result = append(result, FragmentWithLanguage{LangCode: prevLang, Fragment: frag})
				}
			}
			fromPos = curPos
		}
		prevLang = curLang
		for _, tok := range tr.tokens {
			curPos += len(tok)
		}
	}
	if fromPos < len(input) {
		frag := input[fromPos:]
		if frag != "" {
			result = append(result, FragmentWithLanguage{LangCode: prevLang, Fragment: frag})
		}
	}
	return result
}

// TokenWithLanguages is the annotator intermediate.
type TokenWithLanguages struct {
	Token string
	Langs []string
}

func (t TokenWithLanguages) Ambiguous() bool { return len(t.Langs) != 1 }

// GetTokensWithPotentialLanguages is the public entry (Java package-private tests).
func (a *LanguageAnnotator) GetTokensWithPotentialLanguages(input, mainLang string, secondLangs []string) []TokenWithLanguages {
	return a.getTokensWithPotentialLanguages(input, mainLang, secondLangs)
}

// GetTokenRanges is the public entry for sentence-like token range splitting.
func (a *LanguageAnnotator) GetTokenRanges(tokens []TokenWithLanguages) [][]TokenWithLanguages {
	return a.getTokenRanges(tokens)
}

// TokenRangeWithLang is a token span labeled with a language code.
type TokenRangeWithLang struct {
	Tokens []string
	Lang   string
}

// GetTokenRangesWithLang assigns a language to each token range.
func (a *LanguageAnnotator) GetTokenRangesWithLang(tokenRanges [][]TokenWithLanguages, mainLang string, secondLangs []string) []TokenRangeWithLang {
	raw := a.getTokenRangesWithLang(tokenRanges, mainLang, secondLangs)
	out := make([]TokenRangeWithLang, len(raw))
	for i, r := range raw {
		out[i] = TokenRangeWithLang{Tokens: r.tokens, Lang: r.lang}
	}
	return out
}

// TokenRangeString formats ranges like Java getTokenRangeAsString.
func TokenRangeString(ranges [][]TokenWithLanguages) string {
	var b strings.Builder
	for _, r := range ranges {
		b.WriteByte('[')
		for i, t := range r {
			if i > 0 && !isAnnotatorBoundary(t) && t.Token != " " {
				// join without adding spaces — tokens include spaces when present
			}
			b.WriteString(t.Token)
		}
		b.WriteByte(']')
	}
	return b.String()
}

func (a *LanguageAnnotator) getTokensWithPotentialLanguages(input, mainLang string, secondLangs []string) []TokenWithLanguages {
	raw := a.tokenize(input)
	tokens := make([]TokenWithLanguages, 0, len(raw))
	for _, token := range raw {
		if isAnnotatorWord(token) && a.isValid(token, mainLang) {
			tokens = append(tokens, TokenWithLanguages{Token: token, Langs: []string{mainLang}})
		} else {
			tokens = append(tokens, TokenWithLanguages{Token: token})
		}
	}
	for _, second := range secondLangs {
		for i, token := range tokens {
			if isAnnotatorWord(token.Token) && a.isValid(token.Token, second) {
				langs := append(append([]string{}, token.Langs...), second)
				tokens[i] = TokenWithLanguages{Token: token.Token, Langs: langs}
			}
		}
	}
	return tokens
}

func (a *LanguageAnnotator) getTokenRanges(tokens []TokenWithLanguages) [][]TokenWithLanguages {
	var result [][]TokenWithLanguages
	var l []TokenWithLanguages
	inQuote := false
	for _, token := range tokens {
		if isAnnotatorQuote(token.Token) && !inQuote {
			if len(l) > 0 {
				result = append(result, l)
			}
			l = []TokenWithLanguages{token}
		} else if isAnnotatorBoundary(token) {
			l = append(l, token)
			result = append(result, l)
			l = nil
		} else {
			l = append(l, token)
		}
		if isAnnotatorQuote(token.Token) {
			inQuote = !inQuote
		}
	}
	if len(l) > 0 {
		result = append(result, l)
	}
	return result
}

type tokenRangeWithLanguage struct {
	tokens []string
	lang   string
}

func (a *LanguageAnnotator) getTokenRangesWithLang(tokenRanges [][]TokenWithLanguages, mainLang string, secondLangs []string) []tokenRangeWithLanguage {
	var result []tokenRangeWithLanguage
	var prevTopLang string
	for i, tokens := range tokenRanges {
		var topLang string
		allAmbiguous := true
		for _, k := range tokens {
			if !k.Ambiguous() {
				allAmbiguous = false
				break
			}
		}
		if allAmbiguous {
			for j := i + 1; j < len(tokenRanges); j++ {
				nextAllAmb := true
				for _, k := range tokenRanges[j] {
					if !k.Ambiguous() {
						nextAllAmb = false
						break
					}
				}
				if !nextAllAmb {
					topLang = a.getTopLang(mainLang, secondLangs, tokenRanges[j])
					break
				}
			}
		}
		if topLang == "" {
			if len(tokens) < languageAnnotatorMinTokens && prevTopLang != "" {
				topLang = prevTopLang
			} else {
				topLang = a.getTopLang(mainLang, secondLangs, tokens)
			}
		}
		tokenList := make([]string, len(tokens))
		for ti, k := range tokens {
			tokenList[ti] = k.Token
		}
		result = append(result, tokenRangeWithLanguage{tokens: tokenList, lang: topLang})
		prevTopLang = topLang
	}
	return result
}

func (a *LanguageAnnotator) getTopLang(mainLang string, secondLangs []string, tokens []TokenWithLanguages) string {
	counts := map[string]int{}
	for _, t := range tokens {
		for _, lang := range t.Langs {
			counts[lang]++
		}
	}
	// ensure main is considered
	_ = counts[mainLang]
	max := 0
	top := mainLang
	// check main first then seconds then all
	if c := counts[mainLang]; c > max {
		max = c
		top = mainLang
	}
	for _, lang := range secondLangs {
		if c := counts[lang]; c > max {
			max = c
			top = lang
		}
	}
	for lang, c := range counts {
		if c > max {
			max = c
			top = lang
		}
	}
	return top
}

func isAnnotatorBoundary(token TokenWithLanguages) bool {
	return langAnnotatorBoundaryRE.MatchString(token.Token)
}

func isAnnotatorQuote(token string) bool {
	return langAnnotatorQuoteRE.MatchString(token)
}

func isAnnotatorWord(s string) bool {
	// Java: !s.trim().isEmpty() && s.matches("\\w+")
	return javaTrim(s) != "" && langAnnotatorWordRE.MatchString(s)
}

// defaultWordTokenize keeps whitespace and punctuation as separate tokens (simple).
func defaultWordTokenize(input string) []string {
	var out []string
	var b strings.Builder
	flush := func() {
		if b.Len() > 0 {
			out = append(out, b.String())
			b.Reset()
		}
	}
	for _, r := range input {
		if unicode.IsSpace(r) || strings.ContainsRune(`.,?!;:"„“»«()[]`+"\n", r) {
			flush()
			out = append(out, string(r))
			continue
		}
		b.WriteRune(r)
	}
	flush()
	return out
}
