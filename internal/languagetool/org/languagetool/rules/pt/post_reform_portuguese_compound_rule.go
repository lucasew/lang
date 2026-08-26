package pt

import (
	"embed"
	"regexp"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

//go:embed data/post-reform-compounds.txt
var postReformFS embed.FS

// PostReformPortugueseCompoundFilePath is the official resource path Java loads
// (CompoundRuleData("/pt/post-reform-compounds.txt")).
const PostReformPortugueseCompoundFilePath = "/pt/post-reform-compounds.txt"

// PostReformPortugueseCompoundURL ports PostReformPortugueseCompoundRule.getUrl()
// → Tools.getUrl("https://pt.wikipedia.org/wiki/Lista_das_alterações_previstas_pelo_acordo_ortográfico_de_1990").
const PostReformPortugueseCompoundURL = "https://pt.wikipedia.org/wiki/Lista_das_alterações_previstas_pelo_acordo_ortográfico_de_1990"

// Java: private static final Pattern VOWEL = compile("(?i).+[aeiou]$");
// Java: private static final Pattern RS = compile("(?i)^[rs].+");
var (
	postReformVowelRE = regexp.MustCompile(`(?i).+[aeiou]$`)
	postReformRSRE    = regexp.MustCompile(`(?i)^[rs].+`)
)

var (
	postReformOnce sync.Once
	postReformData *rules.CompoundRuleData
)

func loadPostReformCompounds() *rules.CompoundRuleData {
	postReformOnce.Do(func() {
		f, err := postReformFS.Open("data/post-reform-compounds.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		d, err := rules.NewCompoundRuleData(f, PostReformPortugueseCompoundFilePath)
		if err != nil {
			panic(err)
		}
		postReformData = d
	})
	return postReformData
}

// PostReformPortugueseCompoundRule ports org.languagetool.rules.pt.PostReformPortugueseCompoundRule.
type PostReformPortugueseCompoundRule struct {
	*rules.AbstractCompoundRule
}

// NewPostReformPortugueseCompoundRule constructs PT_COMPOUNDS_POST_REFORM.
// Java ctor: four Portuguese messages to AbstractCompoundRule, then
// setCategory(COMPOUNDING), setLocQualityIssueType(Grammar), useSubRuleSpecificIds().
// mergeCompound override is wired via MergeCompoundFn (Java virtual dispatch).
func NewPostReformPortugueseCompoundRule(messages map[string]string) *PostReformPortugueseCompoundRule {
	base := &rules.AbstractCompoundRule{
		Messages:                   messages,
		ID:                         "PT_COMPOUNDS_POST_REFORM",
		Description:                "Erro na formação da palavra composta \"$match\"",
		WithHyphenMessage:          "Esta palavra é hifenizada.",
		WithoutHyphenMessage:       "Esta palavra é composta por justaposição.",
		WithOrWithoutHyphenMessage: "Esta palavra pode ser composta por justaposição ou hifenizada.",
		ShortDesc:                  "Este conjunto forma uma palavra composta.",
		// Java: super.setCategory(Categories.COMPOUNDING.getCategory(messages));
		Category: rules.CatCompounding.GetCategory(messages),
		// Java: setLocQualityIssueType(ITSIssueType.Grammar);
		IssueType: rules.ITSGrammar,
		// Java: getUrl → Tools.getUrl(...)
		URL: PostReformPortugueseCompoundURL,
		// Java Language.getShortCode for StringTools.toId (pt-BR → "pt").
		LanguageCode: "pt",
		// Java AbstractCompoundRule.sentenceStartsWithUpperCase default true
		// (needed so "Super estrela" matches data entry "super estrela").
		SentenceStartsWithUpperCase: true,
		Data:                       loadPostReformCompounds(),
	}
	// Java virtual mergeCompound override (digraph r/s after vowel).
	base.MergeCompoundFn = postReformMergeCompound
	// Java: useSubRuleSpecificIds();
	base.UseSubRuleSpecificIDs()
	return &PostReformPortugueseCompoundRule{AbstractCompoundRule: base}
}

func (r *PostReformPortugueseCompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractCompoundRule.Match(sentence)
}

// postReformMergeCompound ports PostReformPortugueseCompoundRule.mergeCompound:
// if previous part ends in vowel and current starts with r/s, double the first letter
// (ultra+som → ultrassom; anti+semita → antissemita; arqui+rabino → arquirrabino).
func postReformMergeCompound(str string, uncapitalizeMidWords bool) string {
	stringParts := strings.Split(strings.ReplaceAll(str, "-", " "), " ")
	var sb strings.Builder
	for k := 0; k < len(stringParts); k++ {
		if k == 0 {
			sb.WriteString(stringParts[0])
			continue
		}
		// Java: if (VOWEL.matcher(stringParts[k - 1]).matches() && RS.matcher(stringParts[k]).matches())
		if postReformVowelRE.MatchString(stringParts[k-1]) && postReformRSRE.MatchString(stringParts[k]) {
			// Java: stringParts[k] = stringParts[k].charAt(0) + stringParts[k];
			// charAt(0) is UTF-16 unit; r/s are BMP so rune/byte first is fine.
			part := stringParts[k]
			runes := []rune(part)
			stringParts[k] = string(runes[0]) + part
		}
		if uncapitalizeMidWords {
			// Java: StringUtils.uncapitalize(stringParts[k])
			sb.WriteString(tools.LowercaseFirstChar(stringParts[k]))
		} else {
			sb.WriteString(stringParts[k])
		}
	}
	return sb.String()
}
