package ca

import (
	"embed"
	"regexp"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Java CatalanTagger: /ca/ca-ES.dict
const CatalanDictPath = "/ca/ca-ES.dict"

//go:embed data/replace_verbs.txt
var replaceVerbsFS embed.FS

var (
	caAdjPartFS        = regexp.MustCompile(`^VMP00SF.|A[QO].[FC]S.$`)
	caVerb             = regexp.MustCompile(`^V.+$`)
	caPrefixesForVerbs = regexp.MustCompile(`(?i)^(auto)(.*[aeiouàéèíòóïü].+[aeiouàéèíòóïü].*)$`)
	caAdjectiuCompost  = regexp.MustCompile(`(?i)^(.*)o-(.*.*)$`)
	caTresAdjectius    = regexp.MustCompile(`(?i)^(.*)o-(.*)o-(.*.*)$`)
	caDesinencies0     = regexp.MustCompile(`(?i)^(.+?)(a|ada|ades|am|ant|ar|ara|aran|arem|aren|ares|areu|aria|arien|aries|arà|aràs|aré|aríem|aríeu|assen|asses|assin|assis|at|ats|au|ava|aven|aves|e|ec|ega|eguda|egudes|eguem|eguen|eguera|egueren|egueres|egues|eguessen|eguesses|eguessin|eguessis|egueu|egui|eguin|eguis|egut|eguts|egué|eguérem|eguéreu|egués|eguéssem|eguésseu|eguéssim|eguéssiu|eguí|eix|eixem|eixen|eixent|eixeran|eixerem|eixeren|eixeres|eixereu|eixeria|eixerien|eixeries|eixerà|eixeràs|eixeré|eixeríem|eixeríeu|eixes|eixessen|eixesses|eixessin|eixessis|eixeu|eixi|eixia|eixien|eixies|eixin|eixis|eixo|eixé|eixérem|eixéreu|eixés|eixéssem|eixésseu|eixéssim|eixéssiu|eixí|eixíem|eixíeu|em|en|es|esc|esca|escuda|escudes|escut|escuts|esquem|esquen|esquera|esqueren|esqueres|esques|esquessen|esquesses|esquessin|esquessis|esqueu|esqui|esquin|esquis|esqué|esquérem|esquéreu|esqués|esquéssem|esquésseu|esquéssim|esquéssiu|esquí|essen|esses|essin|essis|eu|i|ia|ida|ides|ien|ies|iguem|igueu|im|in|int|ir|ira|iran|irem|iren|ires|ireu|iria|irien|iries|irà|iràs|iré|iríem|iríeu|is|isc|isca|isquen|isques|issen|isses|issin|issis|it|its|iu|ix|ixen|ixes|o|à|àrem|àreu|às|àssem|àsseu|àssim|àssiu|àvem|àveu|èixer|éixer|és|éssem|ésseu|éssim|éssiu|í|íem|íeu|írem|íreu|ís|íssem|ísseu|íssim|íssiu|ïs)$`)
	caDesinencies1     = regexp.MustCompile(`(?i)^(.+)(a|ada|ades|am|ant|ar|ara|aran|arem|aren|ares|areu|aria|arien|aries|arà|aràs|aré|aríem|aríeu|assen|asses|assin|assis|at|ats|au|ava|aven|aves|e|ec|ega|eguda|egudes|eguem|eguen|eguera|egueren|egueres|egues|eguessen|eguesses|eguessin|eguessis|egueu|egui|eguin|eguis|egut|eguts|egué|eguérem|eguéreu|egués|eguéssem|eguésseu|eguéssim|eguéssiu|eguí|eix|eixem|eixen|eixent|eixeran|eixerem|eixeren|eixeres|eixereu|eixeria|eixerien|eixeries|eixerà|eixeràs|eixeré|eixeríem|eixeríeu|eixes|eixessen|eixesses|eixessin|eixessis|eixeu|eixi|eixia|eixien|eixies|eixin|eixis|eixo|eixé|eixérem|eixéreu|eixés|eixéssem|eixésseu|eixéssim|eixéssiu|eixí|eixíem|eixíeu|em|en|es|esc|esca|escuda|escudes|escut|escuts|esquem|esquen|esquera|esqueren|esqueres|esques|esquessen|esquesses|esquessin|esquessis|esqueu|esqui|esquin|esquis|esqué|esquérem|esquéreu|esqués|esquéssem|esquésseu|esquéssim|esquéssiu|esquí|essen|esses|essin|essis|eu|i|ia|ida|ides|ien|ies|iguem|igueu|im|in|int|ir|ira|iran|irem|iren|ires|ireu|iria|irien|iries|irà|iràs|iré|iríem|iríeu|is|isc|isca|isquen|isques|issen|isses|issin|issis|it|its|iu|ix|ixen|ixes|o|à|àrem|àreu|às|àssem|àsseu|àssim|àssiu|àvem|àveu|èixer|éixer|és|éssem|ésseu|éssim|éssiu|í|íem|íeu|írem|íreu|ís|íssem|ísseu|íssim|íssiu|ïs)$`)
	caAltresPrefixos   = []string{"greco", "sino", "italo", "franco", "gal·lo", "luso",
		"germano", "hispano", "anglo", "àrabo", "austro", "belgo"}
	caNoAltresPrefixos = []string{"grego", "xineso", "italiano", "franceso",
		"portugueso", "angleso", "espanyolo", "alemanyo", "arabo", "austríaco", "bèlgico"}
	caAllUpperExceptions = map[string]struct{}{"ARNAU": {}, "CRISTIAN": {}, "TOMÀS": {}}
	caAltresSet          = func() map[string]struct{} {
		m := make(map[string]struct{}, len(caAltresPrefixos))
		for _, s := range caAltresPrefixos {
			m[s] = struct{}{}
		}
		return m
	}()
	caNoAltresSet = func() map[string]struct{} {
		m := make(map[string]struct{}, len(caNoAltresPrefixos))
		for _, s := range caNoAltresPrefixos {
			m[s] = struct{}{}
		}
		return m
	}()

	wrongVerbsOnce sync.Once
	wrongVerbsMap  map[string][]string
)

func loadWrongVerbs() map[string][]string {
	wrongVerbsOnce.Do(func() {
		f, err := replaceVerbsFS.Open("data/replace_verbs.txt")
		if err != nil {
			wrongVerbsMap = map[string][]string{}
			return
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			wrongVerbsMap = map[string][]string{}
			return
		}
		wrongVerbsMap = m
	})
	return wrongVerbsMap
}

// CatalanTagger ports org.languagetool.tagging.ca.CatalanTagger.
type CatalanTagger struct {
	*tagging.BaseTagger
	isValencian bool
}

// NewCatalanTagger ports CatalanTagger for ca-ES (central).
func NewCatalanTagger(wt tagging.WordTagger) *CatalanTagger {
	return NewCatalanTaggerVariant(wt, false)
}

// NewCatalanTaggerValencian ports CatalanTagger for ca-ES-valencia.
func NewCatalanTaggerValencian(wt tagging.WordTagger) *CatalanTagger {
	return NewCatalanTaggerVariant(wt, true)
}

// NewCatalanTaggerVariant ports CatalanTagger(Language) with variant flag.
func NewCatalanTaggerVariant(wt tagging.WordTagger, isValencian bool) *CatalanTagger {
	// Java: super("/ca/ca-ES.dict", Locale("ca"), false)
	return &CatalanTagger{
		BaseTagger:  tagging.NewBaseTagger(wt, CatalanDictPath, "ca", false),
		isValencian: isValencian,
	}
}

func (t *CatalanTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0
	for _, originalWord := range sentenceTokens {
		// Java: typewriter apostrophe hack on originalWord (surface becomes typewriter).
		containsTypographicApostrophe := false
		word := originalWord
		if len(word) > 1 && strings.Contains(word, "’") {
			containsTypographicApostrophe = true
			word = strings.ReplaceAll(word, "’", "'")
		}
		normalizedWord := tools.NormalizeNFC(word)
		lowerWord := strings.ToLower(normalizedWord)
		isLowercase := normalizedWord == lowerWord
		isMixedCase := tools.IsMixedCase(normalizedWord)
		isAllUpper := tools.IsAllUppercase(normalizedWord)

		var analyzed []*languagetool.AnalyzedToken
		// normal case
		for _, tw := range t.TagWordExact(normalizedWord) {
			analyzed = append(analyzed, tagged(word, tw))
		}
		// non-lowercase (not mixed): also lower tags
		if !isLowercase && !isMixedCase {
			for _, tw := range t.TagWordExact(lowerWord) {
				analyzed = append(analyzed, tagged(word, tw))
			}
		}
		// all-uppercase proper nouns (ex. FRANÇA) or ALLUPPERCASE_EXCEPTIONS
		_, allUpperExc := caAllUpperExceptions[normalizedWord]
		if (len(analyzed) == 0 || allUpperExc) && isAllUpper {
			firstUpper := tools.UppercaseFirstChar(lowerWord)
			for _, tw := range t.TagWordExact(firstUpper) {
				analyzed = append(analyzed, tagged(word, tw))
			}
		}
		// additional tagging with prefixes
		if len(analyzed) == 0 && !isMixedCase {
			analyzed = append(analyzed, t.additionalTags(word)...)
		}
		// emoji
		if len(analyzed) == 0 && tools.IsEmoji(word) {
			p, l := "_emoji_", "_emoji_"
			analyzed = append(analyzed, languagetool.NewAnalyzedToken(word, &p, &l))
		}
		// Valencian POS filter
		t.filterAnalyzedTokensInPlace(&analyzed)
		// incorrect verbs
		isIncorrectVerb := false
		if len(analyzed) == 0 {
			tags := t.additionalTagsForIncorrectVerbs(word, lowerWord)
			if len(tags) > 0 {
				analyzed = append(analyzed, tags...)
				isIncorrectVerb = true
			}
		}
		if len(analyzed) == 0 {
			analyzed = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, nil, nil)}
		}
		atr := languagetool.NewAnalyzedTokenReadingsList(analyzed, pos)
		if containsTypographicApostrophe {
			atr.SetTypographicApostrophe(true)
		}
		if isIncorrectVerb {
			atr.SetChunkTags([]string{"_incorrect_verb_"})
		}
		out = append(out, atr)
		pos += tagging.UTF16Len(word)
	}
	return out
}

