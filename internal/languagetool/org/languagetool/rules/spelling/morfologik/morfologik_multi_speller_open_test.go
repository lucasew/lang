package morfologik

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Twin of MorfologikMultiSpeller(binary + plain text paths) isMisspelled OR semantics.
func TestOpenMultiSpellerFromClasspath_EN(t *testing.T) {
	if DiscoverLanguageDict("/en/hunspell/en_US.dict") == "" {
		t.Skip("en_US.dict not in tree")
	}
	// EN prepareLine: form\tNN* → form; no-tab line kept
	prep := func(line string) []string {
		if i := indexHash(line); i >= 0 {
			line = line[:i]
		}
		line = trim(line)
		if line == "" {
			return nil
		}
		if j := indexTab(line); j >= 0 {
			form := trim(line[:j])
			tag := trim(line[j+1:])
			if len(tag) >= 2 && (tag[:2] == "NN" || tag[:2] == "JJ") {
				return []string{form}
			}
			return nil
		}
		return []string{line}
	}
	m := OpenMultiSpellerFromClasspath("/en/hunspell/en_US.dict",
		[]string{"en/hunspell/spelling.txt", "spelling.txt", "en/multiwords.txt"},
		"en/hunspell/spelling_en-US.txt", 1, prep)
	require.NotNil(t, m)
	require.GreaterOrEqual(t, len(m.Spellers), 1)
	// binary accepts
	require.False(t, m.IsMisspelled("software"))
	require.False(t, m.IsMisspelled("behavior"))
	// plain multiwords single-token surface (if present)
	require.False(t, m.IsMisspelled("Booking.com"))
	// unknown
	require.True(t, m.IsMisspelled("sdadsadasxyz"))
	// frequency from binary
	require.Greater(t, m.GetFrequency("the"), 0)
	// Java: convertsCase = binary MorfologikSpeller.convertsCase() (en_US convert-case default true)
	require.True(t, m.ConvertsCase())
}

func TestMultiSpeller_PlainFSA_Suggests(t *testing.T) {
	// Plain runtime FSA (FSABuilder) should findRepl, not only membership.
	m := OpenMultiSpellerFromClasspathWithUser(
		"/xx/none.dict", // no binary — plain only
		nil,
		"",
		1,
		nil,
		[]string{"receive", "recipe", "the", "cat"},
	)
	require.NotNil(t, m)
	// user dict first when user words present
	require.NotEmpty(t, m.UserDictSpellers)
	// suggestions from user FSA
	sugs := m.GetSuggestionsFromUserDicts("recieve")
	require.Contains(t, sugs, "receive", "sugs=%v", sugs)
}

func TestMorfologikSpellerRule_MultiPlainText(t *testing.T) {
	if DiscoverLanguageDict("/en/hunspell/en_US.dict") == "" {
		t.Skip("en_US.dict not in tree")
	}
	m := OpenMultiSpellerFromClasspath("/en/hunspell/en_US.dict",
		[]string{"en/multiwords.txt"},
		"", 1, func(line string) []string {
			// simplified NN* only
			if len(line) == 0 || line[0] == '#' {
				return nil
			}
			parts := splitTab(line)
			if len(parts) < 2 {
				return []string{trim(line)}
			}
			tag := trim(parts[1])
			if len(tag) >= 2 && (tag[:2] == "NN" || tag[:2] == "JJ") {
				return []string{trim(parts[0])}
			}
			return nil
		})
	sp := m.Spellers[0]
	r := NewMorfologikSpellerRule("MORFOLOGIK_RULE_EN_US", "en", "/en/hunspell/en_US.dict", sp)
	r.Multi = m
	if r.SpellingCheckRule != nil {
		r.IgnoreWordsWithLength = 1
	}
	// Booking.com accepted via plain multiwords
	ms, err := r.Match(languagetool.AnalyzePlain("Booking.com"))
	require.NoError(t, err)
	require.Empty(t, ms)
	// still flags true misspellings
	ms, err = r.Match(languagetool.AnalyzePlain("sdadsadasxyz"))
	require.NoError(t, err)
	require.NotEmpty(t, ms)
}

func indexHash(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == '#' {
			return i
		}
	}
	return -1
}
func indexTab(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == '\t' {
			return i
		}
	}
	return -1
}
func trim(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}
func splitTab(s string) []string {
	i := indexTab(s)
	if i < 0 {
		return []string{s}
	}
	return []string{s[:i], s[i+1:]}
}
