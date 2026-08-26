package de

import (
	"regexp"
	"sort"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// prohibitedPair ports ProhibitedCompoundRule.Pair.
type prohibitedPair struct {
	part1, desc1, part2, desc2 string
}

// ProhibitedCompoundRule ports org.languagetool.rules.de.ProhibitedCompoundRule.
//
// Language model: Frequency[token] ports BaseLanguageModel.getCount(token)
// (Java FakeLanguageModel map / Lucene LM). Without Frequency, Match fails closed
// (Java constructor requires non-null LanguageModel).
type ProhibitedCompoundRule struct {
	Messages map[string]string
	// Frequency ports BaseLanguageModel.getCount(word); nil → fail-closed Match.
	Frequency map[string]int64
	// IsMisspelled optional override; nil uses GermanyGerman default spelling rule
	// (FilterDictIsMisspelled after DiscoverGermanHunspellDict wire).
	IsMisspelled func(word string) bool
	// Blacklist extra words (tests); official /de/compound_exceptions.txt always applied.
	Blacklist map[string]struct{}
	// Premium / Category / IssueType / Tags mirror Rule metadata for SpecificIdRule.
	Premium   bool
	Category  *rules.Category
	IssueType rules.ITSIssueType
	Tags      []rules.Tag
	// confusionPair ports evaluation-only single pair (setConfusionPair); nil = normal.
	confusionPair *prohibitedPair
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

// Java HERRN_FRAU
var herrnFrauRE = regexp.MustCompile(`^(Herrn?|Frau|Dr|Prof|Mag|Hr|Fr|Mr|Mrs|Ms|Fräulein)$`)

// Java blacklistRegex — MatchString on wordPart is equivalent to matches(".*"+k+".*")
// for these patterns (anchors in k handled by the pattern itself).
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

// NewProhibitedCompoundRuleWithFrequency ports
// new ProhibitedCompoundRule(messages, new FakeLanguageModel(map), null, language).
func NewProhibitedCompoundRuleWithFrequency(messages map[string]string, freq map[string]int64) *ProhibitedCompoundRule {
	r := NewProhibitedCompoundRule(messages)
	r.Frequency = freq
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

func (r *ProhibitedCompoundRule) GetID() string { return "DE_PROHIBITED_COMPOUNDS" }

func (r *ProhibitedCompoundRule) GetDescription() string {
	return "Markiert wahrscheinlich falsche Komposita wie 'Lehrzeile', wenn 'Leerzeile' häufiger vorkommt."
}

// SetConfusionPair ports setConfusionPair (evaluation: ignore auto pairs, use only this).
// Pass nil to reset.
func (r *ProhibitedCompoundRule) SetConfusionPair(part1, desc1, part2, desc2 string) {
	if r == nil {
		return
	}
	r.confusionPair = &prohibitedPair{part1, desc1, part2, desc2}
}

// ClearConfusionPair resets evaluation pair (Java setConfusionPair(null)).
func (r *ProhibitedCompoundRule) ClearConfusionPair() {
	if r != nil {
		r.confusionPair = nil
	}
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
	// Java: linguServices == null → GermanyGerman.getDefaultSpellingRule().isMisspelled
	ensureProhibitedSpellerWired()
	return FilterDictIsMisspelled(word)
}

func ensureProhibitedSpellerWired() {
	if FilterDictAvailable() {
		return
	}
	if p := DiscoverGermanHunspellDict("DE"); p != "" {
		_ = WireGermanFilterSpeller(p)
	}
}

// getThreshold ports getThreshold() → 0.
func (r *ProhibitedCompoundRule) getThreshold() int64 { return 0 }

// RemoveHyphensAndAdaptCase ports removeHyphensAndAdaptCase.
// Returns "" when Java returns null (no hyphens, or any part length ≤ 1).
func (r *ProhibitedCompoundRule) RemoveHyphensAndAdaptCase(word string) string {
	return removeHyphensAndAdaptCase(word)
}

func removeHyphensAndAdaptCase(word string) string {
	parts := strings.Split(word, "-")
	if len(parts) <= 1 {
		return ""
	}
	var b strings.Builder
	for i, part := range parts {
		if utf16LenDE(part) <= 1 {
			// don't: S-Bahn → Sbahn
			return ""
		}
		if i == 0 {
			b.WriteString(part)
		} else {
			b.WriteString(tools.LowercaseFirstChar(part))
		}
	}
	return b.String()
}

func (r *ProhibitedCompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil || r == nil || r.Frequency == nil {
		// Java requires LanguageModel; without Frequency fail closed (no Prefer invent).
		return nil
	}
	// Ensure pair resources (Aho-Corasick twin) are loaded.
	initProhibitedResources()

	tokens := sentence.GetTokensWithoutWhitespace()
	var ruleMatches []*rules.RuleMatch
	var prev *languagetool.AnalyzedTokenReadings
	for _, readings := range tokens {
		if readings == nil {
			continue
		}
		tmpWord := readings.GetToken()
		// Java: continue without updating prevReadings on name / title skips.
		if prev != nil && prev.HasAnyPartialPosTag("EIG:") && tools.StartsWithUppercase(tmpWord) &&
			(readings.HasAnyPartialPosTag("EIG:") || readings.IsPosTagUnknown()) {
			// assume name, e.g. "Bianca Baalhorn" (avoid: Baalhorn => Ballhorn)
			continue
		}
		if prev != nil && herrnFrauRE.MatchString(prev.GetToken()) {
			// assume name, e.g. "Herr Eiswert" (avoid: Eiswert -> Eiswelt)
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
	// Java: blacklistRegex.stream().noneMatch(k -> wordPart.matches(".*" + k + ".*"))
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
	// Java:
	// if ((readings.isTagged() && !readings.hasPartialPosTag("SUB")) && !readings.hasPosTagStartingWith("EIG:")
	//     || wordPart.length() <= 6)
	skipPOS := readings.IsTagged() && !readings.HasPartialPosTag("SUB") && !readings.HasPosTagStartingWith("EIG:")
	if skipPOS || utf16LenDE(wordPart) <= 6 {
		return partsStartPos + utf16LenDE(wordPart) + 1
	}

	// Collect candidate pairs (Java AhoCorasick parseText → pairMap, or confusionPair eval).
	var candidatePairs []prohibitedPair
	if r.confusionPair == nil {
		for _, stem := range prohibitedStems() {
			if !strings.Contains(wordPart, stem) {
				continue
			}
			// might get duplicates; only one match per word kept
			candidatePairs = append(candidatePairs, prohibitedPairLookup(stem)...)
		}
	} else {
		// Java addAllCaseVariants(candidatePairs, confusionPair)
		candidatePairs = append(candidatePairs, *r.confusionPair)
		uc1 := tools.UppercaseFirstChar(r.confusionPair.part1)
		uc2 := tools.UppercaseFirstChar(r.confusionPair.part2)
		if r.confusionPair.part1 != uc1 || r.confusionPair.part2 != uc2 {
			candidatePairs = append(candidatePairs, prohibitedPair{uc1, r.confusionPair.desc1, uc2, r.confusionPair.desc2})
		}
	}

	type weightedRuleMatch struct {
		weight  int64
		pair    prohibitedPair
		variant string
	}
	var weighted []weightedRuleMatch

	for _, pair := range candidatePairs {
		var variant string
		// Java: wordPart.replaceFirst(pair.part1, pair.part2) — regex first match
		if strings.Contains(wordPart, pair.part1) {
			variant = replaceFirstRegex(wordPart, pair.part1, pair.part2)
		} else if strings.Contains(wordPart, pair.part2) {
			variant = replaceFirstRegex(wordPart, pair.part2, pair.part1)
		} else {
			// Java: partsStartPos += …; continue — only skips this pair
			continue
		}
		wordCount := r.getCount(wordPart)
		variantCount := r.getCount(variant)

		// Java: !blacklist.contains(wordPart) && !isMisspelled(variant) && noneMatch blacklistRegex
		if variantCount > r.getThreshold() && wordCount == 0 &&
			!r.isBlacklistedWord(wordPart) &&
			!r.isMisspelled(variant) &&
			!blacklistMatch(wordPart) {
			weighted = append(weighted, weightedRuleMatch{weight: variantCount, pair: pair, variant: variant})
		}
	}

	if len(weighted) > 0 {
		// Java Collections.sort(weightedMatches) — most popular alternative first
		sort.SliceStable(weighted, func(i, j int) bool {
			return weighted[i].weight > weighted[j].weight
		})
		best := weighted[0]
		msg := "Möglicher Tippfehler: " + best.pair.part1 + "/" + best.pair.part2
		if best.pair.desc1 != "" && best.pair.desc2 != "" {
			msg = "Möglicher Tippfehler. " + tools.UppercaseFirstChar(best.pair.part1) + ": " + best.pair.desc1 +
				", " + tools.UppercaseFirstChar(best.pair.part2) + ": " + best.pair.desc2
		}
		fromPos := readings.GetStartPos() + partsStartPos
		toPos := fromPos + utf16LenDE(wordPart) + toPosCorrection
		// Java SpecificIdRule(toId(RULE_ID_part1_part2, language), desc, isPremium, category, issueType, tags)
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

// replaceFirstRegex ports String.replaceFirst(regex, replacement) for one match.
// Pair stems are plain letters; invalid patterns leave the string unchanged.
func replaceFirstRegex(s, pattern, repl string) string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return s
	}
	loc := re.FindStringIndex(s)
	if loc == nil {
		return s
	}
	return s[:loc[0]] + repl + s[loc[1]:]
}
