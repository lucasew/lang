package ga

import (
	"bufio"
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/dative-plurals.txt
var dativePluralsFS embed.FS

var (
	dativeOnce sync.Once
	dativeMap  map[string][]string
)

// loadDativeSimple ports DativePluralsData.getSimpleReplacements without mutation variants
// (lenition/eclipse/h-prothesis). Base and modernised dative forms map to standard nominative.
func loadDativeSimple() map[string][]string {
	dativeOnce.Do(func() {
		f, err := dativePluralsFS.Open("data/dative-plurals.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m := map[string][]string{}
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || line[0] == '#' {
				continue
			}
			parts := strings.Split(line, ";")
			if len(parts) != 4 {
				continue
			}
			form, formModern := splitColon2(parts[0])
			repl, replModern := splitColon2(parts[3])
			standard := repl
			if replModern != "" {
				standard = replModern
			}
			m[form] = []string{standard}
			if formModern != "" {
				m[formModern] = []string{standard}
			}
		}
		if err := sc.Err(); err != nil {
			panic(err)
		}
		dativeMap = m
	})
	return dativeMap
}

func splitColon2(s string) (a, b string) {
	if i := strings.IndexByte(s, ':'); i >= 0 {
		return s[:i], s[i+1:]
	}
	return s, ""
}

// DativePluralStandardReplaceRule ports org.languagetool.rules.ga.DativePluralStandardReplaceRule.
type DativePluralStandardReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

func NewDativePluralStandardReplaceRule(messages map[string]string) *DativePluralStandardReplaceRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadDativeSimple(),
		CaseSensitive: false,
		CheckLemmas:   false,
		ID:            "GA_DATIVE_PLURALS_STD",
		Description:   "Tuiseal tabharthach iolra",
		ShortMsg:      "Tabharthach iolra",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "Is litriú réamhchaighdeánach (tabharthach iolra) é \"" + tokenStr + "\"; an \"" +
				joinCommaGA(replacements) + "\" a bhí i gceist agat?"
		},
	}
	return &DativePluralStandardReplaceRule{AbstractSimpleReplaceRule: base}
}

func (r *DativePluralStandardReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