func (t *CatalanTagger) filterAnalyzedTokensInPlace(list *[]*languagetool.AnalyzedToken) {
	if list == nil || len(*list) == 0 {
		return
	}
	src := *list
	dst := make([]*languagetool.AnalyzedToken, 0, len(src))
	if t != nil && t.isValencian {
		for _, token := range src {
			if token == nil {
				continue
			}
			posTag := token.GetPOSTag()
			if posTag != nil && len(*posTag) > 0 && (*posTag)[0] == '0' {
				rest := (*posTag)[1:]
				lemma := token.GetLemma()
				var lp *string
				if lemma != nil {
					l := *lemma
					lp = &l
				}
				dst = append(dst, languagetool.NewAnalyzedToken(token.GetToken(), &rest, lp))
			} else {
				dst = append(dst, token)
			}
		}
		*list = dst
		return
	}
	// non-valencian: drop tags starting with 0
	for _, token := range src {
		if token == nil {
			continue
		}
		posTag := token.GetPOSTag()
		if posTag != nil && len(*posTag) > 0 && (*posTag)[0] == '0' {
			continue
		}
		dst = append(dst, token)
	}
	*list = dst
}

func (t *CatalanTagger) additionalTags(word string) []*languagetool.AnalyzedToken {
	if t == nil {
		return nil
	}
	lowerWord := tools.NormalizeNFC(strings.ToLower(word))
	// -ment adverbs
	if strings.HasSuffix(lowerWord, "ment") {
		possibleAdj := strings.TrimSuffix(lowerWord, "ment")
		for _, tw := range t.TagWordExact(possibleAdj) {
			if tw.PosTag != "" && caAdjPartFS.MatchString(tw.PosTag) {
				p, lemma := "RG", lowerWord
				return []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, &p, &lemma)}
			}
		}
	}
	// auto + verb (two vowels pattern)
	if m := caPrefixesForVerbs.FindStringSubmatch(word); m != nil {
		possibleVerb := tools.NormalizeNFC(strings.ToLower(m[2]))
		var out []*languagetool.AnalyzedToken
		for _, tw := range t.TagWordExact(possibleVerb) {
			if tw.Lemma == "nòmer" {
				continue
			}
			if tw.PosTag != "" && caVerb.MatchString(tw.PosTag) {
				p := tw.PosTag
				lemma := strings.ToLower(m[1]) + tw.Lemma
				out = append(out, languagetool.NewAnalyzedToken(word, &p, &lemma))
			}
		}
		return out
	}
	// folklòrico-popular
	if m := caAdjectiuCompost.FindStringSubmatch(word); m != nil {
		adj1 := strings.ToLower(m[1])
		if t.isValidAdjectiveForm(adj1) {
			adj2 := strings.ToLower(m[2])
			for _, tw := range t.TagWordExact(adj2) {
				if tw.PosTag != "" && strings.HasPrefix(tw.PosTag, "A") {
					p := tw.PosTag
					lemma := adj1 + "o-" + tw.Lemma
					return []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, &p, &lemma)}
				}
			}
		}
	}
	// franco-americano-alemany
	if m := caTresAdjectius.FindStringSubmatch(word); m != nil {
		adj1, adj2 := strings.ToLower(m[1]), strings.ToLower(m[2])
		if t.isValidAdjectiveForm(adj1) && t.isValidAdjectiveForm(adj2) {
			adj3 := strings.ToLower(m[3])
			for _, tw := range t.TagWordExact(adj3) {
				if tw.PosTag != "" && strings.HasPrefix(tw.PosTag, "A") {
					p := tw.PosTag
					lemma := adj1 + "o-" + adj2 + "o-" + tw.Lemma
					return []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, &p, &lemma)}
				}
			}
		}
	}
	// ela geminada deprecated chars
	if strings.Contains(word, "\u0140") || strings.Contains(word, "\u013f") {
		possibleWord := strings.ReplaceAll(lowerWord, "\u0140", "l·")
		var out []*languagetool.AnalyzedToken
		for _, tw := range t.TagWordExact(possibleWord) {
			out = append(out, tagged(word, tw))
		}
		return out
	}
	// -iste Valencian
	if t.isValencian && strings.HasSuffix(lowerWord, "iste") {
		possibleAdjNoun := strings.TrimSuffix(lowerWord, "iste") + "ista"
		var out []*languagetool.AnalyzedToken
		for _, tw := range t.TagWordExact(possibleAdjNoun) {
			switch tw.PosTag {
			case "NCCS000":
				p := "NCMS000"
				lemma := possibleAdjNoun
				out = append(out, languagetool.NewAnalyzedToken(word, &p, &lemma))
			case "AQ0CS0":
				p := "AQ0MS0"
				lemma := possibleAdjNoun
				out = append(out, languagetool.NewAnalyzedToken(word, &p, &lemma))
			}
			if len(out) > 0 {
				return out
			}
		}
	}
	return nil
}

