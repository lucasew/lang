package ru

// Russian MultiWordChunker outcome twin (no dedicated Java RussianDisambiguationRuleTest).
// Pattern: PolishDisambiguationRuleTest / SwedishDisambiguationRuleTest#testChunker /
// GalicianDisambiguationRuleTest (GL Go sector).
// Java: MultiWordChunker.getInstance("/ru/multiwords.txt") defaults (false,false,false)
// via RussianHybridDisambiguator.chunker.
// Resources: /ru/multiwords.txt, russian.dict via RussianTagger.
import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	tagru "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ru"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// TestRussianDisambiguationRule_Chunker is the MultiWordChunker stage twin for Russian.
// Java: MultiWordChunker.getInstance("/ru/multiwords.txt") → false,false,false defaults;
// WordTokenizer (core), SRXSentenceTokenizer(Russian/"ru"), RussianTagger; TestTools.myAssert.
// Phrases taken from official /ru/multiwords.txt (not invented).
func TestRussianDisambiguationRule_Chunker(t *testing.T) {
	if tagru.DiscoverRussianPOSDict() == "" {
		t.Skip("russian.dict not in tree")
	}
	tagru.EnsureDefaultRussianTagger()
	require.NotNil(t, tagru.DefaultRussianTagger)
	require.NotNil(t, tagru.DefaultRussianTagger.GetWordTagger())
	require.NotEmpty(t, tagru.RussianPOSDictPath(), "real russian.dict must load")

	disambiguator := loadRussianMultiWordChunker(t)

	// Readings sorted like TestTools.getAsStrings; multiword markers from MultiWordChunker.
	// Entries confirmed in official multiwords.txt: до мажор→NN:Masc, откуда ни возьмись→FR,
	// пиши пропал→FR, черт возьми→CONJ, будь здоров→ADV, в будущем→ADV, до свидания→ADV,
	// во что бы ни стало→ADV, в целом→ADV, Откуда ни возьмись→FR (capitalized form).
	cases := []struct {
		input string
		want  string
	}{
		{
			"до мажор",
			"/[null]SENT_START до/[до мажор]<NN:Masc>|до/[до]PREP  /[null]null мажор/[до мажор]</NN:Masc>|мажор/[мажор]NN:Inanim:Masc:Sin:Nom|мажор/[мажор]NN:Inanim:Masc:Sin:V",
		},
		{
			"до минор",
			"/[null]SENT_START до/[до минор]<NN:Masc>|до/[до]PREP  /[null]null минор/[до минор]</NN:Masc>|минор/[минор]NN:Inanim:Masc:Sin:Nom|минор/[минор]NN:Inanim:Masc:Sin:V",
		},
		{
			"откуда ни возьмись",
			"/[null]SENT_START откуда/[откуда ни возьмись]<FR>|откуда/[откуда]ADV  /[null]null ни/[ни]CONJ|ни/[ни]PARTICLE  /[null]null возьмись/[взяться]VB:IMP:INTR:PFV:Sin:P2|возьмись/[откуда ни возьмись]</FR>",
		},
		{
			"пиши пропал",
			"/[null]SENT_START пиши/[писать]VB:IMP:TRANS:IMPFV:Sin:P2|пиши/[пиши пропал]<FR>  /[null]null пропал/[пиши пропал]</FR>|пропал/[пропасть]VB:Past:INTR:PFV:Masc",
		},
		{
			"черт возьми",
			"/[null]SENT_START черт/[черт возьми]<CONJ>|черт/[черта]NN:Inanim:Fem:PL:R|черт/[чёрт]NN:Anim:Masc:Sin:Nom  /[null]null возьми/[взять]VB:IMP:TRANS:PFV:Sin:P2|возьми/[черт возьми]</CONJ>",
		},
		{
			"будь здоров",
			"/[null]SENT_START будь/[будь здоров]<ADV>|будь/[быть]VB:IMP:INTR:IMPFV:Sin:P2  /[null]null здоров/[будь здоров]</ADV>|здоров/[здоров]NN:Fam:Masc:Sin:Nom|здоров/[здоровый]ADJ:Short:Masc",
		},
		{
			"в будущем",
			"/[null]SENT_START в/[в будущем]<ADV>|в/[в]PREP  /[null]null будущем/[будущее]NN:Inanim:Neut:Sin:P|будущем/[будущий]ADJ:Posit:Masc:P|будущем/[будущий]ADJ:Posit:Neut:P|будущем/[в будущем]</ADV>",
		},
		{
			"до свидания",
			"/[null]SENT_START до/[до свидания]<ADV>|до/[до]PREP  /[null]null свидания/[до свидания]</ADV>|свидания/[свидание]NN:Inanim:Neut:PL:Nom|свидания/[свидание]NN:Inanim:Neut:PL:V|свидания/[свидание]NN:Inanim:Neut:Sin:R",
		},
		{
			"во что бы ни стало",
			"/[null]SENT_START во/[во что бы ни стало]<ADV>|во/[во]PREP  /[null]null что/[что]ADV|что/[что]CONJ|что/[что]PNN:Sin:Nom|что/[что]PNN:Sin:V  /[null]null бы/[бы]PARTICLE  /[null]null ни/[ни]CONJ|ни/[ни]PARTICLE  /[null]null стало/[во что бы ни стало]</ADV>|стало/[стать]VB:Past:INTR:PFV:Neut",
		},
		{
			"в целом",
			"/[null]SENT_START в/[в целом]<ADV>|в/[в]PREP  /[null]null целом/[в целом]</ADV>|целом/[целое]NN:Inanim:Neut:Sin:P|целом/[целый]ADJ:Posit:Masc:P|целом/[целый]ADJ:Posit:Neut:P",
		},
		{
			"до мажор и до минор",
			"/[null]SENT_START до/[до мажор]<NN:Masc>|до/[до]PREP  /[null]null мажор/[до мажор]</NN:Masc>|мажор/[мажор]NN:Inanim:Masc:Sin:Nom|мажор/[мажор]NN:Inanim:Masc:Sin:V  /[null]null и/[и]CONJ|и/[и]INTERJECTION  /[null]null до/[до минор]<NN:Masc>|до/[до]PREP  /[null]null минор/[до минор]</NN:Masc>|минор/[минор]NN:Inanim:Masc:Sin:Nom|минор/[минор]NN:Inanim:Masc:Sin:V",
		},
		{
			"Откуда ни возьмись",
			"/[null]SENT_START Откуда/[Откуда ни возьмись]<FR>|Откуда/[откуда]ADV  /[null]null ни/[ни]CONJ|ни/[ни]PARTICLE  /[null]null возьмись/[Откуда ни возьмись]</FR>|возьмись/[взяться]VB:IMP:INTR:PFV:Sin:P2",
		},
	}
	for _, tc := range cases {
		got := myAssertRussianChunker(tc.input, disambiguator)
		require.Equal(t, tc.want, got, "input=%q", tc.input)
	}
}

