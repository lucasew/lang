package tokenizers

// ArabicWordTokenizer ports org.languagetool.tokenizers.ArabicWordTokenizer.
type ArabicWordTokenizer struct {
	*WordTokenizer
}

func NewArabicWordTokenizer() *ArabicWordTokenizer {
	return &ArabicWordTokenizer{WordTokenizer: NewWordTokenizer()}
}

// GetTokenizingCharacters adds Arabic punctuation: ، ؟ ؛ and hyphen.
func (w *ArabicWordTokenizer) GetTokenizingCharacters() string {
	base := TokenizingCharacters()
	return base + "،؟؛-"
}

func (w *ArabicWordTokenizer) Tokenize(text string) []string {
	if text == "" {
		return nil
	}
	// Use custom character set via local tokenization.
	chars := w.GetTokenizingCharacters()
	set := map[rune]bool{}
	for _, r := range chars {
		set[r] = true
	}
	var out []string
	var cur []rune
	flush := func() {
		if len(cur) > 0 {
			out = append(out, string(cur))
			cur = nil
		}
	}
	for _, r := range text {
		if set[r] {
			flush()
			out = append(out, string(r))
		} else {
			cur = append(cur, r)
		}
	}
	flush()
	return JoinEMailsAndUrls(out)
}
