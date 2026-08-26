package languagetool

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseLanguageModuleProperties(t *testing.T) {
	props, err := ParseLanguageModuleProperties(strings.NewReader(`
# comment
languageClasses=org.languagetool.language.English, org.languagetool.language.German
`))
	require.NoError(t, err)
	require.Len(t, props.LanguageClasses, 2)
	require.Equal(t, "English", ShortClassName(props.LanguageClasses[0]))
}
