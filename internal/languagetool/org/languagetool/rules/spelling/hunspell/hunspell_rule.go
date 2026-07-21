package hunspell

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// HunspellRuleID ports HunspellRule.RULE_ID.
const HunspellRuleID = "HUNSPELL_RULE"

// FileExtension ports HunspellRule.FILE_EXTENSION.
const FileExtension = ".dic"

var nonAlphabeticRE = regexp.MustCompile(`^[^\p{L}]+$`)

// HunspellRule ports org.languagetool.rules.spelling.hunspell.HunspellRule
// with a pluggable HunspellDictionary (native hunspell deferred).
type HunspellRule struct {
	*spelling.SpellingCheckRule
	Dict HunspellDictionary
	// IgnoreTaggedWords skips tokens that already have a real POS tag.
	IgnoreTaggedWords bool
}

func NewHunspellRule(languageCode string, dict HunspellDictionary) *HunspellRule {
	r := &HunspellRule{
		SpellingCheckRule: spelling.NewSpellingCheckRule(HunspellRuleID, "Possible spelling mistake", languageCode),
		Dict:              dict,
	}
	r.IsMisspelled = r.IsMisspelledWord
	// Java SpellingCheckRule.init: ignore/spelling/prohibit word lists for language.
	ApplyDefaultSpellingWordLists(r.SpellingCheckRule)
	return r
}

// IsMisspelledWord ports HunspellRule.isMisspelled.
func (r *HunspellRule) IsMisspelledWord(word string) bool {
	if r == nil || r.Dict == nil {
		return false
	}
	if nonAlphabeticRE.MatchString(word) {
		return false
	}
	return !r.Dict.Spell(word)
}

// Suggest ports dictionary suggestions (empty if dict has none).
func (r *HunspellRule) Suggest(word string) []string {
	if r == nil || r.Dict == nil {
		return nil
	}
	return r.Dict.Suggest(word)
}

// Match flags misspelled tokens in the analyzed sentence.
// Ports HunspellRule.match including wrong-split suggestions (thanky ou → thank you).
func (r *HunspellRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	if sentence == nil || r == nil {
		return nil, nil
	}
	// Java HunspellRule.match: getSentenceWithImmunization(sentence).
	work := sentence
	if r.SpellingCheckRule != nil {
		work = r.SpellingCheckRule.SentenceWithImmunization(sentence)
		r.SpellingCheckRule.MarkMultiWordIgnoreSpelling(work)
	}
	tokens := work.GetTokensWithoutWhitespace()
	var out []*rules.RuleMatch
	var prevTok *languagetool.AnalyzedTokenReadings
	for idx, tok := range tokens {
		// Java: immunized / ignoredBySpeller / isUrl / isEMail
		if spelling.CanBeIgnoredToken(tok) {
			prevTok = tok
			continue
		}
		// Java ignoreToken → ignoreWord
		if r.SpellingCheckRule != nil && r.IgnoreToken(tokens, idx) {
			prevTok = tok
			continue
		}
		w := tok.GetToken()
		// Sentence-end markers with no letters (e.g. bare ".") are skipped; the last
		// content word may still carry IsSentenceEnd in AnalyzePlain and must be checked.
		if w == "" || !hasLetter(w) {
			prevTok = tok
			continue
		}
		// Java getSentenceTextWithoutImmunizedTokens: stringForSpeller strips emoji etc.
		check := tools.StringForSpeller(w)
		check = strings.TrimSpace(check)
		if check == "" || !hasLetter(check) {
			prevTok = tok
			continue
		}
		if r.IgnoreTaggedWords && tok.IsTagged() {
			if r.SpellingCheckRule == nil || !r.IsProhibited(w) {
				prevTok = tok
				continue
			}
		}
		// AcceptWord is true when the word should not be flagged.
		if r.AcceptWord(check) {
			prevTok = tok
			continue
		}
		// Java HunspellRule.match: after isMisspelled, ignorePotentiallyMisspelledWord.
		if r.SpellingCheckRule != nil && r.IgnorePotentiallyMisspelledWord(check) {
			prevTok = tok
			continue
		}

		cleanWord := check
		if strings.HasSuffix(cleanWord, ".") {
			cleanWord = cleanWord[:len(cleanWord)-1]
		}

		// Java wrong-split: rejoin across space using UTF-16 substring/charAt
		if prevTok != nil {
			prevWord := tools.StringForSpeller(prevTok.GetToken())
			prevWord = strings.TrimSpace(prevWord)
			if prevWord != "" {
				if ws := r.tryWrongSplit(sentence, &out, prevWord, prevTok.GetStartPos(), check, tok.GetStartPos(), cleanWord); ws != nil {
					out = append(out, ws)
					prevTok = tok
					continue
				}
			}
		}

		m := rules.NewRuleMatch(r, sentence, tok.GetStartPos(), tok.GetEndPos(),
			"Possible spelling mistake found")
		if sug := r.Suggest(check); len(sug) > 0 {
			// Java: getAdditionalTopSuggestions then filterSuggestions
			if top := spelling.AdditionalTopSuggestions(sug, check); len(top) > 0 {
				sug = append(top, sug...)
			}
			if r.SpellingCheckRule != nil {
				sug = r.SpellingCheckRule.FilterSuggestions(sug)
			}
			if len(sug) > 0 {
				m.SetSuggestedReplacements(sug)
			}
		}
		out = append(out, m)
		prevTok = tok
	}
	return out, nil
}

