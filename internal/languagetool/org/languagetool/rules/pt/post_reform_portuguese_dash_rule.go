package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/post-reform-compounds.txt
var postReformCompoundsFS embed.FS

// postReformCompoundsResourcePath is the official Java resource path loaded by
// PostReformPortugueseDashRule.getCompoundsData (JLanguageTool resource dir).
const postReformCompoundsResourcePath = "/pt/post-reform-compounds.txt"

var (
	postDashOnce     sync.Once
	postDashPatterns []string
)

func loadPostReformDashPatterns() []string {
	postDashOnce.Do(func() {
		// Java: loadCompoundFile("/pt/post-reform-compounds.txt") once (DCL / sync.Once).
		f, err := postReformCompoundsFS.Open("data/post-reform-compounds.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		p, err := rules.LoadDashCompoundPatterns(f)
		if err != nil {
			panic(err)
		}
		postDashPatterns = p
	})
	return postDashPatterns
}

// isPortugueseLetter ports Java PostReformPortugueseDashRule / PreReformPortugueseDashRule
// isBoundary inverse:
//
//	PATTERN = [a-zA-ZÂâÃãÇçÊêÓóÔôÕõü]
//	isBoundary(s) = !PATTERN.matcher(s).matches()
//
// Exactly those code points — no invent extras (ÁàÉé…) and no unicode.IsLetter/Latin.
func isPortugueseLetter(r rune) bool {
	if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
		return true
	}
	switch r {
	case 'Â', 'â', 'Ã', 'ã', 'Ç', 'ç', 'Ê', 'ê', 'Ó', 'ó', 'Ô', 'ô', 'Õ', 'õ', 'ü':
		return true
	}
	return false
}

// PostReformPortugueseDashRule ports org.languagetool.rules.pt.PostReformPortugueseDashRule.
type PostReformPortugueseDashRule struct {
	*rules.AbstractDashRule
}

func NewPostReformPortugueseDashRule(messages map[string]string) *PostReformPortugueseDashRule {
	// Java: super(messages); setLocQualityIssueType(ITSIssueType.Typographical);
	// No setUrl / addExamplePair in Java PT post-reform dash rule.
	base := &rules.AbstractDashRule{
		ID:               "PT_POSAO_DASH_RULE",
		CompoundPatterns: loadPostReformDashPatterns(),
		Message:          "Um travessão foi utilizado em vez de um hífen.",
		Description:      "Travessões no lugar de hífens",
		IsLetter:         isPortugueseLetter,
		IssueType:        rules.ITSTypographical,
	}
	rules.InitDashRuleMeta(base, messages)
	return &PostReformPortugueseDashRule{AbstractDashRule: base}
}

func (r *PostReformPortugueseDashRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractDashRule.Match(sentence)
}
