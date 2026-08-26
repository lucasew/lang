package ar

import (
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	arsynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/ar"
	artag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ar"
)

//go:embed data/verb_trans_to_untrans2.txt
var transVerbFS embed.FS

// ArabicTransVerbRuleID ports ArabicTransVerbRule.AR_VERB_TRANS_INDIRECT_REPLACE
// → "AR_VERB_TRANSITIVE_IINDIRECT" (double-I typo upstream).
const ArabicTransVerbRuleID = "AR_VERB_TRANSITIVE_IINDIRECT"

// ArabicTransVerbFilePath ports ArabicTransVerbRule.FILE_NAME → "/ar/verb_trans_to_untrans2.txt".
const ArabicTransVerbFilePath = "/ar/verb_trans_to_untrans2.txt"

// ArabicTransVerbFileNames ports getFileNames().
var ArabicTransVerbFileNames = []string{ArabicTransVerbFilePath}

var (
	transVerbOnce sync.Once
	transVerbBase *rules.AbstractSimpleReplaceRule2
)

// loadTransVerbBase loads official /ar/verb_trans_to_untrans2.txt once
// (Java AbstractSimpleReplaceRule2.lazyInit via getWrongWords()).
func loadTransVerbBase() *rules.AbstractSimpleReplaceRule2 {
	transVerbOnce.Do(func() {
		f, err := transVerbFS.Open("data/verb_trans_to_untrans2.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		// Field twin of ArabicTransVerbRule + ASR2 defaults:
		// - id AR_VERB_TRANSITIVE_IINDIRECT
		// - description / short / message as Java overrides
		// - suggestions separator " أو "
		// - CaseSensitivy.CI (base default)
		// - Locale "ar"
		// - Categories.MISC
		// - ITSIssueType.Misspelling
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   ArabicTransVerbRuleID,
			// Java getDescription(): leading fatha diacritic on T (U+064E)
			Description:          "َTransitive verbs corrected to indirect transitive",
			ShortMsg:             "أفعال متعدية بحرف، يخطئ في تعديتها",
			MessageTemplate:      "'$match' الفعل خاطئ في التعدية بحرف: $suggestions",
			SuggestionsSeparator: " أو ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "ar",
			SubRuleSpecificIDs:   false,
			Category:             rules.CatMisc.GetCategory(nil),
			IssueType:            rules.ITSMisspelling,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, ArabicTransVerbFilePath); err != nil {
			panic(err)
		}
		// Java English placeholder example pair (as upstream):
		// addExamplePair(Example.wrong("The train arrived <marker>a hour</marker> ago."),
		//   Example.fixed("The train arrived <marker>an hour</marker> ago."));
		base.AddExamplePair(
			rules.Wrong("The train arrived <marker>a hour</marker> ago."),
			rules.Fixed("The train arrived <marker>an hour</marker> ago."),
		)
		transVerbBase = base
	})
	return transVerbBase
}

// ArabicTransVerbRule ports org.languagetool.rules.ar.ArabicTransVerbRule
// (extends AbstractSimpleReplaceRule2; Match fully overridden).
type ArabicTransVerbRule struct {
	*rules.AbstractSimpleReplaceRule2
	// tagger ports ArabicTransVerbRule.tagger (enableNewStylePronounTag; unused by Match).
	tagger *artag.ArabicTagger
	// tagmanager ports ArabicTransVerbRule.tagmanager.
	tagmanager *artag.ArabicTagManager
	// wrongWords ports ArabicTransVerbRule.wrongWords
	// (getWrongWords() list: last map is mFullNoSpace).
	wrongWords []map[string]rules.SuggestionWithMessage
}

