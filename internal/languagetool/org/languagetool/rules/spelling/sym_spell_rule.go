package spelling

import (
	"bufio"
	"fmt"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/languagemodel"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/suggestions"
	symimpl "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/symspell/implementation"
)

// SymSpellRuleID ports SymSpellRule.getId.
const SymSpellRuleID = "SYMSPELL_RULE"

// INITIAL_CAPACITY ports SymSpellRule.INITIAL_CAPACITY.
const INITIAL_CAPACITY = 50000

// SymSpellRule ports org.languagetool.rules.spelling.SymSpellRule.
//
// Java uses static Guava LoadingCaches (expireAfterAccess). Go maps those to
// process-level maps (TTL omitted — pure performance mapping, same as
// CachingWordListLoader). Resource loading uses package GetDataBroker()
// (Java JLanguageTool.getDataBroker()).
type SymSpellRule struct {
	*SpellingCheckRule

	// language ports SpellingCheckRule.language (used for ignore/prohibit caches + paths).
	language languagetool.Language
	// languageModel ports SpellingCheckRule.languageModel (optional; experiment orderer).
	languageModel languagemodel.LanguageModel

	// defaultDictSpeller / userDictSpeller port protected final fields.
	defaultDictSpeller *symimpl.SymSpell
	userDictSpeller    *symimpl.SymSpell

	editDistance int
	verbosity    symimpl.Verbosity
	orderer      suggestions.SuggestionsOrderer
}

// --- static caches (Java LoadingCache<Language, …>) ---

var (
	spellerCacheMu sync.Mutex
	spellerCache   = map[string]*symimpl.SymSpell{} // key: language cache id

	ignoredWordsCacheMu sync.Mutex
	ignoredWordsCache   = map[string]map[string]struct{}{}

	prohibitedWordsCacheMu sync.Mutex
	prohibitedWordsCache   = map[string]map[string]struct{}{}
)

// clearSymSpellRuleCaches clears process-level caches (test helper; no Java twin).
func clearSymSpellRuleCaches() {
	spellerCacheMu.Lock()
	spellerCache = map[string]*symimpl.SymSpell{}
	spellerCacheMu.Unlock()

	ignoredWordsCacheMu.Lock()
	ignoredWordsCache = map[string]map[string]struct{}{}
	ignoredWordsCacheMu.Unlock()

	prohibitedWordsCacheMu.Lock()
	prohibitedWordsCache = map[string]map[string]struct{}{}
	prohibitedWordsCacheMu.Unlock()
}

func languageCacheKey(lang languagetool.Language) string {
	if lang == nil {
		return ""
	}
	// Prefer full code so variants with different dict paths do not collide.
	return lang.GetShortCodeWithCountryAndVariant()
}

// GetSpellingDictBaseDir ports SymSpellRule.getSpellingDictBaseDir.
func GetSpellingDictBaseDir(lang languagetool.Language) string {
	if lang == nil {
		return "/hunspell/"
	}
	return lang.GetShortCode() + "/hunspell/"
}

// getWordList ports private getWordList(lang, file).
// Lines are accepted raw (no comment stripping) — unlike CachingWordListLoader.
func getWordList(lang languagetool.Language, file string) map[string]struct{} {
	base := GetSpellingDictBaseDir(lang)
	paths := []string{base + file}
	words := map[string]struct{}{}
	forEachLineInResources(paths, func(line string) {
		words[line] = struct{}{}
	})
	return words
}

func getIgnoredWordsCached(lang languagetool.Language) map[string]struct{} {
	key := languageCacheKey(lang)
	ignoredWordsCacheMu.Lock()
	if s, ok := ignoredWordsCache[key]; ok {
		ignoredWordsCacheMu.Unlock()
		return s
	}
	ignoredWordsCacheMu.Unlock()

	s := getWordList(lang, "ignore.txt")

	ignoredWordsCacheMu.Lock()
	if existing, ok := ignoredWordsCache[key]; ok {
		ignoredWordsCacheMu.Unlock()
		return existing
	}
	ignoredWordsCache[key] = s
	ignoredWordsCacheMu.Unlock()
	return s
}

func getProhibitedWordsCached(lang languagetool.Language) map[string]struct{} {
	key := languageCacheKey(lang)
	prohibitedWordsCacheMu.Lock()
	if s, ok := prohibitedWordsCache[key]; ok {
		prohibitedWordsCacheMu.Unlock()
		return s
	}
	prohibitedWordsCacheMu.Unlock()

	// Java bug-for-bug: filename is "probibit.txt" (typo), not "prohibit.txt".
	s := getWordList(lang, "probibit.txt")

	prohibitedWordsCacheMu.Lock()
	if existing, ok := prohibitedWordsCache[key]; ok {
		prohibitedWordsCacheMu.Unlock()
		return existing
	}
	prohibitedWordsCache[key] = s
	prohibitedWordsCacheMu.Unlock()
	return s
}

