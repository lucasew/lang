package en

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	tagen "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/en"
	entok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/en"
)

// EnglishForL2SpeakersFalseFriendRule ports
// org.languagetool.rules.en.EnglishForL2SpeakersFalseFriendRule —
// ConfusionProbabilityRule + false-friend message rewrite for L2 mother tongues.
type EnglishForL2SpeakersFalseFriendRule struct {
	*ngrams.ConfusionProbabilityRule
	// MotherTongue short code (e.g. "de") — Java Language motherTongue.
	MotherTongue string
	// Language target short code (e.g. "en").
	Language string
	// Filenames ports getFilenames() confusion set resources under language dir.
	Filenames []string
	// ExampleWrong / ExampleFixed surface for documentation / tests.
	ExampleWrong string
	ExampleFixed string
	// TagWord ports lang.getTagger().tag for isBaseformMatch (nil → no lemma match).
	TagWord func(token string) []languagetool.TokenTag
	// FalseFriendsXML optional path override; empty → discover official false-friends.xml.
	FalseFriendsXML string
}

// Package-level cache ports static motherTongue2rules map.
// l2ConfusionCache ports ConfusionProbabilityRule confSetCache for L2 filenames.
var (
	l2FFMu    sync.Mutex
	l2FFRules = map[string][]*patterns.FalseFriendPatternRule{}

	l2ConfusionMu    sync.Mutex
	l2ConfusionCache = map[string]map[string][]*rules.ConfusionPair{}
)

// NewEnglishForL2SpeakersFalseFriendRule ports the abstract constructor
// (messages, languageModel, motherTongue, lang) with grams=3.
func NewEnglishForL2SpeakersFalseFriendRule(lm ngrams.LanguageModel, motherTongue, lang string) *EnglishForL2SpeakersFalseFriendRule {
	base := ngrams.NewConfusionProbabilityRule(lm, 3)
	// Java: LanguageModelUtils.getGoogleStyleWordTokenizer(language) → language.getWordTokenizer()
	// English.createDefaultWordTokenizer → EnglishWordTokenizer.
	wt := entok.NewEnglishWordTokenizer()
	base.Tokenize = wt.Tokenize
	r := &EnglishForL2SpeakersFalseFriendRule{
		ConfusionProbabilityRule: base,
		MotherTongue:             motherTongue,
		Language:                 lang,
	}
	// Java getMessage override via MessageFor hook.
	base.MessageFor = r.l2Message
	// Java isBaseformMatch: lang.getTagger().tag — EnglishTagger / english.dict.
	r.TagWord = englishL2TagWord
	return r
}

// englishL2TagWord ports lang.getTagger().tag for isBaseformMatch (lazy english.dict).
func englishL2TagWord(token string) []languagetool.TokenTag {
	fn := tagen.EnglishTagWord()
	if fn == nil {
		return nil
	}
	return fn(token)
}

// loadConfusionSetsFromFilenames ports ConfusionProbabilityRule constructor loop:
// for filename in getFilenames(): load "/" + lang + "/" + filename via ConfusionSetLoader.
func (r *EnglishForL2SpeakersFalseFriendRule) loadConfusionSetsFromFilenames() {
	if r == nil || r.ConfusionProbabilityRule == nil {
		return
	}
	merged := map[string][]*rules.ConfusionPair{}
	for _, fn := range r.Filenames {
		if fn == "" {
			continue
		}
		part := loadCachedL2ConfusionPairs(fn)
		for k, v := range part {
			merged[k] = append(merged[k], v...)
		}
	}
	if len(merged) > 0 {
		r.SetWordToPairs(merged)
	}
}

// loadCachedL2ConfusionPairs ports confSetCache.getUnchecked for one L2 confusion file.
func loadCachedL2ConfusionPairs(filename string) map[string][]*rules.ConfusionPair {
	l2ConfusionMu.Lock()
	defer l2ConfusionMu.Unlock()
	if cached, ok := l2ConfusionCache[filename]; ok {
		return cached
	}
	path := discoverENL2ConfusionSet(filename)
	if path == "" {
		l2ConfusionCache[filename] = map[string][]*rules.ConfusionPair{}
		return l2ConfusionCache[filename]
	}
	f, err := os.Open(path)
	if err != nil {
		l2ConfusionCache[filename] = map[string][]*rules.ConfusionPair{}
		return l2ConfusionCache[filename]
	}
	defer f.Close()
	// Java: new ConfusionSetLoader(language) — ShortDescriptionProvider optional; nil = no invent defs.
	m, err := rules.NewConfusionSetLoader(nil).LoadConfusionPairs(f)
	if err != nil || m == nil {
		l2ConfusionCache[filename] = map[string][]*rules.ConfusionPair{}
		return l2ConfusionCache[filename]
	}
	l2ConfusionCache[filename] = m
	return m
}

