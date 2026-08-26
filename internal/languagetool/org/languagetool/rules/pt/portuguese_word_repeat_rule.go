package pt

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PortugueseWordRepeatRule ports org.languagetool.rules.pt.PortugueseWordRepeatRule.
type PortugueseWordRepeatRule struct {
	*rules.WordRepeatRule
}

// Patterns mirror Java PortugueseWordRepeatRule exactly:
//   TAUTONYMS_GENUS / TAUTONYMS_SPECIES — default (case-sensitive) Pattern
//   PRONOUNS — default (case-sensitive) Pattern.compile("mas|n?[ao]s?|se")
//   REDUPLICATED_ADVERBS — Pattern.CASE_INSENSITIVE
// Java matcher.matches() is full-string, so Go patterns are anchored.
var (
	tautonymsGenus = regexp.MustCompile(`A(?:aptos|canthogyrus|chatina|gagus|gama|lburnus|lces|lle|losa|mandava|mazilia|meiva|nableps|nguilla|nhinga|nostomus|nser|nthias|pus|rcinella|riadne|spredo|stacus|vicularia|xis)|B(?:adis|agarius|agre|alanus|anjos|arbatula|arbus|asiliscus|atasio|elobranchus|elone|elonimorphis|idyanus|ison|ombina|oops|rama|rosme|ubo|ucayana|ufo|uteo|utis)|C(?:alamus|alappa|aleta|allichthys|alotes|apoeta|apreolus|aracal|arassius|ardinalis|arduelis|aretta|asuarius|atla|atostomus|ephea|erastes|haca|halcides|handramara|hanos|haos|hinchilla|hiropotes|hitala|hromis|iconia|idaris|inclus|itellus|lelia|occothraustes|ochlearius|oeligena|olius|olumella|oncholepas|onger|onta|onvoluta|ordylus|oscoroba|ossus|otinga|oturnix|rangon|ressida|rex|ricetus|rocuta|rossoptilon|uraeus|yanicterus|ygnus|ymbium|ynoglossus)|D(?:ama|ario|entex|evario|iuca|ives|olabrifera)|E(?:nhydris|nsifera|nsis|rythrinus|xtra)|F(?:alcipennis|eroculus|icus|ragum|rancolinus|urcula)|G(?:agata|albula|allinago|allus|azella|emma|enetta|erbillus|ibberulus|iraffa|lis|lycimeris|lyphis|obio|oliathus|onorynchus|orilla|rapsus|rus|ryllotalpa|uira|ulo)|H(?:ara|arpa|austellum|emilepidotus|eterophyes|imantopus|ippocampus|ippoglossus|ippopus|istrio|istrionicus|oolock|ucho|uso|yaena|ypnale)|I(?:chthyaetus|cterus|dea|guana|ndicator|ndri)|J(?:acana|aculus|anthina)|K(?:achuga|oilofera)|L(?:actarius|agocephalus|agopus|agurus|ambis|emmus|epadogaster|erwa|euciscus|ima|imanda|imosa|iparis|ithognathus|ithophaga|oa|ota|uscinia|utjanus|utra|utraria|ynx)|M(?:acrophyllum|anacus|argaritifera|armota|artes|ascarinus|ashuna|egacephala|elanodera|eles|elo|elolontha|elongena|enidia|ephitis|ercenaria|eretrix|erluccius|eza|icrostoma|ilvus|itella|itra|itu|odiolus|odulus|ola|olossus|olva|onachus|oniliformis|ops|ustelus|yaka|yospalax|yotis)|N(?:aja|aja|angra|asua|atrix|eita|iviventer|otopterus|ycticorax)|O(?:enanthe|gasawarana|liva|phioscincus|plopomus|reotragus|riolus)|P(?:agrus|angasius|apio|auxi|erdix|eriphylla|erna|etaurista|etronia|hocoena|hoenicurus|hoxinus|hycis|ica|ipa|ipile|ipistrellus|ipra|ithecia|lanorbis|lica|oliocephalus|ollachius|ollicipes|orites|orphyrio|orphyrolaema|orpita|orzana|ristis|seudobagarius|udu|uffinus|ungitius|yrrhocorax|yrrhula)|Q(?:uadrula|uelea)|R(?:ama|anina|apa|asbora|attus|edunca|egulus|emora|etropinna|hinobatos|iparia|ita|upicapra|upicola|utilus)|S(?:accolaimus|alamandra|arda|calpellum|cincus|colytus|ephanoides|erinus|odreana|olea|phyraena|pinachia|pirorbis|pirula|prattus|quatina|taphylaea|uiriri|ula|uta|ynodus)|T(?:adorna|andanus|chagra|elescopium|emnurus|erebellum|etradactylus|etrax|herezopolis|hymallus|ibicen|inca|odus|orpedo|rachurus|rachycorystes|rachyrinchus|ricornis|roglodytes|ropheops|ubifex|yrannus)|U(?:mbraculum|ncia)|V(?:anellus|elella|elutina|icugna|illosa|imba|iviparus|olva|ulpes)|X(?:anthocephalus|anthostigma|enopirostris)|Ypiranga|Z(?:ebrus|era|ingel|ingha|oma|onia|ungaro|ygoneura)|Se`)
	tautonymsSpecies = regexp.MustCompile(`a(?:aptos|canthogyrus|chatina|gagus|gama|lburnus|lces|lle|losa|mandava|mazilia|meiva|nableps|nguilla|nhinga|nostomus|nser|nthias|pus|rcinella|riadne|spredo|stacus|vicularia|xis)|b(?:adis|agarius|agre|alanus|anjos|arbatula|arbus|asiliscus|atasio|elobranchus|elone|elonimorphis|idyanus|ison|ombina|oops|rama|rosme|ubo|ucayana|ufo|uteo|utis)|c(?:alamus|alappa|aleta|allichthys|alotes|apoeta|apreolus|aracal|arassius|ardinalis|arduelis|aretta|asuarius|atla|atostomus|ephea|erastes|haca|halcides|handramara|hanos|haos|hinchilla|hiropotes|hitala|hromis|iconia|idaris|inclus|itellus|lelia|occothraustes|ochlearius|oeligena|olius|olumella|oncholepas|onger|onta|onvoluta|ordylus|oscoroba|ossus|otinga|oturnix|rangon|ressida|rex|ricetus|rocuta|rossoptilon|uraeus|yanicterus|ygnus|ymbium|ynoglossus)|d(?:ama|ario|entex|evario|iuca|ives|olabrifera)|e(?:nhydris|nsifera|nsis|rythrinus|xtra)|f(?:alcipennis|eroculus|icus|ragum|rancolinus|urcula)|g(?:agata|albula|allinago|allus|azella|emma|enetta|erbillus|ibberulus|iraffa|lis|lycimeris|lyphis|obio|oliathus|onorynchus|orilla|rapsus|rus|ryllotalpa|uira|ulo)|h(?:ara|arpa|austellum|emilepidotus|eterophyes|imantopus|ippocampus|ippoglossus|ippopus|istrio|istrionicus|oolock|ucho|uso|yaena|ypnale)|i(?:chthyaetus|cterus|dea|guana|ndicator|ndri)|j(?:acana|aculus|anthina)|k(?:achuga|oilofera)|l(?:actarius|agocephalus|agopus|agurus|ambis|emmus|epadogaster|erwa|euciscus|ima|imanda|imosa|iparis|ithognathus|ithophaga|oa|ota|uscinia|utjanus|utra|utraria|ynx)|m(?:acrophyllum|anacus|argaritifera|armota|artes|ascarinus|ashuna|egacephala|elanodera|eles|elo|elolontha|elongena|enidia|ephitis|ercenaria|eretrix|erluccius|eza|icrostoma|ilvus|itella|itra|itu|odiolus|odulus|ola|olossus|olva|onachus|oniliformis|ops|ustelus|yaka|yospalax|yotis)|n(?:aja|aja|angra|asua|atrix|eita|iviventer|otopterus|ycticorax)|o(?:enanthe|gasawarana|liva|phioscincus|plopomus|reotragus|riolus)|p(?:agrus|angasius|apio|auxi|erdix|eriphylla|erna|etaurista|etronia|hocoena|hoenicurus|hoxinus|hycis|ica|ipa|ipile|ipistrellus|ipra|ithecia|lanorbis|lica|oliocephalus|ollachius|ollicipes|orites|orphyrio|orphyrolaema|orpita|orzana|ristis|seudobagarius|udu|uffinus|ungitius|yrrhocorax|yrrhula)|q(?:uadrula|uelea)|r(?:ama|anina|apa|asbora|attus|edunca|egulus|emora|etropinna|hinobatos|iparia|ita|upicapra|upicola|utilus)|s(?:accolaimus|alamandra|arda|calpellum|cincus|colytus|ephanoides|erinus|odreana|olea|phyraena|pinachia|pirorbis|pirula|prattus|quatina|taphylaea|uiriri|ula|uta|ynodus)|t(?:adorna|andanus|chagra|elescopium|emnurus|erebellum|etradactylus|etrax|herezopolis|hymallus|ibicen|inca|odus|orpedo|rachurus|rachycorystes|rachyrinchus|ricornis|roglodytes|ropheops|ubifex|yrannus)|u(?:mbraculum|ncia)|v(?:anellus|elella|elutina|icugna|illosa|imba|iviparus|olva|ulpes)|x(?:anthocephalus|anthostigma|enopirostris)|ypiranga|z(?:ebrus|era|ingel|ingha|oma|onia|ungaro|ygoneura)|se`)
	// Java: Pattern.compile("mas|n?[ao]s?|se") — NOT CASE_INSENSITIVE.
	ptPronouns = regexp.MustCompile(`^(?:mas|n?[ao]s?|se)$`)
	// Java: Pattern.compile("já|logo|fácil", Pattern.CASE_INSENSITIVE)
	ptRedupAdverbs = regexp.MustCompile(`(?i)^(?:já|logo|fácil)$`)
)