func getDefaultDictSpellerCached(lang languagetool.Language) *symimpl.SymSpell {
	key := languageCacheKey(lang)
	spellerCacheMu.Lock()
	if s, ok := spellerCache[key]; ok {
		spellerCacheMu.Unlock()
		return s
	}
	spellerCacheMu.Unlock()

	s := InitDefaultDictSpeller(lang)

	spellerCacheMu.Lock()
	if existing, ok := spellerCache[key]; ok {
		spellerCacheMu.Unlock()
		return existing
	}
	spellerCache[key] = s
	spellerCacheMu.Unlock()
	return s
}

// forEachLineInResources ports SymSpellRule.forEachLineInResources.
// Missing resources are skipped; read errors panic (Java RuntimeException).
func forEachLineInResources(resources []string, function func(string)) {
	broker := GetDataBroker()
	if broker == nil {
		return
	}
	for _, resource := range resources {
		if !broker.ResourceExists(resource) {
			continue
		}
		rc, err := broker.GetFromResourceDirAsStream(resource)
		if err != nil {
			panic(fmt.Errorf("Could not read resource %s: %w", resource, err))
		}
		func() {
			defer rc.Close()
			sc := bufio.NewScanner(rc)
			// Java BufferedReader.readLine handles long lines; allow large lines.
			buf := make([]byte, 0, 64*1024)
			sc.Buffer(buf, 1024*1024)
			for sc.Scan() {
				function(sc.Text())
			}
			if err := sc.Err(); err != nil {
				panic(fmt.Errorf("Could not read resource %s: %w", resource, err))
			}
		}()
	}
}

// InitUserDictSpeller ports SymSpellRule.initUserDictSpeller.
// Returns nil when no custom speller is needed.
func InitUserDictSpeller(config *languagetool.UserConfig) *symimpl.SymSpell {
	if config == nil {
		return nil
	}
	dict := config.GetAcceptedWords()
	if len(dict) == 0 {
		return nil
	}
	speller := symimpl.NewSymSpell(0, 3, -1, 0)
	stage := symimpl.NewSuggestionStage(len(dict))
	for _, word := range dict {
		speller.CreateDictionaryEntry(word, 1, stage)
	}
	speller.CommitStaged(stage)
	return speller
}

// InitDefaultDictSpeller ports SymSpellRule.initDefaultDictSpeller.
func InitDefaultDictSpeller(lang languagetool.Language) *symimpl.SymSpell {
	speller := symimpl.NewSymSpell(INITIAL_CAPACITY, 3, -1, 0)
	// Java: System.out.println("Initializing symspell") — side-effect only; omitted.

	prohibitedWords := getProhibitedWordsCached(lang)

	base := GetSpellingDictBaseDir(lang)
	variant := ""
	if lang != nil {
		variant = lang.GetShortCodeWithCountryAndVariant()
	}
	additional := []string{
		base + "spelling.txt",
		base + "spelling_" + variant + ".txt",
	}
	// Apache StringUtils.replaceOnce(variant, "-", "_") — first '-' only.
	dictPath := base + strings.Replace(variant, "-", "_", 1) + ".dic"
	dict := []string{dictPath}

	stage := symimpl.NewSuggestionStage(100000)
	forEachLineInResources(additional, func(word string) {
		if _, bad := prohibitedWords[word]; !bad {
			speller.CreateDictionaryEntry(word, 1, stage)
		}
	})
	forEachLineInResources(dict, func(line string) {
		split := strings.LastIndex(line, "+")
		if split == -1 || len(line) <= split+1 {
			// Java: throw new IllegalArgumentException(String.format(...))
			panic(fmt.Sprintf("Could not parse frequency dictionary line '%s'.", line))
		}
		word := line[:split]
		freqClass := line[split+1]
		// A - least frequent, Z - most frequent
		freq := int64(1 + int(freqClass) - int('A'))
		if _, bad := prohibitedWords[word]; !bad {
			speller.CreateDictionaryEntry(word, freq, stage)
		}
	})
	speller.CommitStaged(stage)
	return speller
}

