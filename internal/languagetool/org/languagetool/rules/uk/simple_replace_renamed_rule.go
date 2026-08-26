package uk

import (
	"embed"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	taguk "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/uk"
)

// Official rules/uk/replace_renamed.txt (byte-identical to Java
// ExtraDictionaryLoader.loadLists("/uk/replace_renamed.txt")).
//
//go:embed data/replace_renamed.txt
var renamedFS embed.FS

var (
	renamedOnce sync.Once
	// renamedMap ports SimpleReplaceRenamedRule.RENAMED_LIST (LoadLists order of values).
	renamedMap map[string][]string

	// geoPostagPattern ports SimpleReplaceRenamedRule.GEO_POSTAG_PATTERN:
	// Pattern.compile("noun:inanim.*?:prop.*|adj.*") with Matcher.matches().
	geoPostagPattern = regexp.MustCompile(`noun:inanim.*?:prop.*|adj.*`)

	// decomunizationURL ports SimpleReplaceRenamedRule.DECOMUNIZATION_URL.
	decomunizationURL = mustParseURL("https://uk.wikipedia.org/wiki/%D0%A1%D0%BF%D0%B8%D1%81%D0%BE%D0%BA_%D1%82%D0%BE%D0%BF%D0%BE%D0%BD%D1%96%D0%BC%D1%96%D0%B2_%D0%A3%D0%BA%D1%80%D0%B0%D1%97%D0%BD%D0%B8,_%D0%BF%D0%B5%D1%80%D0%B5%D0%B9%D0%BC%D0%B5%D0%BD%D0%BE%D0%B2%D0%B0%D0%BD%D0%B8%D1%85_%D0%B2%D0%BD%D0%B0%D1%81%D0%BB%D1%96%D0%B4%D0%BE%D0%BA_%D0%B4%D0%B5%D0%BA%D0%BE%D0%BC%D1%83%D0%BD%D1%96%D0%B7%D0%B0%D1%86%D1%96%D1%97")
)

func mustParseURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

