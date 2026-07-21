package el

// GreekWordTokenizerImpl ports the surface of the JFlex-generated
// GreekWordTokenizerImpl scanner. Full DFA tables deferred; tokenize is
// one-shot via greekJflexTokenize (same Delim / "ό,τι" rules as the jflex).
type GreekWordTokenizerImpl struct {
	inner *GreekWordTokenizer
}

func NewGreekWordTokenizerImpl() *GreekWordTokenizerImpl {
	return &GreekWordTokenizerImpl{inner: NewGreekWordTokenizer()}
}

// YylexTokenize tokenizes text (Java scanner surface reduced to one-shot tokenize).
// Does not run joinEMailsAndUrls — that is GreekWordTokenizer.Tokenize's job
// (Java: tokenizer loop then joinEMailsAndUrls on the collected list).
func (t *GreekWordTokenizerImpl) YylexTokenize(text string) []string {
	if text == "" {
		return nil
	}
	return greekJflexTokenize(text)
}