func NewPortugueseWordRepeatRule(messages map[string]string) *PortugueseWordRepeatRule {
	base := rules.NewWordRepeatRule(messages)
	base.IDOverride = "PORTUGUESE_WORD_REPEAT_RULE"
	// Java ctor: setCategory(Categories.REPETITIONS.getCategory(messages))
	// (overrides WordRepeatRule's MISC) + setLocQualityIssueType(Duplication) already on base.
	base.Category = rules.CatRepetitions.GetCategory(messages)
	// Java: addExamplePair(é é → é)
	base.AddExamplePair(
		rules.Wrong("Este <marker>é é</marker> apenas uma frase de exemplo."),
		rules.Fixed("Este <marker>é</marker> apenas uma frase de exemplo."),
	)
	r := &PortugueseWordRepeatRule{WordRepeatRule: base}
	// Match path: WordRepeatRule.Ignore calls ExtraIgnore then super.ignore names.
	base.ExtraIgnore = r.ptIgnore
	return r
}

// Ignore exposes ignore for unit tests (Java PortugueseWordRepeatRuleTest).
// Order: PT arms (ExtraIgnore) then WordRepeatRule super names — same as Match.
func (r *PortugueseWordRepeatRule) Ignore(tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	return r.WordRepeatRule.Ignore(tokens, position)
}

