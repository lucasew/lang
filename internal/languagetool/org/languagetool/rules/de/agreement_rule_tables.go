package de

// Tables ported from AgreementRule.java (exact sets).

// agreementModifiers ports MODIFIERS — degree adverbs between DET and ADJ.
var agreementModifiers = map[string]struct{}{
	"zu": {}, "überraschend": {}, "ungeahnt": {}, "absolut": {}, "ausgesprochen": {},
	"außergewöhnlich": {}, "außerordentlich": {}, "äußerst": {}, "besonders": {}, "dringend": {},
	"echt": {}, "einigermaßen": {}, "enorm": {}, "extrem": {}, "fast": {}, "ganz": {},
	"entschieden": {}, "geradezu": {}, "zeitweise": {}, "halbwegs": {}, "höchst": {},
	"komplett": {}, "laufend": {}, "recht": {}, "relativ": {}, "sehr": {}, "total": {},
	"überaus": {}, "ungewöhnlich": {}, "unglaublich": {}, "völlig": {}, "weit": {},
	"wirklich": {}, "gerade": {}, "vereint": {}, "überwiegend": {}, "gewollt": {},
	"angestrengt": {}, "ziemlich": {},
}

// vieleWenige ports VIELE_WENIGE_LOWERCASE — skipSol false for these det/pro forms.
var vieleWenige = map[string]struct{}{
	"manche": {}, "jegliche": {}, "jeglicher": {},
	"andere": {}, "anderer": {}, "anderen": {},
	"sämtliche": {}, "sämtlicher": {},
	"etliche": {}, "etlicher": {},
	"viele": {}, "vieler": {},
	"wenige": {}, "weniger": {},
	"einige": {}, "einiger": {},
	"mehrerer": {}, "mehrere": {},
}

// pronounsToBeIgnored ports PRONOUNS_TO_BE_IGNORED (lowercase keys).
var pronounsToBeIgnored = map[string]struct{}{
	"nichts": {}, "alles": {}, "dies": {}, "ebendies": {},
	"ich": {}, "dir": {}, "dich": {}, "du": {}, "d": {},
	"er": {}, "sie": {}, "es": {}, "wir": {},
	"mich": {}, "mir": {}, "uns": {}, "ihnen": {}, "euch": {},
	"ihm": {}, "ihr": {}, "ihn": {},
	"dessen": {}, "deren": {}, "denen": {}, "sich": {},
	"aller": {}, "allen": {}, "man": {},
	"beide": {}, "beiden": {}, "beider": {},
	"wessen": {}, "a": {}, "alle": {},
	"etwas": {}, "irgendetwas": {}, "irgendwas": {}, "irgendwer": {},
	"was": {}, "wer": {}, "wem": {},
	"jenen": {}, "diejenigen": {},
	"irgendjemand": {}, "irgendjemandes": {},
	"jemand": {}, "jemandes": {},
	"niemand": {}, "niemandes": {},
}

// nounsToBeIgnored ports NOUNS_TO_BE_IGNORED (case-sensitive as in Java).
var nounsToBeIgnored = map[string]struct{}{
	"A": {}, "Prozent": {}, "Wollen": {}, "Gramm": {}, "Kilogramm": {},
	"Flippers": {}, "Standart": {}, "Stellungsname": {}, "Kündigungsscheiben": {},
	"Piepen": {}, "Badlands": {}, "Visual": {}, "Special": {}, "Multiple": {},
	"Chief": {}, "Carina": {}, "Wüstenrot": {}, "Rückgrad": {}, "Rückgrads": {},
	"Anteilname": {}, "Aalen": {}, "Meter": {}, "Boots": {}, "Taxameter": {},
	"Bild": {}, "Emirates": {}, "Uhr": {}, "cm": {}, "km": {}, "Nr": {},
	"KSC": {}, "ANC": {}, "DJK": {}, "RP": {},
}

// relPronounLemmas ports REL_PRONOUN_LEMMAS.
var relPronounLemmas = []string{"der", "welch"}
