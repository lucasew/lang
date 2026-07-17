package commandline

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGolden_PhraseReplace(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Guide tot he Galaxy", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "PHRASE_REPLACE" {
			found = true
			require.Equal(t, "to the", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_SoftPhrasePack(t *testing.T) {
	cases := []struct {
		text, sug string
	}{
		{"This is for all intensive purposes true.", "for all intents and purposes"},
		{"Please nip it in the butt now.", "nip it in the bud"},
		{"I did it on accident.", "by accident"},
		{"They are one in the same.", "one and the same"},
		{"Here is a case and point.", "case in point"},
		{"She waited with baited breath.", "bated breath"},
		{"Give them free reign.", "free rein"},
		{"This is based off of data.", "based on"},
		{"Talk to eachother soon.", "each other"},
		{"Questions in regards to your letter.", "with regard to"},
		{"In regards to your letter.", "With regard to"}, // sentence-initial capital
		{"I did it On Accident.", "By accident"},
		{"That is a mute point.", "moot point"},
		{"Please tow the line.", "toe the line"},
		{"It happened all of the sudden.", "all of a sudden"},
		{"By in large, we agree.", "By and large"}, // sentence-initial capital
		{"That gave me piece of mind.", "peace of mind"},
		{"We must make due with less.", "make do"},
		{"He will pass mustard.", "pass muster"},
		{"They hone in on the target.", "home in on"},
		{"That will wet your appetite.", "whet your appetite"},
		{"In the same vane as before.", "In the same vein"}, // sentence-initial capital
		{"The statue of limitations expired.", "statute of limitations"},
		{"He was the escape goat.", "scapegoat"},
		{"Opportunities are few and far in between.", "few and far between"},
	}
	for _, tc := range cases {
		t.Run(tc.sug, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == "PHRASE_REPLACE" && f.Suggestion == tc.sug {
					found = true
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftMightOf(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "She might of left already.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_MIGHT_OF" {
			found = true
			require.Equal(t, "might have", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_SoftTryAnd(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Please try and finish it.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_TRY_AND" {
			found = true
			require.Equal(t, "try to", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}
