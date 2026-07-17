package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseDataAnnotation(t *testing.T) {
	at, err := ParseDataAnnotation(`{"annotation":[{"text":"See "},{"markup":"<b>"},{"text":"a error"},{"markup":"</b>"},{"text":" here."}]}`)
	require.NoError(t, err)
	require.Equal(t, "See a error here.", at.GetPlainText())
	require.Contains(t, at.GetTextWithMarkup(), "<b>")
}

func TestParseDataAnnotation_InterpretAs(t *testing.T) {
	at, err := ParseDataAnnotation(`{"annotation":[{"text":"Hello"},{"markup":"<p>","interpretAs":"\n\n"},{"text":"World"}]}`)
	require.NoError(t, err)
	require.Equal(t, "Hello\n\nWorld", at.GetPlainText())
}

func TestParseDataAnnotation_BadJSON(t *testing.T) {
	_, err := ParseDataAnnotation(`not-json`)
	require.Error(t, err)
}
