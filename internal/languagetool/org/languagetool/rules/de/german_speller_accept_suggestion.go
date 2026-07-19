package de

import (
	"regexp"
	"strings"
)

// AcceptSuggestion ports GermanSpellerRule.acceptSuggestion + PREVENT_SUGGESTION_PATTERNS.

var preventSuggestionPatterns []*regexp.Regexp

func init() {
	registerPreventSuggestion(`.*(Majonäse|Bravur|Anschovis|Belkanto|Campagne|Frotté|Grisli|Jockei|Joga|Kalvinismus|Kanossa|Kargo|Ketschup|Kollier|Kommunikee|Masurka|Negligee|Nessessär|Poulard|Varietee|Wandalismus|kalvinist|[Ff]ick).*`)
	registerPreventSuggestion(`.+[*_:]in`)
	registerPreventSuggestion(`.+[*_:]innen`)
	registerPreventSuggestion(`.+\szigste[srnm]?`)
	registerPreventSuggestion(`[\wöäüÖÄÜß]+ [a-zöäüß]-[\wöäüÖÄÜß]+`)
	registerPreventSuggestion(`[\wöäüÖÄÜß]+- [\wöäüÖÄÜß]+`)
	registerPreventSuggestion(`[A-ZÄÖÜ][a-zäöüß]+-[a-zäöüß]+-[a-zäöüß]+`)
	registerPreventSuggestion(`[A-ZÄÖÜ][a-zäöüß]+- [a-zäöüßA-ZÄÖÜ\-]+`)
	registerPreventSuggestion(`[A-ZÄÖÜa-zäöüß\-]+ [a-zäöüßA-ZÄÖÜ]-[a-zäöüßA-ZÄÖÜ\-]+`)
	registerPreventSuggestion(`[A-ZÄÖÜa-zäöüß\-]+ [a-zäöüß\-]+-[A-ZÄÖÜ][a-zäöüß\-]+`)
	registerPreventSuggestion(`[\wöäüÖÄÜß]+ -[\wöäüÖÄÜß]+`)
	registerPreventSuggestion(`[A-ZÄÖÜa-zäöüß\-]+\.[A-ZÄÖÜa-zäöüß][A-ZÄÖÜa-zäöüß\-]+`)
	registerPreventSuggestion(`[A-ZÄÖÜa-zäöüß\-]+\.\-[a-zäöüß\-]+`)
	registerPreventSuggestion(`[a-zöäüß]{3,20} [A-ZÄÖÜ][a-zäöüß]{2,20}liche[rnsm]`)
	registerPreventSuggestion(`[A-ZÄÖÜ][a-zäöüß]{2,20}-[a-zäöüß]{2,20}-`)
	registerPreventSuggestion(`[a-zäöüß]{3,20}-[A-ZÄÖÜ][a-zäöüß\-]{2,20}`)
	registerPreventSuggestion(`[a-zäöüß]{3,20}-[A-ZÄÖÜ\-]{2,20}`)
	registerPreventSuggestion(`([skdm]?ein|viel|sitz|sing|web|hör|woh[nl]|kehr|adel|elektiv|wert|wein|wund|wurm|wand|weg|wett|gen|hei[lm]|kenn|vo[rnm]|fein|zu[rm]?|fehl|bei|peil|eckt?|mit|die|das|ehe|für|nur|eure[rn]?|unse?re?|e[sr]|fahr|bar|fern|warn|filz|oft|fort|bot|vote|käse|we[rnm]|was|gie(ss|ß)|haut|band|heiz|merk|mehr|z[äa]hl|knie|zie[lr]|braut|brat|park|reiz|wa[rs]|wo|ma(ß|ss)|kleb|gabel|brat|rast|rang|lesen?|arm|de[rnms]|sämig|sucht?|sägen?|steh|bahn|off|uff|auf|aß|also|anno|dank|back(en?)?|bl[oi]ck|fang|klär|macht?|haken?|[lw]agen?|messe?|bad(en?)?|pack|km|ecken?|bis|tauche?|tr?age?|segeln?|stei[lg]|stahl|da(nn)?|häng(en?)?[bt]oten?|plus|tat|lade?|tasten?|druck|fach|fragen?|lern|mag|facto|magre|bald|bau(en?)?|ich|sei[dtln]|gang|angeln?|[wl]ach|bist|[ge]ilt|warten?|turn|härten?|hold|[hg]alt|holt|angle|angab|ankam|anale?)-[A-ZÄÖÜa-zäöüß\-]+`)
	registerPreventSuggestion(`.+-(gen|tu[etn]|l?ehrt?(en?)?|[fv]iele?n?|gärt?en?|igeln?|nein|ja|d?rum|erb(en?)?|vo[rnm]|vors|hat|gab(en)?|gabs?|gibt|km|geb(en?)?|nu[nr]|gay|kalt(e[snr]?)?|la[gd](en?)?|man|rängen?|nässen?|angle|angeln?|angst|stur(en?)?|oft|wo|wann|was|wer|mengen?|spie(ß|ss)en?|adeln?|näht?en?|ob|beide[rn]?|gärten|zweiten?|hütt?en?|kehrt?en?|h?orten?|messen?|tr[ea]u|trüb|trüben?|senden?|gr[uo]b|feinden?|wie|käsen?|ih[rmn](e[srnm]?)?|grau|trug(en?)?|weil|dass|sein?|zucken?|kanten?|s?ich|getan|hält|bald|ärgern?|fächern?|wart?(en?)?|leid|weit(e[snr]?)?|weiden?|ruf(en?)?|min|im|bin|zicken?|jo|siegeln?|[ao]ha|ganz|zäh|jäh|gehen?|ga[br]|kam|sah|[sr]itzen|kann|mit|ohne|ist|so|war|da[rh]in|über|unter|doof|bis|sie|er|aalen?|[lb]aden?|raten?|die|mit|bis|d[ea]s|eifern?|acker[tn]?|z[iu]cken?|j[oe]|jäh|haha|gerät|[wrbfk]etten?|tja|je|kau|nach|haben?|hab|gaga|kicken?|kick|heil|heilen?|altern?|wänden?|wert(e[rsnm]?)?|werben?|zoom|genug|gehen?|ums?|und|oder|[sn]ah|ha|de[mnsr]|sü(ß|ss)|ringen?|dingen?|seil|au[fs]|gurten?|munden?|eigen|wenden?|regen?|b?rechen?|legen?|fächern?|leger|g[ia]lt|heim|heimen?|[mksdw]?ein|[mksdw]?einen?|erden?|ändern?|ernten?|bänden?|ästen?|arten?|kanten?|eichen?|unken?|wunden?|kunden?|runden?|regeln?|kegeln?|krähen?|zechen?|mähen?|ehren?|ehen?|enden?|eng(e[srn]?)?|gut(e[srn]?)?|zielt?(en?)?|spielt?(en?)?|ätzt?(en?)?|riegeln?|segeln?|engt?|engen?|angeln?|kochen?|[lk]ehren?|festen?|essen?|steuern?|ekeln?|irren?|cum|de|da|du|raus|rein|dort|knien?|hin|zu[rm]?|ritten?|riss|rissen?|[tr]ast(en?)?|rasseln?|hieb|wässern?|putz|hängen?|zinken?|a[bnm]|bisher|schöne?|solo|haken?|dr[üu]ck(en?|tot)?|huren?|pries|hupen?|hüllen?|lang|joa|sei[dt]|weist|üben?|ufern?|iss|steck(en?)?|fort|mal|aal|darf|halt(en?)?|eifern?|van|guck(en?|t)?|ganze?|acht(en?)?|auch|solo|[zs]og|lagern?|baggern?|au|haut?|als|uns|bei[m]?|[dm]ir|dich|uni|ergo|eich(en?)?|spick(en?)?|e[rs]|spielt?|we[hg]|wart|wi[rl]d|neue[rns]?|mithin|tags?|eine[snmr]?|wiesen?|rei[sz]en?|wei[sh]en?|siegen?|sag(en?)?|sitzen?|tagen?|all(en?)?|zahlen?|rügen?|ruhen?|bar|hüben?|hick|arm|armen?|plan(en?)?|[fpl]assen?|per|reg|rinnen?|bringen?|öl(en?)?|alt(en?)?|elf(en?)?|kp|ward|apart|wer[dkt](en?)?|weis(en?)?|sind|mm|wand|wir|licht(en)?|lügen?|loch(en?)?|übel|peu|[wtm]isch(en?)?|fein(e[rns]?)?|a(ß|ss)|mol|neu(en?)?|[dm]ich|rang|obe[nr]|übe[nl]?|maxi?|hart(en?)?|hexen?|ab|zück(en?)?|zurück|köpf(en?)?|band(en?)?|schafft?en?|schalt?en?|giften?|sieben?|seil(en?)?|wehen?|sehen?|s[it]?eht?|stocken?|red|rät|ma(ß|ss)|schämen?|innen?|karren?|wer[tf]en?|werft|loch(en?)?|logen?|gossen?|steil(en?)?|fr?isch(en?)?|d[ea]nn|zelt(en?)?|luv|kauf(en?)?|lasch(en?)?|bei(ß|ss)(en?)?|leihen?|leid(en?)?|[drsl]icht(en?)?|opfern?|[wz]äh[mln]en?|wär(en?)?|À|à|fugen?|la[xs]|zahl(en?)?|[rf]all(en?)?|wichs(en?)?|sog(en?)?|alias|glich(en?)?|würd(en?)?|wärm(en?)?|[rhg]eiz(en?)?|stieren?|teils?|trotz|fahr(en?)?|b[oa]u?[dt](en?)?|kl[öo]n(en?)?|paar|park(en?)?|last|landen?|alle[rnms]?|ad|l[äa]u[ft](en?)?|[ws]äg(en?)?|pasch(en?)?|kehl(en?)?|wohl(en?)?|flucht?(en?)?|zeit|rasa|selben?|mehr(en?)?|gabeln?|ordern?|[cw]ach(en?)?|arg(en?)?|brauch(en?)?|hauch(en?)?|[ms]a(ß|ss)(en?)?|mm?h|zart(e[snmr]?)?|ehrt?(en?)?|de[rn]en|ähm?|hui|hmm?|al|für|[bl]au(en?)?|[lr]ahm(en?)?|[bs]uch(en?)?|[wv]ag(en?)?|[tl]os(en?)?|les(en?)?|str?ahl(en?)?|zäh[mn]t?(en?)?|fest(e[rsnm]?)?|folgt?(en?)?|f[aä]llt?(en?)?|[tr]oll(en?)?|[mf]üllt?(en?)?|[rl]eit(en?)?|ras(en?)?|hall(en?)?|well(en?)?|fra(ß|ss)(en)?|tat(en)?|pah|buh(en?)?|bäh|hör(en?)?|holz(en?)?|reif(e[rsmn]?)?|litt|fort(an)?|härten?|welche[rnsm]?|wegen|fach(en?)?|bog(en?)?|foul(en?)?|löst?(en?)?|lots(en?)?|falls|[bwh][ua]ldige[rsn]?|(st)?reift?(en?)?|t?rei[bh](en?)?|[rb]ück(en?)?|wett(en?)?|t[oü]t(en?)?|[ft]est(en?)?|h[aä]ut(en?)?|knall(en?)?|[dk]ämpft?(en?)?|hört?(en?)?|patt(en?)?|[tw]ollt?en?|[km]g|[bkps]ack(en?)?|[lf]an?d(en?)?|seifen?|tabu|heft(en?)?|forma?|knall(en?)?|[lm]?acht?(en)?|boot(en?)?|lach(en?)?|[hb]i?eb(en?)?|tut(en?)?|tr?öt(e[tn]?)?|[sp]ackt?(en?)?|[klnrd]?eckt?(en?)?|beut(en?)?|top|st?att(en?)?|dien(en?)?|[hl]ieb(en?)?|sät|satt(en?)?|droh(en?)?|[sr]äum(en?)?|zeugt?(en?)?|reu(en?)?|nies(en?)?|[gzf]eigt?(en?)?|gie(ß|ss)(en?)?|sichern?|zog(en?)?|schert?(en?)?|s[tp]r?ickt?(en?)?|seicht(e[srn]?)?|(be)?sorgt?(en?)?|ehelich(en?)?|link(en?)?|wein(en?)?|r?echt|orangen?|blick(en?)?|kling(en?)?|übrig(en?)?|klick(en?)?)`)
	registerPreventSuggestion(`[A-ZÖÄÜa-zöäüß] .+`)
	registerPreventSuggestion(`.+ [a-zöäüßA-ZÖÄÜ]`)
}


