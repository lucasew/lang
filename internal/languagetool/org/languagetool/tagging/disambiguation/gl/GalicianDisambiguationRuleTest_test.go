package gl

// Galician MultiWordChunker outcome twin (no dedicated Java GalicianDisambiguationRuleTest).
// Pattern: PolishDisambiguationRuleTest / SwedishDisambiguationRuleTest#testChunker.
// Java: MultiWordChunker.getInstance("/gl/multiwords.txt") defaults (false,false,false)
// via GalicianHybridDisambiguator.chunker.
// Resources: /gl/multiwords.txt, galician.dict via GalicianTagger.
import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	taggl "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/gl"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// TestGalicianDisambiguationRule_Chunker is the MultiWordChunker stage twin for Galician.
// Java: MultiWordChunker.getInstance("/gl/multiwords.txt") → false,false,false defaults;
// WordTokenizer (core), SRXSentenceTokenizer(Galician/"gl"), GalicianTagger; TestTools.myAssert.
// Phrases taken from official /gl/multiwords.txt (not invented).
func TestGalicianDisambiguationRule_Chunker(t *testing.T) {
	if taggl.DiscoverGalicianPOSDict() == "" {
		t.Skip("galician.dict not in tree")
	}
	taggl.EnsureDefaultGalicianTagger()
	require.NotNil(t, taggl.DefaultGalicianTagger)
	require.NotNil(t, taggl.DefaultGalicianTagger.GetWordTagger())
	require.NotEmpty(t, taggl.GalicianPOSDictPath(), "real galician.dict must load")

	disambiguator := loadGalicianMultiWordChunker(t)

	// Readings sorted like TestTools.getAsStrings; multiword markers from MultiWordChunker.
	// Entries confirmed in official multiwords.txt: abaixo de→SP000, á beira de→SP000,
	// a bordo→RG, a pesar de→CS, aínda que→CS, por enriba de→SP000, acerca de→SP000.
	cases := []struct {
		input string
		want  string
	}{
		{
			"abaixo de",
			"/[null]SENT_START abaixo/[abaixar]VMIP1S0|abaixo/[abaixo de]<SP000>|abaixo/[abaixo]RG  /[null]null de/[abaixo de]</SP000>|de/[de]NCMS000|de/[de]SPS00",
		},
		{
			"á beira de",
			"/[null]SENT_START á/[a]SPS00:DA0FS0|á/[á beira de]<SP000>|á/[á]NCFS000  /[null]null beira/[beira]NCFS000  /[null]null de/[de]NCMS000|de/[de]SPS00|de/[á beira de]</SP000>",
		},
		{
			"a bordo",
			"/[null]SENT_START a/[a bordo]<RG>|a/[a]SPS00|a/[o]DA0FS0|a/[o]PP3FSA00  /[null]null bordo/[a bordo]</RG>|bordo/[bordar]VMIP1S0|bordo/[bordo]NCMS000",
		},
		{
			"a pesar de",
			"/[null]SENT_START a/[a pesar de]<CS>|a/[a]SPS00|a/[o]DA0FS0|a/[o]PP3FSA00  /[null]null pesar/[pesar]VMN0000|pesar/[pesar]VMN01S0|pesar/[pesar]VMN03S0|pesar/[pesar]VMSF1S0|pesar/[pesar]VMSF3S0  /[null]null de/[a pesar de]</CS>|de/[de]NCMS000|de/[de]SPS00",
		},
		{
			"aínda que",
			"/[null]SENT_START aínda/[aínda que]<CS>|aínda/[aínda]CS|aínda/[aínda]RG  /[null]null que/[aínda que]</CS>|que/[que]CS|que/[que]DE0CN0|que/[que]DT0CN0|que/[que]NCMS000|que/[que]PE0CN000|que/[que]PR0CN000|que/[que]PT0CN000",
		},
		{
			"por enriba de",
			"/[null]SENT_START por/[por enriba de]<SP000>|por/[por]SPS00  /[null]null enriba/[enriba]RG  /[null]null de/[de]NCMS000|de/[de]SPS00|de/[por enriba de]</SP000>",
		},
		{
			"acerca de",
			"/[null]SENT_START acerca/[acerca de]<SP000>|acerca/[acercar]VMIP3S0|acerca/[acercar]VMM02S0  /[null]null de/[acerca de]</SP000>|de/[de]NCMS000|de/[de]SPS00",
		},
		{
			"abaixo de algo",
			"/[null]SENT_START abaixo/[abaixar]VMIP1S0|abaixo/[abaixo de]<SP000>|abaixo/[abaixo]RG  /[null]null de/[abaixo de]</SP000>|de/[de]NCMS000|de/[de]SPS00  /[null]null algo/[algo]PI0CN000|algo/[algo]RG",
		},
		{
			"está á beira de casa",
			"/[null]SENT_START está/[estar]VMIP3S0|está/[estar]VMM02S0  /[null]null á/[a]SPS00:DA0FS0|á/[á beira de]<SP000>|á/[á]NCFS000  /[null]null beira/[beira]NCFS000  /[null]null de/[de]NCMS000|de/[de]SPS00|de/[á beira de]</SP000>  /[null]null casa/[casa]NCFS000|casa/[casar]VMIP3S0|casa/[casar]VMM02S0",
		},
	}
	for _, tc := range cases {
		got := myAssertGalicianChunker(tc.input, disambiguator)
		require.Equal(t, tc.want, got, "input=%q", tc.input)
	}
}

