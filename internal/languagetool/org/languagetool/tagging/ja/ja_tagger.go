package ja

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// JapaneseDictPath is retained for twins that assert the Java resource path.
// Japanese analysis does not load a Morfologik dict (Java JapaneseTagger parses
// Sen-encoded tokens from JapaneseWordTokenizer).
const JapaneseDictPath = "/ja/ja.dict"

// JapaneseTagger ports tagging.ja.JapaneseTagger.
// Each input token is "surface POS lemma" (spaces) from JapaneseWordTokenizer.
type JapaneseTagger struct{}

func NewJapaneseTagger() *JapaneseTagger { return &JapaneseTagger{} }

// GetDictionaryPath matches BaseTagger twins (path only; no dict load).
func (t *JapaneseTagger) GetDictionaryPath() string { return JapaneseDictPath }

// Tag ports JapaneseTagger.tag: split each encoded token into AnalyzedToken.
func (t *JapaneseTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0
	for _, word := range sentenceTokens {
		at := asAnalyzedToken(word)
		out = append(out, languagetool.NewAnalyzedTokenReadingsAt(at, pos))
		// Java: pos += at.getToken().length() — surface UTF-16 units for BMP JA.
		pos += tokenizers.UTF16Len(at.GetToken())
	}
	return out
}

// CreateNullToken ports JapaneseTagger.createNullToken.
func (t *JapaneseTagger) CreateNullToken(token string, startPos int) *languagetool.AnalyzedTokenReadings {
	return languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken(token, nil, nil), startPos)
}

// CreateToken ports JapaneseTagger.createToken (lemma null).
func (t *JapaneseTagger) CreateToken(token, posTag string) *languagetool.AnalyzedToken {
	p := posTag
	return languagetool.NewAnalyzedToken(token, &p, nil)
}

// asAnalyzedToken ports JapaneseTagger.asAnalyzedToken.
func asAnalyzedToken(word string) *languagetool.AnalyzedToken {
	// Java: word.split(" ") — default limit 0 discards trailing empty strings.
	parts := strings.Split(word, " ")
	// Match Java Pattern.split trailing-empty discard for parity on edge cases.
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	if len(parts) != 3 {
		// Java returns new AnalyzedToken(" ", null, null) for malformed rows.
		return languagetool.NewAnalyzedToken(" ", nil, nil)
	}
	pos := parts[1]
	lemma := parts[2]
	return languagetool.NewAnalyzedToken(parts[0], &pos, &lemma)
}