// NewArabicTransVerbRule ports ArabicTransVerbRule(ResourceBundle messages).
func NewArabicTransVerbRule(messages map[string]string) *ArabicTransVerbRule {
	base := loadTransVerbBase()
	r := *base
	r.Messages = messages
	// Java: super.setCategory(Categories.MISC.getCategory(messages));
	r.Category = rules.CatMisc.GetCategory(messages)
	// Java: setLocQualityIssueType(ITSIssueType.Misspelling);
	r.IssueType = rules.ITSMisspelling

	// Java: tagger = new ArabicTagger(); tagger.enableNewStylePronounTag();
	// Field is unused by Match (lemma/POS come from the sentence analysis path);
	// keep a private instance so we do not mutate DefaultArabicTagger.
	artag.EnsureDefaultArabicTagger()
	var tagger *artag.ArabicTagger
	if artag.DefaultArabicTagger != nil && artag.DefaultArabicTagger.GetWordTagger() != nil {
		tagger = artag.NewArabicTagger(artag.DefaultArabicTagger.GetWordTagger())
	} else {
		tagger = artag.NewArabicTagger(nil)
	}
	tagger.EnableNewStylePronounTag()

	// Java: tagmanager = new ArabicTagManager();
	// synthesizer is ArabicSynthesizer.INSTANCE (static call site in generateNewForm).
	// Ensure INSTANCE is loadable when official dict is present.
	_ = arsynth.Instance()

	return &ArabicTransVerbRule{
		AbstractSimpleReplaceRule2: &r,
		tagger:                     tagger,
		tagmanager:                 artag.NewArabicTagManager(),
		// Java: wrongWords = getWrongWords();
		wrongWords: r.GetWrongWords(),
	}
}

// GetFileNames ports ArabicTransVerbRule.getFileNames().
func (r *ArabicTransVerbRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{ArabicTransVerbFilePath}
}