// tryWrongSplit ports HunspellRule wrong-split join (UTF-16 charAt/substring).
// If a match is returned, the previous match covering prevFrom is removed when present.
func (r *HunspellRule) tryWrongSplit(
	sentence *languagetool.AnalyzedSentence,
	ruleMatches *[]*rules.RuleMatch,
	prevWord string,
	prevFrom int,
	word string,
	wordFrom int,
	cleanWord string,
) *rules.RuleMatch {
	if r == nil || prevWord == "" || word == "" {
		return nil
	}
	// "thanky ou" → "thank you"
	if pu := utf16.Encode([]rune(prevWord)); len(pu) >= 1 {
		sugg1a := string(utf16.Decode(pu[:len(pu)-1]))
		sugg1b := cutOffDotHun(string(utf16.Decode(pu[len(pu)-1:])) + word)
		if sugg1a != "" && sugg1b != "" &&
			!r.IsMisspelledWord(sugg1a) && !r.IsMisspelledWord(sugg1b) {
			return r.createWrongSplitMatch(sentence, ruleMatches, wordFrom, cleanWord, sugg1a, sugg1b, prevFrom)
		}
	}
	// "than kyou" → "thank you"
	if wu := utf16.Encode([]rune(word)); len(wu) > 1 {
		sugg2a := prevWord + string(utf16.Decode(wu[:1]))
		sugg2b := cutOffDotHun(string(utf16.Decode(wu[1:])))
		if sugg2a != "" && sugg2b != "" &&
			!r.IsMisspelledWord(sugg2a) && !r.IsMisspelledWord(sugg2b) {
			return r.createWrongSplitMatch(sentence, ruleMatches, wordFrom, cleanWord, sugg2a, sugg2b, prevFrom)
		}
	}
	return nil
}

// createWrongSplitMatch ports SpellingCheckRule.createWrongSplitMatch.
func (r *HunspellRule) createWrongSplitMatch(
	sentence *languagetool.AnalyzedSentence,
	ruleMatches *[]*rules.RuleMatch,
	pos int,
	coveredWord, suggestion1, suggestion2 string,
	prevPos int,
) *rules.RuleMatch {
	if ruleMatches != nil && len(*ruleMatches) > 0 {
		last := (*ruleMatches)[len(*ruleMatches)-1]
		if last != nil && last.GetFromPos() == prevPos {
			*ruleMatches = (*ruleMatches)[:len(*ruleMatches)-1]
		}
	}
	// Java: prevPos .. pos + coveredWord.length() (UTF-16)
	to := pos + len(utf16.Encode([]rune(coveredWord)))
	m := rules.NewRuleMatch(r, sentence, prevPos, to, "Possible spelling mistake found")
	m.SetType(rules.RuleMatchTypeUnknownWord)
	m.SetSuggestedReplacements([]string{strings.TrimSpace(suggestion1 + " " + suggestion2)})
	return m
}

func cutOffDotHun(s string) string {
	if strings.HasSuffix(s, ".") {
		return s[:len(s)-1]
	}
	return s
}

func hasLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}
