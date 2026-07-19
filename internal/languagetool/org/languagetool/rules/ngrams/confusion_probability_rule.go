package ngrams

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// ConfusionRuleID ports deprecated ConfusionProbabilityRule.RULE_ID.
const ConfusionRuleID = "CONFUSION_RULE"

// MinCoverage ports ConfusionProbabilityRule.MIN_COVERAGE.
const MinCoverage = 0.5

// MinProb ports ConfusionProbabilityRule.MIN_PROB (0.0).
const MinProb = 0.0

var realWordRE = regexp.MustCompile(`^\p{L}+$`)

// ConfusionProbabilityRule ports org.languagetool.rules.ngrams.ConfusionProbabilityRule.
// Match scores confusion pairs via LanguageModel; nil LM → no matches (fail-closed).
type ConfusionProbabilityRule struct {
	LM             LanguageModel
	Grams          int
	Exceptions     []string
	WordToPairs    map[string][]*rules.ConfusionPair
	DefaultOff     bool
	RuleIDOverride string
	// Messages optional Tools.i18n keys (statistics_suggest_short_desc, statistics_rule_description, …).
	// When nil, MessagesBundle.properties English defaults are used (not invent DE).
	Messages map[string]string
	// Tokenize defaults to WordTokenizer (Google-style = language word tokenizer).
	Tokenize func(string) []string
	// IsException optional sentence-level skip (German overrides).
	IsException func(sentenceText string, startPos, endPos int) bool
	// IsCoveredByAntiPattern optional skip (Java immunization / anti-patterns).
	IsCoveredByAntiPattern func(sentence *languagetool.AnalyzedSentence, startPos, endPos int) bool
	// IsCommonWord defaults to \w+; German uses [\\wöäüßÖÄÜ]+.
	IsCommonWord func(token string) bool
	// Message builders optional; defaults to simple German/English-neutral text.
	MessageFor func(textStr, better *rules.ConfusionString) string
	// Premium / Category / IssueType / Tags for SpecificIdRule (Java Rule metadata).
	Premium   bool
	Category  *rules.Category
	IssueType rules.ITSIssueType
	Tags      []rules.Tag
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

// AddExamplePair ports Rule.addExamplePair.
func (r *ConfusionProbabilityRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *ConfusionProbabilityRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *ConfusionProbabilityRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

func NewConfusionProbabilityRule(lm LanguageModel, grams int) *ConfusionProbabilityRule {
	if grams <= 0 {
		grams = 3
	}
	return &ConfusionProbabilityRule{
		LM:    lm,
		Grams: grams,
		// Java: setCategory(Categories.TYPOS); setLocQualityIssueType(NonConformance).
		// Messages filled by language wrappers (e.g. German) after construct.
		Category:  rules.CatTypos.GetCategory(nil),
		IssueType: rules.ITSNonConformance,
	}
}

// InitConfusionProbabilityMeta refreshes TYPOS category name from a message bundle.
func InitConfusionProbabilityMeta(r *ConfusionProbabilityRule, messages map[string]string) {
	if r == nil {
		return
	}
	r.Category = rules.CatTypos.GetCategory(messages)
	if r.IssueType == "" {
		r.IssueType = rules.ITSNonConformance
	}
}

func (r *ConfusionProbabilityRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *ConfusionProbabilityRule) GetLocQualityIssueType() rules.ITSIssueType {
	if r == nil || r.IssueType == "" {
		return rules.ITSNonConformance
	}
	return r.IssueType
}

func (r *ConfusionProbabilityRule) GetID() string {
	if r != nil && r.RuleIDOverride != "" {
		return r.RuleIDOverride
	}
	return ConfusionRuleID
}

// GetDescription ports getDescription() (rule-level; SpecificIdRule carries pair description).
func (r *ConfusionProbabilityRule) GetDescription() string {
	return r.msgOr("statistics_rule_description", "Detects potentially wrong usage of \"{0}\" instead of \"{1}\"")
}

// EstimateContextForSureMatch ports estimateContextForSureMatch → grams.
func (r *ConfusionProbabilityRule) EstimateContextForSureMatch() int {
	if r == nil || r.Grams <= 0 {
		return 3
	}
	return r.Grams
}

func (r *ConfusionProbabilityRule) msgOr(key, def string) string {
	if r != nil && r.Messages != nil {
		if s := r.Messages[key]; s != "" {
			return s
		}
	}
	return def
}

// pairDescription ports private getDescription(word1, word2).
func (r *ConfusionProbabilityRule) pairDescription(word1, word2 string) string {
	tmpl := r.msgOr("statistics_rule_description", "Detects potentially wrong usage of \"{0}\" instead of \"{1}\"")
	return strings.NewReplacer("{0}", word1, "{1}", word2).Replace(tmpl)
}

// shortDesc ports statistics_suggest_short_desc.
func (r *ConfusionProbabilityRule) shortDesc() string {
	return r.msgOr("statistics_suggest_short_desc", "Possible word confusion")
}

func (r *ConfusionProbabilityRule) SetWordToPairs(m map[string][]*rules.ConfusionPair) {
	r.WordToPairs = m
}

// SetConfusionPair ports setConfusionPair (tests): one pair indexed by both terms.
func (r *ConfusionProbabilityRule) SetConfusionPair(pair *rules.ConfusionPair) {
	if r == nil || pair == nil {
		return
	}
	r.WordToPairs = map[string][]*rules.ConfusionPair{}
	for _, w := range pair.GetTerms() {
		r.WordToPairs[w.GetString()] = []*rules.ConfusionPair{pair}
	}
}

// IsLocalException reports whether text contains a configured exception phrase covering a span.
// Soft helper for tests; Match uses covers() like Java.
func (r *ConfusionProbabilityRule) IsLocalException(text string) bool {
	if r == nil || text == "" {
		return false
	}
	low := strings.ToLower(text)
	for _, ex := range r.Exceptions {
		if ex != "" && strings.Contains(low, strings.ToLower(ex)) {
			return true
		}
	}
	return false
}

// PairsFor returns confusion pairs for a surface word (exact then lowercase key).
func (r *ConfusionProbabilityRule) PairsFor(word string) []*rules.ConfusionPair {
	if r == nil || r.WordToPairs == nil {
		return nil
	}
	if p := r.WordToPairs[word]; len(p) > 0 {
		return p
	}
	return r.WordToPairs[strings.ToLower(word)]
}

func (r *ConfusionProbabilityRule) tokenize() TokenizerFunc {
	if r != nil && r.Tokenize != nil {
		return r.Tokenize
	}
	wt := tokenizers.NewWordTokenizer()
	return wt.Tokenize
}

func (r *ConfusionProbabilityRule) commonWord(token string) bool {
	if r != nil && r.IsCommonWord != nil {
		return r.IsCommonWord(token)
	}
	return regexp.MustCompile(`^\w+$`).MatchString(token)
}

// Match ports ConfusionProbabilityRule.match. Nil LM or empty pairs → nil (fail-closed).
func (r *ConfusionProbabilityRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.LM == nil || sentence == nil || len(r.WordToPairs) == 0 {
		return nil
	}
	if r.Grams != 3 && r.Grams != 4 {
		return nil
	}
	text := sentence.GetText()
	tok := r.tokenize()
	tokens := GetGoogleTokens(text, true, tok)
	if len(tokens) == 2 {
		// only _START_ + one token — no real context
		return nil
	}
	var matches []*rules.RuleMatch
	realWordBefore := false
	for pos, googleToken := range tokens {
		token := googleToken.Token
		confusionPairs := r.WordToPairs[token]
		uppercase := false
		if confusionPairs == nil && token != "" {
			first, _ := utf8.DecodeRuneInString(token)
			if unicode.IsUpper(first) && !realWordBefore && isRealWord(token) {
				confusionPairs = r.WordToPairs[tools.LowercaseFirstChar(token)]
				uppercase = true
			}
		}
		if isRealWord(token) {
			realWordBefore = true
		}
		if confusionPairs == nil {
			continue
		}
		for _, confusionPair := range confusionPairs {
			if confusionPair == nil {
				continue
			}
			var pairs []*rules.ConfusionString
			if uppercase {
				pairs = confusionPair.GetUppercaseFirstCharTerms()
			} else {
				pairs = confusionPair.GetTerms()
			}
			if len(pairs) != 2 {
				continue
			}
			better := r.getBetterAlternativeOrNull(pos, tokens, pairs, confusionPair.GetFactor(), tok)
			if better == nil {
				continue
			}
			if r.IsException != nil && r.IsException(text, googleToken.StartPos, googleToken.EndPos) {
				continue
			}
			if !confusionPair.IsBidirectional() && better.GetString() == pairs[0].GetString() {
				// only A -> B: if better is term1, skip (Java: betterAlternative equals pairs.get(0))
				continue
			}
			if pos > 0 && tokens[pos-1].Token == GoogleSentenceStart && pos+1 < len(tokens) &&
				tokens[pos+1].Token != "" && !r.commonWord(tokens[pos+1].Token) {
				continue
			}
			if r.isLocalExceptionCovering(text, googleToken.StartPos, googleToken.EndPos) {
				continue
			}
			if r.IsCoveredByAntiPattern != nil && r.IsCoveredByAntiPattern(sentence, googleToken.StartPos, googleToken.EndPos) {
				continue
			}
			stringFromText := getConfusionString(pairs, googleToken.Token)
			msg := r.message(stringFromText, better)
			term1 := confusionPair.GetTerm1().GetString()
			term2 := confusionPair.GetTerm2().GetString()
			// Java: SpecificIdRule(getId()+"_"+cleanId(term1)+"_"+cleanId(term2), getDescription(term1,term2), …)
			id := r.GetID() + "_" + cleanConfusionID(term1) + "_" + cleanConfusionID(term2)
			desc := r.pairDescription(term1, term2)
			cat := r.Category
			if cat == nil {
				cat = rules.NewCategory(rules.CategoryTypos, "Typos")
			}
			issue := r.IssueType
			if issue == "" {
				issue = rules.ITSNonConformance
			}
			idRule := rules.NewSpecificIdRule(id, desc, r.Premium, cat, issue, r.Tags)
			rm := rules.NewRuleMatch(idRule, sentence, googleToken.StartPos, googleToken.EndPos, msg)
			rm.ShortMessage = r.shortDesc()
			rm.SetSuggestedReplacement(better.GetString())
			matches = append(matches, rm)
		}
	}
	return matches
}

func (r *ConfusionProbabilityRule) message(textStr, better *rules.ConfusionString) string {
	if r.MessageFor != nil {
		return r.MessageFor(textStr, better)
	}
	if textStr == nil || better == nil {
		return r.shortDesc()
	}
	// Java getMessage branches on optional ConfusionString descriptions; without
	// descriptions falls through to statistics_suggest3_new (MessagesBundle defaults).
	sug, wrong := better.GetString(), textStr.GetString()
	sugDesc, wrongDesc := "", ""
	if d := better.GetDescription(); d != nil {
		sugDesc = *d
	}
	if d := textStr.GetDescription(); d != nil {
		wrongDesc = *d
	}
	// Defaults from MessagesBundle.properties (not invented German).
	switch {
	case sugDesc != "" && wrongDesc != "":
		tmpl := r.msgOr("statistics_suggest1_new",
			"Please check whether ''{0}'' ({1}) might be the correct word here instead of ''{2}'' ({3}).")
		return strings.NewReplacer("{0}", sug, "{1}", sugDesc, "{2}", wrong, "{3}", wrongDesc).Replace(tmpl)
	case wrongDesc != "":
		tmpl := r.msgOr("statistics_suggest4_new",
			"Please check whether ''{0}'' might be the correct word here instead of ''{1}'' ({2}).")
		return strings.NewReplacer("{0}", sug, "{1}", wrong, "{2}", wrongDesc).Replace(tmpl)
	case sugDesc != "":
		tmpl := r.msgOr("statistics_suggest2_new",
			"Please check whether ''{0}'' ({1}) might be the correct word here instead of ''{2}''.")
		return strings.NewReplacer("{0}", sug, "{1}", sugDesc, "{2}", wrong).Replace(tmpl)
	default:
		tmpl := r.msgOr("statistics_suggest3_new",
			"Please check whether ''{0}'' might be the correct word here instead of ''{1}''.")
		return strings.NewReplacer("{0}", sug, "{1}", wrong).Replace(tmpl)
	}
}

func (r *ConfusionProbabilityRule) getBetterAlternativeOrNull(pos int, tokens []GoogleToken, pairs []*rules.ConfusionString, factor int64, tokenize TokenizerFunc) *rules.ConfusionString {
	if len(pairs) != 2 || pos < 0 || pos >= len(tokens) {
		return nil
	}
	token := tokens[pos]
	other := alternativeTerm(pairs, token.Token)
	if other == nil {
		return nil
	}
	word := token.Token
	var p1, p2 float64
	switch r.Grams {
	case 3:
		p1 = Get3gramProbabilityFor(r.LM, pos, tokens, word, tokenize)
		p2 = Get3gramProbabilityFor(r.LM, pos, tokens, other.GetString(), tokenize)
	case 4:
		p1 = Get4gramProbabilityFor(r.LM, pos, tokens, word, tokenize)
		p2 = Get4gramProbabilityFor(r.LM, pos, tokens, other.GetString(), tokenize)
	default:
		return nil
	}
	if p2 >= MinProb && p2 > p1*float64(factor) {
		return other
	}
	return nil
}

func (r *ConfusionProbabilityRule) isLocalExceptionCovering(text string, startPos, endPos int) bool {
	low := strings.ToLower(text)
	for _, exception := range r.Exceptions {
		if exception == "" {
			continue
		}
		exLow := strings.ToLower(exception)
		from := 0
		for {
			idx := strings.Index(low[from:], exLow)
			if idx < 0 {
				break
			}
			exStart := from + idx
			exEnd := exStart + len(exLow) // Java: exStart + exception.length() on original; ASCII exceptions only
			// Java: covers(exStart, exEnd, startPos, endPos)
			if covers(exStart, exEnd, startPos, endPos) {
				return true
			}
			from = exEnd
			if exEnd <= exStart {
				break
			}
		}
	}
	return false
}

func covers(exceptionStartPos, exceptionEndPos, startPos, endPos int) bool {
	return exceptionStartPos <= startPos && exceptionEndPos >= endPos
}

func isRealWord(token string) bool {
	return realWordRE.MatchString(token)
}

func alternativeTerm(pairs []*rules.ConfusionString, token string) *rules.ConfusionString {
	for _, s := range pairs {
		if s != nil && s.GetString() != token {
			return s
		}
	}
	return nil
}

func getConfusionString(pairs []*rules.ConfusionString, token string) *rules.ConfusionString {
	for _, s := range pairs {
		if s != nil && strings.EqualFold(s.GetString(), token) {
			return s
		}
	}
	if len(pairs) > 0 {
		return pairs[0]
	}
	return nil
}

func cleanConfusionID(id string) string {
	id = strings.ToUpper(id)
	r := strings.NewReplacer("Ä", "AE", "Ü", "UE", "Ö", "OE")
	return r.Replace(id)
}
