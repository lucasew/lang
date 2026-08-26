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

//go:embed data/inflected_one_word.txt
var inflectedOneWordFS embed.FS

// ArabicInflectedOneWordReplaceRuleID ports
// ArabicInflectedOneWordReplaceRule.AR_INFLECTED_ONE_WORD_REPLACE → "AR_INFLECTED_ONE_WORD".
const ArabicInflectedOneWordReplaceRuleID = "AR_INFLECTED_ONE_WORD"

// ArabicInflectedOneWordFilePath ports ArabicInflectedOneWordReplaceRule.FILE_NAME
// → "/ar/inflected_one_word.txt".
const ArabicInflectedOneWordFilePath = "/ar/inflected_one_word.txt"

// ArabicInflectedOneWordFileNames ports getFileNames().
var ArabicInflectedOneWordFileNames = []string{ArabicInflectedOneWordFilePath}

var (
	inflectedOnce sync.Once
	inflectedBase *rules.AbstractSimpleReplaceRule2
)

// loadInflectedOneWordBase loads official /ar/inflected_one_word.txt once
// (Java AbstractSimpleReplaceRule2.lazyInit via getWrongWords()).
func loadInflectedOneWordBase() *rules.AbstractSimpleReplaceRule2 {
	inflectedOnce.Do(func() {
		f, err := inflectedOneWordFS.Open("data/inflected_one_word.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		// Field twin of ArabicInflectedOneWordReplaceRule + ASR2 defaults:
		// - id AR_INFLECTED_ONE_WORD
		// - description / short / message as Java overrides
		// - suggestions separator " أو "
		// - CaseSensitivy.CI (base default)
		// - Locale "ar"
		// - Categories.MISC (super + super.setCategory(MISC) in ctor)
		// - ITSIssueType.Inconsistency
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   ArabicInflectedOneWordReplaceRuleID,
			Description:          "قاعدة تطابق الكلمات التي يجب تجنبها وتقترح تصويبا لها",
			// Java getShort(): double space after أن
			ShortMsg:             "خطأ، يفضل أن  يقال:",
			// Java getMessage() (ASR2 API; Match builds a different msg)
			MessageTemplate:      " لا تقل '$match' بل قل: $suggestions",
			SuggestionsSeparator: " أو ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "ar",
			SubRuleSpecificIDs:   false,
			Category:             rules.CatMisc.GetCategory(nil),
			IssueType:            rules.ITSInconsistency,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, ArabicInflectedOneWordFilePath); err != nil {
			panic(err)
		}
		// Java:
		// addExamplePair(Example.wrong("أجريت <marker>أبحاثا</marker> في المخبر"),
		//                Example.fixed("أجريت <marker>بحوثا</marker> في المخبر."));
		base.AddExamplePair(
			rules.Wrong("أجريت <marker>أبحاثا</marker> في المخبر"),
			rules.Fixed("أجريت <marker>بحوثا</marker> في المخبر."),
		)
		inflectedBase = base
	})
	return inflectedBase
}

// ArabicInflectedOneWordReplaceRule ports
// org.languagetool.rules.ar.ArabicInflectedOneWordReplaceRule
// (extends AbstractSimpleReplaceRule2; Match fully overridden).
type ArabicInflectedOneWordReplaceRule struct {
	*rules.AbstractSimpleReplaceRule2
	// tagger ports ArabicInflectedOneWordReplaceRule.tagger
	// (constructed with enableNewStylePronounTag; unused by Match — Java stores it).
	tagger *artag.ArabicTagger
	// synthesizer ports ArabicInflectedOneWordReplaceRule.synthesizer.
	synthesizer *arsynth.ArabicSynthesizer
	// wrongWords ports ArabicInflectedOneWordReplaceRule.wrongWords
	// (getWrongWords() list: last map is mFullNoSpace).
	wrongWords []map[string]rules.SuggestionWithMessage
}

// NewArabicInflectedOneWordReplaceRule ports
// ArabicInflectedOneWordReplaceRule(ResourceBundle messages).
func NewArabicInflectedOneWordReplaceRule(messages map[string]string) *ArabicInflectedOneWordReplaceRule {
	base := loadInflectedOneWordBase()
	r := *base
	r.Messages = messages
	// Java: super.setCategory(Categories.MISC.getCategory(messages));
	r.Category = rules.CatMisc.GetCategory(messages)
	// Java: setLocQualityIssueType(ITSIssueType.Inconsistency);
	r.IssueType = rules.ITSInconsistency

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

	// Java: synthesizer = new ArabicSynthesizer(new Arabic());
	synth := arsynth.Instance()
	if synth == nil {
		// Fail-soft for environments without arabic_synth.dict: still construct
		// like Java would after loading — empty Instance means no forms.
		synth = arsynth.NewArabicSynthesizer(nil)
	}

	return &ArabicInflectedOneWordReplaceRule{
		AbstractSimpleReplaceRule2: &r,
		tagger:                     tagger,
		synthesizer:                synth,
		// Java: wrongWords = getWrongWords();
		wrongWords: r.GetWrongWords(),
	}
}