// discoverENL2ConfusionSet finds official /en/<filename> (Java resource dir).
func discoverENL2ConfusionSet(filename string) string {
	if filename == "" {
		return ""
	}
	if p := os.Getenv("LANG_EN_L2_CONFUSION_DIR"); p != "" {
		cand := filepath.Join(p, filename)
		if st, err := os.Stat(cand); err == nil && !st.IsDir() {
			return cand
		}
	}
	relPaths := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "en",
			"src", "main", "resources", "org", "languagetool", "resource", "en", filename),
		filepath.Join("testdata", "upstream", "en", "resource", filename),
	}
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}
	dir := wd
	for i := 0; i < 14; i++ {
		for _, rel := range relPaths {
			p := filepath.Join(dir, rel)
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				return p
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// GetFilenames ports getFilenames().
func (r *EnglishForL2SpeakersFalseFriendRule) GetFilenames() []string {
	if r == nil {
		return nil
	}
	return append([]string(nil), r.Filenames...)
}

// AddExamplePair ports Rule.addExamplePair and sets ExampleWrong/Fixed surfaces.
func (r *EnglishForL2SpeakersFalseFriendRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil || r.ConfusionProbabilityRule == nil {
		return
	}
	r.ConfusionProbabilityRule.AddExamplePair(incorrect, correct)
	r.ExampleWrong = incorrect.GetExample()
	r.ExampleFixed = correct.GetExample()
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *EnglishForL2SpeakersFalseFriendRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || r.ConfusionProbabilityRule == nil {
		return nil
	}
	return r.ConfusionProbabilityRule.GetIncorrectExamples()
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *EnglishForL2SpeakersFalseFriendRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || r.ConfusionProbabilityRule == nil {
		return nil
	}
	return r.ConfusionProbabilityRule.GetCorrectExamples()
}

// Match ports ConfusionProbabilityRule.match (ngram pairs; nil LM → no hits).
func (r *EnglishForL2SpeakersFalseFriendRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.ConfusionProbabilityRule == nil {
		return nil
	}
	return r.ConfusionProbabilityRule.Match(sentence)
}

// l2Message ports EnglishForL2SpeakersFalseFriendRule.getMessage.
func (r *EnglishForL2SpeakersFalseFriendRule) l2Message(textStr, better *rules.ConfusionString) string {
	if r == nil || textStr == nil {
		return ""
	}
	s := textStr.GetString()
	for _, rule := range r.getFalseFriendRules() {
		if rule == nil || rule.PatternRule == nil {
			continue
		}
		for _, pt := range rule.Tokens {
			if pt == nil {
				continue
			}
			if s == pt.Token || r.isBaseformMatch(textStr, pt) {
				if msg := rule.GetMessage(); msg != "" {
					return msg
				}
			}
		}
	}
	// super.getMessage — temporarily clear MessageFor to avoid recursion
	mf := r.MessageFor
	r.MessageFor = nil
	msg := r.ConfusionProbabilityRule.Message(textStr, better)
	r.MessageFor = mf
	return msg
}

// isBaseformMatch ports isBaseformMatch (inflected pattern token + tagger lemmas).
func (r *EnglishForL2SpeakersFalseFriendRule) isBaseformMatch(textString *rules.ConfusionString, patternToken *patterns.PatternToken) bool {
	if r == nil || textString == nil || patternToken == nil || !patternToken.IsInflected() {
		return false
	}
	if r.TagWord == nil {
		return false
	}
	want := patternToken.Token
	for _, tt := range r.TagWord(textString.GetString()) {
		if tt.Lemma == want {
			return true
		}
	}
	return false
}

// getFalseFriendRules ports getRules() with motherTongue2rules cache.
func (r *EnglishForL2SpeakersFalseFriendRule) getFalseFriendRules() []*patterns.FalseFriendPatternRule {
	if r == nil || r.MotherTongue == "" {
		return nil
	}
	key := r.MotherTongue
	l2FFMu.Lock()
	defer l2FFMu.Unlock()
	if cached, ok := l2FFRules[key]; ok {
		return cached
	}
	path := r.FalseFriendsXML
	if path == "" {
		path = discoverFalseFriendsXML()
	}
	if path == "" {
		l2FFRules[key] = nil
		return nil
	}
	f, err := os.Open(path)
	if err != nil {
		// Java: throw RuntimeException — fail closed with empty rules (do not invent)
		l2FFRules[key] = nil
		return nil
	}
	defer f.Close()
	// Java: FalseFriendRuleLoader("\"{0}\" ({1}) means {2} ({3}).", "Did you maybe mean {0}?")
	loader := patterns.NewFalseFriendRuleLoader(
		`"{0}" ({1}) means {2} ({3}).`,
		"Did you maybe mean {0}?",
	)
	lang := r.Language
	if lang == "" {
		lang = "en"
	}
	rules, err := loader.GetRulesFromReader(f, lang, r.MotherTongue)
	if err != nil {
		l2FFRules[key] = nil
		return nil
	}
	l2FFRules[key] = rules
	return rules
}

// discoverFalseFriendsXML finds official false-friends.xml (same roots as Java data broker).
func discoverFalseFriendsXML() string {
	candidates := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-core", "src", "main", "resources", "org", "languagetool", "rules", "false-friends.xml"),
		filepath.Join("testdata", "upstream", "false-friends.xml"),
	}
	// walk up from CWD
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}
	dir := wd
	for i := 0; i < 8; i++ {
		for _, rel := range candidates {
			p := filepath.Join(dir, rel)
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				return p
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// ClearL2FalseFriendRuleCache clears the motherTongue→rules cache (tests).
func ClearL2FalseFriendRuleCache() {
	l2FFMu.Lock()
	defer l2FFMu.Unlock()
	l2FFRules = map[string][]*patterns.FalseFriendPatternRule{}
}

// --- concrete language variants ---

// NewEnglishForGermansFalseFriendRule ports EnglishForGermansFalseFriendRule metadata.
func NewEnglishForGermansFalseFriendRule() *EnglishForL2SpeakersFalseFriendRule {
	return NewEnglishForGermansFalseFriendRuleWithLM(nil)
}

// NewEnglishForGermansFalseFriendRuleWithLM ports the full Java constructor.
func NewEnglishForGermansFalseFriendRuleWithLM(lm ngrams.LanguageModel) *EnglishForL2SpeakersFalseFriendRule {
	r := NewEnglishForL2SpeakersFalseFriendRule(lm, "de", "en")
	r.RuleIDOverride = "EN_FOR_DE_SPEAKERS_FALSE_FRIENDS"
	r.Filenames = []string{"confusion_sets_l2_de.txt"}
	r.loadConfusionSetsFromFilenames()
	r.AddExamplePair(
		rules.Wrong("My <marker>handy</marker> is broken."),
		rules.Fixed("My <marker>phone</marker> is broken."),
	)
	return r
}

// NewEnglishForFrenchFalseFriendRule ports EnglishForFrenchFalseFriendRule.
func NewEnglishForFrenchFalseFriendRule() *EnglishForL2SpeakersFalseFriendRule {
	return NewEnglishForFrenchFalseFriendRuleWithLM(nil)
}

func NewEnglishForFrenchFalseFriendRuleWithLM(lm ngrams.LanguageModel) *EnglishForL2SpeakersFalseFriendRule {
	r := NewEnglishForL2SpeakersFalseFriendRule(lm, "fr", "en")
	r.RuleIDOverride = "EN_FOR_FR_SPEAKERS_FALSE_FRIENDS"
	r.Filenames = []string{"confusion_sets_l2_fr.txt"}
	// Java ConfusionProbabilityRule ctor: load /en/confusion_sets_l2_fr.txt
	r.loadConfusionSetsFromFilenames()
	r.AddExamplePair(
		rules.Wrong("She will <marker>achieve</marker> her task."),
		rules.Fixed("She will <marker>complete</marker> her task."),
	)
	return r
}

// NewEnglishForSpaniardsFalseFriendRule ports EnglishForSpaniardsFalseFriendRule.
func NewEnglishForSpaniardsFalseFriendRule() *EnglishForL2SpeakersFalseFriendRule {
	return NewEnglishForSpaniardsFalseFriendRuleWithLM(nil)
}

func NewEnglishForSpaniardsFalseFriendRuleWithLM(lm ngrams.LanguageModel) *EnglishForL2SpeakersFalseFriendRule {
	r := NewEnglishForL2SpeakersFalseFriendRule(lm, "es", "en")
	r.RuleIDOverride = "EN_FOR_ES_SPEAKERS_FALSE_FRIENDS"
	r.Filenames = []string{"confusion_sets_l2_es.txt"}
	r.loadConfusionSetsFromFilenames()
	r.AddExamplePair(
		rules.Wrong("The factory will <marker>realize</marker> computer chips."),
		rules.Fixed("The factory will <marker>produce</marker> computer chips."),
	)
	return r
}

// NewEnglishForDutchmenFalseFriendRule ports EnglishForDutchmenFalseFriendRule.
func NewEnglishForDutchmenFalseFriendRule() *EnglishForL2SpeakersFalseFriendRule {
	return NewEnglishForDutchmenFalseFriendRuleWithLM(nil)
}

func NewEnglishForDutchmenFalseFriendRuleWithLM(lm ngrams.LanguageModel) *EnglishForL2SpeakersFalseFriendRule {
	r := NewEnglishForL2SpeakersFalseFriendRule(lm, "nl", "en")
	r.RuleIDOverride = "EN_FOR_NL_SPEAKERS_FALSE_FRIENDS"
	r.Filenames = []string{"confusion_sets_l2_nl.txt"}
	r.loadConfusionSetsFromFilenames()
	r.AddExamplePair(
		rules.Wrong("The <marker>want</marker> should be painted green."),
		rules.Fixed("The <marker>wall</marker> should be painted green."),
	)
	return r
}
