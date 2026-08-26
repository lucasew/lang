package languagetool

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"

// TagWordFromMap adapts a tagging.MapWordTagger into a TagWord inject for Analyze.
func TagWordFromMap(m tagging.MapWordTagger) func(token string) []TokenTag {
	if m == nil {
		return nil
	}
	return func(token string) []TokenTag {
		tw := m.Tag(token)
		if len(tw) == 0 {
			// try lowercase soft
			tw = m.Tag(toLowerASCII(token))
		}
		if len(tw) == 0 {
			return nil
		}
		out := make([]TokenTag, 0, len(tw))
		for _, w := range tw {
			out = append(out, TokenTag{POS: w.GetPosTag(), Lemma: w.GetLemma()})
		}
		return out
	}
}

func toLowerASCII(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}
