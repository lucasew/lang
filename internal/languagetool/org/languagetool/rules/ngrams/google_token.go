package ngrams

import (
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

// GetGoogleTokens tokenizes sentence for Google ngram style (skip whitespace tokens).
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
		if !tools.IsWhitespace(token) {
			result = append(result, NewGoogleToken(token, startPos, startPos+len(token)))
		}
		startPos += len(token)
	}
	return result
}

// GetGoogleTokensFromSentence adds POS tags when span matches a single LT token trivially.
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
		if !tools.IsWhitespace(token) {
			endPos := startPos + len(token)
			pos := findOriginalAnalyzedTokens(sentence, startPos, endPos)
			result = append(result, NewGoogleTokenWithPOS(token, startPos, endPos, pos))
		}
		startPos += len(token)
	}
	return result
}

func findOriginalAnalyzedTokens(sentence *languagetool.AnalyzedSentence, startPos, endPos int) []*languagetool.AnalyzedToken {
	// Exact span match first.
	for _, tr := range sentence.GetTokens() {
		if tr == nil {
			continue
		}
		if tr.GetStartPos() == startPos && tr.GetEndPos() == endPos {
			return tr.GetReadings()
		}
	}
	// Soft: token surface equality when positions differ (tokenizer vs analyzer).
	// Prefer non-blank tokens whose GetToken matches the substring length span loosely.
	for _, tr := range sentence.GetTokensWithoutWhitespace() {
		if tr == nil {
			continue
		}
		// cover when analyzer start aligns and length matches end-start
		if tr.GetStartPos() == startPos && tr.GetEndPos()-tr.GetStartPos() == endPos-startPos {
			return tr.GetReadings()
		}
	}
	return nil
}

// GetGoogleTokensForString ports GoogleTokenUtil.getGoogleTokensForString.
func GetGoogleTokensForString(sentence string, addStartToken bool, tokenize TokenizerFunc) []string {
	var out []string
	for _, t := range GetGoogleTokens(sentence, addStartToken, tokenize) {
		out = append(out, t.Token)
	}
	return out
}
