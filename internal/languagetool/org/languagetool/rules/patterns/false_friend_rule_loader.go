package patterns

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// FalseFriendRuleLoader ports org.languagetool.rules.patterns.FalseFriendRuleLoader
// for a simplified false-friends XML subset.
type FalseFriendRuleLoader struct {
	FalseFriendHint string
	FalseFriendSugg string
	// SuggestionMap is rule ID → translation strings (mother-tongue side).
	SuggestionMap map[string][]string
}

func NewFalseFriendRuleLoader(hint, sugg string) *FalseFriendRuleLoader {
	return &FalseFriendRuleLoader{
		FalseFriendHint: hint,
		FalseFriendSugg: sugg,
		SuggestionMap:   map[string][]string{},
	}
}

// GetRulesFromReader loads FalseFriendPatternRules for textLang when mother tongue is motherLang.
// Only rule entries whose pattern lang matches textLang and that have a translation for motherLang
// are returned (same pairing idea as Java FalseFriendRuleHandler).
func (l *FalseFriendRuleLoader) GetRulesFromReader(r io.Reader, textLang, motherLang string) ([]*FalseFriendPatternRule, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return l.parse(data, textLang, motherLang)
}

func (l *FalseFriendRuleLoader) GetRulesFromString(xmlStr, textLang, motherLang string) ([]*FalseFriendPatternRule, error) {
	return l.GetRulesFromReader(strings.NewReader(xmlStr), textLang, motherLang)
}

type ffRoot struct {
	XMLName    xml.Name      `xml:"rules"`
	RuleGroups []ffRuleGroup `xml:"rulegroup"`
	Rules      []ffRule      `xml:"rule"`
}

type ffRuleGroup struct {
	ID    string   `xml:"id,attr"`
	Rules []ffRule `xml:"rule"`
}

type ffRule struct {
	ID           string          `xml:"id,attr"`
	Pattern      ffPattern       `xml:"pattern"`
	Translations []ffTranslation `xml:"translation"`
}

type ffPattern struct {
	Lang   string    `xml:"lang,attr"`
	Tokens []ffToken `xml:"token"`
}

type ffToken struct {
	Inflected string `xml:"inflected,attr"`
	Regexp    string `xml:"regexp,attr"`
	Content   string `xml:",chardata"`
}

type ffTranslation struct {
	Lang    string `xml:"lang,attr"`
	Content string `xml:",chardata"`
}

func (l *FalseFriendRuleLoader) parse(data []byte, textLang, motherLang string) ([]*FalseFriendPatternRule, error) {
	var root ffRoot
	if err := xml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("false friends XML: %w", err)
	}
	textLang = baseLang(textLang)
	motherLang = baseLang(motherLang)
	l.SuggestionMap = map[string][]string{}

	var out []*FalseFriendPatternRule
	process := func(groupID string, xr ffRule) {
		id := xr.ID
		if id == "" {
			id = groupID
		}
		if id == "" {
			return
		}
		patLang := baseLang(xr.Pattern.Lang)
		// collect mother-tongue translations for suggestion map
		var motherTranslations []string
		for _, tr := range xr.Translations {
			if baseLang(tr.Lang) == motherLang {
				motherTranslations = append(motherTranslations, strings.TrimSpace(tr.Content))
			}
		}
		if len(motherTranslations) > 0 {
			l.SuggestionMap[id] = append(l.SuggestionMap[id], motherTranslations...)
		}
		// only emit rule when pattern is in text language and mother translations exist
		if patLang != textLang || len(motherTranslations) == 0 {
			return
		}
		if patLang == motherLang {
			return
		}
		var tokens []*PatternToken
		for _, xt := range xr.Pattern.Tokens {
			content := strings.TrimSpace(xt.Content)
			pt := NewPatternToken(content, false, strings.EqualFold(xt.Regexp, "yes"), strings.EqualFold(xt.Inflected, "yes"))
			tokens = append(tokens, pt)
		}
		// format message with hint template placeholders {0}=pattern, {1}=translations, {2}=mother name
		msg := l.FalseFriendHint
		if msg == "" {
			msg = "Possible false friend: {0} → {1}"
		}
		// Java FalseFriendRuleHandler.formatTranslations: "a", "b"
		msg = strings.ReplaceAll(msg, "{0}", tokensAsString(tokens))
		msg = strings.ReplaceAll(msg, "{1}", formatFFTranslations(motherTranslations))
		msg = strings.ReplaceAll(msg, "{2}", motherLang)
		// suggestions
		suggMsg := l.FalseFriendSugg
		if suggMsg == "" {
			suggMsg = strings.Join(motherTranslations, ", ")
		}
		rule := NewFalseFriendPatternRule(id, textLang, tokens, "False friend: "+id, msg, "")
		// stash suggestions on message suffix as <suggestion> for matcher consumers
		var sb strings.Builder
		sb.WriteString(msg)
		for _, s := range motherTranslations {
			if s == "" {
				continue
			}
			sb.WriteString(" <suggestion>")
			sb.WriteString(s)
			sb.WriteString("</suggestion>")
		}
		rule.Message = sb.String()
		_ = suggMsg
		out = append(out, rule)
	}
	for _, g := range root.RuleGroups {
		for _, xr := range g.Rules {
			process(g.ID, xr)
		}
	}
	for _, xr := range root.Rules {
		process(xr.ID, xr)
	}
	return out, nil
}

func baseLang(code string) string {
	code = strings.TrimSpace(code)
	if i := strings.IndexByte(code, '-'); i >= 0 {
		return code[:i]
	}
	return code
}

func tokensAsString(tokens []*PatternToken) string {
	parts := make([]string, 0, len(tokens))
	for _, t := range tokens {
		if t != nil {
			parts = append(parts, t.Token)
		}
	}
	return strings.Join(parts, " ")
}

// formatFFTranslations ports FalseFriendRuleHandler.formatTranslations:
// each translation wrapped in quotes, joined by ", ".
func formatFFTranslations(trs []string) string {
	parts := make([]string, 0, len(trs))
	for _, t := range trs {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		parts = append(parts, `"`+t+`"`)
	}
	return strings.Join(parts, ", ")
}
