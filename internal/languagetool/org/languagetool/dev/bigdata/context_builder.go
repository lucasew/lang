package bigdata

// ContextBuilder ports org.languagetool.dev.bigdata.ContextBuilder.
type ContextBuilder struct {
	StartMarker string
	EndMarker   string
}

func NewContextBuilder() *ContextBuilder {
	return &ContextBuilder{StartMarker: "_START_", EndMarker: "_END_"}
}

// GetContext returns tokens around pos (index into tokens) of half-width contextSize.
// tokens typically come from AnalyzedSentence.GetTokensWithoutWhitespace() surfaces.
func (c *ContextBuilder) GetContext(tokens []string, pos, contextSize int) []string {
	if c == nil {
		c = NewContextBuilder()
	}
	if pos < 0 || pos >= len(tokens) {
		return nil
	}
	left := c.leftContext(tokens, pos, contextSize)
	right := c.rightContext(tokens, pos, contextSize)
	out := make([]string, 0, len(left)+1+len(right))
	out = append(out, left...)
	out = append(out, tokens[pos])
	out = append(out, right...)
	return out
}

func (c *ContextBuilder) leftContext(tokens []string, pos, contextSize int) []string {
	var l []string
	for i := pos - 1; i >= 0 && len(l) < contextSize; i-- {
		if i == 0 {
			l = append([]string{c.StartMarker}, l...)
		} else {
			l = append([]string{tokens[i]}, l...)
		}
	}
	return l
}

func (c *ContextBuilder) rightContext(tokens []string, pos, contextSize int) []string {
	var l []string
	for i := pos + 1; i <= len(tokens) && len(l) < contextSize; i++ {
		if i == len(tokens) {
			l = append(l, c.EndMarker)
		} else {
			l = append(l, tokens[i])
		}
	}
	return l
}
