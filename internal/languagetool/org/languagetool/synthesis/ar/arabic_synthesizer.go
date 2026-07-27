package ar

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	artag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ar"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const (
	ArabicSynthDict = "/ar/arabic_synth.dict"
	ArabicTagsFile  = "/ar/arabic_tags.txt"
)

// ArabicSynthesizer ports org.languagetool.synthesis.ar.ArabicSynthesizer.
type ArabicSynthesizer struct {
	*synthesis.BaseSynthesizer
	tagmanager *artag.ArabicTagManager
	// tagger ports ArabicSynthesizer.tagger (used by setEnclitic*/inflectLemmaLike).
	tagger *artag.ArabicTagger
}

func NewArabicSynthesizer(manual *synthesis.ManualSynthesizer) *ArabicSynthesizer {
	base := synthesis.NewBaseSynthesizer("ar", manual)
	base.ResourceFileName = ArabicSynthDict
	base.TagFileName = ArabicTagsFile
	return &ArabicSynthesizer{
		BaseSynthesizer: base,
		tagmanager:      artag.NewArabicTagManager(),
		// Java: private final ArabicTagger tagger = new ArabicTagger();
		// Dict-backed DefaultArabicTagger is preferred; set lazily in getTagger().
		tagger: nil,
	}
}

// INSTANCE ports ArabicSynthesizer.INSTANCE (lazy official load via Instance()).
// Prefer Instance(); INSTANCE is kept in sync on successful load for Java-field naming.
var INSTANCE *ArabicSynthesizer

// bindINSTANCE is called from Instance() once so INSTANCE matches Java singleton field name.
func bindINSTANCE(s *ArabicSynthesizer) {
	INSTANCE = s
}

// Synthesize ports ArabicSynthesizer.synthesize(token, posTag):
// lookup lemma|posTag then correctStem each form.
func (s *ArabicSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	return s.synthesizeExact(token, posTag)
}

// SynthesizeRE ports ArabicSynthesizer.synthesize(token, posTag, posTagRegExp).
// When regexp: correctTag the pattern, match possibleTags, lookup, correctStem with original posTag.
func (s *ArabicSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, posTagRegExp bool) ([]string, error) {
	if token == nil {
		return nil, nil
	}
	if posTag != "" && posTagRegExp {
		return s.synthesizeRegexp(token, posTag)
	}
	return s.synthesizeExact(token, posTag)
}

func (s *ArabicSynthesizer) synthesizeExact(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	if token == nil {
		return nil, nil
	}
	lemma := lemmaOf(token)
	if lemma == "" {
		return nil, nil
	}
	raw := s.lookupLemmaTag(lemma, posTag)
	out := make([]string, 0, len(raw))
	for _, wd := range raw {
		out = append(out, s.CorrectStem(wd, posTag))
	}
	return out, nil
}

func (s *ArabicSynthesizer) synthesizeRegexp(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	myPosTag := s.CorrectTag(posTag)
	if myPosTag == "" {
		return nil, nil
	}
	re, err := regexp.Compile("^(?:" + myPosTag + ")$")
	if err != nil {
		return nil, err
	}
	lemma := lemmaOf(token)
	if lemma == "" {
		return nil, nil
	}
	var results []string
	for _, tag := range s.possibleTags() {
		if !re.MatchString(tag) {
			continue
		}
		for _, wd := range s.lookupLemmaTag(lemma, tag) {
			// Java: correctStem(wd, posTag) with original (uncorrected) posTag
			results = append(results, s.CorrectStem(wd, posTag))
		}
	}
	return results, nil
}

func lemmaOf(token *languagetool.AnalyzedToken) string {
	if token == nil {
		return ""
	}
	if token.GetLemma() != nil && *token.GetLemma() != "" {
		return *token.GetLemma()
	}
	return token.GetToken()
}

func (s *ArabicSynthesizer) lookupLemmaTag(lemma, posTag string) []string {
	if s == nil || s.BaseSynthesizer == nil {
		return nil
	}
	// Mirror Base lookupForms without removal filter first — Arabic correctStem is the adjuster.
	var out []string
	if s.Lookup != nil {
		out = append(out, s.Lookup(lemma, posTag)...)
	}
	if s.Manual != nil {
		out = append(out, s.Manual.Lookup(lemma, posTag)...)
	}
	// Drop removed forms like BaseSynthesizer
	if s.Removal != nil {
		filtered := out[:0]
		for _, f := range out {
			removed := false
			for _, r := range s.Removal.Lookup(lemma, posTag) {
				if r == f {
					removed = true
					break
				}
			}
			if !removed {
				filtered = append(filtered, f)
			}
		}
		out = filtered
	}
	return out
}

