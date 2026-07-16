package el

// GreekWordTokenizerImpl ports the JFlex-generated GreekWordTokenizerImpl.
// Full DFA deferred; delegates to GreekWordTokenizer rune-class splitting.
type GreekWordTokenizerImpl struct {
	inner *GreekWordTokenizer
}

func NewGreekWordTokenizerImpl() *GreekWordTokenizerImpl {
	return &GreekWordTokenizerImpl{inner: NewGreekWordTokenizer()}
}

// YylexTokenize tokenizes text (Java scanner surface reduced to one-shot tokenize).
func (t *GreekWordTokenizerImpl) YylexTokenize(text string) []string {
	if t == nil || t.inner == nil {
		return NewGreekWordTokenizer().Tokenize(text)
	}
	return t.inner.Tokenize(text)
}
