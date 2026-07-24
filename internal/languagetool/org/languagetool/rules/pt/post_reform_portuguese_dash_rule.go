package pt

import (
	"embed"
	"sync"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/post-reform-compounds.txt
var postReformCompoundsFS embed.FS

var (
	postDashOnce     sync.Once
	postDashPatterns []string
)

func loadPostReformDashPatterns() []string {
	postDashOnce.Do(func() {
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

func isPortugueseLetter(r rune) bool {
	if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
		return true
	}
	// From Java PostReformPortugueseDashRule PATTERN + common accents.
	switch r {
	case 'Â', 'â', 'Ã', 'ã', 'Ç', 'ç', 'Ê', 'ê', 'Ó', 'ó', 'Ô', 'ô', 'Õ', 'õ',
		'ü', 'Ü', 'Á', 'á', 'À', 'à', 'É', 'é', 'Í', 'í', 'Ú', 'ú', 'È', 'è':
		return true
	}
	return unicode.Is(unicode.Latin, r) && unicode.IsLetter(r)
}

// PostReformPortugueseDashRule ports org.languagetool.rules.pt.PostReformPortugueseDashRule.
type PostReformPortugueseDashRule struct {
	*rules.AbstractDashRule
}

func NewPostReformPortugueseDashRule(messages map[string]string) *PostReformPortugueseDashRule {
	base := &rules.AbstractDashRule{
		ID:               "PT_POSAO_DASH_RULE",
		CompoundPatterns: loadPostReformDashPatterns(),
		Message:          "Um travessão foi utilizado em vez de um hífen.",
		Description:      "Travessões no lugar de hífens",
		IsLetter:         isPortugueseLetter,
	}
	rules.InitDashRuleMeta(base, messages)
	return &PostReformPortugueseDashRule{AbstractDashRule: base}
}

func (r *PostReformPortugueseDashRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractDashRule.Match(sentence)
}