func (t *CatalanTagger) wordformHasPostag(wordform, postag string) bool {
	for _, tw := range t.TagWordExact(wordform) {
		if tw.PosTag == postag {
			return true
		}
	}
	return false
}

func (t *CatalanTagger) isValidAdjectiveForm(wordStem string) bool {
	if _, bad := caNoAltresSet[wordStem+"o"]; bad {
		return false
	}
	if t.wordformHasPostag(wordStem+"a", "AQ0FS0") {
		return true
	}
	_, ok := caAltresSet[wordStem+"o"]
	return ok
}

func (t *CatalanTagger) additionalTagsForIncorrectVerbs(originalWord, lowerWord string) []*languagetool.AnalyzedToken {
	// en- prefix special cases
	if strings.HasPrefix(lowerWord, "en") {
		taggedWords := t.TagWordExact(lowerWord[2:])
		var selected []tagging.TaggedWord
		lemma := ""
		for _, tw := range taggedWords {
			if strings.HasPrefix(tw.PosTag, "V") {
				selected = append(selected, tagging.NewTaggedWord("en"+tw.Lemma, tw.PosTag))
				lemma = "en" + tw.Lemma
			}
		}
		if len(selected) > 0 && (lemma == "enfotre" || lemma == "enriure") {
			return t.asAnalyzedTokenListWithLemma(originalWord, lemma, selected)
		}
	}
	var additional []*languagetool.AnalyzedToken
	for _, pattern := range []*regexp.Regexp{caDesinencies0, caDesinencies1} {
		m := pattern.FindStringSubmatch(lowerWord)
		if m == nil {
			continue
		}
		baseLexeme := m[1]
		desinence := m[2]
		adjustedLexeme := baseLexeme
		lexemes := []string{baseLexeme}
		if strings.HasPrefix(desinence, "e") || strings.HasPrefix(desinence, "é") ||
			strings.HasPrefix(desinence, "i") || strings.HasPrefix(desinence, "ï") {
			adjustedLexeme = adjustLexemeForSoftVowel(baseLexeme)
			if adjustedLexeme != baseLexeme {
				lexemes = append(lexemes, adjustedLexeme)
			}
		}
		if strings.HasPrefix(desinence, "ï") {
			desinence = "i" + desinence[len("ï"):]
		}
		additional = t.tryTag(originalWord, adjustedLexeme+"ar", "cant"+desinence)
		for _, lex := range lexemes {
			if len(additional) == 0 {
				additional = t.tryTag(originalWord, lex+"ir", "serv"+desinence)
			}
			if len(additional) == 0 && strings.HasSuffix(lex, "g") {
				additional = t.tryTag(originalWord, lex+"uir", "serv"+desinence)
			}
		}
		if len(additional) == 0 {
			eixer := baseLexeme + "èixer"
			additional = t.tryTag(originalWord, eixer, "con"+desinence)
			if len(additional) == 0 {
				additional = t.tryTag(originalWord, eixer, "desmer"+desinence)
			}
		}
		if len(additional) > 0 {
			break
		}
	}
	return additional
}

