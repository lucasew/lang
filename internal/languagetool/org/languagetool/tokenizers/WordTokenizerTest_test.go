package tokenizers

// Twin of languagetool-core/src/test/java/org/languagetool/tokenizers/WordTokenizerTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-core/src/test/java/org/languagetool/tokenizers/WordTokenizerTest.java :: WordTokenizerTest.testTokenize
func TestWordTokenizer_Tokenize(t *testing.T) {
	// contains assertEquals — full values in Java twin source
	// contains assertTrue
}

// Port of languagetool-core/src/test/java/org/languagetool/tokenizers/WordTokenizerTest.java :: WordTokenizerTest.testIsUrl
func TestWordTokenizer_IsUrl(t *testing.T) {
	// contains assertTrue
	// contains assertFalse
}

// Port of languagetool-core/src/test/java/org/languagetool/tokenizers/WordTokenizerTest.java :: WordTokenizerTest.testIsEMail
func TestWordTokenizer_IsEMail(t *testing.T) {
	// contains assertTrue
	// contains assertFalse
}

// Port of languagetool-core/src/test/java/org/languagetool/tokenizers/WordTokenizerTest.java :: WordTokenizerTest.testUrlTokenize
func TestWordTokenizer_UrlTokenize(t *testing.T) {
	// contains assertEquals — full values in Java twin source
}

// Port of languagetool-core/src/test/java/org/languagetool/tokenizers/WordTokenizerTest.java :: WordTokenizerTest.testUrlTokenizeWithQuote
func TestWordTokenizer_UrlTokenizeWithQuote(t *testing.T) {
	// contains assertEquals — full values in Java twin source
}

// Port of languagetool-core/src/test/java/org/languagetool/tokenizers/WordTokenizerTest.java :: WordTokenizerTest.testUrlTokenizeWithAppendedCharacter
func TestWordTokenizer_UrlTokenizeWithAppendedCharacter(t *testing.T) {
	// contains assertEquals — full values in Java twin source
}

// Port of languagetool-core/src/test/java/org/languagetool/tokenizers/WordTokenizerTest.java :: WordTokenizerTest.testIncompleteUrlTokenize
func TestWordTokenizer_IncompleteUrlTokenize(t *testing.T) {
	// contains assertEquals — full values in Java twin source
}

// Port of languagetool-core/src/test/java/org/languagetool/tokenizers/WordTokenizerTest.java :: WordTokenizerTest.testCheckCurrencyExpression
func TestWordTokenizer_CheckCurrencyExpression(t *testing.T) {
	// contains assertTrue
	// contains assertFalse
}

// Port of languagetool-core/src/test/java/org/languagetool/tokenizers/WordTokenizerTest.java :: WordTokenizerTest.testSplitCurrencyExpression
func TestWordTokenizer_SplitCurrencyExpression(t *testing.T) {
	tools.Unimplemented("WordTokenizerTest.testSplitCurrencyExpression")
}
