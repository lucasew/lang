package bitext

import (
	"io"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// FalseFriendsAsBitextLoader ports
// org.languagetool.rules.patterns.bitext.FalseFriendsAsBitextLoader.
// Pairs false-friend pattern rules that share an ID for motherTongue↔language.
type FalseFriendsAsBitextLoader struct {
	// LoadRules loads false-friend AbstractPatternRules for (textLang, motherLang) pairing.
	// When nil, uses patterns.FalseFriendRuleLoader.
	LoadRules func(r io.Reader, textLang, motherLang string) ([]*patterns.FalseFriendPatternRule, error)
}

func NewFalseFriendsAsBitextLoader() *FalseFriendsAsBitextLoader {
	return &FalseFriendsAsBitextLoader{}
}

// GetFalseFriendsAsBitext builds BitextPatternRules from two directional false-friend loads.
func (l *FalseFriendsAsBitextLoader) GetFalseFriendsAsBitext(
	xmlReader1, xmlReader2 io.Reader,
	motherTongue, language string,
) ([]*BitextPatternRule, error) {
	load := l.LoadRules
	if load == nil {
		load = func(r io.Reader, textLang, motherLang string) ([]*patterns.FalseFriendPatternRule, error) {
			return patterns.NewFalseFriendRuleLoader("", "").GetRulesFromReader(r, textLang, motherLang)
		}
	}
	// rules1: mother as text, language as mother (src side)
	rules1, err := load(xmlReader1, motherTongue, language)
	if err != nil {
		return nil, err
	}
	rules2, err := load(xmlReader2, language, motherTongue)
	if err != nil {
		return nil, err
	}
	srcByID := map[string]*patterns.FalseFriendPatternRule{}
	for _, r := range rules1 {
		if r == nil {
			continue
		}
		srcByID[r.GetID()] = r
	}
	var out []*BitextPatternRule
	for _, trg := range rules2 {
		if trg == nil {
			continue
		}
		src, ok := srcByID[trg.GetID()]
		if !ok {
			continue
		}
		// FalseFriendPatternRule embeds PatternRule which implements Match
		br := NewBitextPatternRule(src.PatternRule, trg.PatternRule)
		br.SetSourceLanguage(motherTongue)
		out = append(out, br)
	}
	return out, nil
}