// TestRussianHybridDisambiguator_ChunkerLoad wires official multiwords into the hybrid
// (Java field: MultiWordChunker.getInstance("/ru/multiwords.txt")).
// Isolates the multiword stage (Rules nil) — XML stage is covered by rules/ru XmlRule tests.
func TestRussianHybridDisambiguator_ChunkerLoad(t *testing.T) {
	if tagru.DiscoverRussianPOSDict() == "" {
		t.Skip("russian.dict not in tree")
	}
	chunker := loadRussianMultiWordChunker(t)
	// Multiword-only path: do not run eager XmlRuleDisambiguator stage here.
	h := NewRussianHybridDisambiguatorWithStages(chunker, nil)
	got := myAssertRussianChunker("до мажор", h)
	want := "/[null]SENT_START до/[до мажор]<NN:Masc>|до/[до]PREP  /[null]null мажор/[до мажор]</NN:Masc>|мажор/[мажор]NN:Inanim:Masc:Sin:Nom|мажор/[мажор]NN:Inanim:Masc:Sin:V"
	require.Equal(t, want, got)
}

// ruMultiwordsPath resolves Java resource /ru/multiwords.txt under inspiration.
func ruMultiwordsPath(t *testing.T) string {
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
		"inspiration/languagetool/languagetool-language-modules/ru/src/main/resources/org/languagetool/resource/ru/multiwords.txt")
	_, err = os.Stat(p)
	require.NoError(t, err, "Java /ru/multiwords.txt resource must exist")
	return p
}

// loadRussianMultiWordChunker ports MultiWordChunker.getInstance("/ru/multiwords.txt")
// defaults: allowFirstCapitalized=false, allowAllUppercase=false, allowTitlecase=false.
func loadRussianMultiWordChunker(t *testing.T) *disambiguation.MultiWordChunker {
	t.Helper()
	c, err := LoadRussianMultiWordChunkerFromPath(ruMultiwordsPath(t))
	require.NoError(t, err)
	return c
}

// myAssertRussianChunker ports Java TestTools.myAssert(input, expected, WordTokenizer,
// SRXSentenceTokenizer(Russian), RussianTagger, MultiWordChunker).
// Format: token/[lemma]POS readings sorted and joined by '|', tokens joined by space;
// null lemma/POS print as the literal "null" (Java string concat of null).
func myAssertRussianChunker(input string, dis disambiguation.Disambiguator) string {
	tagru.EnsureDefaultRussianTagger()
	tagger := tagru.DefaultRussianTagger
	wt := tokenizers.NewWordTokenizer()
	st := tokenizers.NewSRXSentenceTokenizer("ru")
	var out strings.Builder
	for _, sentence := range st.Tokenize(input) {
		tokens := wt.Tokenize(sentence)
		var noWS []string
		for _, tok := range tokens {
			if russianTestToolsIsWord(tok) {
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
			if russianTestToolsIsWord(tokenStr) {
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
		out.WriteString(formatRussianMyAssertSentence(finalSentence))
	}
	return out.String()
}

// russianTestToolsIsWord ports TestTools.isWord: any letter or digit → word token.
func russianTestToolsIsWord(token string) bool {
	for _, r := range token {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

// formatRussianMyAssertSentence ports TestTools.getAsStrings + join for one sentence.
func formatRussianMyAssertSentence(sent *languagetool.AnalyzedSentence) string {
	if sent == nil {
		return ""
	}
	var parts []string
	for _, tr := range sent.GetTokens() {
		var readings []string
		for _, r := range tr.GetReadings() {
			if r != nil {
				readings = append(readings, russianTestToolsGetAsString(r))
			}
		}
		// Java Collections.sort — force stable order across lexicon versions
		sort.Strings(readings)
		parts = append(parts, strings.Join(readings, "|"))
	}
	return strings.Join(parts, " ")
}

// russianTestToolsGetAsString ports TestTools.getAsString: token/[lemma]POS with null literals.
func russianTestToolsGetAsString(tok *languagetool.AnalyzedToken) string {
	lemma, pos := "null", "null"
	if tok.GetLemma() != nil {
		lemma = *tok.GetLemma()
	}
	if tok.GetPOSTag() != nil {
		pos = *tok.GetPOSTag()
	}
	return tok.GetToken() + "/[" + lemma + "]" + pos
}