func registerPreventSuggestion(pat string) {
	re, err := regexp.Compile("^(?:" + pat + ")$")
	if err != nil {
		// try without full-string wrap for patterns that already match full
		re, err = regexp.Compile(pat)
		if err != nil {
			return
		}
	}
	preventSuggestionPatterns = append(preventSuggestionPatterns, re)
}

// AcceptSuggestion ports GermanSpellerRule.acceptSuggestion.
func (r *GermanSpellerRule) AcceptSuggestion(s string) bool {
	if s == "" {
		return false
	}
	for _, re := range preventSuggestionPatterns {
		if re.MatchString(s) {
			return false
		}
	}
	if strings.Contains(s, "--") {
		return false
	}
	if strings.HasSuffix(s, "roulett") || strings.HasSuffix(s, "-s") ||
		strings.HasSuffix(s, " de") || strings.HasSuffix(s, " en") ||
		strings.HasSuffix(s, " Artigen") || strings.HasSuffix(s, " Artige") ||
		strings.HasSuffix(s, " artigen") || strings.HasSuffix(s, " artiges") ||
		strings.HasSuffix(s, " artiger") || strings.HasSuffix(s, " artige") ||
		strings.HasSuffix(s, " artig") || strings.HasSuffix(s, " gen") ||
		strings.HasSuffix(s, " ehe") || strings.HasSuffix(s, " ende") ||
		strings.HasSuffix(s, " enden") || strings.HasSuffix(s, " enge") ||
		strings.HasSuffix(s, " förmig") || strings.HasSuffix(s, " förmige") ||
		strings.HasSuffix(s, " förmigen") || strings.HasSuffix(s, " förmiger") ||
		strings.HasSuffix(s, " förmiges") {
		return false
	}
	// Java rejects suggestions starting with these prefixes
	if strings.HasPrefix(s, "Doppel ") || strings.HasPrefix(s, "Kombi ") {
		return false
	}
	return true
}

