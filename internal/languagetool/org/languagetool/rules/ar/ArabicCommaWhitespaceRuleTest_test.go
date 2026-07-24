package ar

// Twin of languagetool-language-modules/ar/src/test/java/org/languagetool/rules/ar/ArabicCommaWhitespaceRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

func TestArabicCommaWhitespaceRule_Rule(t *testing.T) {
	rule := NewArabicCommaWhitespaceRule(nil)
	// Java: JLanguageTool(Languages.getLanguageForShortCode("ar")) → ArabicWordTokenizer
	wt := tokenizers.NewArabicWordTokenizer()
	assertMatches := func(text string, n int) {
		t.Helper()
		sent := languagetool.AnalyzeWithTokenizer(text, wt)
		require.Equal(t, n, len(rule.Match(sent)), "text=%q", text)
	}
	// correct
	assertMatches("هذه جملة تجريبية.", 0)
	assertMatches("هذه, هي, جملة التجربة.", 0)
	assertMatches("قل (كيت وكيت) تجربة!.", 0)
	assertMatches("تكلف €2,45.", 0)
	assertMatches("ثمنها 50,- يورو", 0)
	assertMatches("جملة مع علامات الحذف ...", 0)
	assertMatches("هذه صورة: .5 وهي صحيحة.", 0)
	assertMatches("هذه $1,000,000.", 0)
	assertMatches("هذه 1,5.", 0)
	assertMatches("هذا ,,فحص''.", 0)
	assertMatches("نفّذ ./validate.sh لفحص الملف.", 0)
	assertMatches("هذه,\u00A0حقا,\u00A0فراغ غير فاصل.", 0)

	// errors — Arabic comma (Java assertMatches counts)
	assertMatches("هذه،جملة للتجربة.", 1)
	assertMatches("هذه ، جملة للتجربة.", 1)
	assertMatches("هذه ،تجربة جملة.", 2)
	// Leading Arabic comma: tokenizer splits "،" + "هذه" → 2 matches like Java
	assertMatches("،هذه جملة للتجربة.", 2)
}

func TestArabicCommaWhitespaceRule_IDAndComma(t *testing.T) {
	rule := NewArabicCommaWhitespaceRule(nil)
	require.Equal(t, "ARABIC_COMMA_PARENTHESIS_WHITESPACE", rule.GetID())
	require.Equal(t, "،", rule.GetCommaCharacter())
}
