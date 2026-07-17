package commandline

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGolden_SoftXMLDisambig_IgnoreProductNames(t *testing.T) {
	if DiscoverEnglishIgnoreSpellingList(nil) == "" && DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("soft disambiguation data not found")
	}
	// product/tech names from en-ignore-spelling.txt should not spell-flag
	for _, text := range []string{
		"I use an iPhone every day.",
		"We push code to GitHub daily.",
		"OpenAI models are popular.",
		"We deploy with Kubernetes and Docker.",
		"I prefer TypeScript over JavaScript.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftXML_WillRunFilter(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" || DiscoverEnglishPOSDict(nil) == "" {
		t.Skip("need en-soft.xml and english.dict")
	}
	var out bytes.Buffer
	err := CoreTagHook(&out, "I will run tomorrow.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	s := out.String()
	// run should show VB (possibly only VB after filter)
	require.Contains(t, s, "run/")
	require.Contains(t, s, "VB")
}

func TestGolden_SoftXML_CanCanFilter(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" || DiscoverEnglishPOSDict(nil) == "" {
		t.Skip("need en-soft.xml and english.dict")
	}
	var out bytes.Buffer
	err := CoreTagHook(&out, "They can can fish.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	s := out.String()
	// second "can" filtered to VB
	require.Contains(t, s, "can/")
	require.Contains(t, s, "VB")
}

func TestGolden_ImmunizeAsapNoSpell(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft.xml not found")
	}
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Please reply asap thanks.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	for _, f := range findings {
		require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
	}
}

func TestGolden_SoftXML_ShouldRunFilter(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" || DiscoverEnglishPOSDict(nil) == "" {
		t.Skip("need en-soft.xml and english.dict")
	}
	var out bytes.Buffer
	err := CoreTagHook(&out, "You should run more.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	s := out.String()
	require.Contains(t, s, "run/")
	require.Contains(t, s, "VB")
}

func TestGolden_SoftXML_ModalMakeFilter(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" || DiscoverEnglishPOSDict(nil) == "" {
		t.Skip("need en-soft.xml and english.dict")
	}
	var out bytes.Buffer
	err := CoreTagHook(&out, "They will make dinner.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	s := out.String()
	require.Contains(t, s, "make/")
	require.Contains(t, s, "VB")
}

func TestGolden_ImmunizeBtwIrlNoSpell(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft.xml not found")
	}
	for _, text := range []string{
		"Send that btw.",
		"We met irl yesterday.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestParseOptions_DisambigPaths(t *testing.T) {
	p := &CommandLineParser{}
	opts, err := p.ParseOptions([]string{
		"-l", "en",
		"--ignore-spelling-file", "/tmp/ign.txt",
		"--disambiguation-file", "/tmp/dis.xml",
		"-",
	})
	require.NoError(t, err)
	require.Equal(t, "/tmp/ign.txt", opts.GetIgnoreSpellingFile())
	require.Equal(t, "/tmp/dis.xml", opts.GetDisambiguationFile())
}

func TestGolden_SoftMultiwords_NewZealand(t *testing.T) {
	// default embedded multiwords include New Zealand even without file
	var out bytes.Buffer
	err := CoreTagHook(&out, "I fly to New Zealand tomorrow.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	s := out.String()
	require.Contains(t, s, "New")
	require.True(t, strings.Contains(s, "NNP") || strings.Contains(s, "NP"), s)
}

func TestGolden_SoftMultiwords_SanFrancisco(t *testing.T) {
	if DiscoverEnglishMultiwords(nil) == "" {
		t.Skip("multiwords not found")
	}
	var out bytes.Buffer
	err := CoreTagHook(&out, "I fly to San Francisco tomorrow.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	s := out.String()
	require.Contains(t, s, "San")
	// multiword should attach NNP-family tags
	require.True(t,
		strings.Contains(s, "NNP") || strings.Contains(s, "B-N") || strings.Contains(s, "E-N") || strings.Contains(s, "B-NP"),
		s,
	)
}