func (t *CatalanTagger) tryTag(originalWord, infinitive, conjugated string) []*languagetool.AnalyzedToken {
	wrong := loadWrongVerbs()
	if _, ok := wrong[infinitive]; !ok {
		return nil
	}
	return t.asAnalyzedTokenListWithLemma(originalWord, infinitive, t.TagWordExact(conjugated))
}

func adjustLexemeForSoftVowel(lexeme string) string {
	switch {
	case strings.HasSuffix(lexeme, "c"):
		return lexeme[:len(lexeme)-len("c")] + "ç"
	case strings.HasSuffix(lexeme, "qu"):
		return lexeme[:len(lexeme)-len("qu")] + "c"
	case strings.HasSuffix(lexeme, "g"):
		return lexeme[:len(lexeme)-len("g")] + "j"
	case strings.HasSuffix(lexeme, "gü"):
		return lexeme[:len(lexeme)-len("gü")] + "gu"
	case strings.HasSuffix(lexeme, "gu"):
		return lexeme[:len(lexeme)-len("gu")] + "g"
	default:
		return lexeme
	}
}

func (t *CatalanTagger) asAnalyzedTokenListWithLemma(word, lemma string, taggedWords []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	var aTokenList []*languagetool.AnalyzedToken
	for _, tw := range taggedWords {
		if strings.HasPrefix(tw.PosTag, "V") {
			p := tw.PosTag
			l := lemma
			aTokenList = append(aTokenList, languagetool.NewAnalyzedToken(word, &p, &l))
		}
	}
	return aTokenList
}

func tagged(surface string, tw tagging.TaggedWord) *languagetool.AnalyzedToken {
	var pos, lemma *string
	if tw.PosTag != "" {
		p := tw.PosTag
		pos = &p
	}
	if tw.Lemma != "" {
		l := tw.Lemma
		lemma = &l
	}
	return languagetool.NewAnalyzedToken(surface, pos, lemma)
}
