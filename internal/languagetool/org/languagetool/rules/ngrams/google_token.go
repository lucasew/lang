package ngrams

import (
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// GoogleSentenceStart ports LanguageModel.GOOGLE_SENTENCE_START (_START_).
const GoogleSentenceStart = "_START_"

// TokenizerFunc tokenizes text into tokens (including whitespace tokens).
type TokenizerFunc func(text string) []string

// GoogleToken ports org.languagetool.rules.ngrams.GoogleToken.
type GoogleToken struct {
	Token    string
	StartPos int
	EndPos   int
	PosTags  []*languagetool.AnalyzedToken
}

func NewGoogleToken(token string, startPos, endPos int) GoogleToken {
	// Google indexes typographic apostrophe as ASCII '
	if token == "’" {
		token = "'"
	}
	return GoogleToken{Token: token, StartPos: startPos, EndPos: endPos}
}

func NewGoogleTokenWithPOS(token string, startPos, endPos int, posTags []*languagetool.AnalyzedToken) GoogleToken {
	g := NewGoogleToken(token, startPos, endPos)
	g.PosTags = append([]*languagetool.AnalyzedToken(nil), posTags...)
	return g
}

func (g GoogleToken) IsWhitespace() bool {
	return tools.IsWhitespace(g.Token)
}

func (g GoogleToken) String() string { return g.Token }

// javaStringLen ports Java String.length() (UTF-16 code units).
func javaStringLen(s string) int {
	return len(utf16.Encode([]rune(s)))
}

// GetGoogleTokens tokenizes sentence for Google ngram style (skip whitespace tokens).
// Positions are UTF-16 indices (Java: startPos += token.length()).
func GetGoogleTokens(sentence string, addStartToken bool, tokenize TokenizerFunc) []GoogleToken {
	if tokenize == nil {
		panic("tokenizer required")
	}
	var result []GoogleToken
	if addStartToken {
		result = append(result, NewGoogleToken(GoogleSentenceStart, 0, 0))
	}
	tokens := tokenize(sentence)
	startPos := 0
	for _, token := range tokens {
		tLen := javaStringLen(token)
		if !tools.IsWhitespace(token) {
			result = append(result, NewGoogleToken(token, startPos, startPos+tLen))
		}
		startPos += tLen
	}
	return result
}

// GetGoogleTokensFromSentence adds POS tags when span matches a single LT token trivially.
// Ports GoogleToken.getGoogleTokens(AnalyzedSentence, …).
func GetGoogleTokensFromSentence(sentence *languagetool.AnalyzedSentence, addStartToken bool, tokenize TokenizerFunc) []GoogleToken {
	if sentence == nil {
		return GetGoogleTokens("", addStartToken, tokenize)
	}
	text := sentence.GetText()
	var result []GoogleToken
	if addStartToken {
		result = append(result, NewGoogleToken(GoogleSentenceStart, 0, 0))
	}
	tokens := tokenize(text)
	startPos := 0
	for _, token := range tokens {
		tLen := javaStringLen(token)
		if !tools.IsWhitespace(token) {
			endPos := startPos + tLen
			pos := findOriginalAnalyzedTokens(sentence, startPos, endPos)
			result = append(result, NewGoogleTokenWithPOS(token, startPos, endPos, pos))
		}
		startPos += tLen
	}
	return result
}

// findOriginalAnalyzedTokens ports GoogleToken.findOriginalAnalyzedTokens:
// exact start/end match on tokensWithoutWhitespace only (no soft surface invent).
func findOriginalAnalyzedTokens(sentence *languagetool.AnalyzedSentence, startPos, endPos int) []*languagetool.AnalyzedToken {
	if sentence == nil {
		return nil
	}
	var out []*languagetool.AnalyzedToken
	for _, tr := range sentence.GetTokensWithoutWhitespace() {
		if tr == nil {
			continue
		}
		if tr.GetStartPos() == startPos && tr.GetEndPos() == endPos {
			// Java HashSet of readings; order is not API-stable — keep GetReadings order.
			out = append(out, tr.GetReadings()...)
		}
	}
	return out
}

// GetGoogleTokensForString ports GoogleTokenUtil.getGoogleTokensForString.
func GetGoogleTokensForString(sentence string, addStartToken bool, tokenize TokenizerFunc) []string {
	var out []string
	for _, t := range GetGoogleTokens(sentence, addStartToken, tokenize) {
		out = append(out, t.Token)
	}
	return out
}
