package commandline

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Official multiwords + disambiguation.xml must resolve (submodule or vendored).
func TestDiscoverOfficialDisambigResources(t *testing.T) {
	mw := DiscoverEnglishMultiwords(nil)
	require.NotEmpty(t, mw, "en multiwords.txt")
	require.NotContains(t, mw, "soft.txt", "must not use soft multiword invent list")

	xml := DiscoverLanguageDisambiguationXML(nil, "en")
	require.NotEmpty(t, xml, "en disambiguation.xml")
	require.NotContains(t, xml, "soft.xml")
	require.Contains(t, xml, "disambiguation.xml")

	// global optional but preferred when present
	_ = DiscoverGlobalDisambiguationXML(nil)
	_ = DiscoverSpellingGlobal(nil)
}

func TestRegisterEnglishHybridDisambiguator(t *testing.T) {
	lt, err := configureCoreLT("en", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	require.NotNil(t, lt)
	// When official resources exist, hybrid is installed.
	if DiscoverEnglishMultiwords(nil) != "" || DiscoverLanguageDisambiguationXML(nil, "en") != "" {
		require.NotNil(t, lt.Disambiguator, "EnglishHybridDisambiguator should be wired")
	}
	// Smoke: analyze does not panic
	sents := lt.Analyze("New York is big.")
	require.NotEmpty(t, sents)
}

func TestRegisterHybridDisambiguator_FR_ES_PT(t *testing.T) {
	for _, lang := range []string{"fr", "es", "pt"} {
		t.Run(lang, func(t *testing.T) {
			lt, err := configureCoreLT(lang, &CommandLineOptions{Language: lang})
			require.NoError(t, err)
			// Official multiwords or disambiguation.xml should yield a hybrid.
			if DiscoverLanguageMultiwords(nil, lang) != "" || DiscoverLanguageDisambiguationXML(nil, lang) != "" {
				require.NotNil(t, lt.Disambiguator, "hybrid for %s", lang)
			}
			// Analyze must not panic / wipe tokens
			sents := lt.Analyze("Bonjour le monde.")
			if lang == "es" {
				sents = lt.Analyze("Hola mundo.")
			}
			if lang == "pt" {
				sents = lt.Analyze("Olá mundo.")
			}
			require.NotEmpty(t, sents)
		})
	}
}

func TestRegisterHybridDisambiguator_DE_CA_NL(t *testing.T) {
	for _, lang := range []string{"de", "ca", "nl"} {
		t.Run(lang, func(t *testing.T) {
			// DE multitoken lists are large; allow longer timeout via package default.
			lt, err := configureCoreLT(lang, &CommandLineOptions{Language: lang})
			require.NoError(t, err)
			require.NotNil(t, lt.Disambiguator, "hybrid for %s", lang)
			text := "Hallo Welt."
			switch lang {
			case "ca":
				text = "Hola món."
			case "nl":
				text = "Hallo wereld."
			}
			sents := lt.Analyze(text)
			require.NotEmpty(t, sents)
		})
	}
}

func TestRegisterHybridDisambiguator_UK(t *testing.T) {
	// Official multiwords and/or disambiguation.xml under inspiration.
	require.NotEmpty(t, DiscoverLanguageMultiwords(nil, "uk"), "uk multiwords.txt")
	require.NotEmpty(t, DiscoverLanguageDisambiguationXML(nil, "uk"), "uk disambiguation.xml")

	lt, err := configureCoreLT("uk", &CommandLineOptions{Language: "uk"})
	require.NoError(t, err)
	require.NotNil(t, lt.Disambiguator, "UkrainianHybridDisambiguator should be wired")

	// Smoke: analyze does not panic
	sents := lt.Analyze("Це просте речення.")
	require.NotEmpty(t, sents)
}

func TestRegisterHybridDisambiguator_RU(t *testing.T) {
	// Official multiwords and/or disambiguation.xml under inspiration.
	require.NotEmpty(t, DiscoverLanguageMultiwords(nil, "ru"), "ru multiwords.txt")
	require.NotEmpty(t, DiscoverLanguageDisambiguationXML(nil, "ru"), "ru disambiguation.xml")

	lt, err := configureCoreLT("ru", &CommandLineOptions{Language: "ru"})
	require.NoError(t, err)
	require.NotNil(t, lt.Disambiguator, "RussianHybridDisambiguator should be wired")

	sents := lt.Analyze("Это простое предложение.")
	require.NotEmpty(t, sents)
}

func TestRegisterHybridDisambiguator_PL_SV_GL_GA(t *testing.T) {
	cases := []struct {
		lang string
		text string
	}{
		{"pl", "To jest test."},
		{"sv", "Detta är ett test."},
		{"gl", "Isto é unha proba."},
		{"ga", "Is tástáil é seo."},
	}
	for _, tc := range cases {
		t.Run(tc.lang, func(t *testing.T) {
			if DiscoverLanguageMultiwords(nil, tc.lang) == "" && DiscoverLanguageDisambiguationXML(nil, tc.lang) == "" {
				t.Skipf("no official multiwords/disambiguation for %s", tc.lang)
			}
			lt, err := configureCoreLT(tc.lang, &CommandLineOptions{Language: tc.lang})
			require.NoError(t, err)
			require.NotNil(t, lt.Disambiguator, "hybrid for %s", tc.lang)
			sents := lt.Analyze(tc.text)
			require.NotEmpty(t, sents)
		})
	}
}

func TestRegisterHybridDisambiguator_IT(t *testing.T) {
	// Java ItalianRuleDisambiguator: XmlRuleDisambiguator only (no multiwords).
	require.NotEmpty(t, DiscoverLanguageDisambiguationXML(nil, "it"), "it disambiguation.xml")
	lt, err := configureCoreLT("it", &CommandLineOptions{Language: "it"})
	require.NoError(t, err)
	require.NotNil(t, lt.Disambiguator, "ItalianRuleDisambiguator should be wired")
	sents := lt.Analyze("Questo è un test.")
	require.NotEmpty(t, sents)
}

func TestRegisterHybridDisambiguator_AR_SR(t *testing.T) {
	cases := []struct {
		lang string
		text string
	}{
		{"ar", "هذا اختبار."},
		{"sr", "Ovo je test."},
	}
	for _, tc := range cases {
		t.Run(tc.lang, func(t *testing.T) {
			if DiscoverLanguageMultiwords(nil, tc.lang) == "" && DiscoverLanguageDisambiguationXML(nil, tc.lang) == "" {
				t.Skipf("no official multiwords/disambiguation for %s", tc.lang)
			}
			lt, err := configureCoreLT(tc.lang, &CommandLineOptions{Language: tc.lang})
			require.NoError(t, err)
			require.NotNil(t, lt.Disambiguator, "hybrid for %s", tc.lang)
			// Official POS dicts ship in inspiration for ar/sr (sr: ekavian path).
			p := DiscoverLanguagePOSDict(nil, tc.lang)
			require.NotEmpty(t, p, "POS dict for %s", tc.lang)
			require.NotNil(t, lt.TagWord, "POS tagger for %s from %s", tc.lang, p)
			sents := lt.Analyze(tc.text)
			require.NotEmpty(t, sents)
		})
	}
}

func TestRegisterHybridDisambiguator_RO(t *testing.T) {
	// Java Romanian.createDefaultDisambiguator: XmlRuleDisambiguator only.
	require.NotEmpty(t, DiscoverLanguageDisambiguationXML(nil, "ro"), "ro disambiguation.xml")
	lt, err := configureCoreLT("ro", &CommandLineOptions{Language: "ro"})
	require.NoError(t, err)
	require.NotNil(t, lt.Disambiguator, "Romanian XmlRuleDisambiguator should be wired")
	if p := DiscoverLanguagePOSDict(nil, "ro"); p != "" {
		require.NotNil(t, lt.TagWord, "POS tagger for ro from %s", p)
	}
	sents := lt.Analyze("Aceasta este un test.")
	require.NotEmpty(t, sents)
}

func TestRegisterXmlOnlyDisambiguator_DA_EL_BR_EO_KM(t *testing.T) {
	// Java createDefaultDisambiguator → XmlRuleDisambiguator(this) only.
	cases := []struct {
		lang string
		text string
	}{
		{"da", "Dette er en test."},
		{"el", "Αυτή είναι μια δοκιμή."},
		{"br", "Un taol-arnod eo."},
		{"eo", "Ĉi tio estas testo."},
		{"km", "នេះគឺជាការធ្វើតេស្ត។"},
	}
	for _, tc := range cases {
		t.Run(tc.lang, func(t *testing.T) {
			if DiscoverLanguageDisambiguationXML(nil, tc.lang) == "" {
				t.Skipf("no official disambiguation.xml for %s", tc.lang)
			}
			lt, err := configureCoreLT(tc.lang, &CommandLineOptions{Language: tc.lang})
			require.NoError(t, err)
			require.NotNil(t, lt.Disambiguator, "XmlRuleDisambiguator for %s", tc.lang)
			if p := DiscoverLanguagePOSDict(nil, tc.lang); p != "" {
				require.NotNil(t, lt.TagWord, "POS tagger for %s from %s", tc.lang, p)
			}
			sents := lt.Analyze(tc.text)
			require.NotEmpty(t, sents)
		})
	}
}

func TestDiscoverLanguageMultiwords_UK(t *testing.T) {
	p := DiscoverLanguageMultiwords(nil, "uk")
	require.NotEmpty(t, p)
	// official multiwords.txt or vendored *-multiwords-upstream.txt
	require.True(t,
		strings.Contains(p, "multiwords.txt") || strings.Contains(p, "multiwords-upstream"),
		"path=%s", p)
	require.NotContains(t, p, "soft")
}
