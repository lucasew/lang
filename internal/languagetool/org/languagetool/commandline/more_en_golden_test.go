package commandline

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGolden_WouldOfMustOf(t *testing.T) {
	for _, tc := range []struct {
		text, rule, sug string
	}{
		{"I would of gone.", "EN_WOULD_OF", "would have"},
		{"You must of seen it.", "EN_MUST_OF", "must have"},
	} {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
					require.Equal(t, "grammar", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_IrregardlessPicky(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Irregardless of that.", &CommandLineOptions{Language: "en", Level: "PICKY"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_IRREGARDLESS" {
			found = true
			require.Equal(t, "regardless", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_TooLongSentence(t *testing.T) {
	words := make([]string, 45)
	for i := range words {
		words[i] = "word"
	}
	words[0] = "Word"
	text := strings.Join(words, " ") + "."
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "TOO_LONG_SENTENCE" {
			found = true
			require.Equal(t, "style", f.Type)
			require.Equal(t, "note", f.Severity)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_SoftYourYoure(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Your welcome here.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_SOFT_YOUR_YOU_RE" {
			found = true
			require.Equal(t, "grammar", f.Type)
			require.Equal(t, "error", f.Severity)
			require.Equal(t, "You're welcome", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_SoftItsIts(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Its a test.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_SOFT_ITS_IT_S" {
			found = true
			require.Equal(t, "grammar", f.Type)
			require.Equal(t, "error", f.Severity)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_TooLongParagraph(t *testing.T) {
	// enough sentences/words to trip TOO_LONG_PARAGRAPH
	var sents []string
	for i := 0; i < 20; i++ {
		sents = append(sents, "This is sentence number and filler words enough.")
	}
	text := strings.Join(sents, " ")
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "TOO_LONG_PARAGRAPH" {
			found = true
			require.Equal(t, "style", f.Type)
			require.Equal(t, "note", f.Severity)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_EnglishWordRepeatBeginning(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "I went home. I ate dinner. I slept well.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "ENGLISH_WORD_REPEAT_BEGINNING_RULE" {
			found = true
			require.Equal(t, "duplication", f.Type)
			require.Equal(t, "warning", f.Severity)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_SoftTheirTheyre(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Their going home now.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_SOFT_THEIR_THEY_RE" {
			found = true
			require.Equal(t, "grammar", f.Type)
			require.Contains(t, f.URL, "lang=en")
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_SoftThenThan(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "This is better then that.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_SOFT_THEN_THAN" {
			found = true
			require.Equal(t, "grammar", f.Type)
		}
	}
	require.True(t, found, "%+v", findings)
}
