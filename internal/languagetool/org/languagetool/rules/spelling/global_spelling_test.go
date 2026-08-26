package spelling

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of GlobalSpellingTest.avoidSomeWords — synthetic oracle + official resource.
func TestGlobalSpelling_AvoidSomeWords(t *testing.T) {
	// --- Synthetic cases (same prohibited lists / messages as Java) ---
	require.NoError(t, AvoidSomeWords([]string{
		"# comment",
		"Microsoft Entra",
		"log4j",
		"car2go",
		"#Dnipro",          // commented-out form is allowed (Java splits on #)
		"#Fyodor Dostoevsky",
	}))

	err := AvoidSomeWords([]string{"Dnipro"})
	require.Error(t, err)
	require.Equal(t, "Do not use 'Dnipro' in global_spelling.txt. It is not a valid spelling for all languages.", err.Error())

	err = AvoidSomeWords([]string{"Dnepr"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "Dnepr")

	err = AvoidSomeWords([]string{"Leo Tolstoy"})
	require.Error(t, err)
	require.Equal(t, "Do not use 'Tolstoy' in global_spelling.txt. It is not a valid spelling for all languages.", err.Error())

	err = AvoidSomeWords([]string{"Tolstoi"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "Tolstoi")

	err = AvoidSomeWords([]string{"Dostoevsky"})
	require.Error(t, err)
	require.Equal(t, "Do not use 'Dostoevsky' in global_spelling.txt. It is not a valid spelling for all languages.", err.Error())

	// Trailing comment on an allowed line (Java: parts[0].trim() before checks).
	require.NoError(t, AvoidSomeWords([]string{"Microsoft Entra # brand"}))
	// Expression match is full entry only (not a substring token check for Dnipro).
	require.NoError(t, AvoidSomeWords([]string{"something Dnipro else"}))

	// --- Official spelling_global.txt (Java GLOBAL_SPELLING_FILE) ---
	// Must load the same core resource Java DataBroker serves; no soft-skip green.
	require.Equal(t, "spelling_global.txt", GlobalSpellingFile, "SpellingCheckRule.GLOBAL_SPELLING_FILE")
	path := DiscoverSpellingGlobal()
	require.NotEmpty(t, path, "official %s must be discoverable under repo (inspiration/.../resource/spelling_global.txt)", GlobalSpellingFile)
	require.True(t, strings.Contains(path, "spelling_global.txt"), "path=%s", path)

	lines, err := LoadGlobalSpellingFileLines()
	require.NoError(t, err)
	require.NotEmpty(t, lines, "official file must have content")
	require.NoError(t, AvoidSomeWords(lines), "path=%s", path)
	require.NoError(t, AvoidSomeWordsOnOfficialFile())
}

func TestGlobalSpelling_ConstantMatchesJava(t *testing.T) {
	// SpellingCheckRule.GLOBAL_SPELLING_FILE = "spelling_global.txt"
	require.Equal(t, "spelling_global.txt", GlobalSpellingFile)
	// getAdditionalSpellingFileNames includes GLOBAL_SPELLING_FILE for bare rules.
	r := NewSpellingCheckRule("HUNSPELL_RULE", "spell", "xx")
	require.Contains(t, r.GetAdditionalSpellingFileNames(), GlobalSpellingFile)
}