func (s *ArabicSynthesizer) possibleTags() []string {
	if s == nil || s.BaseSynthesizer == nil {
		return nil
	}
	if len(s.PossibleTags) > 0 {
		return s.PossibleTags
	}
	if s.Manual != nil {
		var tags []string
		for t := range s.Manual.GetPossibleTags() {
			tags = append(tags, t)
		}
		return tags
	}
	return nil
}

func (s *ArabicSynthesizer) tm() *artag.ArabicTagManager {
	if s == nil {
		return artag.NewArabicTagManager()
	}
	if s.tagmanager == nil {
		s.tagmanager = artag.NewArabicTagManager()
	}
	return s.tagmanager
}

// CorrectTag ports ArabicSynthesizer.correctTag (used by getPosTagCorrection + regexp synth).
// null postag → empty (Java null).
func (s *ArabicSynthesizer) CorrectTag(postag string) string {
	if postag == "" {
		return ""
	}
	tm := s.tm()
	// remove attached pronouns (via conj clear) — Java setConjunction(mypostag, "-")
	mypostag := tm.SetConjunction(postag, "-")
	// remove Alef Lam definite article
	mypostag = tm.SetDefinite(mypostag, "-")
	// change all pronouns to one kind
	mypostag = tm.UnifyPronounTag(mypostag)
	return mypostag
}

// GetPosTagCorrection ports ArabicSynthesizer.getPosTagCorrection → correctTag.
func (s *ArabicSynthesizer) GetPosTagCorrection(posTag string) string {
	return s.CorrectTag(posTag)
}

// CorrectStem ports ArabicSynthesizer.correctStem — adjust form for attached
// pronouns / definite / jar / conjunction prefixes from the original posTag.
func (s *ArabicSynthesizer) CorrectStem(stem, postag string) string {
	if postag == "" {
		return stem
	}
	tm := s.tm()
	correctStem := stem
	if tm.IsAttached(postag) {
		// Java StringUtils.removeEnd(correctStem, "ه")
		correctStem = strings.TrimSuffix(correctStem, "ه")
	}
	if tm.IsDefinite(postag) {
		correctStem = tm.GetDefinitePrefix(postag) + correctStem
	}
	if tm.HasJar(postag) {
		correctStem = tm.GetJarPrefix(postag) + correctStem
	}
	if tm.HasConjunction(postag) {
		correctStem = tm.GetConjunctionPrefix(postag) + correctStem
	}
	return correctStem
}

// InflectMafoulMutlq ports ArabicSynthesizer.inflectMafoulMutlq (static morph rule).
func InflectMafoulMutlq(word string) string {
	if word == "" {
		return word
	}
	teh := string(tools.ArabicTehMarbuta)
	if strings.HasSuffix(word, teh) {
		return word + string(tools.ArabicFathatan)
	}
	return word + string(tools.ArabicFathatan) + string(tools.ArabicAlef)
}

// InflectAdjectiveTanwinNasb ports ArabicSynthesizer.inflectAdjectiveTanwinNasb.
func InflectAdjectiveTanwinNasb(word string, feminin bool) string {
	if word == "" {
		return word
	}
	teh := string(tools.ArabicTehMarbuta)
	if feminin {
		if strings.HasSuffix(word, teh) {
			return word + string(tools.ArabicFathatan)
		}
		return word + teh + string(tools.ArabicFathatan)
	}
	// masculine: strip teh marbuta if present
	if strings.HasSuffix(word, teh) {
		return strings.TrimSuffix(word, teh)
	}
	return word + string(tools.ArabicFathatan) + string(tools.ArabicAlef)
}

// Instance methods match Java instance call sites (same as static helpers).
func (s *ArabicSynthesizer) InflectMafoulMutlq(word string) string {
	return InflectMafoulMutlq(word)
}

func (s *ArabicSynthesizer) InflectAdjectiveTanwinNasb(word string, feminin bool) string {
	return InflectAdjectiveTanwinNasb(word, feminin)
}

func (s *ArabicSynthesizer) getTagger() *artag.ArabicTagger {
	if s != nil && s.tagger != nil {
		return s.tagger
	}
	// Prefer official dict-backed DefaultArabicTagger when available.
	artag.EnsureDefaultArabicTagger()
	if artag.DefaultArabicTagger != nil {
		return artag.DefaultArabicTagger
	}
	return artag.NewArabicTagger(nil)
}

// SetEnclitic ports ArabicSynthesizer.setEnclitic — first form of setEncliticMultiple.
func (s *ArabicSynthesizer) SetEnclitic(token *languagetool.AnalyzedToken, suffix string) string {
	list := s.SetEncliticMultiple(token, suffix)
	if len(list) == 0 {
		return ""
	}
	return list[0]
}

