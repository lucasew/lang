package uk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestAllowFullTagMatch_pronLeft(t *testing.T) {
	tag := func(w string) []tagging.TaggedWord {
		switch w {
		case "хто", "Хто":
			return []tagging.TaggedWord{{Lemma: "хто", PosTag: "noun:anim:m:v_naz:pron:int"}}
		case "небудь":
			return []tagging.TaggedWord{{Lemma: "небудь", PosTag: "part"}}
		case "вгору":
			return []tagging.TaggedWord{{Lemma: "вгору", PosTag: "adv"}}
		case "вниз":
			return []tagging.TaggedWord{{Lemma: "вниз", PosTag: "adv"}}
		}
		return nil
	}
	require.False(t, AllowFullTagMatch("хто-небудь", tag))
	require.True(t, AllowFullTagMatch("вгору-вниз", tag))
}

func TestAllowFullTagMatch_rightPart(t *testing.T) {
	tag := func(w string) []tagging.TaggedWord {
		switch w {
		case "лікар":
			return []tagging.TaggedWord{{Lemma: "лікар", PosTag: "noun:anim:m:v_naz"}}
		case "коли":
			return []tagging.TaggedWord{{Lemma: "коли", PosTag: "conj:subord"}}
		case "гомеопат":
			return []tagging.TaggedWord{{Lemma: "гомеопат", PosTag: "noun:anim:m:v_naz"}}
		}
		return nil
	}
	// відносини-коли style: right conj → no tagMatch
	require.False(t, AllowFullTagMatch("лікар-коли", tag))
	require.True(t, AllowFullTagMatch("лікар-гомеопат", tag))
}

func TestAllowFullTagMatch_shortLeft(t *testing.T) {
	tag := func(w string) []tagging.TaggedWord {
		switch w {
		case "га", "Га":
			return []tagging.TaggedWord{{Lemma: "га", PosTag: "intj"}}
		case "м":
			return []tagging.TaggedWord{{Lemma: "м", PosTag: "noun:inanim:m:v_naz:abbr"}}
		case "б":
			return []tagging.TaggedWord{{Lemma: "б", PosTag: "part"}}
		}
		return nil
	}
	// short left intj allowed
	require.True(t, AllowFullTagMatch("га-га", tag))
	// short left non-intj blocked
	require.False(t, AllowFullTagMatch("м-б", tag))
}

func TestDynamicFinalOAdj(t *testing.T) {
	tag := func(w string) []tagging.TaggedWord {
		switch w {
		case "західний", "Західний":
			return []tagging.TaggedWord{{Lemma: "західний", PosTag: "adj:m:v_naz"}}
		case "південно":
			// LEFT_O_ADJ_INVALID but oAdjDictOK may fail — use leftOAdj list word
			return nil
		}
		return nil
	}
	// австро is in LEFT_O_ADJ
	tag = func(w string) []tagging.TaggedWord {
		if w == "угорський" {
			return []tagging.TaggedWord{{Lemma: "угорський", PosTag: "adj:m:v_naz"}}
		}
		return nil
	}
	rs := DynamicFinalOAdjReadings("австро-угорський", tag)
	require.NotEmpty(t, rs)
	require.Contains(t, *rs[0].GetPOSTag(), "adj")
}
