package sr

// Twin of languagetool-language-modules/sr/src/test/java/org/languagetool/tagging/sr/EkavianTaggerTest.java

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func requireDefaultEkavian(t *testing.T) *SerbianTagger {
	t.Helper()
	if DiscoverEkavianPOSDict() == "" {
		t.Skip("ekavian/serbian.dict not in tree")
	}
	EnsureDefaultEkavianTagger()
	require.NotNil(t, DefaultEkavianTagger)
	require.NotNil(t, DefaultEkavianTagger.SerbianTagger)
	require.NotNil(t, DefaultEkavianTagger.GetWordTagger())
	return DefaultEkavianTagger.SerbianTagger
}

// Twin of EkavianTaggerTest.testTaggerRaditi
func TestEkavianTagger_TaggerRaditi(t *testing.T) {
	tagger := requireDefaultEkavian(t)
	// to work
	assertHasLemmaAndPos(t, tagger, "радим", "радити", "GL:GV:PZ:1L:0J")
	// Глаголски прилог садашњи
	assertHasLemmaAndPos(t, tagger, "радећи", "радити", "PL:PN")
}

// Twin of EkavianTaggerTest.testTaggerJesam
func TestEkavianTagger_TaggerJesam(t *testing.T) {
	tagger := requireDefaultEkavian(t)
	assertHasLemmaAndPos(t, tagger, "је", "јесам", "GL:PM:PZ:3L:0J")
	assertHasLemmaAndPos(t, tagger, "јеси", "јесам", "GL:PM:PZ:2L:0J")
	assertHasLemmaAndPos(t, tagger, "смо", "јесам", "GL:PM:PZ:1L:0M")
}

// Twin of EkavianTaggerTest.testTagger (Java TestTools.myAssert cases).
func TestEkavianTagger_Tagger(t *testing.T) {
	tagger := requireDefaultEkavian(t)
	require.Equal(t, EkavianDictionaryPath, tagger.GetDictionaryPath())
	require.NotEmpty(t, EkavianPOSDictPath(), "real ekavian serbian.dict must load")

	// Exact expected strings from Java EkavianTaggerTest.testTagger (1:1).
	cases := []struct {
		input string
		want  string
	}{
		{
			"Данас је леп дан.",
			"Данас/[данас]PL:GN:PO -- је/[јесам]GL:PM:PZ:3L:0J -- леп/[леп]PR:OP:PO:MU:0J:AK:ST|леп/[леп]PR:OP:PO:MU:0J:NO:NE|леп/[леп]PR:OP:PO:MU:0J:VO:NE|леп/[лепак]PR:OP:PO:MU:0J:VO:NE -- дан/[дан]IM:ZA:MU:0J:AK:ST|дан/[дан]IM:ZA:MU:0J:NO|дан/[дан]PR:OP:PO:MU:0J:AK:ST|дан/[дан]PR:OP:PO:MU:0J:NO:NE|дан/[дан]PR:OP:PO:MU:0J:VO:NE|дан/[дати]PR:PC:PO:MU:0J:AK:ST|дан/[дати]PR:PC:PO:MU:0J:NO:NE|дан/[дати]PR:PC:PO:MU:0J:VO:NE",
		},
		{
			// Note: first letter of "Oво" is Latin O in Java source (not Cyrillic О).
			"Oво је велика кућа.",
			"Oво/[null]null -- је/[јесам]GL:PM:PZ:3L:0J -- велика/[велик]PR:OP:PO:MU:0J:AK:ZI|велика/[велик]PR:OP:PO:MU:0J:GE:NE|велика/[велик]PR:OP:PO:SR:0J:GE:NE|велика/[велик]PR:OP:PO:SR:0M:AK:OR|велика/[велик]PR:OP:PO:SR:0M:NO:OR|велика/[велик]PR:OP:PO:SR:0M:VO:OR|велика/[велик]PR:OP:PO:ZE:0J:NO:OR|велика/[велик]PR:OP:PO:ZE:0J:VO:OR -- кућа/[кућа]IM:ZA:ZE:0J:NO|кућа/[кућа]IM:ZA:ZE:0M:GE",
		},
		{
			"Растао сам поред Дунава.",
			"Растао/[растати]GL:GV:RA:0:0J:MU|Растао/[расти]GL:GV:RA:0:0J:MU -- сам/[сам]PR:OP:PO:MU:0J:AK:ST|сам/[сам]PR:OP:PO:MU:0J:NO:NE|сам/[сам]PR:OP:PO:MU:0J:VO:NE|сам/[јесам]GL:PM:PZ:1L:0J -- поред/[поред]PE:GE|поред/[поред]PL:GN:PO -- Дунава/[Дунав]IM:VL:MU:0J:GE|Дунава/[Дунав]IM:VL:MU:0M:GE",
		},
		{
			"Србијом је владао Петар I, краљ ослободилац.",
			"Србијом/[Србија]IM:VL:ZE:0J:IN -- је/[јесам]GL:PM:PZ:3L:0J -- владао/[владати]GL:GV:RA:0:0J:MU -- Петар/[Петар]IM:VL:MU:0J:NO|Петар/[Петар]IM:VL:MU:0J:NO:ZI -- I/[I]BR:RI:ON|I/[i]BR:RI:ON|I/[i]RE:MO|I/[i]UZ|I/[i]VE:SA -- краљ/[краљ]IM:ZA:MU:0J:NO -- ослободилац/[ослободилац]IM:ZA:MU:0J:NO",
		},
		{
			"Луђа кућа.",
			"Луђа/[луд]PR:OP:KM:SR:0M:AK:OR|Луђа/[луд]PR:OP:KM:SR:0M:NO:OR|Луђа/[луд]PR:OP:KM:SR:0M:VO:OR|Луђа/[луд]PR:OP:KM:ZE:0J:NO:OR|Луђа/[луд]PR:OP:KM:ZE:0J:VO:OR|Луђа/[луђи]PR:OP:PO:SR:0M:AK:OR|Луђа/[луђи]PR:OP:PO:SR:0M:NO:OR|Луђа/[луђи]PR:OP:PO:SR:0M:VO:OR|Луђа/[луђи]PR:OP:PO:ZE:0J:NO:OR|Луђа/[луђи]PR:OP:PO:ZE:0J:VO:OR -- кућа/[кућа]IM:ZA:ZE:0J:NO|кућа/[кућа]IM:ZA:ZE:0M:GE",
		},
		{
			// Proof that Ekavian tagger does not tag Jekavian words
			"Ала је лијеп овај свијет, ондје поток, овдје цвијет.",
			"Ала/[ала]IM:ZA:ZE:0J:NO|Ала/[ала]IM:ZA:ZE:0M:GE -- је/[јесам]GL:PM:PZ:3L:0J -- лијеп/[null]null -- овај/[овај]ZM:PK:0:MU:0J:AK:ST|овај/[овај]ZM:PK:0:MU:0J:NO -- свијет/[null]null -- ондје/[null]null -- поток/[поток]IM:ZA:MU:0J:AK:ST|поток/[поток]IM:ZA:MU:0J:NO -- овдје/[null]null -- цвијет/[null]null",
		},
	}
	for _, tc := range cases {
		got := myAssertTagger(tagger, tc.input)
		require.Equal(t, tc.want, got, "input=%q", tc.input)
	}
}
