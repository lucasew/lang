package morfologik

// Twin of languagetool-core/src/test/java/org/languagetool/rules/spelling/morfologik/MorfologikMultiSpellerTest.java
// Uses official /xx/spelling/test.dict + test2.txt (inspiration languagetool-core test resources).
// No AddWord / Suggestions map invent as dictionary proof.

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/stretchr/testify/require"
)

const (
	javaMultiTestDictClasspath = "/xx/spelling/test.dict"
	javaMultiTestPlainClasspath = "/xx/spelling/test2.txt"
)

// getSpeller ports MorfologikMultiSpellerTest.getSpeller:
//
//	new MorfologikMultiSpeller("/xx/spelling/test.dict", singletonList("/xx/spelling/test2.txt"), null, 1)
func getSpeller(t *testing.T) *MorfologikMultiSpeller {
	t.Helper()
	dictPath := DiscoverLanguageDict(javaMultiTestDictClasspath)
	require.NotEmpty(t, dictPath, "official %s not found under inspiration/languagetool", javaMultiTestDictClasspath)
	plainPath := spelling.DiscoverSpellingResource(strings.TrimPrefix(javaMultiTestPlainClasspath, "/"))
	require.NotEmpty(t, plainPath, "official %s not found under inspiration/languagetool", javaMultiTestPlainClasspath)

	speller, err := NewMorfologikMultiSpellerFromPaths(
		javaMultiTestDictClasspath,
		[]string{javaMultiTestPlainClasspath},
		1,
	)
	require.NoError(t, err)
	require.NotNil(t, speller)
	// binary + plain-text spellers (Java: spellers = [binary, plain])
	require.GreaterOrEqual(t, len(speller.Spellers), 2, "binary + plain-text dicts required")
	// No Suggestions map invent on either component
	for i, s := range speller.Spellers {
		if s == nil {
			continue
		}
		require.Empty(t, s.Suggestions, "speller[%d] must not use Suggestions map invent", i)
	}
	return speller
}

// Port of MorfologikMultiSpellerTest.testIsMisspelled
func TestMorfologikMultiSpeller_IsMisspelled(t *testing.T) {
	speller := getSpeller(t)
	// from test.dict:
	require.False(t, speller.IsMisspelled("wordone"))
	require.False(t, speller.IsMisspelled("wordtwo"))
	// from test2.txt:
	require.False(t, speller.IsMisspelled("Abc"))
	require.False(t, speller.IsMisspelled("wordthree"))
	require.False(t, speller.IsMisspelled("wordfour"))
	require.False(t, speller.IsMisspelled("üblich"))
	require.False(t, speller.IsMisspelled("schön"))
	require.False(t, speller.IsMisspelled("Fön"))
	require.False(t, speller.IsMisspelled("Fün"))
	require.False(t, speller.IsMisspelled("Fän"))
	// from both test.dict and test2.txt:
	require.False(t, speller.IsMisspelled("Häuser"))
	// not in any of the files:
	require.True(t, speller.IsMisspelled("notthere"))
	require.True(t, speller.IsMisspelled("Fun"))
	require.True(t, speller.IsMisspelled("Füns"))
	require.True(t, speller.IsMisspelled("AFün"))
}

// Port of MorfologikMultiSpellerTest.testGetSuggestions
// Java List.toString() form: [Fän, Fön, Fün] — order matters.
func TestMorfologikMultiSpeller_GetSuggestions(t *testing.T) {
	speller := getSpeller(t)
	require.Equal(t, "[]", javaStringList(speller.GetSuggestions("wordone"))) // non-misspelled
	require.Equal(t, "[wordone]", javaStringList(speller.GetSuggestions("wordones")))
	require.Equal(t, "[Abc]", javaStringList(speller.GetSuggestions("Abd")))
	require.Equal(t, "[Fän, Fön, Fün]", javaStringList(speller.GetSuggestions("Fxn")))
	require.Equal(t, "[Häuser]", javaStringList(speller.GetSuggestions("Häusers")))
}

// Port of MorfologikMultiSpellerTest.testInvalidFileName
func TestMorfologikMultiSpeller_InvalidFileName(t *testing.T) {
	_, err := NewMorfologikMultiSpellerFromPaths("/xx/spelling/test.dict.README", []string{"/xx/spelling/test2.txt"}, 1)
	require.Error(t, err)
	require.Error(t, ValidateMultiSpellerDictPath("/xx/spelling/test.dict.README"))
}

// Port of MorfologikMultiSpellerTest.testInvalidFile
func TestMorfologikMultiSpeller_InvalidFile(t *testing.T) {
	_, err := NewMorfologikMultiSpellerFromPaths("/xx/spelling/no-such-file", []string{"/xx/spelling/test2.txt"}, 1)
	require.Error(t, err)
}

// javaStringList ports Java List.toString() for string lists used in MultiSpellerTest assertions.
func javaStringList(xs []string) string {
	if len(xs) == 0 {
		return "[]"
	}
	return "[" + strings.Join(xs, ", ") + "]"
}
