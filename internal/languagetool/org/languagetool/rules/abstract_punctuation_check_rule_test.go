package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// punctCheckDouble is a thin test double of AbstractPunctuationCheckRule.
// Hooks mirror UK PunctuationCheckRule join patterns only for outcome ideas;
// this is not the UK language sector.
type punctCheckDouble struct {
	*AbstractPunctuationCheckRule
}

// newPunctCheckDouble wires hooks that treat single-char ".,!?: -" as punctuation
// and allow only a few "good" joins (..., !.., ?.., !!!, ???, ?!!, ,- with optional trailing spaces)
// similar to UK patterns for outcome-fidelity twins.
func newPunctCheckDouble() *punctCheckDouble {
	base := &AbstractPunctuationCheckRule{
		// isPunctuation: single token matching ^[.,!?: -]$
		IsPunctuation: func(token string) bool {
			if len(token) != 1 {
				return false
			}
			switch token[0] {
			case '.', ',', '!', '?', ':', ' ', '-':
				return true
			}
			return false
		},
		// isPunctsJoinOk: good joins only (not ".." or ",," alone)
		IsPunctsJoinOk: func(tokens string) bool {
			switch tokens {
			case "...", "!..", "?..", "!!!", "???", "?!!", ",-", ":-":
				return true
			}
			// also allow internal-style ", " / ": " / " - " style joins used as good in UK
			if tokens == ", " || tokens == ": " || tokens == " - " || tokens == " " {
				return true
			}
			return false
		},
	}
	InitPunctuationCheckMeta(base, nil)
	return &punctCheckDouble{AbstractPunctuationCheckRule: base}
}

func TestAbstractPunctuationCheckRule_Metadata(t *testing.T) {
	r := newPunctCheckDouble()
	require.Equal(t, "PUNCTUATION_GENERIC_CHECK", r.GetID())
	require.Equal(t, "Use of unusual combination of punctuation characters", r.GetDescription())
	require.NotNil(t, r.GetCategory())
	require.True(t, r.GetCategory().GetID().Equals(CategoryPunctuation))
}

func TestAbstractPunctuationCheckRule_SinglePunctuationNoMatch(t *testing.T) {
	r := newPunctCheckDouble()
	// single punctuation token → tkns.length() < 2 → 0 matches
	require.Empty(t, r.Match(languagetool.AnalyzePlain("ok.")))
	require.Empty(t, r.Match(languagetool.AnalyzePlain("Hi!")))
	require.Empty(t, r.Match(languagetool.AnalyzePlain("a,b")))
}

func TestAbstractPunctuationCheckRule_GoodJoinZeroMatches(t *testing.T) {
	r := newPunctCheckDouble()
	// "..." is isPunctsJoinOk → 0 matches
	require.Empty(t, r.Match(languagetool.AnalyzePlain("a...b")))
	// "!.." good
	require.Empty(t, r.Match(languagetool.AnalyzePlain("x!..")))
	// "?.." good
	require.Empty(t, r.Match(languagetool.AnalyzePlain("q?..")))
	// "!!!" good
	require.Empty(t, r.Match(languagetool.AnalyzePlain("wow!!!")))
}

func TestAbstractPunctuationCheckRule_BadDoubleDot(t *testing.T) {
	r := newPunctCheckDouble()
	// "a..b" → tokens ".", "." → tkns ".." not join-ok → 1 match
	sent := languagetool.AnalyzePlain("a..b")
	ms := r.Match(sent)
	require.Len(t, ms, 1)
	require.Equal(t, 1, ms[0].GetFromPos())
	require.Equal(t, 3, ms[0].GetToPos()) // from 1 + tkns.length() 2
	require.Equal(t, "bad duplication or combination of punctuation signs", ms[0].GetMessage())
	require.Equal(t, "Punctuation problem", ms[0].GetShortMessage())
	require.Equal(t, []string{"."}, ms[0].GetSuggestedReplacements())
	require.Equal(t, r.AbstractPunctuationCheckRule, ms[0].GetRule())
}

func TestAbstractPunctuationCheckRule_BadDoubleComma(t *testing.T) {
	r := newPunctCheckDouble()
	// "x,,y" → ",," bad
	sent := languagetool.AnalyzePlain("x,,y")
	ms := r.Match(sent)
	require.Len(t, ms, 1)
	require.Equal(t, 1, ms[0].GetFromPos())
	require.Equal(t, 3, ms[0].GetToPos())
	require.Equal(t, []string{","}, ms[0].GetSuggestedReplacements())
}

func TestAbstractPunctuationCheckRule_LastTokenPunctuationPath(t *testing.T) {
	r := newPunctCheckDouble()
	// run at end of sentence: "end.." — last token is punct, Java falls through without continue
	sent := languagetool.AnalyzePlain("end..")
	ms := r.Match(sent)
	require.Len(t, ms, 1)
	require.Equal(t, 3, ms[0].GetFromPos())
	require.Equal(t, 5, ms[0].GetToPos())
	require.Equal(t, []string{"."}, ms[0].GetSuggestedReplacements())

	// only punctuation at end / whole text
	ms = r.Match(languagetool.AnalyzePlain(".."))
	require.Len(t, ms, 1)
	require.Equal(t, 0, ms[0].GetFromPos())
	require.Equal(t, 2, ms[0].GetToPos())
	require.Equal(t, []string{"."}, ms[0].GetSuggestedReplacements())
}