// GetFileNames ports ArabicInflectedOneWordReplaceRule.getFileNames().
func (r *ArabicInflectedOneWordReplaceRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{ArabicInflectedOneWordFilePath}
}

// GetShort ports ArabicInflectedOneWordReplaceRule.getShort().
func (r *ArabicInflectedOneWordReplaceRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports ArabicInflectedOneWordReplaceRule.getMessage() (ASR2 API).
func (r *ArabicInflectedOneWordReplaceRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetSuggestionsSeparator ports getSuggestionsSeparator() → " أو ".
func (r *ArabicInflectedOneWordReplaceRule) GetSuggestionsSeparator() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	if r.SuggestionsSeparator == "" {
		return ", "
	}
	return r.SuggestionsSeparator
}

// GetLocaleCode ports getLocale() → Locale("ar").
func (r *ArabicInflectedOneWordReplaceRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports ArabicInflectedOneWordReplaceRule.match (fully overridden vs ASR2).
func (r *ArabicInflectedOneWordReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || sentence == nil {
		return nil
	}
	// Java: if (wrongWords.size() == 0)
	if len(r.wrongWords) == 0 {
		return nil
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	var ruleMatches []*rules.RuleMatch
	// Java iterates all tokens including SENT_START (comment says "ignoring" but loop does not skip).
	for _, token := range tokens {
		if token == nil {
			continue
		}
		for _, wordTok := range token.GetReadings() {
			if wordTok == nil {
				continue
			}
			if !r.isCandidateWord(wordTok) {
				continue
			}
			// get suggestions
			var propositions []string
			sugMsg := ""
			propositionsWithMessage := r.getSuggestedWords(wordTok)
			if propositionsWithMessage != nil {
				propositions = strings.Split(propositionsWithMessage.Suggestion, "|")
				sugMsg = propositionsWithMessage.Message
				// Java: sugMsg = sugMsg != null ? sugMsg : "";
			}

			// generate suggestion according to suggested word
			var replacement strings.Builder
			var forms []string
			for _, proposition := range propositions {
				// Java does not trim propositions before inflect
				inflectedWordList := r.inflectSuggestedWords(proposition, wordTok)
				for _, w := range inflectedWordList {
					replacement.WriteString("<suggestion>")
					replacement.WriteString(w)
					replacement.WriteString("</suggestion>&nbsp;")
					forms = append(forms, w)
				}
			}
			// Java message (exact spacing):
			// "' الكلمة خاطئة " + token.getToken() + " ' ،" + sugMsg + ". استعمل  " + replacement
			msg := "' الكلمة خاطئة " + token.GetToken() + " ' ،" + sugMsg + ". استعمل  " + replacement.String()
			// Java short: "خطأ في استعمال كلمة:" + sugMsg
			shortMsg := "خطأ في استعمال كلمة:" + sugMsg
			rm := rules.NewRuleMatch(r, sentence, token.GetStartPos(), token.GetEndPos(), msg)
			rm.ShortMessage = shortMsg
			// Java RuleMatch ctor extracts <suggestion>…</suggestion> into replacements
			if len(forms) > 0 {
				rm.SetSuggestedReplacements(forms)
			}
			ruleMatches = append(ruleMatches, rm)
		}
	}
	return ruleMatches
}

// isCandidateWord ports isCandidateWord.
func (r *ArabicInflectedOneWordReplaceRule) isCandidateWord(mytoken *languagetool.AnalyzedToken) bool {
	return r.getSuggestedWords(mytoken) != nil
}

// getSuggestedWords ports getSuggestedWords: requires non-null POSTag; lemma in last wrongWords map.
func (r *ArabicInflectedOneWordReplaceRule) getSuggestedWords(mytoken *languagetool.AnalyzedToken) *rules.SuggestionWithMessage {
	if mytoken == nil || r == nil || len(r.wrongWords) == 0 {
		return nil
	}
	wordPostag := mytoken.GetPOSTag()
	// Java: if (wordPostag != null)
	if wordPostag == nil {
		return nil
	}
	wordLemma := ""
	if mytoken.GetLemma() != nil {
		wordLemma = *mytoken.GetLemma()
	}
	// lookup in WrongWords (last map = mFullNoSpace)
	// Java: wrongWords.get(wrongWords.size() - 1).get(wordLemma) — lemma as-is
	// (CI lowercased keys on load; Arabic letters are case-invariant).
	last := r.wrongWords[len(r.wrongWords)-1]
	if swm, ok := last[wordLemma]; ok {
		cp := swm
		return &cp
	}
	return nil
}

// inflectSuggestedWords ports inflectSuggestedWords → synthesizer.inflectLemmaLike.
func (r *ArabicInflectedOneWordReplaceRule) inflectSuggestedWords(targetLemma string, sourcetoken *languagetool.AnalyzedToken) []string {
	if r == nil || r.synthesizer == nil {
		// Java always has synthesizer; without resources mirror tag-miss path.
		return []string{"[" + targetLemma + "]"}
	}
	return r.synthesizer.InflectLemmaLike(targetLemma, sourcetoken)
}
