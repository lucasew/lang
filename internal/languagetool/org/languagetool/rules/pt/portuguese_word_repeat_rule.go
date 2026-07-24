package pt

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PortugueseWordRepeatRule ports org.languagetool.rules.pt.PortugueseWordRepeatRule.
type PortugueseWordRepeatRule struct {
	*rules.WordRepeatRule
}

var (
	tautonymsGenus   = regexp.MustCompile(`A(?:aptos|canthogyrus|chatina|gagus|gama|lburnus|lces|lle|losa|mandava|mazilia|meiva|nableps|nguilla|nhinga|nostomus|nser|nthias|pus|rcinella|riadne|spredo|stacus|vicularia|xis)|B(?:adis|agarius|agre|alanus|anjos|arbatula|arbus|asiliscus|atasio|elobranchus|elone|elonimorphis|idyanus|ison|ombina|oops|rama|rosme|ubo|ucayana|ufo|uteo|utis)|C(?:alamus|alappa|aleta|allichthys|alotes|apoeta|apreolus|aracal|arassius|ardinalis|arduelis|aretta|asuarius|atla|atostomus|ephea|erastes|haca|halcides|handramara|hanos|haos|hinchilla|hiropotes|hitala|hromis|iconia|idaris|inclus|itellus|lelia|occothraustes|ochlearius|oeligena|olius|olumella|oncholepas|onger|onta|onvoluta|ordylus|oscoroba|ossus|otinga|oturnix|rangon|ressida|rex|ricetus|rocuta|rossoptilon|uraeus|yanicterus|ygnus|ymbium|ynoglossus)|D(?:ama|ario|entex|evario|iuca|ives|olabrifera)|E(?:nhydris|nsifera|nsis|rythrinus|xtra)|F(?:alcipennis|eroculus|icus|ragum|rancolinus|urcula)|G(?:agata|albula|allinago|allus|azella|emma|enetta|erbillus|ibberulus|iraffa|lis|lycimeris|lyphis|obio|oliathus|onorynchus|orilla|rapsus|rus|ryllotalpa|uira|ulo)|H(?:ara|arpa|austellum|emilepidotus|eterophyes|imantopus|ippocampus|ippoglossus|ippopus|istrio|istrionicus|oolock|ucho|uso|yaena|ypnale)|I(?:chthyaetus|cterus|dea|guana|ndicator|ndri)|J(?:acana|aculus|anthina)|K(?:achuga|oilofera)|L(?:actarius|agocephalus|agopus|agurus|ambis|emmus|epadogaster|erwa|euciscus|ima|imanda|imosa|iparis|ithognathus|ithophaga|oa|ota|uscinia|utjanus|utra|utraria|ynx)|M(?:acrophyllum|anacus|argaritifera|armota|artes|ascarinus|ashuna|egacephala|elanodera|eles|elo|elolontha|elongena|enidia|ephitis|ercenaria|eretrix|erluccius|eza|icrostoma|ilvus|itella|itra|itu|odiolus|odulus|ola|olossus|olva|onachus|oniliformis|ops|ustelus|yaka|yospalax|yotis)|N(?:aja|aja|angra|asua|atrix|eita|iviventer|otopterus|ycticorax)|O(?:enanthe|gasawarana|liva|phioscincus|plopomus|reotragus|riolus)|P(?:agrus|angasius|apio|auxi|erdix|eriphylla|erna|etaurista|etronia|hocoena|hoenicurus|hoxinus|hycis|ica|ipa|ipile|ipistrellus|ipra|ithecia|lanorbis|lica|oliocephalus|ollachius|ollicipes|orites|orphyrio|orphyrolaema|orpita|orzana|ristis|seudobagarius|udu|uffinus|ungitius|yrrhocorax|yrrhula)|Q(?:uadrula|uelea)|R(?:ama|anina|apa|asbora|attus|edunca|egulus|emora|etropinna|hinobatos|iparia|ita|upicapra|upicola|utilus)|S(?:accolaimus|alamandra|arda|calpellum|cincus|colytus|ephanoides|erinus|odreana|olea|phyraena|pinachia|pirorbis|pirula|prattus|quatina|taphylaea|uiriri|ula|uta|ynodus)|T(?:adorna|andanus|chagra|elescopium|emnurus|erebellum|etradactylus|etrax|herezopolis|hymallus|ibicen|inca|odus|orpedo|rachurus|rachycorystes|rachyrinchus|ricornis|roglodytes|ropheops|ubifex|yrannus)|U(?:mbraculum|ncia)|V(?:anellus|elella|elutina|icugna|illosa|imba|iviparus|olva|ulpes)|X(?:anthocephalus|anthostigma|enopirostris)|Ypiranga|Z(?:ebrus|era|ingel|ingha|oma|onia|ungaro|ygoneura)|Se`)
	tautonymsSpecies = regexp.MustCompile(`a(?:aptos|canthogyrus|chatina|gagus|gama|lburnus|lces|lle|losa|mandava|mazilia|meiva|nableps|nguilla|nhinga|nostomus|nser|nthias|pus|rcinella|riadne|spredo|stacus|vicularia|xis)|b(?:adis|agarius|agre|alanus|anjos|arbatula|arbus|asiliscus|atasio|elobranchus|elone|elonimorphis|idyanus|ison|ombina|oops|rama|rosme|ubo|ucayana|ufo|uteo|utis)|c(?:alamus|alappa|aleta|allichthys|alotes|apoeta|apreolus|aracal|arassius|ardinalis|arduelis|aretta|asuarius|atla|atostomus|ephea|erastes|haca|halcides|handramara|hanos|haos|hinchilla|hiropotes|hitala|hromis|iconia|idaris|inclus|itellus|lelia|occothraustes|ochlearius|oeligena|olius|olumella|oncholepas|onger|onta|onvoluta|ordylus|oscoroba|ossus|otinga|oturnix|rangon|ressida|rex|ricetus|rocuta|rossoptilon|uraeus|yanicterus|ygnus|ymbium|ynoglossus)|d(?:ama|ario|entex|evario|iuca|ives|olabrifera)|e(?:nhydris|nsifera|nsis|rythrinus|xtra)|f(?:alcipennis|eroculus|icus|ragum|rancolinus|urcula)|g(?:agata|albula|allinago|allus|azella|emma|enetta|erbillus|ibberulus|iraffa|lis|lycimeris|lyphis|obio|oliathus|onorynchus|orilla|rapsus|rus|ryllotalpa|uira|ulo)|h(?:ara|arpa|austellum|emilepidotus|eterophyes|imantopus|ippocampus|ippoglossus|ippopus|istrio|istrionicus|oolock|ucho|uso|yaena|ypnale)|i(?:chthyaetus|cterus|dea|guana|ndicator|ndri)|j(?:acana|aculus|anthina)|k(?:achuga|oilofera)|l(?:actarius|agocephalus|agopus|agurus|ambis|emmus|epadogaster|erwa|euciscus|ima|imanda|imosa|iparis|ithognathus|ithophaga|oa|ota|uscinia|utjanus|utra|utraria|ynx)|m(?:acrophyllum|anacus|argaritifera|armota|artes|ascarinus|ashuna|egacephala|elanodera|eles|elo|elolontha|elongena|enidia|ephitis|ercenaria|eretrix|erluccius|eza|icrostoma|ilvus|itella|itra|itu|odiolus|odulus|ola|olossus|olva|onachus|oniliformis|ops|ustelus|yaka|yospalax|yotis)|n(?:aja|aja|angra|asua|atrix|eita|iviventer|otopterus|ycticorax)|o(?:enanthe|gasawarana|liva|phioscincus|plopomus|reotragus|riolus)|p(?:agrus|angasius|apio|auxi|erdix|eriphylla|erna|etaurista|etronia|hocoena|hoenicurus|hoxinus|hycis|ica|ipa|ipile|ipistrellus|ipra|ithecia|lanorbis|lica|oliocephalus|ollachius|ollicipes|orites|orphyrio|orphyrolaema|orpita|orzana|ristis|seudobagarius|udu|uffinus|ungitius|yrrhocorax|yrrhula)|q(?:uadrula|uelea)|r(?:ama|anina|apa|asbora|attus|edunca|egulus|emora|etropinna|hinobatos|iparia|ita|upicapra|upicola|utilus)|s(?:accolaimus|alamandra|arda|calpellum|cincus|colytus|ephanoides|erinus|odreana|olea|phyraena|pinachia|pirorbis|pirula|prattus|quatina|taphylaea|uiriri|ula|uta|ynodus)|t(?:adorna|andanus|chagra|elescopium|emnurus|erebellum|etradactylus|etrax|herezopolis|hymallus|ibicen|inca|odus|orpedo|rachurus|rachycorystes|rachyrinchus|ricornis|roglodytes|ropheops|ubifex|yrannus)|u(?:mbraculum|ncia)|v(?:anellus|elella|elutina|icugna|illosa|imba|iviparus|olva|ulpes)|x(?:anthocephalus|anthostigma|enopirostris)|ypiranga|z(?:ebrus|era|ingel|ingha|oma|onia|ungaro|ygoneura)|se`)
	ptPronouns       = regexp.MustCompile(`(?i)^(?:mas|n?[ao]s?|se)$`)
	ptRedupAdverbs   = regexp.MustCompile(`(?i)^(?:já|logo|fácil)$`)
)

