package sr

// Twin of languagetool-language-modules/sr/src/test/java/org/languagetool/tagging/sr/JekavianTaggerTest.java

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func requireDefaultJekavian(t *testing.T) *SerbianTagger {
	t.Helper()
	if DiscoverJekavianPOSDict() == "" {
		t.Skip("jekavian/serbian.dict not in tree")
	}
	EnsureDefaultJekavianTagger()
	require.NotNil(t, DefaultJekavianTagger)
	require.NotNil(t, DefaultJekavianTagger.SerbianTagger)
	require.NotNil(t, DefaultJekavianTagger.GetWordTagger())
	return DefaultJekavianTagger.SerbianTagger
}

// Twin of JekavianTaggerTest.testTaggerJesam
func TestJekavianTagger_TaggerJesam(t *testing.T) {
	tagger := requireDefaultJekavian(t)
	assertHasLemmaAndPos(t, tagger, "је", "јесам", "GL:PM:PZ:3L:0J")
	assertHasLemmaAndPos(t, tagger, "јеси", "јесам", "GL:PM:PZ:2L:0J")
	assertHasLemmaAndPos(t, tagger, "смо", "јесам", "GL:PM:PZ:1L:0M")
}

// Twin of JekavianTaggerTest.testTaggerSvijet
func TestJekavianTagger_TaggerSvijet(t *testing.T) {
	tagger := requireDefaultJekavian(t)
	assertHasLemmaAndPos(t, tagger, "цвијете", "цвијет", "IM:ZA:MU:0J:VO")
	assertHasLemmaAndPos(t, tagger, "цвијетом", "цвијет", "IM:ZA:MU:0J:IN")
}

// Twin of JekavianTaggerTest.testTagger (Java TestTools.myAssert cases).
func TestJekavianTagger_Tagger(t *testing.T) {
	tagger := requireDefaultJekavian(t)
	require.Equal(t, JekavianDictionaryPath, tagger.GetDictionaryPath())
	require.NotEmpty(t, JekavianPOSDictPath(), "real jekavian serbian.dict must load")

	// Exact expected strings from Java JekavianTaggerTest.testTagger (1:1).
	cases := []struct {
		input string
		want  string
	}{
		{
			"Ово је лијеп цвијет.",
			"Ово/[овај]ZM:PK:0:SR:0J:AK|Ово/[овај]ZM:PK:0:SR:0J:NO -- је/[јесам]GL:PM:PZ:3L:0J -- лијеп/[лијеп]PR:OP:PO:MU:0J:AK:ST|лијеп/[лијеп]PR:OP:PO:MU:0J:NO:NE|лијеп/[лијеп]PR:OP:PO:MU:0J:VO:NE -- цвијет/[цвијет]IM:ZA:MU:0J:AK:ST|цвијет/[цвијет]IM:ZA:MU:0J:NO",
		},
		{
			// Proof that Jekavian tagger does not tag Ekavian words
			"Ала је леп овај свет, онде поток, овде свет.",
			"Ала/[ала]IM:ZA:ZE:0J:NO|Ала/[ала]IM:ZA:ZE:0M:GE -- је/[јесам]GL:PM:PZ:3L:0J -- леп/[лепак]PR:OP:PO:MU:0J:VO:NE -- овај/[овај]ZM:PK:0:MU:0J:AK:ST|овај/[овај]ZM:PK:0:MU:0J:NO -- свет/[свет]PR:OP:PO:MU:0J:AK:ST|свет/[свет]PR:OP:PO:MU:0J:NO:NE|свет/[свет]PR:OP:PO:MU:0J:VO:NE -- онде/[null]null -- поток/[поток]IM:ZA:MU:0J:AK:ST|поток/[поток]IM:ZA:MU:0J:NO -- овде/[null]null -- свет/[свет]PR:OP:PO:MU:0J:AK:ST|свет/[свет]PR:OP:PO:MU:0J:NO:NE|свет/[свет]PR:OP:PO:MU:0J:VO:NE",
		},
	}
	for _, tc := range cases {
		got := myAssertTagger(tagger, tc.input)
		require.Equal(t, tc.want, got, "input=%q", tc.input)
	}
}
