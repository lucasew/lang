package chunking

// germanRegex2 is one Java GermanChunker.REGEXES2 entry.
type germanRegex2 struct {
	pattern   string
	phrase    germanPhraseType
	overwrite bool
}

type germanPhraseType int

const (
	phraseNP germanPhraseType = iota // B-NP / I-NP (REGEXES1 only)
	phraseNPS
	phraseNPP
	phrasePP
)

func (p germanPhraseType) tagName() string {
	switch p {
	case phraseNPS:
		return "NPS"
	case phraseNPP:
		return "NPP"
	case phrasePP:
		return "PP"
	default:
		return ""
	}
}

// Auto-transcribed from Java GermanChunker.REGEXES2 (order + overwrite flags).
var germanRegexes2 = []germanRegex2{
	{`<pos=ADJ> <,> <chunk=B-NP> <chunk=I-NP>* <und|sowie> <NP>`, phraseNPP, false},
	{`<chunk=B-NP & !regex=jede[rs]?> <chunk=I-NP>* <und|sowie> <pos=ADV>? <NP>`, phraseNPP, false},
	{`<pos=ADJ> <und|sowie> <chunk=B-NP & !pos=PLU> <chunk=I-NP>*`, phraseNPS, true},
	{`<deren> <chunk=B-NP & !pos=PLU> <und|sowie> <chunk=B-NP>*`, phraseNPS, true},
	{`<pos=EIG> <und> <pos=EIG>`, phraseNPP, false},
	{`<pos=ART> <pos=ADJ> <und|sowie> (<pos=ADJ>|<pos=PA2>) <chunk=I-NP & !pos=PLU>+`, phraseNPS, true},
	{`<chunk=B-NP & !pos=PLU> <chunk=I-NP>* <und|sowie> <keine> <chunk=I-NP>+`, phraseNPS, true},
	{`<NP> <und|sowie> <pos=ART> <pos=PA1> <pos=SUB>`, phraseNPP, true},
	{`<eins|eines> <chunk=B-NP> <chunk=I-NP>+`, phraseNPS, false},
	{`<ich|du|er|sie|es|wir|ihr|sie> <und|oder|sowie> <NP>`, phraseNPP, false},
	{`<sowohl> <NP> <als> <auch> <NP>`, phraseNPP, false},
	{`<sowohl> <pos=EIG> <als> <auch> <pos=EIG>`, phraseNPP, false},
	{`<sowohl> <ich|du|er|sie|es|wir|ihr|sie> <als> <auch> <NP>`, phraseNPP, false},
	{`<pos=SUB> <und|oder|sowie> <chunk=B-NP & !ihre> <chunk=I-NP>*`, phraseNPP, false},
	{`<weder> <pos=SUB> <noch> <pos=SUB>`, phraseNPP, false},
	{`<zwei|drei|vier|fünf|sechs|sieben|acht|neun|zehn|elf|zwölf> <chunk=I-NP>`, phraseNPP, false},
	{`<chunk=B-NP> <pos=PRP> <NP> <chunk=B-NP & pos=SIN> <chunk=I-NP>*`, phraseNPS, false},
	{`<chunk=B-NP> <pos=PRP> <NP> <chunk=B-NP & pos=PLU> <chunk=I-NP>*`, phraseNPP, false},
	{`<chunk=B-NP> <pos=PRP> <NP> <pos=PA2> <chunk=B-NP & !pos=PLU> <chunk=I-NP>*`, phraseNPS, false},
	{`<chunk=B-NP> <pos=PRP> <NP> <pos=PA2> <chunk=B-NP & !pos=SIN> <chunk=I-NP>*`, phraseNPP, false},
	{`<Herr|Frau> <und> <Herr|Frau> <pos=EIG>*`, phraseNPP, false},
	{`<chunk=B-NP & !pos=ZAL & !pos=PLU & !chunk=NPP & !einige & !(regex=&prozent;)> <chunk=I-NP & !pos=PLU & !und>*`, phraseNPS, false},
	{`<chunk=B-NP & !pos=SIN & !chunk=NPS & !Ellen> <chunk=I-NP & !pos=SIN>*`, phraseNPP, false},
	{`<chunk=NPS> <pos=PRO> <pos=ADJ> <pos=ADJ> <NP>`, phraseNPS, false},
	{`<regex=eine[rs]?> <der> <am> <pos=ADJ> <pos=PA2> <NP>`, phraseNPS, false},
	{`<regex=eine[rs]?> <der> <beiden> <pos=ADJ>* <pos=SUB>`, phraseNPS, false},
	{`<regex=eine[rs]?> <seiner|ihrer> <pos=PA1> <pos=SUB>`, phraseNPS, false},
	{`<regex=[\d,.]+> <&prozent;>`, phraseNPS, false},
	{`<regex=[\d,.]+> <&prozent;>`, phraseNPP, false},
	{`<dass> <sie> <wie> <NP>`, phraseNPP, false},
	{`<pos=PLU> <die> <Regel>`, phraseNPP, false},
	{`<chunk=B-NP & pos=SIN> <chunk=I-NP & pos=SIN>* <,> <die> <pos=ADV>+ <chunk=NPS>+`, phraseNPS, false},
	{`<chunk=B-NP & pos=PLU> <chunk=I-NP & pos=PLU>* <,> <die> <pos=ADV>+ <chunk=NPS>+`, phraseNPP, false},
	{`<der|die|das> <pos=ADJ> <der> <pos=PA1> <pos=SUB>`, phraseNPS, false},
	{`<pos=SUB & pos=PLU> <der> <pos=PA1> <pos=SUB>`, phraseNPP, false},
	{`<der|die|das> <pos=ADJ> <der> <pos=PRO>? <pos=SUB>`, phraseNPS, false},
	{`<chunk=NPS & !einige> <chunk=NPP & (pos=GEN |pos=ZAL)>+`, phraseNPS, true},
	{`<chunk=NPP> <chunk=NPS & pos=GEN>+`, phraseNPP, true},
	{`<chunk=NPS>+ <und> <chunk=NP[SP] & (pos=GEN | pos=ADV)>+`, phraseNPS, true},
	{`<chunk=NPS>+ <der> <pos=ADV> <pos=PA2> <chunk=I-NP>`, phraseNPS, true},
	{`<chunk=NPS>+ <der> (<pos=ADJ>|<pos=ZAL>) <NP>`, phraseNPS, true},
	{`<chunk=NPS>+ <der> <NP>`, phraseNPS, true},
	{`<chunk=NPS>+ <der> <pos=ADJ> <pos=ADV> <pos=PA2> <NP>`, phraseNPS, true},
	{`<chunk=NPS>+ <pos=PRO:POS> <pos=ADJ> <NP>`, phraseNPS, true},
	{`<der|das> <pos=ADJ> <der> <pos=ZAL> <NP>`, phraseNPS, true},
	{`<eine> <menge> <NP>+`, phraseNPP, true},
	{`<er|sie|es> <und> <NP> <NP>`, phraseNPP, false},
	{`<laut> <regex=.*>{0,3} <Quellen>`, phrasePP, true},
	{`<pos=PRP> <pos=ART:> <pos=ADV>* <pos=ADJ> <NP>`, phrasePP, true},
	{`<pos=PRP> <chunk=NPP>+ <,> <NP>`, phrasePP, true},
	{`<pos=PRP> <chunk=NPP>+`, phrasePP, true},
	{`<pos=PRP> <der> <chunk=NPP>+`, phrasePP, false},
	{`<pos=PRP> <NP>`, phrasePP, false},
	{`<pos=PRP> <NP> <pos=ADJ> <und|oder|bzw.> <NP>`, phrasePP, false},
	{`<pos=PRP> (<NP>)+`, phrasePP, false},
	{`<pos=PRP> <chunk=B-NP> <pos=ADV> <NP>`, phrasePP, false},
	{`<pos=PRP> <pos=ADV> <pos=ZAL> <chunk=B-NP>`, phrasePP, false},
	{`<pos=PRP> <pos=PRO> <NP>`, phrasePP, false},
	{`<pos=PRP> <pos=ADJ> <und|oder|sowie> <NP>`, phrasePP, false},
	{`<pos=PRP> <pos=ADV> <regex=\d+> <NP>`, phrasePP, false},
	{`<pos=PRP> <pos=PA1> <NP>`, phrasePP, false},
	{`<pos=PRP> <pos=ADJ> <pos=PA1> <NP>`, phrasePP, false},
	{`<pos=PRP> <NP> <NP> <und|oder> <NP>`, phrasePP, false},
	{`<pos=PRP> <pos=ADV> <pos=ADJ> <NP>`, phrasePP, false},
	{`<pos=PRP> <pos=ADJ:PRD:GRU> <pos=ZAL> <NP>`, phrasePP, false},
	{`<die> <pos=ADJ> <Sekunden|Minuten|Stunden|Tage|Wochen|Monate|Jahre|Jahrzehnte|Jahrhunderte> (<NP>)?`, phrasePP, false},
	{`<die> <pos=ADJ> <pos=ZAL> <Sekunden|Minuten|Stunden|Tage|Wochen|Monate|Jahre|Jahrzehnte|Jahrhunderte> (<NP>)?`, phrasePP, false},
	{`<regex=(vor)?letzte[sn]?> <Woche|Monat|Jahr|Jahrzehnt|Jahrhundert>`, phrasePP, false},
	{`<für> <in> <pos=EIG> <pos=PA1> <pos=SUB> <und> <pos=SUB>`, phrasePP, true},
	{`<chunk=NPP> <zwischen> <pos=EIG> <und|sowie> <NP>`, phraseNPP, false},
	{`<,> <die|welche> <NP> <chunk=NPS & pos=GEN>+`, phraseNPP, false},
	{`<NP> <,> <NP> <,> <NP>`, phraseNPP, false},
	{`<NP> <,> <NP> <,> <wie> <auch> <chunk=NPS>+`, phraseNPP, false},
}

