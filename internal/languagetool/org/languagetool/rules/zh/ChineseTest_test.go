package zh

// Twin of ChineseTest.testLanguage — analyze smoke (full demo-rule list deferred).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of ChineseTest.testLanguage
func TestChinese_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("zh")
	require.Equal(t, "zh", lt.GetLanguageCode())
	sents := lt.Analyze("这是一个测试。")
	require.NotEmpty(t, sents)
}