// GetShort ports ArabicTransVerbRule.getShort().
func (r *ArabicTransVerbRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports ArabicTransVerbRule.getMessage() (ASR2 API; Match builds a different msg).
func (r *ArabicTransVerbRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetSuggestionsSeparator ports getSuggestionsSeparator() → " أو ".
func (r *ArabicTransVerbRule) GetSuggestionsSeparator() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	if r.SuggestionsSeparator == "" {
		return ", "
	}
	return r.SuggestionsSeparator
}

// GetLocaleCode ports getLocale() → Locale("ar").
func (r *ArabicTransVerbRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports ArabicTransVerbRule.match (fully overridden vs ASR2).
func (r *ArabicTransVerbRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || sentence == nil {
		return nil
	}
	// Java: if (wrongWords.size() == 0)
	if len(r.wrongWords) == 0 {
		return nil
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	var ruleMatches []*rules.RuleMatch
	prevTokenIndex := 0
	for i := 1; i < len(tokens); i++ { // ignoring token 0, i.e., SENT_START
		token := tokens[i]
		if token == nil {
			continue
		}
		var prevToken *languagetool.AnalyzedTokenReadings
		var prevTokenStr string
		// Java: prevTokenStr = prevTokenIndex > 0 ? tokens[prevTokenIndex].getToken() : null;
		// Go: empty string stands for Java null (never enter block when prev unset).
		if prevTokenIndex > 0 {
			prevToken = tokens[prevTokenIndex]
			if prevToken != nil {
				prevTokenStr = prevToken.GetToken()
			}
		}

		if prevTokenStr != "" && prevToken != nil {
			// test if the first token is a verb
			isAttachedVerbTransitive := r.isAttachedTransitiveVerb(prevToken)

			// test if the preposition token is suitable for verb token (previous)
			prepositions := r.getProperPrepositionForTransitiveVerb(prevToken)

			isRightPreposition := r.isRightPreposition(token, prepositions)

			// the verb is attached and the next token is not the suitable preposition
			// we give the correct new form
			if isAttachedVerbTransitive && !isRightPreposition {
				verb := r.getCorrectVerbForm(tokens[prevTokenIndex])
				// generate suggestion according to suggested prepositions
				// FIXME: test all suggestions (Java)
				newPreposition := prepositions[0]
				preposition := r.getCorrectPrepositionForm(newPreposition, prevToken)

				replacement := verb + " " + preposition
				// Java exact spacing:
				// "قل <suggestion>" + replacement + "</suggestion> بدلا من '" + prevTokenStr + "' لأنّ الفعل " +
				//   " متعد بحرف  ."
				msg := "قل <suggestion>" + replacement + "</suggestion> بدلا من '" + prevTokenStr + "' لأنّ الفعل " +
					" متعد بحرف  ."
				// Java:
				// new RuleMatch(this, sentence,
				//   tokens[prevTokenIndex].getStartPos(), tokens[prevTokenIndex].getEndPos(),  // from/to
				//   tokens[prevTokenIndex].getStartPos(), token.getEndPos(),                   // pattern
				//   msg, "خطأ في الفعل المتعدي بحرف");
				fromPos := tokens[prevTokenIndex].GetStartPos()
				toPos := tokens[prevTokenIndex].GetEndPos()
				patternFrom := fromPos
				patternTo := token.GetEndPos()
				rm := rules.NewRuleMatch(r, sentence, fromPos, toPos, msg)
				rm.SetPatternPosition(patternFrom, patternTo)
				rm.ShortMessage = "خطأ في الفعل المتعدي بحرف"
				// Java RuleMatch ctor extracts <suggestion>…</suggestion> into replacements
				rm.SetSuggestedReplacement(replacement)
				ruleMatches = append(ruleMatches, rm)
			}
		}

		if r.isAttachedTransitiveVerb(token) {
			prevTokenIndex = i
		} else {
			prevTokenIndex = 0
		}
	}
	return ruleMatches
}

// isAttachedTransitiveVerb ports isAttachedTransitiveVerb:
// any reading with non-null POSTag whose lemma is in last wrongWords map.
func (r *ArabicTransVerbRule) isAttachedTransitiveVerb(mytoken *languagetool.AnalyzedTokenReadings) bool {
	if mytoken == nil || r == nil || len(r.wrongWords) == 0 {
		return false
	}
	last := r.wrongWords[len(r.wrongWords)-1]
	for _, verbTok := range mytoken.GetReadings() {
		if verbTok == nil {
			continue
		}
		// Java: if (verbPostag != null)
		if verbTok.GetPOSTag() == nil {
			continue
		}
		verbLemma := ""
		if verbTok.GetLemma() != nil {
			verbLemma = *verbTok.GetLemma()
		}
		// Java: wrongWords.get(wrongWords.size() - 1).get(verbLemma) — lemma as-is
		// (CI lowercased keys on load; Arabic letters are case-invariant).
		if _, ok := last[verbLemma]; ok {
			return true
		}
	}
	return false
}

// getProperPrepositionForTransitiveVerb ports getProperPrepositionForTransitiveVerb.
func (r *ArabicTransVerbRule) getProperPrepositionForTransitiveVerb(mytoken *languagetool.AnalyzedTokenReadings) []string {
	if mytoken == nil || r == nil || len(r.wrongWords) == 0 {
		return nil
	}
	last := r.wrongWords[len(r.wrongWords)-1]
	for _, verbTok := range mytoken.GetReadings() {
		if verbTok == nil {
			continue
		}
		if verbTok.GetPOSTag() == nil {
			continue
		}
		verbLemma := ""
		if verbTok.GetLemma() != nil {
			verbLemma = *verbTok.GetLemma()
		}
		if swm, ok := last[verbLemma]; ok {
			// Java: Arrays.asList(verbLemmaMatch.getSuggestion().split("\\|"));
			return strings.Split(swm.Suggestion, "|")
		}
	}
	return nil
}

// isRightPreposition ports isRightPreposition: next token first reading lemma in list.
func (r *ArabicTransVerbRule) isRightPreposition(nextToken *languagetool.AnalyzedTokenReadings, prepositionList []string) bool {
	// Java:
	// String nextTokenStr = nextToken.getReadings().get(0).getLemma();
	// return prepositionList.contains(nextTokenStr);
	if nextToken == nil {
		return false
	}
	rds := nextToken.GetReadings()
	if len(rds) == 0 || rds[0] == nil {
		return false
	}
	var nextTokenStr *string
	nextTokenStr = rds[0].GetLemma() // may be nil (Java null)
	for _, p := range prepositionList {
		if nextTokenStr == nil {
			// Java List.contains(null) is true only if list has a null element — never here.
			continue
		}
		if p == *nextTokenStr {
			return true
		}
	}
	return false
}

// getCorrectVerbForm ports getCorrectVerbForm → generateUnattachedNewForm.
func (r *ArabicTransVerbRule) getCorrectVerbForm(token *languagetool.AnalyzedTokenReadings) string {
	return r.generateUnattachedNewForm(token)
}

// getCorrectPrepositionForm ports getCorrectPrepositionForm → generateAttachedNewForm.
func (r *ArabicTransVerbRule) getCorrectPrepositionForm(prepositionLemma string, prevtoken *languagetool.AnalyzedTokenReadings) string {
	return r.generateAttachedNewForm(prepositionLemma, prevtoken)
}

// generateNewForm ports generateNewForm: setFlag PRONOUN, optional OPTION 'D', synthesize.
func (r *ArabicTransVerbRule) generateNewForm(word, posTag string, flag rune) string {
	if r == nil {
		return ""
	}
	tm := r.tagmanager
	if tm == nil {
		tm = artag.NewArabicTagManager()
	}
	// generate new from word form
	newposTag := tm.SetFlag(posTag, "PRONOUN", flag)
	// FIXME: remove the specific flag for option D (Java)
	if flag != '-' {
		newposTag = tm.SetFlag(newposTag, "OPTION", 'D')
	}
	// generate the new preposition according to modified postag
	// Java: AnalyzedToken prepAToken = new AnalyzedToken(word, newposTag, word);
	prepAToken := languagetool.NewAnalyzedTokenStr(word, newposTag, word, false, false)
	// Java: ArabicSynthesizer.INSTANCE.synthesize(prepAToken, newposTag);
	synth := arsynth.Instance()
	if synth == nil {
		// fail-soft without official synth dict: empty forms (Java would NPE if INSTANCE missing)
		return ""
	}
	newwordList, _ := synth.Synthesize(prepAToken, newposTag)
	newWord := ""
	if len(newwordList) != 0 {
		newWord = newwordList[0]
	}
	return newWord
}

// generateUnattachedNewForm ports generateUnattachedNewForm (flag '-').
func (r *ArabicTransVerbRule) generateUnattachedNewForm(token *languagetool.AnalyzedTokenReadings) string {
	if token == nil {
		return ""
	}
	rds := token.GetReadings()
	if len(rds) == 0 || rds[0] == nil {
		return ""
	}
	// Java: token.getReadings().get(0).getLemma() / getPOSTag()
	lemma := ""
	if rds[0].GetLemma() != nil {
		lemma = *rds[0].GetLemma()
	}
	postag := ""
	if rds[0].GetPOSTag() != nil {
		postag = *rds[0].GetPOSTag()
	}
	return r.generateNewForm(lemma, postag, '-')
}

// generateAttachedNewForm ports generateAttachedNewForm:
// postag "PR-;---;---" + pronoun flag from previous token.
func (r *ArabicTransVerbRule) generateAttachedNewForm(prepositionLemma string, prevtoken *languagetool.AnalyzedTokenReadings) string {
	// FIXME ; generate multiple cases (Java)
	postag := "PR-;---;---"
	if prevtoken == nil {
		return r.generateNewForm(prepositionLemma, postag, '-')
	}
	rds := prevtoken.GetReadings()
	if len(rds) == 0 || rds[0] == nil {
		return r.generateNewForm(prepositionLemma, postag, '-')
	}
	prevPosTag := ""
	if rds[0].GetPOSTag() != nil {
		prevPosTag = *rds[0].GetPOSTag()
	}
	tm := r.tagmanager
	if tm == nil {
		tm = artag.NewArabicTagManager()
	}
	flag := tm.GetFlag(prevPosTag, "PRONOUN")
	return r.generateNewForm(prepositionLemma, postag, flag)
}