// ptIgnore ports PortugueseWordRepeatRule.ignore before super.ignore(...).
// Java order:
//  1. wordRepetitionOf("blá"|"se"|"sapiens"|"tuk")
//  2. isGenus(prev) && isSpecies(cur)
//  3. isHyphenated && isPronoun(cur)
//  4. isReduplicatedAdverb(cur)
//  5. super.ignore (via WordRepeatRule.Ignore after ExtraIgnore)
func (r *PortugueseWordRepeatRule) ptIgnore(tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	if wordRepetitionOfPT("blá", tokens, position) {
		return true // "blá blá"
	}
	if wordRepetitionOfPT("se", tokens, position) {
		return true // "se se"
	}
	if wordRepetitionOfPT("sapiens", tokens, position) {
		return true // "Homo sapiens sapiens"
	}
	if wordRepetitionOfPT("tuk", tokens, position) {
		return true // "tuk tuk"
	}
	if position > 0 &&
		tautonymsGenus.MatchString(tokens[position-1].GetToken()) &&
		tautonymsSpecies.MatchString(tokens[position].GetToken()) {
		return true // e.g. Vulpes vulpes / Aaptos aaptos
	}
	if isHyphenated(tokens, position) && isPronoun(tokens[position]) {
		return true // e.g. "Coloquem-na na sala." when split around hyphen
	}
	if position >= 0 && position < len(tokens) && isReduplicatedAdverb(tokens[position]) {
		return true // e.g. "Logo logo"
	}
	return false
}

// wordRepetitionOfPT ports WordRepeatRule.wordRepetitionOf: Token.equals (case-sensitive).
func wordRepetitionOfPT(word string, tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	return position > 0 &&
		tokens[position-1].GetToken() == word &&
		tokens[position].GetToken() == word
}

// isHyphenated ports PortugueseWordRepeatRule.isHyphenated:
// tokens[position-2].getToken().equals("-") && !(tokens[position-1].isWhitespaceBefore())
// Bounds guard avoids panic when position < 2 (Java would AIOOBE; Match never hits that path).
func isHyphenated(tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	if position < 2 {
		return false
	}
	return tokens[position-2].GetToken() == "-" && !tokens[position-1].IsWhitespaceBefore()
}

// isPronoun ports PortugueseWordRepeatRule.isPronoun (case-sensitive PRONOUNS).
func isPronoun(token *languagetool.AnalyzedTokenReadings) bool {
	if token == nil {
		return false
	}
	return ptPronouns.MatchString(token.GetToken())
}

// isReduplicatedAdverb ports PortugueseWordRepeatRule.isReduplicatedAdverb
// (CASE_INSENSITIVE; only checks tokens[position], not prev).
func isReduplicatedAdverb(token *languagetool.AnalyzedTokenReadings) bool {
	if token == nil {
		return false
	}
	return ptRedupAdverbs.MatchString(token.GetToken())
}
