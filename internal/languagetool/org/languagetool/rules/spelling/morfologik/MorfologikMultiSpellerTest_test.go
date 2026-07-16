package morfologik

// Twin of languagetool-core/src/test/java/org/languagetool/rules/spelling/morfologik/MorfologikMultiSpellerTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func testMultiSpeller() *MorfologikMultiSpeller {
	// test.dict words
	main := NewMorfologikSpeller("/xx/spelling/test.dict", 1)
	for _, w := range []string{"wordone", "wordtwo", "Häuser"} {
		main.AddWord(w)
	}
	main.Suggestions["wordones"] = []string{"wordone"}
	main.Suggestions["Häusers"] = []string{"Häuser"}

	// test2.txt words
	plain := NewMorfologikSpeller("/xx/spelling/test2.txt", 1)
	for _, w := range []string{"Abc", "wordthree", "wordfour", "üblich", "schön", "Fön", "Fün", "Fän", "Häuser"} {
		plain.AddWord(w)
	}
	plain.Suggestions["Abd"] = []string{"Abc"}
	plain.Suggestions["Fxn"] = []string{"Fän", "Fön", "Fün"}
	return NewMorfologikMultiSpeller(main, plain)
}

// Port of MorfologikMultiSpellerTest.testIsMisspelled
func TestMorfologikMultiSpeller_IsMisspelled(t *testing.T) {
	speller := testMultiSpeller()
	require.False(t, speller.IsMisspelled("wordone"))
	require.False(t, speller.IsMisspelled("wordtwo"))
	require.False(t, speller.IsMisspelled("Abc"))
	require.False(t, speller.IsMisspelled("wordthree"))
	require.False(t, speller.IsMisspelled("wordfour"))
	require.False(t, speller.IsMisspelled("üblich"))
	require.False(t, speller.IsMisspelled("schön"))
	require.False(t, speller.IsMisspelled("Fön"))
	require.False(t, speller.IsMisspelled("Fün"))
	require.False(t, speller.IsMisspelled("Fän"))
	require.False(t, speller.IsMisspelled("Häuser"))
	require.True(t, speller.IsMisspelled("notthere"))
	require.True(t, speller.IsMisspelled("Fun"))
	require.True(t, speller.IsMisspelled("Füns"))
	require.True(t, speller.IsMisspelled("AFün"))
}

// Port of MorfologikMultiSpellerTest.testGetSuggestions
func TestMorfologikMultiSpeller_GetSuggestions(t *testing.T) {
	speller := testMultiSpeller()
	require.Empty(t, speller.GetSuggestions("wordone"))
	require.Equal(t, []string{"wordone"}, speller.GetSuggestions("wordones"))
	require.Equal(t, []string{"Abc"}, speller.GetSuggestions("Abd"))
	require.Equal(t, []string{"Fän", "Fön", "Fün"}, speller.GetSuggestions("Fxn"))
	require.Equal(t, []string{"Häuser"}, speller.GetSuggestions("Häusers"))
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
