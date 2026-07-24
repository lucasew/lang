package ja

import (
	"strings"
	"sync"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

// JapaneseWordTokenizer ports org.languagetool.tokenizers.ja.JapaneseWordTokenizer.
//
// Java constructs SenFactory.getStringTagger(null, false) and returns each
// morpheme as:
//
//	surface + " " + partOfSpeech + " " + basicForm
//
// where basicForm is the surface when the morpheme basic form is "*".
//
// Go uses kagome with the IPA dictionary (same MeCab IPADIC POS inventory as
// lucene-gosen/ipadic). This is an allowed engine substitute only while
// Java-visible token lists match (JapaneseWordTokenizerTest).
type JapaneseWordTokenizer struct{}

// NewJapaneseWordTokenizer ports JapaneseWordTokenizer().
func NewJapaneseWordTokenizer() *JapaneseWordTokenizer {
	return &JapaneseWordTokenizer{}
}

var (
	// stringTagger equivalent: shared kagome tokenizer (Java holds StringTagger).
	kagomeOnce sync.Once
	kagomeTok  *tokenizer.Tokenizer
	kagomeErr  error
)

func stringTagger() (*tokenizer.Tokenizer, error) {
	kagomeOnce.Do(func() {
		// OmitBosEos: Sen analyze does not emit BOS/EOS surfaces in the LT output.
		kagomeTok, kagomeErr = tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	})
	return kagomeTok, kagomeErr
}

// Tokenize ports JapaneseWordTokenizer.tokenize.
// On analyzer failure Java returns the (empty) list accumulated so far.
func (t *JapaneseWordTokenizer) Tokenize(text string) []string {
	ret := make([]string, 0)
	tok, err := stringTagger()
	if err != nil || tok == nil {
		return ret
	}
	// Java: synchronized (stringTagger) { analyze + format }
	// Kagome Analyze is safe for concurrent use after construction.
	tokens := tok.Analyze(text, tokenizer.Normal)
	for _, tk := range tokens {
		if tk.Class == tokenizer.DUMMY {
			continue
		}
		surface := tk.Surface
		if surface == "" {
			continue
		}
		// Java: token.getMorpheme().getPartOfSpeech()
		pos := ipaPOS(tk.Features())
		// Java: if basicForm equalsIgnoreCase("*") → surface, else basicForm
		basicForm := surface
		if bf, ok := tk.BaseForm(); ok && bf != "" && !strings.EqualFold(bf, "*") {
			basicForm = bf
		}
		ret = append(ret, surface+" "+pos+" "+basicForm)
	}
	return ret
}

// ipaPOS joins IPADIC feature slots 0..3 into Sen-style "名詞-代名詞-一般"
// (stops at first empty or "*" sub-classification, matching lucene-gosen POS).
func ipaPOS(feats []string) string {
	if len(feats) == 0 {
		return ""
	}
	n := 4
	if len(feats) < n {
		n = len(feats)
	}
	parts := make([]string, 0, n)
	for i := 0; i < n; i++ {
		f := feats[i]
		if f == "" || f == "*" {
			break
		}
		parts = append(parts, f)
	}
	return strings.Join(parts, "-")
}
