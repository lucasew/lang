package tokenizers

// Tokenizer ports org.languagetool.tokenizers.Tokenizer.
type Tokenizer interface {
	Tokenize(text string) []string
}

// CompoundWordTokenizer ports org.languagetool.tokenizers.CompoundWordTokenizer.
// Same surface as Tokenizer — splits compounds into parts.
type CompoundWordTokenizer interface {
	Tokenizer
}

// FuncTokenizer adapts a function to Tokenizer.
type FuncTokenizer func(text string) []string

func (f FuncTokenizer) Tokenize(text string) []string {
	if f == nil {
		return nil
	}
	return f(text)
}

// Ensure WordTokenizer implements Tokenizer.
var _ Tokenizer = (*WordTokenizer)(nil)
