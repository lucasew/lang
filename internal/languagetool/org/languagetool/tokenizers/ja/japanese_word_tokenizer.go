package ja

import (
	"strings"
	"sync"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

// JapaneseWordTokenizer ports tokenizers.ja.JapaneseWordTokenizer.
//
// Java uses lucene-gosen (Sen) StringTagger and returns each morpheme as:
//
//	surface + " " + partOfSpeech + " " + basicForm
//
// We use kagome with the IPA dictionary (same MeCab IPADIC tag inventory as
// lucene-gosen/ipadic) so soft JA goldens see real POS/lemma, not invented
// surface lists or pattern-matcher char-cover hacks.
type JapaneseWordTokenizer struct {
	// Segment optional custom segmenter (tests / inject). When set, its strings
	// are returned as-is (may already be Java-encoded "surface POS lemma").
	Segment func(text string) []string
}

func NewJapaneseWordTokenizer() *JapaneseWordTokenizer { return &JapaneseWordTokenizer{} }

var (
	kagomeOnce sync.Once
	kagomeTok  *tokenizer.Tokenizer
	kagomeErr  error
)

func kagome() (*tokenizer.Tokenizer, error) {
	kagomeOnce.Do(func() {
		kagomeTok, kagomeErr = tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	})
	return kagomeTok, kagomeErr
}

func (t *JapaneseWordTokenizer) Tokenize(text string) []string {
	if t != nil && t.Segment != nil {
		return t.Segment(text)
	}
	return tokenizeWithKagome(text)
}

func tokenizeWithKagome(text string) []string {
	if text == "" {
		return nil
	}
	tok, err := kagome()
	if err != nil || tok == nil {
		// Last resort: single unknown chunk (Java Sen returns empty on failure).
		return nil
	}
	tokens := tok.Analyze(text, tokenizer.Normal)
	out := make([]string, 0, len(tokens))
	for _, tk := range tokens {
		if tk.Class == tokenizer.DUMMY {
			continue
		}
		surface := tk.Surface
		if surface == "" {
			continue
		}
		pos := ipaPOS(tk.Features())
		basic := surface
		if bf, ok := tk.BaseForm(); ok && bf != "" && bf != "*" {
			basic = bf
		}
		// Java: token.getSurface() + " " + getPartOfSpeech() + " " + basicForm
		out = append(out, surface+" "+pos+" "+basic)
	}
	return out
}

// ipaPOS joins Kagome/IPADIC feature slots into Sen-style "名詞-代名詞-一般".
func ipaPOS(feats []string) string {
	if len(feats) == 0 {
		return ""
	}
	var parts []string
	// IPADIC: 品詞, 品詞細分類1, 品詞細分類2, 品詞細分類3
	n := 4
	if len(feats) < n {
		n = len(feats)
	}
	for i := 0; i < n; i++ {
		f := feats[i]
		if f == "" || f == "*" {
			break
		}
		parts = append(parts, f)
	}
	return strings.Join(parts, "-")
}
