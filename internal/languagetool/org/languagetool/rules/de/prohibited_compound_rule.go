package de

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// prohibitedPair ports ProhibitedCompoundRule.Pair (all-lowercase parts in source list).
type prohibitedPair struct {
	part1, desc1, part2, desc2 string
}

// ProhibitedCompoundRule ports org.languagetool.rules.de.ProhibitedCompoundRule.
// Language model: Frequency[token] count (Java BaseLanguageModel.getCount).
// Without Frequency, Match fails closed (Java requires LanguageModel).
type ProhibitedCompoundRule struct {
	Messages map[string]string
	// Frequency ports BaseLanguageModel.getCount(word); nil → fail-closed Match.
	Frequency map[string]int64
	// IsMisspelled optional; nil uses FilterDictIsMisspelled.
	IsMisspelled func(word string) bool
	// Blacklist extra words (compound_exceptions.txt); optional.
	Blacklist map[string]struct{}
	// Premium / Category / IssueType / Tags mirror Rule metadata for SpecificIdRule (Java isPremium/getCategory/…).
	Premium   bool
	Category  *rules.Category
	IssueType rules.ITSIssueType
	Tags      []rules.Tag
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

var herrnFrauRE = regexp.MustCompile(`^(Herrn?|Frau|Dr|Prof|Mag|Hr|Fr|Mr|Mrs|Ms|Fräulein)$`)

var prohibitedBlacklistRegexes = []*regexp.Regexp{
	regexp.MustCompile(`Lande(basis|basen|region|gebiets?|gebieten?|regionen|betriebs?|betrieben?|offizieren?|bereichs?|bereichen?|einrichtung|einrichtungen|massen?|plans?|versuchs?|versuchen?)`),
	regexp.MustCompile(`Model(vertrags?|verträgen?|erfahrung|erfahrungen|szene|welt)`),
	regexp.MustCompile(`(Raum|Surf|Jazz|Herbst|Gymnastik|Normal)schuhen?`),
	regexp.MustCompile(`preis`),
	regexp.MustCompile(`reisähnlich(e|e[nmrs])?`),
	regexp.MustCompile(`neugestartet(e|e[nmrs])?`),
	regexp.MustCompile(`reisender`),
	regexp.MustCompile(`[a-zöäüß]+sender`),
	regexp.MustCompile(`gra(ph|f)ische?`),
	regexp.MustCompile(`gra(ph|f)ische[rsnm]`),
	regexp.MustCompile(`gra(ph|f)s?$`),
	regexp.MustCompile(`gra(ph|f)en`),
	regexp.MustCompile(`gra(ph|f)in`),
	regexp.MustCompile(`gra(ph|f)ik`),
	regexp.MustCompile(`gra(ph|f)ie`),
	regexp.MustCompile(`Gra(ph|f)its?`),
	regexp.MustCompile(`.+gra(ph|f)its?`),
}

func NewProhibitedCompoundRule(messages map[string]string) *ProhibitedCompoundRule {
	r := &ProhibitedCompoundRule{
		Messages: messages,
		// Java: super.setCategory(Categories.TYPOS.getCategory(messages)); ITSIssueType.Uncategorized default.
		Category:  rules.CatTypos.GetCategory(messages),
		IssueType: rules.ITSUncategorized,
	}
	// Java: Lehrzeile → Leerzeile
	r.AddExamplePair(
		rules.Wrong("Da steht eine <marker>Lehrzeile</marker> zu viel."),
		rules.Fixed("Da steht eine <marker>Leerzeile</marker> zu viel."),
	)
	return r
}

// AddExamplePair ports Rule.addExamplePair.
func (r *ProhibitedCompoundRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *ProhibitedCompoundRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *ProhibitedCompoundRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// NewProhibitedCompoundRuleWithFrequency ports FakeLanguageModel constructor path.
func NewProhibitedCompoundRuleWithFrequency(messages map[string]string, freq map[string]int64) *ProhibitedCompoundRule {
	r := NewProhibitedCompoundRule(messages)
	r.Frequency = freq
	return r
}

func (r *ProhibitedCompoundRule) GetID() string { return "DE_PROHIBITED_COMPOUNDS" }

func (r *ProhibitedCompoundRule) GetDescription() string {
	return "Markiert wahrscheinlich falsche Komposita wie 'Lehrzeile', wenn 'Leerzeile' häufiger vorkommt."
}

func (r *ProhibitedCompoundRule) getCount(word string) int64 {
	if r == nil || r.Frequency == nil {
		return 0
	}
	return r.Frequency[word]
}

func (r *ProhibitedCompoundRule) isMisspelled(word string) bool {
	if r != nil && r.IsMisspelled != nil {
		return r.IsMisspelled(word)
	}
	return FilterDictIsMisspelled(word)
}

func (r *ProhibitedCompoundRule) getThreshold() int64 { return 0 }

func removeHyphensAndAdaptCase(word string) string {
	// Java returns null when no hyphens or short parts; Go uses "" for none
	if !strings.Contains(word, "-") {
		return ""
	}
	parts := strings.Split(word, "-")
	for _, p := range parts {
		if utf16LenDE(p) <= 1 {
			return ""
		}
	}
	var b strings.Builder
	for i, p := range parts {
		if i == 0 {
			b.WriteString(p)
			continue
		}
		b.WriteString(tools.LowercaseFirstChar(p))
	}
	return b.String()
}

func (r *ProhibitedCompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil || r == nil || r.Frequency == nil {
		// Java requires LanguageModel; without Frequency fail closed (no Prefer invent).
		return nil
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	var ruleMatches []*rules.RuleMatch
	var prev *languagetool.AnalyzedTokenReadings
	for _, readings := range tokens {
		if readings == nil {
			continue
		}
		tmpWord := readings.GetToken()
		if prev != nil && prev.HasPartialPosTag("EIG:") && tools.StartsWithUppercase(tmpWord) &&
			(readings.HasPartialPosTag("EIG:") || !readings.IsTagged()) {
			// assume name (e.g. Bianca Baalhorn); isPosTagUnknown → !IsTagged for plain
			// With AnalyzePlain IsTagged false always — skip only when prev EIG (morph).
			if prev.HasPartialPosTag("EIG:") {
				prev = readings
				continue
			}
		}
		if prev != nil && herrnFrauRE.MatchString(prev.GetToken()) {
			prev = readings
			continue
		}
		wordsParts := strings.Split(tmpWord, "-")
		partsStartPos := 0
		for _, wordPart := range wordsParts {
			partsStartPos = r.getMatches(sentence, &ruleMatches, readings, partsStartPos, wordPart, 0)
		}
		noHyphens := removeHyphensAndAdaptCase(tmpWord)
		if noHyphens != "" {
			r.getMatches(sentence, &ruleMatches, readings, 0, noHyphens, utf16LenDE(tmpWord)-utf16LenDE(noHyphens))
		}
		prev = readings
	}
	return ruleMatches
}

func blacklistMatch(wordPart string) bool {
	for _, re := range prohibitedBlacklistRegexes {
		if re.MatchString(wordPart) {
			return true
		}
	}
	return false
}

func (r *ProhibitedCompoundRule) getMatches(
	sentence *languagetool.AnalyzedSentence,
	ruleMatches *[]*rules.RuleMatch,
	readings *languagetool.AnalyzedTokenReadings,
	partsStartPos int,
	wordPart string,
	toPosCorrection int,
) int {
	// Java: tagged non-SUB (except EIG) skip; length <= 6 skip
	if readings.IsTagged() && !readings.HasPartialPosTag("SUB") && !readings.HasPosTagStartingWith("EIG:") ||
		utf16LenDE(wordPart) <= 6 {
		// AnalyzePlain: IsTagged false → don't skip by POS
		if readings.IsTagged() && !readings.HasPartialPosTag("SUB") && !readings.HasPosTagStartingWith("EIG:") {
			return partsStartPos + utf16LenDE(wordPart) + 1
		}
		if utf16LenDE(wordPart) <= 6 {
			return partsStartPos + utf16LenDE(wordPart) + 1
		}
	}

	type cand struct {
		pair    prohibitedPair
		variant string
		weight  int64
	}
	var weighted []cand

	// Collect candidate pairs where either part appears in wordPart
	for _, pair := range AllProhibitedPairs() {
		// case variants: try lc pair and first-upper variants
		variants := []prohibitedPair{pair}
		uc1, uc2 := tools.UppercaseFirstChar(pair.part1), tools.UppercaseFirstChar(pair.part2)
		if pair.part1 != uc1 || pair.part2 != uc2 {
			variants = append(variants, prohibitedPair{uc1, pair.desc1, uc2, pair.desc2})
		}
		for _, p := range variants {
			var variant string
			if strings.Contains(wordPart, p.part1) {
				variant = strings.Replace(wordPart, p.part1, p.part2, 1)
			} else if strings.Contains(wordPart, p.part2) {
				variant = strings.Replace(wordPart, p.part2, p.part1, 1)
			} else {
				continue
			}
			if variant == "" || variant == wordPart {
				continue
			}
			wordCount := r.getCount(wordPart)
			variantCount := r.getCount(variant)

			if r.isBlacklistedWord(wordPart) {
				continue
			}
			if variantCount > r.getThreshold() && wordCount == 0 &&
				!blacklistMatch(wordPart) && !r.isMisspelled(variant) {
				weighted = append(weighted, cand{pair: p, variant: variant, weight: variantCount})
			}
		}
	}

	if len(weighted) > 0 {
		// sort by weight desc
		for i := 0; i < len(weighted); i++ {
			for j := i + 1; j < len(weighted); j++ {
				if weighted[j].weight > weighted[i].weight {
					weighted[i], weighted[j] = weighted[j], weighted[i]
				}
			}
		}
		best := weighted[0]
		msg := "Möglicher Tippfehler: " + best.pair.part1 + "/" + best.pair.part2
		if best.pair.desc1 != "" && best.pair.desc2 != "" {
			msg = "Möglicher Tippfehler. " + tools.UppercaseFirstChar(best.pair.part1) + ": " + best.pair.desc1 +
				", " + tools.UppercaseFirstChar(best.pair.part2) + ": " + best.pair.desc2
		}
		fromPos := readings.GetStartPos() + partsStartPos
		toPos := fromPos + utf16LenDE(wordPart) + toPosCorrection
		// clamp to token end for hyphenated partials
		if toPos > readings.GetEndPos() {
			toPos = readings.GetEndPos()
		}
		// Java: SpecificIdRule(toId(RULE_ID_part1_part2), desc, isPremium, category, issueType, tags)
		// then new RuleMatch(idRule, …) + setSuggestedReplacement (no shortMessage).
		id := tools.ToId(r.GetID()+"_"+best.pair.part1+"_"+best.pair.part2, "de")
		desc := "Markiert wahrscheinlich falsche Komposita mit Teilwort '" +
			tools.UppercaseFirstChar(best.pair.part1) + "' statt '" +
			tools.UppercaseFirstChar(best.pair.part2) + "' und umgekehrt"
		cat := r.Category
		if cat == nil {
			cat = rules.NewCategory(rules.CategoryTypos, "Typos")
		}
		issue := r.IssueType
		if issue == "" {
			issue = rules.ITSUncategorized
		}
		idRule := rules.NewSpecificIdRule(id, desc, r.Premium, cat, issue, r.Tags)
		rm := rules.NewRuleMatch(idRule, sentence, fromPos, toPos, msg)
		rm.SetSuggestedReplacement(best.variant)
		*ruleMatches = append(*ruleMatches, rm)
	}
	return partsStartPos + utf16LenDE(wordPart) + 1
}

func (r *ProhibitedCompoundRule) isBlacklistedWord(wordPart string) bool {
	if r != nil && r.Blacklist != nil {
		if _, ok := r.Blacklist[wordPart]; ok {
			return true
		}
	}
	if _, ok := ProhibitedCompoundExceptions()[wordPart]; ok {
		return true
	}
	return false
}
