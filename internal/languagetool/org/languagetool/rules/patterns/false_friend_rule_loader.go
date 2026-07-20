package patterns

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
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

// Official EN MessagesBundle keys (Java FalseFriendRuleLoader(Language) loads mother-tongue
// MessagesBundle; when callers pass empty strings, use EN defaults — not invent 2-arg templates).
const (
	messagesFalseFriendHint = `Hint: "{0}" ({1}) means {2} ({3}).`
	messagesFalseFriendSugg = `Did you mean {0}?`
)

func NewFalseFriendRuleLoader(hint, sugg string) *FalseFriendRuleLoader {
	if hint == "" {
		hint = messagesFalseFriendHint
	}
	if sugg == "" {
		sugg = messagesFalseFriendSugg
	}
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
	Inflected     string `xml:"inflected,attr"`
	Regexp        string `xml:"regexp,attr"`
	CaseSensitive string `xml:"case_sensitive,attr"`
	Negate        string `xml:"negate,attr"`
	Postag        string `xml:"postag,attr"`
	PostagRegexp  string `xml:"postag_regexp,attr"`
	// SpaceBefore ports spacebefore="yes|no" (rare on false friends).
	SpaceBefore string `xml:"spacebefore,attr"`
	// Skip ports skip="N" (Java PatternToken skip; e.g. skip="-1" in false-friends.xml).
	Skip    string `xml:"skip,attr"`
	Content string `xml:",chardata"`
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
			// SENT_START and empty surface + postag-only tokens are valid Java
			cs := strings.EqualFold(xt.CaseSensitive, "yes")
			re := strings.EqualFold(xt.Regexp, "yes")
			inf := strings.EqualFold(xt.Inflected, "yes")
			pt := NewPatternToken(content, cs, re, inf)
			if strings.EqualFold(xt.Negate, "yes") {
				pt.SetNegation(true)
			}
			if pos := strings.TrimSpace(xt.Postag); pos != "" {
				pt.SetPosToken(PosToken{
					PosTag: pos,
					Regexp: strings.EqualFold(xt.PostagRegexp, "yes"),
					Negate: false,
				})
			}
			if sb := strings.TrimSpace(xt.SpaceBefore); sb != "" {
				pt.SetWhitespaceBefore(strings.EqualFold(sb, "yes"))
			}
			if sk := strings.TrimSpace(xt.Skip); sk != "" {
				if n, err := strconv.Atoi(sk); err == nil {
					pt.SetSkipNext(n)
				}
			}
			tokens = append(tokens, pt)
		}
		// Java FalseFriendRuleHandler: MessageFormat(falseFriendHint).format(
		//   tokensAsString, englishName(textLang), formatTranslations, englishName(mother))
		tokensStr := strings.ReplaceAll(tokensAsString(tokens), "|", "/")
		transStr := FormatTranslations(motherTranslations)
		// NewFalseFriendRuleLoader always sets non-empty hint/sugg (EN MessagesBundle defaults).
		h := NewFalseFriendRuleHandler(textLang, motherLang, l.FalseFriendHint)
		msg := h.FormatHint(tokensStr, englishLangName(textLang), transStr, englishLangName(motherLang))
		// Java appends MessageFormat(falseFriendSugg) with joined <suggestion> list
		suggMsg := l.FalseFriendSugg
		rule := NewFalseFriendPatternRule(id, textLang, tokens, "False friend: "+id, msg, "")
		var sb strings.Builder
		sb.WriteString(msg)
		// Prefer Java-style suggestion suffix when template present
		if strings.Contains(suggMsg, "{0}") {
			var formatted []string
			for _, s := range motherTranslations {
				if s == "" || strings.EqualFold(s, tokensStr) {
					continue
				}
				formatted = append(formatted, "<suggestion>"+s+"</suggestion>")
			}
			if len(formatted) > 0 {
				joined := strings.Join(formatted, ", ")
				sb.WriteString(" ")
				sb.WriteString(strings.ReplaceAll(suggMsg, "{0}", joined))
			}
		} else {
			for _, s := range motherTranslations {
				if s == "" {
					continue
				}
				sb.WriteString(" <suggestion>")
				sb.WriteString(s)
				sb.WriteString("</suggestion>")
			}
		}
		rule.Message = sb.String()
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

// englishLangName ports englishMessages.getString(lang.getShortCode()) for common codes.
func englishLangName(code string) string {
	switch baseLang(code) {
	case "en":
		return "English"
	case "de":
		return "German"
	case "fr":
		return "French"
	case "es":
		return "Spanish"
	case "nl":
		return "Dutch"
	case "pt":
		return "Portuguese"
	case "it":
		return "Italian"
	case "pl":
		return "Polish"
	case "ru":
		return "Russian"
	case "ca":
		return "Catalan"
	default:
		return code
	}
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
