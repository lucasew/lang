package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestMatchState_FilterReadings_PostagRegexp(t *testing.T) {
	// Keep only NN readings (Java FILTER with postag regexp).
	nn := "NN"
	nns := "NNS"
	vb := "VB"
	lemma := "dog"
	atr := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("dogs", &nn, &lemma),
		languagetool.NewAnalyzedToken("dogs", &nns, &lemma),
		languagetool.NewAnalyzedToken("dogs", &vb, &lemma),
	}, 0)
	m := NewMatch("NN|NNS", "", true, "", "", CaseNone, false, false, IncludeNone)
	ms := m.CreateStateWithSynth(nil, atr)
	out := ms.FilterReadings()
	require.NotNil(t, out)
	var tags []string
	for _, r := range out.GetReadings() {
		if r != nil && r.GetPOSTag() != nil {
			tags = append(tags, *r.GetPOSTag())
		}
	}
	require.Contains(t, tags, "NN")
	require.Contains(t, tags, "NNS")
	require.NotContains(t, tags, "VB")
}

func TestMatchState_FilterReadings_ExactPostag(t *testing.T) {
	nn := "NN"
	vb := "VB"
	lem := "run"
	atr := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("run", &nn, &lem),
		languagetool.NewAnalyzedToken("run", &vb, &lem),
	}, 0)
	m := NewMatch("VB", "", false, "", "", CaseNone, false, false, IncludeNone)
	ms := m.CreateStateWithSynth(nil, atr)
	out := ms.FilterReadings()
	require.NotNil(t, out)
	// getNewToken: one VB reading per tagged source reading
	for _, r := range out.GetReadings() {
		if r.GetPOSTag() != nil && *r.GetPOSTag() == languagetool.SentenceEndTagName {
			continue
		}
		require.Equal(t, "VB", *r.GetPOSTag())
	}
}

func TestMatchState_FilterReadings_NoPosTag(t *testing.T) {
	atr := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("foo", nil, nil))
	m := NewMatch("", "", false, "", "", CaseNone, false, false, IncludeNone)
	ms := m.CreateStateWithSynth(nil, atr)
	out := ms.FilterReadings()
	require.Equal(t, atr, out) // unchanged
}