// TestGalicianHybridDisambiguator_ChunkerLoad wires official multiwords into the hybrid
// (Java field: MultiWordChunker.getInstance("/gl/multiwords.txt")).
func TestGalicianHybridDisambiguator_ChunkerLoad(t *testing.T) {
	if taggl.DiscoverGalicianPOSDict() == "" {
		t.Skip("galician.dict not in tree")
	}
	chunker := loadGalicianMultiWordChunker(t)
	h := NewGalicianHybridDisambiguator()
	h.Chunker = chunker
	// multiwords-only path (Rules nil): same as chunker stage
	got := myAssertGalicianChunker("abaixo de", h)
	want := "/[null]SENT_START abaixo/[abaixar]VMIP1S0|abaixo/[abaixo de]<SP000>|abaixo/[abaixo]RG  /[null]null de/[abaixo de]</SP000>|de/[de]NCMS000|de/[de]SPS00"
	require.Equal(t, want, got)
}

// glMultiwordsPath resolves Java resource /gl/multiwords.txt under inspiration.
func glMultiwordsPath(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
	p := filepath.Join(dir,
		"inspiration/languagetool/languagetool-language-modules/gl/src/main/resources/org/languagetool/resource/gl/multiwords.txt")
	_, err = os.Stat(p)
	require.NoError(t, err, "Java /gl/multiwords.txt resource must exist")
	return p
}

// loadGalicianMultiWordChunker ports MultiWordChunker.getInstance("/gl/multiwords.txt")
// defaults: allowFirstCapitalized=false, allowAllUppercase=false, allowTitlecase=false.
func loadGalicianMultiWordChunker(t *testing.T) *disambiguation.MultiWordChunker {
	t.Helper()
	f, err := os.Open(glMultiwordsPath(t))
	require.NoError(t, err)
	defer f.Close()
	c, err := OpenGalicianMultiWordChunker(f)
	require.NoError(t, err)
	return c
}

// myAssertGalicianChunker ports Java TestTools.myAssert(input, expected, WordTokenizer,
// SRXSentenceTokenizer(Galician), GalicianTagger, MultiWordChunker).
// Format: token/[lemma]POS readings sorted and joined by '|', tokens joined by space;
// null lemma/POS print as the literal "null" (Java string concat of null).
func myAssertGalicianChunker(input string, dis disambiguation.Disambiguator) string {
	taggl.EnsureDefaultGalicianTagger()
	tagger := taggl.DefaultGalicianTagger
	wt := tokenizers.NewWordTokenizer()
	st := tokenizers.NewSRXSentenceTokenizer("gl")
	var out strings.Builder
	for _, sentence := range st.Tokenize(input) {
		tokens := wt.Tokenize(sentence)
		var noWS []string
		for _, tok := range tokens {
			if galicianTestToolsIsWord(tok) {
				noWS = append(noWS, tok)
			}
		}
		aTokens := tagger.Tag(noWS)
		tokenArray := make([]*languagetool.AnalyzedTokenReadings, 0, len(tokens)+1)
		ss := languagetool.SentenceStartTagName
		tokenArray = append(tokenArray, languagetool.NewAnalyzedTokenReadingsAt(
			languagetool.NewAnalyzedToken("", &ss, nil), 0))
		startPos := 0
		noWSCount := 0
		for _, tokenStr := range tokens {
			var posTag *languagetool.AnalyzedTokenReadings
			if galicianTestToolsIsWord(tokenStr) {
				posTag = aTokens[noWSCount]
				posTag.SetStartPos(startPos)
				noWSCount++
			} else {
				// Java BaseTagger.createNullToken / tagger.createNullToken
				posTag = languagetool.NewAnalyzedTokenReadingsAt(
					languagetool.NewAnalyzedToken(tokenStr, nil, nil), startPos)
			}
			tokenArray = append(tokenArray, posTag)
			startPos += tokenizers.UTF16Len(tokenStr)
		}
		finalSentence := languagetool.NewAnalyzedSentence(tokenArray)
		if dis != nil {
			finalSentence = dis.Disambiguate(finalSentence)
		}
		out.WriteString(formatGalicianMyAssertSentence(finalSentence))
	}
	return out.String()
}

// galicianTestToolsIsWord ports TestTools.isWord: any letter or digit → word token.
func galicianTestToolsIsWord(token string) bool {
	for _, r := range token {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

// formatGalicianMyAssertSentence ports TestTools.getAsStrings + join for one sentence.
func formatGalicianMyAssertSentence(sent *languagetool.AnalyzedSentence) string {
	if sent == nil {
		return ""
	}
	var parts []string
	for _, tr := range sent.GetTokens() {
		var readings []string
		for _, r := range tr.GetReadings() {
			if r != nil {
				readings = append(readings, galicianTestToolsGetAsString(r))
			}
		}
		// Java Collections.sort — force stable order across lexicon versions
		sort.Strings(readings)
		parts = append(parts, strings.Join(readings, "|"))
	}
	return strings.Join(parts, " ")
}

// galicianTestToolsGetAsString ports TestTools.getAsString: token/[lemma]POS with null literals.
func galicianTestToolsGetAsString(tok *languagetool.AnalyzedToken) string {
	lemma, pos := "null", "null"
	if tok.GetLemma() != nil {
		lemma = *tok.GetLemma()
	}
	if tok.GetPOSTag() != nil {
		pos = *tok.GetPOSTag()
	}
	return tok.GetToken() + "/[" + lemma + "]" + pos
}