// loadRenamed ports static RENAMED_LIST = ExtraDictionaryLoader.loadLists("/uk/replace_renamed.txt").
func loadRenamed() map[string][]string {
	renamedOnce.Do(func() {
		f, err := renamedFS.Open("data/replace_renamed.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := LoadLists(f)
		if err != nil {
			panic(err)
		}
		renamedMap = m
	})
	return renamedMap
}

// SimpleReplaceRenamedRule ports org.languagetool.rules.uk.SimpleReplaceRenamedRule.
// Match is lemma + GEO POS only (Java PosTagHelper.hasPosTag + RENAMED_LIST);
// without tags/lemmas, fails closed like untagged analysis (no surface invent).
// Java: setLocQualityIssueType(ITSIssueType.Style).
type SimpleReplaceRenamedRule struct {
	messages  map[string]string
	IssueType rules.ITSIssueType
}

// NewSimpleReplaceRenamedRule ports SimpleReplaceRenamedRule(ResourceBundle).
func NewSimpleReplaceRenamedRule(messages map[string]string) *SimpleReplaceRenamedRule {
	_ = loadRenamed()
	return &SimpleReplaceRenamedRule{
		messages:  messages,
		IssueType: rules.ITSStyle,
	}
}

func (r *SimpleReplaceRenamedRule) GetID() string { return "UK_SIMPLE_REPLACE_RENAMED" }

func (r *SimpleReplaceRenamedRule) GetDescription() string {
	return "Пропозиція поточної назви для перейменованих власних назв"
}

// GetLocQualityIssueType ports Rule.getLocQualityIssueType (Java Style).
func (r *SimpleReplaceRenamedRule) GetLocQualityIssueType() rules.ITSIssueType {
	if r == nil || r.IssueType == "" {
		return rules.ITSStyle
	}
	return r.IssueType
}

// Match ports SimpleReplaceRenamedRule.match.
func (r *SimpleReplaceRenamedRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || sentence == nil {
		return nil
	}
	m := loadRenamed()
	tokens := sentence.GetTokensWithoutWhitespace()
	var ruleMatches []*rules.RuleMatch

	for _, tokenReadings := range tokens {
		if tokenReadings == nil {
			continue
		}

		// Java LinkedHashSet: insertion order of first-seen renamed lemmas.
		var renamedLemmas []string
		seen := map[string]bool{}
		overlapClear := false

		for _, reading := range tokenReadings.GetReadings() {
			if reading == nil {
				continue
			}
			// Java: if SENTENCE_END_TAGNAME.equals(reading.getPOSTag()) continue;
			pos := reading.GetPOSTag()
			if pos != nil && *pos == languagetool.SentenceEndTagName {
				continue
			}

			lemmaPtr := reading.GetLemma()
			// Java: if (lemma != null) { ... } — null lemma: skip reading, no clear.
			if lemmaPtr == nil {
				continue
			}
			lemma := *lemmaPtr

			// Java: RENAMED_LIST.containsKey(lemma) && PosTagHelper.hasPosTag(reading, GEO_POSTAG_PATTERN)
			if _, inList := m[lemma]; inList && taguk.HasPosTagToken(reading, geoPostagPattern) {
				if !seen[lemma] {
					seen[lemma] = true
					renamedLemmas = append(renamedLemmas, lemma)
				}
			} else {
				// overlaps with normal lemma
				renamedLemmas = nil
				overlapClear = true
				break
			}
		}

		if overlapClear || len(renamedLemmas) == 0 {
			continue
		}

		info := ""
		var replacements []string
		for _, lemma := range renamedLemmas {
			repl := m[lemma]
			if len(repl) == 0 {
				// Java: repl.get(0) would throw; official lists always have ≥1 value field.
				continue
			}
			// Java: replacements.add(repl.get(0));
			replacements = append(replacements, repl[0])
			// Java: for (int i = 1; i < repl.size() - 1; i++) replacements.add(repl.get(i));
			for i := 1; i < len(repl)-1; i++ {
				replacements = append(replacements, repl[i])
			}
			// kinda cheating - getting first explanation we find
			if info == "" && len(repl) > 1 {
				info = repl[len(repl)-1]
			}
		}

		// Java createRuleMatch(tokenReadings, replacements, renamedLemmas.iterator().next(), info, sentence)
		msgLemma := renamedLemmas[0]
		rm := r.createRuleMatch(tokenReadings, replacements, msgLemma, info, sentence)
		ruleMatches = append(ruleMatches, rm)
	}
	return ruleMatches
}

// getMessage ports SimpleReplaceRenamedRule.getMessage.
func (r *SimpleReplaceRenamedRule) getMessage(tokenStr, info string) string {
	msg := "«" + tokenStr + "» було перейменовано"
	if info != "" {
		msg += " (" + info + ")"
	}
	return msg
}

// createRuleMatch ports SimpleReplaceRenamedRule.createRuleMatch.
// Note: Java param is named `msg` but holds the first renamed lemma (tokenStr for getMessage).
func (r *SimpleReplaceRenamedRule) createRuleMatch(
	readings *languagetool.AnalyzedTokenReadings,
	replacements []string,
	msgLemma string,
	info string,
	sentence *languagetool.AnalyzedSentence,
) *rules.RuleMatch {
	msg := r.getMessage(msgLemma, info)
	// Java: new RuleMatch(this, sentence, start, end, msg, "Перейменована назва")
	potential := rules.NewRuleMatch(r, sentence, readings.GetStartPos(), readings.GetEndPos(), msg)
	potential.ShortMessage = "Перейменована назва"
	potential.SetSuggestedReplacements(replacements)
	if strings.Contains(info, "декомуніз") {
		// Java: potentialRuleMatch.setUrl(DECOMUNIZATION_URL)
		potential.URL = decomunizationURL.String()
	}
	return potential
}
