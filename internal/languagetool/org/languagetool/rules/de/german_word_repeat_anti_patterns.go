package de

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"

// GermanWordRepeatAntiPatterns ports GermanWordRepeatRule.ANTI_PATTERNS (59/59).
// Token-only and POS patterns; POS ones need tagged input (fail-closed otherwise).
// matchInflectedForms uses lemma readings (Java); no surface-list invent.
var GermanWordRepeatAntiPatterns = [][]*patterns.PatternToken{
	{patterns.CsToken("please"), patterns.CsToken("please"), patterns.CsToken("please")},
	{patterns.CsToken("Late"), patterns.CsToken("Late"), patterns.CsToken("Show")},
	{
		patterns.CsToken("Wenn"), patterns.CsToken("hinter"), patterns.CsToken("Robben"),
		patterns.CsToken("Robben"), patterns.CsToken("robben"), patterns.CsToken(","),
		patterns.CsToken("robben"), patterns.CsToken("Robben"), patterns.CsToken("Robben"),
		patterns.CsToken("hinterher"),
	},
	{patterns.TokenRegex("tägliche(n|m|s)?"), patterns.CsToken("klein"), patterns.CsToken("klein")},
	{patterns.CsToken("Bora"), patterns.CsToken("Bora")},
	{patterns.CsToken("Tuk"), patterns.CsToken("Tuk")},
	{patterns.CsToken("Miu"), patterns.CsToken("Miu")},
	{patterns.Token("Moin"), patterns.Token("Moin")},
	{patterns.Token("Na"), patterns.Token("na")},
	{patterns.Token("la"), patterns.Token("la")},
	{patterns.CsToken("Fragen"), patterns.CsToken("fragen")},
	{patterns.Token("ha"), patterns.Token("ha")},
	{patterns.Token("teils"), patterns.Token("teils")},
	{patterns.Token("Marsch"), patterns.Token("Marsch")},
	{patterns.Token("hip"), patterns.Token("hip"), patterns.Token("hurra")},
	{patterns.Token("möp"), patterns.Token("möp")},
	{patterns.Token("gout"), patterns.Token("gout")},
	{patterns.Token("piep"), patterns.Token("piep")},
	{patterns.Token("bla"), patterns.Token("bla")},
	{patterns.Token("blah"), patterns.Token("blah")},
	{patterns.Token("oh"), patterns.Token("oh")},
	{patterns.Token("klopf"), patterns.Token("klopf")},
	{patterns.Token("ne"), patterns.Token("ne")},
	{patterns.Token("Fakten"), patterns.Token("Fakten"), patterns.Token("Fakten")},
	{patterns.Token("Top"), patterns.Token("Top"), patterns.Token("Top")},
	{patterns.Token("Toi"), patterns.Token("Toi"), patterns.Token("Toi")},
	{patterns.Token("und"), patterns.Token("und"), patterns.Token("und")},
	{patterns.Token("man"), patterns.Token("man"), patterns.Token("man")},
	{
		patterns.TokenRegex("wenn|falls"), patterns.Token("das"), patterns.Token("das"),
		patterns.Token("nächste"), patterns.Token("mal"),
	},
	{patterns.Token("Arbeit"), patterns.Token("Arbeit"), patterns.Token("Arbeit")},
	{patterns.TokenRegex(`\*|:|\/`), patterns.Token("in"), patterns.Token("in")},
	{patterns.Token("Üben"), patterns.Token("Üben"), patterns.Token("Üben")},
	{patterns.Token("cha"), patterns.Token("cha")},
	{patterns.Token("zack"), patterns.Token("zack")},
	{patterns.Token("sapiens"), patterns.Token("sapiens")},
	{patterns.Token("peng"), patterns.Token("peng")},
	{patterns.Token("bye"), patterns.Token("bye")},
	{patterns.Token("nicht"), patterns.Token("nicht"), patterns.Token("kommunizieren")},
	{patterns.Token("Dee"), patterns.TokenRegex("Dees?")},
	{patterns.Token("Phi"), patterns.Token("Phi")},
	{
		patterns.TokenRegex(`,|wei(ß|ss)|nicht`), patterns.Token("wer"), patterns.Token("wer"),
		patterns.TokenRegex("war|ist|sein"),
	},
	// POS-dependent (need tags)
	{
		patterns.TokenRegex(`ist|war|wäre?|für|dass`), patterns.Token("das"), patterns.Token("das"),
		patterns.PosRegex(`.*SUB:.*NEU.*`),
	},
	{
		patterns.TokenRegex(`ist|war|wäre?|für|dass`), patterns.Token("das"), patterns.Token("das"),
		patterns.PosRegex(`ADJ:.*`), patterns.PosRegex(`.*SUB:.*NEU.*`),
	},
	{
		patterns.TokenRegex(`ist|war|wäre?|für|dass`), patterns.Token("das"), patterns.Token("das"),
		patterns.PosRegex(`ADJ:.*NEU.*`), patterns.PosRegex(`UNKNOWN`),
	},
	{
		patterns.TokenRegex(`als|wenn`), patterns.PosRegex(`(PRO|EIG):.*`),
		patterns.Token("das"), patterns.Token("das"),
		patterns.PosRegex(`ADJ:.*NEU.*`), patterns.PosRegex(`.*SUB:.*NEU.*`),
	},
	{
		patterns.TokenRegex(`werden|würden|sollt?en|müsst?en|könnt?en`),
		patterns.Token("sie"), patterns.Token("sie"),
		patterns.PosRegex(`VER:1:PLU:.*`),
	},
	{
		// "Falls das das Problem ist, …"
		patterns.TokenRegex(`wenn|falls|ob`), patterns.Token("das"), patterns.Token("das"),
		patterns.PosRegex(`SUB:NOM:SIN:NEU.*`),
		patterns.NewPatternTokenBuilder().TokenRegex("sein|haben").MatchInflectedForms().Build(),
	},
	{
		// "Falls das das neue Problem ist, …"
		patterns.TokenRegex(`wenn|falls|ob`), patterns.Token("das"), patterns.Token("das"),
		patterns.PosRegex(`(ADJ|PA[12]).*NEU.*`), patterns.PosRegex(`SUB:NOM:SIN:NEU.*`),
		patterns.NewPatternTokenBuilder().TokenRegex("sein|haben").MatchInflectedForms().Build(),
	},
	{
		// "wie Honda und Samsung, die die Bezahlung ..."
		patterns.CsToken(","),
		patterns.NewPatternTokenBuilder().CsToken("der").MatchInflectedForms().Build(),
		patterns.NewPatternTokenBuilder().CsToken("der").MatchInflectedForms().Build(),
	},
	{
		// "Alle die die"
		patterns.TokenRegex(`alle|nur|obwohl|lediglich|für|zwar|aber`),
		patterns.NewPatternTokenBuilder().CsToken("die").Build(),
		patterns.NewPatternTokenBuilder().CsToken("die").Build(),
	},
	{
		// "Haben die die Elemente ..."
		patterns.PosRegex(`PKT|SENT_START|KON:NEB`),
		patterns.TokenRegex(`haben|hatten`),
		patterns.NewPatternTokenBuilder().CsToken("die").Build(),
		patterns.NewPatternTokenBuilder().CsToken("die").Build(),
		patterns.PosRegex(`.*SUB.*PLU.*|UNKNOWN`),
	},
	{
		// "und ob die die Währungen ..."
		patterns.PosRegex(`PKT|SENT_START|KON:NEB`),
		patterns.TokenRegex(`ob|falls`),
		patterns.NewPatternTokenBuilder().CsToken("die").Build(),
		patterns.NewPatternTokenBuilder().CsToken("die").Build(),
		patterns.PosRegex(`.*SUB.*PLU.*|UNKNOWN`),
	},
	{
		// "Das Haus, in das das Kind läuft."
		patterns.CsToken(","), patterns.PosRegex(`PRP:.+`),
		patterns.NewPatternTokenBuilder().CsToken("der").MatchInflectedForms().Build(),
		patterns.NewPatternTokenBuilder().CsToken("der").MatchInflectedForms().Build(),
	},
	{patterns.CsToken("Leben"), patterns.CsToken("leben")},
	{patterns.CsToken("Stellen"), patterns.CsToken("stellen")},
	{patterns.Token("die"), patterns.CsToken("ferne"), patterns.CsToken("Ferne")},
	{patterns.CsToken("Essen"), patterns.CsToken("essen")},
	{patterns.TokenRegex(`^[_]+$`), patterns.TokenRegex(`^[_]+$`)},
	{
		patterns.PosRegex(`VER:.*[123]:.+|PKT|ADV:INR`),
		patterns.CsToken("ihr"), patterns.CsToken("ihr"),
		patterns.PosRegex(`SUB.+`),
	},
}
