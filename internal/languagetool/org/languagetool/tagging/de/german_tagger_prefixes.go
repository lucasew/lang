package de

// Code derived 1:1 from GermanTagger.java prefix arrays (no invent).

// prefixesSeparableVerbs ports GermanTagger.prefixesSeparableVerbs.
var prefixesSeparableVerbsJava = []string{
	"gegeneinander", "durcheinander", "nebeneinander", "übereinander", "aufeinander", "auseinander", "beieinander", "aneinander",
	"ineinander", "zueinander", "gegenüber", "beisammen", "hernieder", "rückwärts", "wiederauf", "wiederein",
	"wiederher", "zufrieden", "zwangsvor", "entgegen", "hinunter", "abhanden", "aufrecht", "aufwärts",
	"auswärts", "beiseite", "danieder", "drauflos", "einwärts", "herunter", "hindurch", "verrückt",
	"vorwärts", "zunichte", "zusammen", "zwangsum", "zwischen", "abseits", "abwärts", "entlang",
	"hinfort", "ähnlich", "daneben", "general", "herüber", "hierher", "hierhin", "hinüber",
	"schwarz", "trocken", "überein", "vorlieb", "vorüber", "wichtig", "zurecht", "zuwider",
	"hinweg", "allein", "besser", "daheim", "doppel", "feinst", "fertig", "herauf",
	"heraus", "herbei", "hinauf", "hinaus", "hinein", "kaputt", "kennen", "kürzer",
	"mittag", "nieder", "runter", "sicher", "sitzen", "voraus", "vorbei", "vorweg",
	"weiter", "wieder", "zugute", "zurück", "zwangs", "abend", "blank", "brust",
	"dahin", "davon", "drauf", "drein", "durch", "einig", "empor", "grund",
	"herum", "höher", "klein", "knapp", "krank", "krumm", "kugel", "näher",
	"neben", "offen", "preis", "rüber", "ruhig", "statt", "still", "übrig",
	"umher", "unter", "voran", "zweck", "acht", "drei", "fehl", "feil",
	"fort", "frei", "groß", "hand", "hart", "heim", "hier", "hoch",
	"klar", "lahm", "miss", "nach", "nahe", "quer", "rauf", "raus",
	"rein", "rück", "satt", "stoß", "teil", "über", "voll", "wach",
	"wahr", "warm", "wert", "wohl", "auf", "aus", "bei", "ehe",
	"ein", "eis", "end", "her", "hin", "los", "maß", "mit",
	"out", "ran", "rum", "tot", "vor", "weg", "weh", "ab",
	"an", "da", "um", "zu",
}

// prefixesVerbs ports GermanTagger.prefixesVerbs (separable + non-separable).
var prefixesVerbsJava = []string{
	"gegeneinander", "durcheinander", "nebeneinander", "übereinander", "aufeinander", "auseinander", "beieinander", "aneinander",
	"ineinander", "zueinander", "gegenüber", "beisammen", "hernieder", "rückwärts", "wiederauf", "wiederein",
	"wiederher", "zufrieden", "zwangsvor", "entgegen", "hinunter", "abhanden", "aufrecht", "aufwärts",
	"auswärts", "beiseite", "danieder", "drauflos", "einwärts", "herunter", "hindurch", "verrückt",
	"vorwärts", "zunichte", "zusammen", "zwangsum", "zwischen", "abseits", "abwärts", "entlang",
	"hinfort", "ähnlich", "daneben", "general", "herüber", "hierher", "hierhin", "hinüber",
	"schwarz", "trocken", "überein", "vorlieb", "vorüber", "wichtig", "zurecht", "zuwider",
	"hinweg", "hinter", "allein", "besser", "daheim", "doppel", "feinst", "fertig",
	"herauf", "heraus", "herbei", "hinauf", "hinaus", "hinein", "kaputt", "kennen",
	"kürzer", "mittag", "nieder", "runter", "sicher", "sitzen", "voraus", "vorbei",
	"vorweg", "weiter", "wieder", "zugute", "zurück", "zwangs", "abend", "blank",
	"brust", "dahin", "davon", "drauf", "drein", "durch", "einig", "empor",
	"grund", "herum", "höher", "klein", "knapp", "krank", "krumm", "kugel",
	"näher", "neben", "offen", "preis", "rüber", "ruhig", "statt", "still",
	"übrig", "umher", "unter", "voran", "zweck", "miss", "acht", "drei",
	"fehl", "feil", "fort", "frei", "groß", "hand", "hart", "heim",
	"hier", "hoch", "klar", "lahm", "nach", "nahe", "quer", "rauf",
	"raus", "rein", "rück", "satt", "stoß", "teil", "über", "voll",
	"wach", "wahr", "warm", "wert", "wohl", "emp", "ent", "ver",
	"zer", "auf", "aus", "bei", "ehe", "ein", "eis", "end",
	"her", "hin", "los", "maß", "mit", "out", "ran", "rum",
	"tot", "vor", "weg", "weh", "be", "er", "un", "ab",
	"an", "da", "um", "zu",
}

// prefixesSeparableVerbsLongest is longest-first for startsWithAny / strip.
var prefixesSeparableVerbsLongestList = longestFirstCopy(prefixesSeparableVerbsJava)

// prefixesVerbsLongestList is longest-first for stripLongestPrefix.
var prefixesVerbsLongestList = longestFirstCopy(prefixesVerbsJava)

func longestFirstCopy(in []string) []string {
	out := append([]string(nil), in...)
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if len(out[j]) > len(out[i]) {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

// domainTLDs ports the domain-ignore TLD list in GermanTagger.tag (com|net|…).
var domainTLDs = map[string]struct{}{
	"com": {},
	"net": {},
	"org": {},
	"de":  {},
	"at":  {},
	"ch":  {},
	"fr":  {},
	"uk":  {},
	"gov": {},
}

// separablePrefixSet for exact first-part membership (Java prfxs.contains).
var separablePrefixSet map[string]struct{}

func init() {
	separablePrefixSet = make(map[string]struct{}, len(prefixesSeparableVerbsJava))
	for _, p := range prefixesSeparableVerbsJava {
		separablePrefixSet[p] = struct{}{}
	}
}

func isExactSeparablePrefix(p string) bool {
	_, ok := separablePrefixSet[p]
	return ok
}