// initParameters ports SymSpellRule.initParameters (experiment hooks).
func (r *SymSpellRule) initParameters() {
	if r == nil {
		return
	}
	sc := suggestions.GetSuggestionsChanges()
	if sc == nil {
		return
	}
	exp := sc.GetCurrentExperiment()
	if exp == nil {
		return
	}
	if exp.Parameters != nil {
		if v, ok := exp.Parameters["candidates"]; ok && v != nil {
			candidatesParam := fmt.Sprint(v)
			r.verbosity = verbosityValueOf(candidatesParam)
		}
		if v, ok := exp.Parameters["editDistance"]; ok && v != nil {
			r.editDistance = experimentInt(v)
		}
	}
	if suggestions.IsRunningExperiment("SymSpell+NewSuggestionsOrderer") {
		r.orderer = suggestions.NewSuggestionsOrdererFeatureExtractor(
			languageModelHook(r.languageModel),
		)
	}
}

// verbosityValueOf ports SymSpell.Verbosity.valueOf (enum names Top/Closest/All).
func verbosityValueOf(name string) symimpl.Verbosity {
	switch name {
	case "Top":
		return symimpl.VerbosityTop
	case "Closest":
		return symimpl.VerbosityClosest
	case "All":
		return symimpl.VerbosityAll
	default:
		// Java Enum.valueOf throws IllegalArgumentException
		panic(fmt.Sprintf("No enum constant org.languagetool.rules.spelling.symspell.implementation.SymSpell.Verbosity.%s", name))
	}
}

func experimentInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case int32:
		return int(n)
	case int64:
		return int(n)
	case float64:
		return int(n)
	case float32:
		return int(n)
	default:
		var i int
		_, err := fmt.Sscan(fmt.Sprint(v), &i)
		if err != nil {
			panic(fmt.Sprintf("editDistance parameter not an integer: %v", v))
		}
		return i
	}
}

// languageModelHook adapts languagemodel.LanguageModel to LanguageModelHook.
// Count is only available on models that expose GetCountToken (BaseLanguageModel twin).
func languageModelHook(lm languagemodel.LanguageModel) suggestions.LanguageModelHook {
	if lm == nil {
		return nil
	}
	return lmHook{lm: lm}
}

type lmHook struct {
	lm languagemodel.LanguageModel
}

func (h lmHook) PseudoProbability(tokens []string) float64 {
	if h.lm == nil {
		return 0
	}
	return h.lm.GetPseudoProbability(tokens).GetProb()
}

func (h lmHook) Count(word string) int64 {
	type counter interface {
		GetCountToken(string) int64
	}
	if c, ok := h.lm.(counter); ok {
		return c.GetCountToken(word)
	}
	return 0
}

// NewSymSpellRule ports SymSpellRule(messages, language, userConfig).
func NewSymSpellRule(lang languagetool.Language, userConfig *languagetool.UserConfig) *SymSpellRule {
	return NewSymSpellRuleFull(lang, userConfig, nil, nil)
}

// NewSymSpellRuleWithAlts ports SymSpellRule(..., altLanguages).
func NewSymSpellRuleWithAlts(lang languagetool.Language, userConfig *languagetool.UserConfig, altLanguages []languagetool.Language) *SymSpellRule {
	return NewSymSpellRuleFull(lang, userConfig, altLanguages, nil)
}

// NewSymSpellRuleFull ports SymSpellRule(..., altLanguages, languageModel).
// messages (ResourceBundle) is unused in the Go twin (same as other spell-rule ports).
func NewSymSpellRuleFull(
	lang languagetool.Language,
	userConfig *languagetool.UserConfig,
	altLanguages []languagetool.Language,
	languageModel languagemodel.LanguageModel,
) *SymSpellRule {
	_ = altLanguages // Java stores on Rule base; not required for SymSpellRule.match
	code := ""
	if lang != nil {
		code = lang.GetShortCode()
	}
	scr := NewSpellingCheckRule(SymSpellRuleID, "Spell checking rule using SymSpell algorithm", code)
	if userConfig != nil {
		// Java SpellingCheckRule ctor: wordsToBeIgnored.addAll(userConfig.getAcceptedWords())
		scr.ApplyUserAcceptedWords(userConfig.GetAcceptedWords())
	}
	r := &SymSpellRule{
		SpellingCheckRule: scr,
		language:          lang,
		languageModel:     languageModel,
		editDistance:      3,
		verbosity:         symimpl.VerbosityClosest,
	}
	r.initParameters()
	r.defaultDictSpeller = getDefaultDictSpellerCached(lang)
	r.userDictSpeller = InitUserDictSpeller(userConfig)
	return r
}

// DefaultDictSpeller returns the Java defaultDictSpeller field.
func (r *SymSpellRule) DefaultDictSpeller() *symimpl.SymSpell {
	if r == nil {
		return nil
	}
	return r.defaultDictSpeller
}

