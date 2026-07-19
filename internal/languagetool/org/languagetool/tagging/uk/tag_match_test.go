package uk

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func tok(surface, pos, lemma string) *languagetool.AnalyzedToken {
	p, l := pos, lemma
	return languagetool.NewAnalyzedToken(surface, &p, &l)
}

func TestTagMatch_equalVerb(t *testing.T) {
	left := []*languagetool.AnalyzedToken{tok("жило", "verb:imperf:past:n", "жити")}
	right := []*languagetool.AnalyzedToken{tok("було", "verb:imperf:past:n", "бути")}
	rs := TagMatch("жило-було", left, right)
	require.NotEmpty(t, rs)
	require.Equal(t, "verb:imperf:past:n", *rs[0].GetPOSTag())
	require.Equal(t, "жити-бути", *rs[0].GetLemma())
}

func TestTagMatch_nounNounAgreed(t *testing.T) {
	left := []*languagetool.AnalyzedToken{tok("пане", "noun:anim:m:v_naz", "пан")}
	right := []*languagetool.AnalyzedToken{tok("товаришу", "noun:anim:m:v_naz", "товариш")}
	rs := TagMatch("пане-товаришу", left, right)
	require.NotEmpty(t, rs)
	require.Contains(t, *rs[0].GetPOSTag(), "noun:anim:m:v_naz")
}

func TestTagMatch_skipInanimVKly(t *testing.T) {
	left := []*languagetool.AnalyzedToken{tok("рибо", "noun:inanim:f:v_kly", "риба")}
	right := []*languagetool.AnalyzedToken{tok("полювання", "noun:inanim:n:v_naz", "полювання")}
	require.Nil(t, TagMatch("рибо-полювання", left, right))
}

func TestTagMatch_juniorSenior(t *testing.T) {
	left := []*languagetool.AnalyzedToken{tok("Буш", "noun:anim:m:v_naz:prop:lname", "Буш")}
	right := []*languagetool.AnalyzedToken{tok("молодший", "adj:m:v_naz", "молодший")}
	rs := TagMatch("Буш-молодший", left, right)
	require.NotEmpty(t, rs)
}

func TestTagMatch_daysPluralExtra(t *testing.T) {
	left := []*languagetool.AnalyzedToken{tok("понеділок", "noun:inanim:m:v_naz", "понеділок")}
	right := []*languagetool.AnalyzedToken{tok("вівторок", "noun:inanim:m:v_naz", "вівторок")}
	rs := TagMatch("понеділок-вівторок", left, right)
	require.NotEmpty(t, rs)
	hasP := false
	for _, r := range rs {
		if r.GetPOSTag() != nil && strings.Contains(*r.GetPOSTag(), ":p:") {
			hasP = true
		}
	}
	require.True(t, hasP)
}

func TestFullTagMatchViaTagMatch_dict(t *testing.T) {
	tag := func(w string) []tagging.TaggedWord {
		switch strings.ToLower(w) {
		case "жило":
			return []tagging.TaggedWord{{Lemma: "жити", PosTag: "verb:imperf:past:n"}}
		case "було":
			return []tagging.TaggedWord{{Lemma: "бути", PosTag: "verb:imperf:past:n"}}
		}
		return nil
	}
	rs := FullTagMatchReadings("жило-було", tag)
	require.NotEmpty(t, rs)
	require.Equal(t, "verb:imperf:past:n", *rs[0].GetPOSTag())
	require.Equal(t, "жити-бути", *rs[0].GetLemma())
}
