package zh

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// ChineseDictPath is retained for twins; Chinese analysis does not load Morfologik.
// Java ChineseTagger only parses HanLP-encoded "surface/pos" tokens.
const ChineseDictPath = "/zh/zh.dict"

// ChineseTagger ports tagging.zh.ChineseTagger.
// Each input token is "surface/pos" from ChineseWordTokenizer (HanLP Term.toString).
type ChineseTagger struct{}

func NewChineseTagger() *ChineseTagger { return &ChineseTagger{} }

// GetDictionaryPath matches twins that assert the Java resource path.
func (t *ChineseTagger) GetDictionaryPath() string { return ChineseDictPath }

// Tag ports ChineseTagger.tag: split each encoded token into AnalyzedToken.
func (t *ChineseTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0
	for _, word := range sentenceTokens {
		at := asAnalyzedToken(word)
		out = append(out, languagetool.NewAnalyzedTokenReadingsAt(at, pos))
		// Java: pos += at.getToken().length() (UTF-16 units for BMP CJK)
		pos += tokenizers.UTF16Len(at.GetToken())
	}
	return out
}

// asAnalyzedToken ports ChineseTagger.asAnalyzedToken.
// Java always uses parts[1] as POS (including HanLP unknown "x") — do not invent nil POS.
func asAnalyzedToken(word string) *languagetool.AnalyzedToken {
	if !strings.Contains(word, "/") {
		return languagetool.NewAnalyzedToken(" ", nil, nil)
	}
	// Java:
	// if parts[0].equals("") && parts[parts.length-1].equals("w")
	//   return new AnalyzedToken(word.substring(0, word.length()-2), last char, null)
	parts := strings.Split(word, "/")
	if parts[0] == "" && parts[len(parts)-1] == "w" {
		p := "w"
		surface := word[:len(word)-2]
		return languagetool.NewAnalyzedToken(surface, &p, nil)
	}
	surface := parts[0]
	posTag := parts[1]
	// Java: new AnalyzedToken(parts[0], parts[1], null)
	p := posTag
	return languagetool.NewAnalyzedToken(surface, &p, nil)
}
