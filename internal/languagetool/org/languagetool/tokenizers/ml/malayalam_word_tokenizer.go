package ml

// MalayalamWordTokenizer ports tokenizers.ml.MalayalamWordTokenizer.
type MalayalamWordTokenizer struct{}

func NewMalayalamWordTokenizer() *MalayalamWordTokenizer { return &MalayalamWordTokenizer{} }

const malayalamDelims = "\u0020\u00A0\u115f\u1160\u1680" +
	",.;()[]{}!?:\"'’‘„“”…\\/\t\n"

func (w *MalayalamWordTokenizer) Tokenize(text string) []string {
	if text == "" {
		return nil
	}
	set := map[rune]bool{}
	for _, r := range malayalamDelims {
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
	return out
}
