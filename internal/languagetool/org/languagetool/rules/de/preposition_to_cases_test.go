package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// javaActivePrepKeys is the set of keys from PrepositionToCases static add(...) calls
// that are active in Java (commented //add lines for "in" and "ob" excluded). Count: 113.
var javaActivePrepKeys = []string{
	"ab", "abseits", "abzüglich", "an", "angesichts", "anhand", "anlässlich", "anstatt", "anstelle",
	"auf", "aufgrund", "aus", "ausgangs", "ausgenommen", "ausschließlich", "ausweislich", "außer",
	"außerhalb", "bar", "behufs", "bei", "beidseits", "betreffend", "betreffs", "bezüglich", "binnen",
	"bis", "bis auf", "dank", "diesseits", "durch", "eingangs", "eingedenk", "einschließlich",
	"entgegen", "entlang", "entsprechend", "exklusive", "fern", "fernab", "für", "gegen", "gegenüber",
	"gemäß", "getreu", "halber", "hinsichtlich", "hinter", "in puncto", "infolge", "inklusive",
	"inmitten", "innerhalb", "innert", "je", "jenseits", "kontra", "kraft", "laut", "längs",
	"längsseits", "mangels", "mit", "mitsamt", "mittels", "nach", "namens", "neben", "nebst", "nächst",
	"oberhalb", "ohne", "per", "pro", "rücksichtlich", "samt", "seit", "seitens", "seitlich", "statt",
	"trotz", "um", "unbeschadet", "uneingedenk", "ungeachtet", "unter", "unterhalb", "unweit",
	"vermittels", "vermittelst", "vermöge", "versus", "via", "vis-à-vis", "von", "vor", "vorbehaltlich",
	"wegen", "weitab", "wider", "während", "zeit", "zu", "zufolge", "zugunsten", "zuliebe",
	"zuungunsten", "zuwider", "zuzüglich", "zwecks", "zwischen", "à", "über",
}

func TestPrepositionToCases_KeySetMatchesJava(t *testing.T) {
	require.Len(t, javaActivePrepKeys, 113)
	require.Len(t, PrepositionCases, 113)
	for _, k := range javaActivePrepKeys {
		_, ok := PrepositionCases[k]
		require.True(t, ok, "missing Java key %q", k)
	}
	for k := range PrepositionCases {
		found := false
		for _, jk := range javaActivePrepKeys {
			if jk == k {
				found = true
				break
			}
		}
		require.True(t, found, "invent key not in Java: %q", k)
	}
}

func TestCasesForPreposition(t *testing.T) {
	c := CasesForPreposition("mit")
	require.Equal(t, []GrammaticalCase{CaseDat}, c)
	// lowercasing twin of prep.toLowerCase()
	require.Equal(t, []GrammaticalCase{CaseDat}, CasesForPreposition("Mit"))
	c = CasesForPreposition("an")
	require.Equal(t, []GrammaticalCase{CaseDat, CaseAkk}, c)
}

func TestCasesForPreposition_GegenAkkOnly(t *testing.T) {
	// Java: add("gegen", /*Dat,*/ Akk) → Akk only
	require.Equal(t, []GrammaticalCase{CaseAkk}, CasesForPreposition("gegen"))
	require.Equal(t, []GrammaticalCase{CaseAkk}, CasesForPreposition("Gegen"))
}

func TestCasesForPreposition_InAndObAbsent(t *testing.T) {
	// Java commented out: //add("in", Dat, Akk) and //add("ob", Gen)
	_, hasIn := PrepositionCases["in"]
	_, hasOb := PrepositionCases["ob"]
	require.False(t, hasIn)
	require.False(t, hasOb)
	// getOrDefault → empty list (not invent cases)
	in := CasesForPreposition("in")
	ob := CasesForPreposition("ob")
	require.Empty(t, in)
	require.Empty(t, ob)
	require.NotNil(t, in)
	require.NotNil(t, ob)
}

func TestCasesForPreposition_MultiCaseLists(t *testing.T) {
	require.Equal(t, []GrammaticalCase{CaseDat, CaseAkk}, CasesForPreposition("an"))
	require.Equal(t, []GrammaticalCase{CaseDat, CaseGen}, CasesForPreposition("dank"))
	require.Equal(t, []GrammaticalCase{CaseDat, CaseAkk, CaseGen}, CasesForPreposition("entlang"))
	require.Equal(t, []GrammaticalCase{CaseGen, CaseDat}, CasesForPreposition("laut"))
	require.Equal(t, []GrammaticalCase{CaseDat, CaseAkk}, CasesForPreposition("auf"))
	require.Equal(t, []GrammaticalCase{CaseDat, CaseAkk}, CasesForPreposition("hinter"))
	require.Equal(t, []GrammaticalCase{CaseDat, CaseAkk}, CasesForPreposition("neben"))
	require.Equal(t, []GrammaticalCase{CaseDat, CaseAkk}, CasesForPreposition("unter"))
	require.Equal(t, []GrammaticalCase{CaseDat, CaseAkk}, CasesForPreposition("vor"))
	require.Equal(t, []GrammaticalCase{CaseDat, CaseAkk}, CasesForPreposition("zwischen"))
	require.Equal(t, []GrammaticalCase{CaseDat, CaseAkk}, CasesForPreposition("über"))
}

func TestCasesForPreposition_UnknownEmpty(t *testing.T) {
	// Java: getOrDefault(..., new ArrayList<>()) — empty, not null
	c := CasesForPreposition("xyzzy")
	require.Empty(t, c)
	require.NotNil(t, c)
	require.Equal(t, []GrammaticalCase{}, c)
}