// SetEncliticMultiple ports ArabicSynthesizer.setEncliticMultiple.
func (s *ArabicSynthesizer) SetEncliticMultiple(token *languagetool.AnalyzedToken, suffix string) []string {
	if token == nil {
		return nil
	}
	postagPtr := token.GetPOSTag()
	postag := ""
	if postagPtr != nil {
		postag = *postagPtr
	}
	word := token.GetToken()
	defaultWordlist := []string{"(" + word + ")"}
	// Java: if (postag.isEmpty()) — null would NPE; Go treats nil as empty
	if postag == "" {
		return defaultWordlist
	}
	// flag is by default '-', if suffix non-empty => 'H'
	flag := rune('-')
	if suffix != "" {
		flag = 'H'
	}
	tagger := s.getTagger()
	tm := s.tm()
	// save procletic
	procletic := tagger.GetProclitic(token)
	// set enclitic flag
	newposTag := tm.SetFlag(postag, "PRONOUN", flag)
	// adjust procletics
	newposTag = tm.SetProcleticFlags(newposTag)
	// synthesis => lookup for stems with similar postag and has enclitic flag
	lemma := lemmaOf(token)
	lemPtr := lemma
	newToken := languagetool.NewAnalyzedToken(lemma, &newposTag, &lemPtr)
	newwordList, _ := s.Synthesize(newToken, newposTag)

	var wordlist []string
	if len(newwordList) != 0 {
		for _, stem := range newwordList {
			newWord := ""
			// We replace the Heh, because the Tag dictionary represent only Heh Pronouns
			if tm.HasPronoun(newposTag) && flag == 'H' {
				if strings.HasSuffix(stem, "ي") {
					// if the stem is ended by Yeh for 1st person pronoun
					if suffix == "ي" {
						newWord = procletic + stem
					} else {
						newWord = ""
					}
				} else if strings.HasSuffix(stem, "ه") {
					stem = strings.TrimSuffix(stem, "ه")
					newWord = procletic + stem + suffix
				} else {
					newWord = procletic + stem + suffix
				}
			} else {
				newWord = procletic + stem
			}
			if newWord != "" {
				wordlist = append(wordlist, newWord)
			}
		}
	} else {
		// no word generated
		wordlist = append(wordlist, "("+word+")")
	}
	if len(wordlist) == 0 {
		return defaultWordlist
	}
	return wordlist
}

// InflectLemmaLike ports ArabicSynthesizer.inflectLemmaLike — generate forms of
// targetLemma matching source token's morphology (proclitic/enclitic + merged POS).
func (s *ArabicSynthesizer) InflectLemmaLike(targetLemma string, sourcetoken *languagetool.AnalyzedToken) []string {
	if s == nil {
		return []string{"[" + targetLemma + "]"}
	}
	tagger := s.getTagger()
	// Java: tagger.tag(targetLemma) — single-token Tag
	tokenReadList := tagger.TagSingle(targetLemma)
	if tokenReadList == nil || !tokenReadList.HasLemma(targetLemma) {
		return []string{"[" + targetLemma + "]"}
	}
	sourcePostag := ""
	if sourcetoken != nil && sourcetoken.GetPOSTag() != nil {
		sourcePostag = *sourcetoken.GetPOSTag()
	}
	// get affixes
	prefix := tagger.GetProclitic(sourcetoken)
	suffix := tagger.GetEnclitic(sourcetoken)

	tm := s.tm()
	var wordlist []string
	for _, currentToken := range tokenReadList.GetReadings() {
		if currentToken == nil {
			continue
		}
		// if the lemma is not equals to given one, continue
		curLemma := ""
		if currentToken.GetLemma() != nil {
			curLemma = *currentToken.GetLemma()
		}
		if targetLemma != curLemma {
			continue
		}
		postagLemma := ""
		if currentToken.GetPOSTag() != nil {
			postagLemma = *currentToken.GetPOSTag()
		}
		mergedPostag := tm.MergePosTag(sourcePostag, postagLemma)
		// construct word = prefix + targetLemma
		word := prefix + targetLemma
		lem := targetLemma
		token := languagetool.NewAnalyzedToken(word, &mergedPostag, &lem)
		wordlist = append(wordlist, s.SetEncliticMultiple(token, suffix)...)
	}
	// remove duplicates (Java: new ArrayList<>(new HashSet<>(wordlist)))
	seen := make(map[string]struct{}, len(wordlist))
	out := make([]string, 0, len(wordlist))
	for _, w := range wordlist {
		if _, ok := seen[w]; ok {
			continue
		}
		seen[w] = struct{}{}
		out = append(out, w)
	}
	return out
}

var _ synthesis.Synthesizer = (*ArabicSynthesizer)(nil)
