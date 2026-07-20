package chunking

import (
	"path/filepath"
	"regexp"
	"unicode"
	"unicode/utf8"
)

// TokenizerME ports opennlp.tools.tokenize.TokenizerME for OpenNLP 1.5 en-token.bin.
// Outcomes: "T" = split, "F" = no split. useAlphaNumericOptimization=true (manifest).
type TokenizerME struct {
	model                       *GISModel
	useAlphaNumericOptimization bool
	alphanumeric                *regexp.Regexp
}

// NewTokenizerME loads an OpenNLP tokenizer model zip (token.model).
func NewTokenizerME(modelPath string) (*TokenizerME, error) {
	m, err := LoadGISModelFromZip(modelPath)
	if err != nil {
		return nil, err
	}
	return &TokenizerME{
		model:                       m,
		useAlphaNumericOptimization: true,
		alphanumeric:                regexp.MustCompile(`^[A-Za-z0-9]+$`),
	}, nil
}

// Tokenize returns token strings (OpenNLP TokenizerME.tokenize).
func (t *TokenizerME) Tokenize(s string) []string {
	spans := t.TokenizePos(s)
	out := make([]string, len(spans))
	for i, sp := range spans {
		out[i] = s[sp.start:sp.end]
	}
	return out
}

type charSpan struct{ start, end int }

// TokenizePos ports TokenizerME.tokenizePos (whitespace pre-split + maxent splits).
// Index arithmetic matches Java String (UTF-16 code units) for BMP/ASCII English text
// after LanguageTool replaces ’ with ' before calling the tokenizer.
func (t *TokenizerME) TokenizePos(d string) []charSpan {
	if t == nil || t.model == nil {
		return nil
	}
	ws := whitespaceTokenizePos(d)
	var newTokens []charSpan
	for _, s := range ws {
		tok := d[s.start:s.end]
		// Java: tok.length() < 2 (UTF-16 length; English path is BMP)
		if utf16Len(tok) < 2 {
			newTokens = append(newTokens, s)
			continue
		}
		if t.useAlphaNumericOptimization && t.alphanumeric.MatchString(tok) {
			newTokens = append(newTokens, s)
			continue
		}
		start := s.start
		end := s.end
		origStart := s.start
		for j := origStart + 1; j < end; j++ {
			// Context index is offset within the whitespace token (Java char index).
			// For ASCII, byte offset == char index.
			ctx := DefaultTokenContext(tok, j-origStart)
			probs := t.model.Eval(ctx)
			best := t.model.BestOutcome(probs)
			if best == "T" { // TokenizerME.SPLIT
				newTokens = append(newTokens, charSpan{start, j})
				start = j
			}
		}
		newTokens = append(newTokens, charSpan{start, end})
	}
	return newTokens
}

// DefaultTokenContext ports DefaultTokenContextGenerator.createContext.
// index is the character offset within the whitespace-token string for a
// potential split (between index-1 and index). English models use ASCII features.
func DefaultTokenContext(sentence string, index int) []string {
	preds := make([]string, 0, 24)
	prefix := sentence[:index]
	suffix := sentence[index:]
	preds = append(preds, "p="+prefix, "s="+suffix)
	if index > 0 {
		addCharPreds("p1", charAt(sentence, index-1), &preds)
		if index > 1 {
			addCharPreds("p2", charAt(sentence, index-2), &preds)
			preds = append(preds, "p21="+string([]byte{sentence[index-2], sentence[index-1]}))
		} else {
			preds = append(preds, "p2=bok")
		}
		preds = append(preds, "p1f1="+string([]byte{sentence[index-1], sentence[index]}))
	} else {
		preds = append(preds, "p1=bok")
	}
	addCharPreds("f1", charAt(sentence, index), &preds)
	if index+1 < len(sentence) {
		addCharPreds("f2", charAt(sentence, index+1), &preds)
		preds = append(preds, "f12="+string([]byte{sentence[index], sentence[index+1]}))
	} else {
		preds = append(preds, "f2=bok")
	}
	if len(sentence) > 0 && sentence[0] == '&' && sentence[len(sentence)-1] == ';' {
		preds = append(preds, "cc")
	}
	return preds
}

func charAt(s string, i int) rune {
	if i < 0 || i >= len(s) {
		return 0
	}
	return rune(s[i])
}

func addCharPreds(key string, c rune, preds *[]string) {
	*preds = append(*preds, key+"="+string(c))
	if unicode.IsLetter(c) {
		*preds = append(*preds, key+"_alpha")
		if unicode.IsUpper(c) {
			*preds = append(*preds, key+"_caps")
		}
	} else if unicode.IsDigit(c) {
		*preds = append(*preds, key+"_num")
	} else if openNLPIsWhitespace(c) {
		*preds = append(*preds, key+"_ws")
	} else {
		switch c {
		case '.', '?', '!':
			*preds = append(*preds, key+"_eos")
		case '`', '"', '\'':
			*preds = append(*preds, key+"_quote")
		case '[', '{', '(':
			*preds = append(*preds, key+"_lp")
		case ']', '}', ')':
			*preds = append(*preds, key+"_rp")
		}
	}
}

// openNLPIsWhitespace ports StringUtil.isWhitespace.
func openNLPIsWhitespace(c rune) bool {
	if unicode.IsSpace(c) || unicode.Is(unicode.Zs, c) {
		return true
	}
	return c == '\u00A0'
}

// whitespaceTokenizePos ports WhitespaceTokenizer.tokenizePos.
// Byte offsets for ASCII; rune-aware for general UTF-8 so spans stay valid.
func whitespaceTokenizePos(d string) []charSpan {
	var tokens []charSpan
	tokStart := -1
	inTok := false
	for i := 0; i < len(d); {
		r, size := utf8.DecodeRuneInString(d[i:])
		if openNLPIsWhitespace(r) {
			if inTok {
				tokens = append(tokens, charSpan{tokStart, i})
				inTok = false
				tokStart = -1
			}
		} else if !inTok {
			tokStart = i
			inTok = true
		}
		i += size
	}
	if inTok {
		tokens = append(tokens, charSpan{tokStart, len(d)})
	}
	return tokens
}

func utf16Len(s string) int {
	// Java String.length(); for BMP-only English tokens equals rune count.
	return utf8.RuneCountInString(s)
}

// DiscoverOpenNLPTokenModel finds en-token.bin under third_party (walk-up).
func DiscoverOpenNLPTokenModel() string {
	return walkUpFindFile(filepath.Join("third_party", "opennlp-models", "en-token.bin"))
}
