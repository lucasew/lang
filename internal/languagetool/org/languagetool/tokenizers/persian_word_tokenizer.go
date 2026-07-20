package tokenizers

// PersianWordTokenizer ports org.languagetool.tokenizers.PersianWordTokenizer.
// Java: super.getTokenizingCharacters() + "،؟؛"
type PersianWordTokenizer struct {
	*WordTokenizer
}

func NewPersianWordTokenizer() *PersianWordTokenizer {
	return &PersianWordTokenizer{WordTokenizer: NewWordTokenizer()}
}

func (w *PersianWordTokenizer) GetTokenizingCharacters() string {
	return TokenizingCharacters() + "،؟؛"
}

func (w *PersianWordTokenizer) Tokenize(text string) []string {
	if text == "" {
		return nil
	}
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