func TestAbstractPunctuationCheckRule_NonPunctResetsAccumulation(t *testing.T) {
	r := newPunctCheckDouble()
	// "a.b" — single dots separated by non-punct → each run length 1 → 0 matches
	require.Empty(t, r.Match(languagetool.AnalyzePlain("a.b")))
	// "a.b.c" still singles
	require.Empty(t, r.Match(languagetool.AnalyzePlain("a.b.c")))
}

func TestAbstractPunctuationCheckRule_MultiRuns(t *testing.T) {
	r := newPunctCheckDouble()
	// two bad runs: "a..b,,c"
	sent := languagetool.AnalyzePlain("a..b,,c")
	ms := r.Match(sent)
	require.Len(t, ms, 2)
	// ".." at 1..3
	require.Equal(t, 1, ms[0].GetFromPos())
	require.Equal(t, 3, ms[0].GetToPos())
	require.Equal(t, []string{"."}, ms[0].GetSuggestedReplacements())
	// ",," at 4..6
	require.Equal(t, 4, ms[1].GetFromPos())
	require.Equal(t, 6, ms[1].GetToPos())
	require.Equal(t, []string{","}, ms[1].GetSuggestedReplacements())
}

func TestAbstractPunctuationCheckRule_HandBuiltTokens(t *testing.T) {
	r := newPunctCheckDouble()
	// Hand-built sentence with SENT_START + tokens; positions in UTF-16 units
	ss := languagetool.SentenceStartTagName
	start := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("", &ss, nil), 0)
	t0 := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("hi", nil, nil), 0)
	t1 := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken(":", nil, nil), 2)
	t2 := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("-", nil, nil), 3)
	// ":-" is join-ok in our double → 0 matches
	sentOK := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, t0, t1, t2})
	require.Empty(t, r.Match(sentOK))

	// ".." hand-built bad
	d1 := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken(".", nil, nil), 10)
	d2 := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken(".", nil, nil), 11)
	sentBad := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, d1, d2})
	ms := r.Match(sentBad)
	require.Len(t, ms, 1)
	require.Equal(t, 10, ms[0].GetFromPos())
	require.Equal(t, 12, ms[0].GetToPos())
	require.Equal(t, []string{"."}, ms[0].GetSuggestedReplacements())
}

func TestAbstractPunctuationCheckRule_SpanUsesUTF16Length(t *testing.T) {
	// Even for ASCII punct, to = from + tkns.length() (Java UTF-16 units).
	// Verify javaUTF16Len / javaSubstring01 helpers match Java String for BMP.
	require.Equal(t, 2, javaUTF16Len(".."))
	require.Equal(t, 3, javaUTF16Len("..."))
	require.Equal(t, ".", javaSubstring01(".."))
	require.Equal(t, ",", javaSubstring01(",,"))
	// supplementary: length 2, substring(0,1) is high surrogate decoded alone —
	// for completeness of helpers (rule typically never sees this as punct)
	emoji := "😀"
	require.Equal(t, 2, javaUTF16Len(emoji))
}

func TestAbstractPunctuationCheckRule_CustomHooksOnlyDotComma(t *testing.T) {
	// Narrow hooks: only "." and "," are punctuation; space and letters reset.
	r := &AbstractPunctuationCheckRule{
		IsPunctuation: func(token string) bool {
			return token == "." || token == ","
		},
		IsPunctsJoinOk: func(tokens string) bool {
			return tokens == "..."
		},
	}
	InitPunctuationCheckMeta(r, nil)

	// space between punctuation tokens resets accumulation → singles → 0
	// "a. .b" tokens: ".", " ", "." — space not punct → each run length 1
	require.Empty(t, r.Match(languagetool.AnalyzePlain("a. .b")))

	// adjacent bad ".." still matches
	ms := r.Match(languagetool.AnalyzePlain("a..b"))
	require.Len(t, ms, 1)
	require.Equal(t, []string{"."}, ms[0].GetSuggestedReplacements())
}

func TestAbstractPunctuationCheckRule_InitMetaIdDescriptionCategory(t *testing.T) {
	r := &AbstractPunctuationCheckRule{
		IsPunctuation:  func(string) bool { return false },
		IsPunctsJoinOk: func(string) bool { return true },
	}
	InitPunctuationCheckMeta(r, nil)
	require.Equal(t, "PUNCTUATION_GENERIC_CHECK", r.GetID())
	require.Equal(t, "Use of unusual combination of punctuation characters", r.GetDescription())
	require.True(t, r.GetCategory().GetID().Equals(CategoryPunctuation))
	// empty text / no punct → 0
	require.Empty(t, r.Match(languagetool.AnalyzePlain("hello")))
}