// UserDictSpeller returns the Java userDictSpeller field.
func (r *SymSpellRule) UserDictSpeller() *symimpl.SymSpell {
	if r == nil {
		return nil
	}
	return r.userDictSpeller
}

// EditDistance returns the current editDistance (after initParameters).
func (r *SymSpellRule) EditDistance() int {
	if r == nil {
		return 0
	}
	return r.editDistance
}

// Verbosity returns the current Verbosity (after initParameters).
func (r *SymSpellRule) Verbosity() symimpl.Verbosity {
	if r == nil {
		return 0
	}
	return r.verbosity
}

// Orderer returns the SuggestionsOrderer (may be nil).
func (r *SymSpellRule) Orderer() suggestions.SuggestionsOrderer {
	if r == nil {
		return nil
	}
	return r.orderer
}

// GetID ports getId.
func (r *SymSpellRule) GetID() string {
	if r == nil || r.SpellingCheckRule == nil {
		return SymSpellRuleID
	}
	return r.SpellingCheckRule.GetID()
}

// GetDescription ports getDescription.
func (r *SymSpellRule) GetDescription() string {
	if r == nil || r.SpellingCheckRule == nil {
		return "Spell checking rule using SymSpell algorithm"
	}
	return r.SpellingCheckRule.GetDescription()
}

// IsMisspelled ports SymSpellRule.isMisspelled — Java always throws RuntimeException("not implemented").
// Match path does not call this; use candidates instead.
func (r *SymSpellRule) IsMisspelled(word string) bool {
	_ = word
	panic("not implemented")
}

// Match ports SymSpellRule.match.
//
//	for token : sentence.getTokensWithoutWhitespace():
//	  skip sentenceStart | immunized | ignoredBySpeller | nonWord
//	  skip ignoredWords.contains(word)   // ignore.txt cache only
//	  candidates = filterCandidates(getSpellerMatches(default))
//	  userCandidates = getSpellerMatches(user)   // NOT filterCandidates'd
//	  if both empty → "Misspelling or unknown word!"
//	  else if first != word (default and user) → "Misspelling!" + suggestions
func (r *SymSpellRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	if sentence == nil || r == nil {
		return nil, nil
	}
	var out []*rules.RuleMatch
	ignoredWords := getIgnoredWordsCached(r.language)
	for _, tok := range sentence.GetTokensWithoutWhitespace() {
		if tok == nil || tok.IsSentenceStart() || tok.IsImmunized() || tok.IsIgnoredBySpeller() || tok.IsNonWord() {
			continue
		}
		word := tok.GetToken()
		if _, skip := ignoredWords[word]; skip {
			continue
		}
		candidates := r.filterCandidates(r.getSpellerMatches(word, r.defaultDictSpeller))
		userCandidates := r.getSpellerMatches(word, r.userDictSpeller)

		var m *rules.RuleMatch
		if len(candidates) == 0 && len(userCandidates) == 0 {
			m = rules.NewRuleMatch(r, sentence, tok.GetStartPos(), tok.GetEndPos(),
				"Misspelling or unknown word!")
		} else if !(len(candidates) > 0 && candidates[0] == word ||
			len(userCandidates) > 0 && userCandidates[0] == word) {
			m = rules.NewRuleMatch(r, sentence, tok.GetStartPos(), tok.GetEndPos(),
				"Misspelling!")
			AddSuggestionsToRuleMatchStrings(word, userCandidates, candidates, r.orderer, m)
		}
		if m != nil {
			out = append(out, m)
		}
	}
	return out, nil
}

// filterCandidates ports SymSpellRule.filterCandidates — drop ignore + prohibited (cache sets).
func (r *SymSpellRule) filterCandidates(candidates []string) []string {
	if len(candidates) == 0 {
		// Java Collectors.toList on empty stream → empty list (not null).
		return []string{}
	}
	ignoredWords := getIgnoredWordsCached(r.language)
	prohibitedWords := getProhibitedWordsCached(r.language)
	out := make([]string, 0, len(candidates))
	for _, c := range candidates {
		if _, bad := ignoredWords[c]; bad {
			continue
		}
		if _, bad := prohibitedWords[c]; bad {
			continue
		}
		out = append(out, c)
	}
	return out
}

// getSpellerMatches ports SymSpellRule.getSpellerMatches.
func (r *SymSpellRule) getSpellerMatches(word string, speller *symimpl.SymSpell) []string {
	if speller == nil {
		return []string{}
	}
	// Java: speller.lookup(word, verbosity, editDistance) — no invent clamp.
	candidatesData := speller.LookupMax(word, r.verbosity, r.editDistance)
	out := make([]string, 0, len(candidatesData))
	for _, candidate := range candidatesData {
		out = append(out, candidate.Term)
	}
	return out
}
