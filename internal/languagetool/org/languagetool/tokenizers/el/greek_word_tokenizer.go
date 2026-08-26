package el

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// GreekWordTokenizer ports org.languagetool.tokenizers.el.GreekWordTokenizer.
// Java: extends WordTokenizer; overrides tokenize only — synchronized JFlex loop
// GreekWordTokenizerImpl.getNextToken until YYEOF, then joinEMailsAndUrls.
// Does not override getTokenizingCharacters (inherited base set unused by tokenize).
type GreekWordTokenizer struct {
	*tokenizers.WordTokenizer
	tokenizer *GreekWordTokenizerImpl
}

// NewGreekWordTokenizer ports new GreekWordTokenizer().
func NewGreekWordTokenizer() *GreekWordTokenizer {
	return &GreekWordTokenizer{
		WordTokenizer: tokenizers.NewWordTokenizer(),
		tokenizer:     NewGreekWordTokenizerImpl(),
	}
}

// Tokenize ports GreekWordTokenizer.tokenize.
// Java:
//
//	List<String> tokens = new ArrayList<>();
//	synchronized (tokenizer) {
//	  tokenizer.yyreset(new StringReader(text));
//	  while (tokenizer.getNextToken() != GreekWordTokenizerImpl.YYEOF) {
//	    tokens.add(tokenizer.getText());
//	  }
//	}
//	return joinEMailsAndUrls(tokens);
func (t *GreekWordTokenizer) Tokenize(text string) []string {
	var tokens []string
	// Java synchronizes on the shared scanner; single-threaded Go twin needs no lock.
	t.tokenizer.Yyreset(text)
	for t.tokenizer.GetNextToken() != YYEOF {
		tokens = append(tokens, t.tokenizer.GetText())
	}
	return tokenizers.JoinEMailsAndUrls(tokens)
}
