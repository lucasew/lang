package commandline

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGolden_WhitespaceRule(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "This  has double spaces.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "WHITESPACE_RULE" {
			found = true
			require.Equal(t, "whitespace", f.Type)
			require.Equal(t, "warning", f.Severity)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_DoublePunctuation(t *testing.T) {
	var buf bytes.Buffer
	// rule flags two consecutive dots (not ???/!!!; ellipsis ... is ignored by design)
	_, err := CoreGoldenHook(&buf, "This is a test sentence.. More text.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "DOUBLE_PUNCTUATION" {
			found = true
			require.Equal(t, "whitespace", f.Type) // SoftRuleMeta PUNCT → typography/whitespace family
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_CommaWhitespace(t *testing.T) {
	var buf bytes.Buffer
	// missing space after comma
	_, err := CoreGoldenHook(&buf, "Hello,world.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "COMMA_PARENTHESIS_WHITESPACE" {
			found = true
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_SentenceWhitespace(t *testing.T) {
	var buf bytes.Buffer
	// missing space after sentence-ending period
	_, err := CoreGoldenHook(&buf, "This is a text.And there's the next sentence.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "SENTENCE_WHITESPACE" {
			found = true
			require.Equal(t, "whitespace", f.Type)
			require.Equal(t, "warning", f.Severity)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_WhitespacePunctuation(t *testing.T) {
	var buf bytes.Buffer
	// space before colon
	_, err := CoreGoldenHook(&buf, "Wait : now", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "WHITESPACE_PUNCTUATION" {
			found = true
			require.Equal(t, "whitespace", f.Type)
			require.Equal(t, "warning", f.Severity)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_WhitespaceParagraphBegin(t *testing.T) {
	var buf bytes.Buffer
	// leading paragraph whitespace (may also fire WHITESPACE_RULE)
	_, err := CoreGoldenHook(&buf, "  Hello world.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "WHITESPACE_PARAGRAPH_BEGIN" {
			found = true
			require.Equal(t, "whitespace", f.Type)
			require.Equal(t, "warning", f.Severity)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_WhitespaceParagraphEnd(t *testing.T) {
	var buf bytes.Buffer
	// trailing whitespace at paragraph end
	_, err := CoreGoldenHook(&buf, "Hello world.  ", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "WHITESPACE_PARAGRAPH" {
			found = true
			require.Equal(t, "whitespace", f.Type)
			require.Equal(t, "warning", f.Severity)
		}
	}
	require.True(t, found, "%+v", findings)
}