func NewPortugueseWordRepeatRule(messages map[string]string) *PortugueseWordRepeatRule {
	base := rules.NewWordRepeatRule(messages)
	base.IDOverride = "PORTUGUESE_WORD_REPEAT_RULE"
	// Java: é é → é
	base.AddExamplePair(
		rules.Wrong("Este <marker>é é</marker> apenas uma frase de exemplo."),
		rules.Fixed("Este <marker>é</marker> apenas uma frase de exemplo."),
	)
	r := &PortugueseWordRepeatRule{WordRepeatRule: base}
	base.ExtraIgnore = r.ptIgnore
	return r
}

// Ignore exposes ignore for unit tests (Java PortugueseWordRepeatRuleTest).
func (r *PortugueseWordRepeatRule) Ignore(tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	return r.ptIgnore(tokens, position) || r.WordRepeatRule.Ignore(tokens, position)
}

func (r *PortugueseWordRepeatRule) ptIgnore(tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	if position <= 0 {
		return false
	}
	if wordRep(tokens, position, "blá") || wordRep(tokens, position, "se") ||
		wordRep(tokens, position, "sapiens") || wordRep(tokens, position, "tuk") {
		return true
	}
	if tautonymsGenus.MatchString(tokens[position-1].GetToken()) &&
		tautonymsSpecies.MatchString(tokens[position].GetToken()) {
		return true
	}
	if isHyphenated(tokens, position) && ptPronouns.MatchString(tokens[position].GetToken()) {
		return true
	}
	if ptRedupAdverbs.MatchString(tokens[position].GetToken()) &&
		strings.EqualFold(tokens[position-1].GetToken(), tokens[position].GetToken()) {
		return true
	}
	return false
}

func wordRep(tokens []*languagetool.AnalyzedTokenReadings, position int, word string) bool {
	return position > 0 &&
		strings.EqualFold(tokens[position-1].GetToken(), word) &&
		strings.EqualFold(tokens[position].GetToken(), word)
}

func isHyphenated(tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	if position < 2 {
		return false
	}
	return tokens[position-2].GetToken() == "-" && !tokens[position-1].IsWhitespaceBefore()
}
