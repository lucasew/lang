package tokenizers

// SentenceTokenizer ports org.languagetool.tokenizers.SentenceTokenizer.
type SentenceTokenizer interface {
	Tokenizer
	// SetSingleLineBreaksMarksParagraph controls paragraph detection via line breaks.
	SetSingleLineBreaksMarksParagraph(lineBreakParagraphs bool)
	// SingleLineBreaksMarksPara reports the current paragraph mode.
	SingleLineBreaksMarksPara() bool
}
