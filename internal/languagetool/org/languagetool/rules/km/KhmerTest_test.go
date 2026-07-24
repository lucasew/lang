package km

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestKhmer_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("km")
	require.Equal(t, "km", lt.GetLanguageCode())
	require.NotEmpty(t, lt.Analyze(`នេះជាអត្ថបទ។`))
}
