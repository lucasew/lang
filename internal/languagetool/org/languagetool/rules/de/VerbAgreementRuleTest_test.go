package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/VerbAgreementRuleTest.java
// Java: JLanguageTool(de-DE) + VerbAgreementRule(messages, German); rule.match(lt.analyzeText(s)).
// Go: NewJLanguageTool("de-DE") + WireGermanTagWord (SentenceTagger = GermanTagger.Tag) +
//     MatchList(lt.Analyze(s)) — real morph path, no invent POS inject.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// atrWithPOSMulti builds a token with multiple POS readings (Java Morphy ambiguity).
// Shared by other DE rule tests in this package.
func atrWithPOSMulti(token, lemma string, tags ...string) *languagetool.AnalyzedTokenReadings {
	if len(tags) == 0 {
		return atrWithPOS(token, "", lemma)
	}
	atr := atrWithPOS(token, tags[0], lemma)
	for _, tag := range tags[1:] {
		tt, ll := tag, lemma
		atr.AddReading(languagetool.NewAnalyzedToken(token, &tt, &ll), "")
	}
	return atr
}

func vaSetup(t *testing.T) (*languagetool.JLanguageTool, *VerbAgreementRule) {
	t.Helper()
	lt := languagetool.NewJLanguageTool("de-DE")
	if !WireGermanTagWord(lt) {
		t.Skip("GermanTagger resources unavailable")
	}
	_ = WireGermanDisambiguator(lt)
	rule := NewVerbAgreementRule(nil)
	if synth := openDiscoveredGermanSynthesizer(); synth != nil {
		rule = rule.WithSynth(synth)
	}
	return lt, rule
}

// vaMatch ports rule.match(lt.analyzeText(s)).
func vaMatch(lt *languagetool.JLanguageTool, rule *VerbAgreementRule, s string) []*rules.RuleMatch {
	return rule.MatchList(lt.Analyze(s))
}

func vaAssertGood(t *testing.T, lt *languagetool.JLanguageTool, rule *VerbAgreementRule, s string) {
	t.Helper()
	ms := vaMatch(lt, rule, s)
	require.Equal(t, 0, len(ms), "Got > 0 matches for %q: %v", s, ms)
}

func vaAssertBad(t *testing.T, lt *languagetool.JLanguageTool, rule *VerbAgreementRule, s string, n int) {
	t.Helper()
	ms := vaMatch(lt, rule, s)
	require.Equal(t, n, len(ms), "Did not find %d match(es) in sentence %q (got %d)", n, s, len(ms))
}

func vaAssertBadMsg(t *testing.T, lt *languagetool.JLanguageTool, rule *VerbAgreementRule, s, sub string) {
	t.Helper()
	ms := vaMatch(lt, rule, s)
	require.Equal(t, 1, len(ms), "expected 1 match for %q", s)
	require.Contains(t, ms[0].GetMessage(), sub, "Got error %q, expected substring %q", ms[0].GetMessage(), sub)
}

// vaAssertBadSugs ports assertBad(input, expectedMatches, expectedSuggestions...).
// When two matches and first has empty suggestions, use second (Java TODO comment).
func vaAssertBadSugs(t *testing.T, lt *languagetool.JLanguageTool, rule *VerbAgreementRule, s string, n int, expected []string) {
	t.Helper()
	ms := vaMatch(lt, rule, s)
	require.Equal(t, n, len(ms), "Did not find %d match(es) in sentence %q (got %d)", n, s, len(ms))
	if len(expected) == 0 {
		return
	}
	match := ms[0]
	if len(ms) > 1 && len(match.GetSuggestedReplacements()) == 0 {
		match = ms[1]
	}
	got := match.GetSuggestedReplacements()
	// Java HashSet pre-order makes equal-Levenshtein tie-break non-portable across runtimes.
	// Require exact first (closest) and multiset equality for the full list.
	require.NotEmpty(t, got, "suggestions for %q", s)
	require.Equal(t, expected[0], got[0], "first suggestion for %q", s)
	require.ElementsMatch(t, expected, got, "suggestion multiset for %q\n got=%v\nwant=%v", s, got, expected)
}

// TestVerbAgreementRule_Meta ports rule id/description/examples/anti-patterns.
func TestVerbAgreementRule_Meta(t *testing.T) {
	r := NewVerbAgreementRule(nil)
	require.Equal(t, "DE_VERBAGREEMENT", r.GetID())
	require.Contains(t, r.GetDescription(), "Kongruenz")
	require.Equal(t, 0, r.EstimateContextForSureMatch())
	require.Equal(t, 0, r.MinToCheckParagraph())
	require.NotEmpty(t, r.GetIncorrectExamples())
	require.NotEmpty(t, r.GetCorrectExamples())
	require.NotNil(t, r.GetCategory())
	// ANTI_PATTERNS: 96 patterns (Java asList count)
	require.Equal(t, 96, len(VerbAgreementAntiPatterns))
	require.Equal(t, 96, len(r.GetAntiPatterns()))
}

// TestVerbAgreementRule_SuggestionSorting ports testSuggestionSorting.
func TestVerbAgreementRule_SuggestionSorting(t *testing.T) {
	lt, rule := vaSetup(t)
	require.NotNil(t, rule.Synth, "German synthesizer required for suggestion sorting twin")
	ms := vaMatch(lt, rule, "Wir nenne ihn mal „wild“.")
	require.Equal(t, 1, len(ms))
	require.Equal(t,
		[]string{"Wir nennen", "Wir nennten", "Er nenne", "Sie nenne", "Wir nannten", "Ich nenne", "Es nenne"},
		ms[0].GetSuggestedReplacements(),
	)
}

// TestVerbAgreementRule_Positions ports testPositions.
func TestVerbAgreementRule_Positions(t *testing.T) {
	lt, rule := vaSetup(t)

	ms := vaMatch(lt, rule, "Du erreichst ich unter 12345")
	require.Equal(t, 1, len(ms))
	require.Equal(t, 3, ms[0].GetFromPos())
	require.Equal(t, 16, ms[0].GetToPos())

	ms = vaMatch(lt, rule, "Hallo Karl. Du erreichst ich unter 12345")
	require.Equal(t, 1, len(ms))
	require.Equal(t, 12+3, ms[0].GetFromPos())
	require.Equal(t, 12+16, ms[0].GetToPos())

	ms = vaMatch(lt, rule, "Ihr könnt das Training abbrechen, weil es nichts bringen wird. Er geht los und sagt dabei: Werde ich machen.")
	require.Equal(t, 1, len(ms))
	require.Equal(t, 97, ms[0].GetFromPos())
	require.Equal(t, 107, ms[0].GetToPos())

	// Was a FP, now fixed
	ms = vaMatch(lt, rule, "Mir ist bewusst, dass viele Menschen wie du empfinden.")
	require.Equal(t, 0, len(ms))
}

// TestVerbAgreementRule_WrongVerb ports testWrongVerb (assertGood + assertBad matrix).
func TestVerbAgreementRule_WrongVerb(t *testing.T) {
	lt, rule := vaSetup(t)

	// correct sentences (Java assertGood):
	goods := []string{
		"*runterguck* das ist aber tief",
		"Weder Peter noch ich wollen das.",
		"Du bist in dem Moment angekommen, als ich gegangen bin.",
		"Kümmere du dich mal nicht darum!",
		"Ich weiß, was ich tun werde, falls etwas geschehen sollte.",
		"...die dreißig Jahre jünger als ich ist.",
		"Ein Mann wie ich braucht einen Hut.",
		"Egal, was er sagen wird, ich habe meine Entscheidung getroffen.",
		"Du Beharrst darauf, dein Wörterbuch hätte recht, hast aber von den Feinheiten des Japanischen keine Ahnung!",
		"Bin gleich wieder da.",
		"Wobei ich äußerst vorsichtig bin.",
		"Es ist klar, dass ich äußerst vorsichtig mit den Informationen umgehe",
		"Es ist klar, dass ich äußerst vorsichtig bin.",
		"Wobei er äußerst selten darüber spricht.",
		"Wobei er äußerst selten über seine erste Frau spricht.",
		"Das Wort „schreibst“ ist schön.",
		"Die Jagd nach bin Laden.",
		"Die Unterlagen solltet ihr gründlich durcharbeiten.",
		"Er reagierte äußerst negativ.",
		"Max und ich sollten das machen.",
		"Osama bin Laden stammt aus Saudi-Arabien.",
		"Solltet ihr das machen?",
		"Dann beende du den Auftrag und bring sie ihrem Vater.",
		"- Wirst du ausflippen?",
		"Ein Geschenk, das er einst von Aphrodite erhalten hatte.",
		"Wenn ich sterben sollte, wer würde sich dann um die Katze kümmern?",
		"Wenn er sterben sollte, wer würde sich dann um die Katze kümmern?",
		"Wenn sie sterben sollte, wer würde sich dann um die Katze kümmern?",
		"Wenn es sterben sollte, wer würde sich dann um die Katze kümmern?",
		"Wenn ihr sterben solltet, wer würde sich dann um die Katze kümmern?",
		"Wenn wir sterben sollten, wer würde sich dann um die Katze kümmern?",
		"Dafür erhielten er sowie der Hofgoldschmied Theodor Heiden einen Preis.",
		"Probst wurde deshalb in den Medien gefeiert.",
		"/usr/bin/firefox",
		"Das sind Leute, die viel mehr als ich wissen.",
		"Das ist mir nicht klar, kannst ja mal beim Kunden nachfragen.",
		"So tes\u00ADtest Du das mit dem soft hyphen.",
		"Viele Brunnen in Italiens Hauptstadt sind bereits abgeschaltet.",
		"„Werde ich tun!“",
		"Könntest dir mal eine Scheibe davon abschneiden!",
		"Müsstest dir das mal genauer anschauen.",
		"Kannst ein neues Release machen.",
		"Sie fragte: „Muss ich aussagen?“",
		"„Können wir bitte das Thema wechseln, denn ich möchte ungern darüber reden?“",
		"Er sagt: „Willst du behaupten, dass mein Sohn euch liebt?“",
		"Kannst mich gerne anrufen.",
		"Kannst ihn gerne anrufen.",
		"Kannst sie gerne anrufen.",
		"Aber wie ich sehe, benötigt ihr Nachschub.",
		"Wie ich sehe, benötigt ihr Nachschub.",
		"Einer wie du kennt doch bestimmt viele Studenten.",
		"Für Sie mache ich eine Ausnahme.",
		"Ohne sie hätte ich das nicht geschafft.",
		"Ohne Sie hätte ich das nicht geschafft.",
		"Ich hoffe du auch.",
		"Ich hoffe ihr auch.",
		"Wird hoffen du auch.",
		"Hab einen schönen Tag!",
		"Tom traue ich mehr als Maria.",
		"Tom kenne ich nicht besonders gut, dafür aber seine Frau.",
		"Tom habe ich heute noch nicht gesehen.",
		"Tom bezahle ich gut.",
		"Tom werde ich nicht noch mal um Hilfe bitten.",
		"Tom konnte ich überzeugen, nicht aber Maria.",
		"Mach du mal!",
		"Das bekomme ich nicht hin.",
		"Dies betreffe insbesondere Nietzsches Aussagen zu Kant und der Evolutionslehre.",
		"❌Du fühlst Dich unsicher?",
		"Bringst nicht einmal so etwas Einfaches zustande!",
		"Bekommst sogar eine Sicherheitszulage",
		"Dallun sagte nur, dass er gleich kommen wird und legte wieder auf.",
		"Tinne, Elvis und auch ich werden gerne wiederkommen!",
		"Du bist Lehrer und weißt diese Dinge nicht?",
		"Die Frage lautet: Bist du bereit zu helfen?",
		"Ich will nicht so wie er enden.",
		"Das heißt, wir geben einander oft nach als gute Freunde, ob wir gleich nicht einer Meinung sind.",
		"Wir seh'n uns in Berlin.",
		"Bist du bereit, darüber zu sprechen?",
		"Bist du schnell eingeschlafen?",
		"Im Gegenzug bin ich bereit, beim Türkischlernen zu helfen.",
		"Das habe ich lange gesucht.",
		"Dann solltest du schnell eine Nummer der sexy Omas wählen.",
		"Vielleicht würdest du bereit sein, ehrenamtlich zu helfen.",
		"Werde nicht alt, egal wie lange du lebst.",
		"Du bist hingefallen und hast dir das Bein gebrochen.",
		"Mögest du lange leben!",
		"Planst du lange hier zu bleiben?",
		"Du bist zwischen 11 und 12 Jahren alt und spielst gern Fußball bzw. möchtest damit anfangen?",
		"Ein großer Hadithwissenschaftler, Scheich Şemseddin Mehmed bin Muhammed-ül Cezri, kam in der Zeit von Mirza Uluğ Bey nach Semerkant.",
		"Die Prüfbescheinigung bekommst du gleich nach der bestanden Prüfung vom Prüfer.",
		"Du bist sehr schön und brauchst überhaupt gar keine Schminke zu verwenden.",
		"Ist das so schnell, wie du gehen kannst?",
		"Egal wie lange du versuchst, die Leute davon zu überzeugen",
		"Du bist verheiratet und hast zwei Kinder.",
		"Du bist aus Berlin und wohnst in Bonn.",
		"Sie befestigen die Regalbretter vermittelst dreier Schrauben.",
		"Meine Familie & ich haben uns ein neues Auto gekauft.",
		"Der Bescheid lasse im übrigen die Abwägungen vermissen, wie die Betriebsprüfung zu den Sachverhaltsbeurteilungen gelange, die den von ihr bekämpften Bescheiden zugrundegelegt worden seien.",
		"Die Bildung des Samens erfolgte laut Alkmaion im Gehirn, von wo aus er durch die Adern in den Hoden gelange.",
		"Michael Redmond (geb. 1963, USA).",
		"Würd mich sehr freuen drüber.",
		"Es würd' ein jeder Doktor sein, wenn's Wissen einging wie der Wein.",
		"Bald merkte er, dass er dank seines Talents nichts mehr in der österreichischen Jazzszene lernen konnte.",
		"»Alles, was wir dank dieses Projektes sehen werden, wird für uns neu sein«, so der renommierte Bienenforscher.",
		"Und da wir äußerst Laissez-faire sind, kann man das auch machen.",
		"Duzen, jemanden mit Du anreden, eine Sitte, die bei allen alten Völkern üblich war.",
		"Schreibtischtäter wie Du sind doch eher selten.",
		"Nee, geh du!",
	}
	for _, s := range goods {
		vaAssertGood(t, lt, rule, s)
	}

	// incorrect sentences (Java assertBad):
	bads := []string{
		"Als Borcarbid weißt es eine hohe Härte auf.",
		"Das greift auf Vorläuferinstitutionen bist auf die Zeit von 1234 zurück.",
		"Die Eisenbahn dienst überwiegend dem Güterverkehr.",
		"Die Unterlagen solltest ihr gründlich durcharbeiten.",
		"Peter bin nett.",
		"Weiter befindest sich im Osten die Gemeinde Dorf.",
		"Ich geht jetzt nach Hause, weil ich schon zu spät bin.",
		"„Du muss gehen.“",
		"Du weiß es doch.",
		"Sie sagte zu mir: „Du muss gehen.“",
		"„Ich müsst alles machen.“",
		"„Ich könnt mich sowieso nicht verstehen.“",
		"Er sagte düster: Ich brauchen mich nicht böse angucken.",
		"David sagte düster: Ich brauchen mich nicht böse angucken.",
		"Ich setzet mich auf den weichen Teppich und kreuzte die Unterschenkel wie ein Japaner.",
		"Ich brauchen einen Karren mit zwei Ochsen.",
		"Ich haben meinen Ohrring fallen lassen.",
		"Ich stehen Ihnen gerne für Rückfragen zur Verfügung.",
	}
	for _, s := range bads {
		vaAssertBad(t, lt, rule, s, 1)
	}
}

// TestVerbAgreementRule_WrongVerbSubject ports testWrongVerbSubject.
func TestVerbAgreementRule_WrongVerbSubject(t *testing.T) {
	lt, rule := vaSetup(t)

	goods := []string{
		"Auch morgen lebe ich.",
		"Auch morgen leben wir noch.",
		"Auch morgen lebst du.",
		"Auch morgen lebt er.",
		"Auch wenn du leben möchtest.",
		"auf der er sieben Jahre blieb.",
		"Das absolute Ich ist nicht mit dem individuellen Geist zu verwechseln.",
		"Das Ich ist keine Einbildung",
		"Das lyrische Ich ist verzweifelt.",
		"Den Park, von dem er äußerst genaue Karten zeichnete.",
		"Der auffälligste Ring ist der erster Ring, obwohl er verglichen mit den anderen Ringen sehr schwach erscheint.",
		"Der Fehler, falls er bestehen sollte, ist schwerwiegend.",
		"Der Vorfall, bei dem er einen Teil seines Vermögens verloren hat, ist lange vorbei.",
		"Diese Lösung wurde in der 64'er beschrieben, kam jedoch nie.",
		"Die Theorie, mit der ich arbeiten konnte.",
		"Du bist nett.",
		"Du kannst heute leider nicht kommen.",
		"Du lebst.",
		"Du wünschst dir so viel.",
		"Er geht zu ihr.",
		"Er ist nett.",
		"Er kann heute leider nicht kommen.",
		"Er lebt.",
		"Er wisse nicht, ob er lachen oder weinen solle.",
		"Er und du leben.",
		"Er und ich leben.",
		"Falls er bestehen sollte, gehen sie weg.",
		"Heere, des Gottes der Schlachtreihen Israels, den du verhöhnt hast.",
		"Ich bin",
		"Ich bin Frankreich!",
		"Ich bin froh, dass ich arbeiten kann.",
		"Ich bin nett.",
		"‚ich bin tot‘",
		"Ich kann heute leider nicht kommen.",
		"Ich lebe.",
		"Lebst du?",
		"Morgen kommen du und ich.",
		"Morgen kommen er, den ich sehr mag, und ich.",
		"Morgen kommen er und ich.",
		"Morgen kommen ich und sie.",
		"Morgen kommen wir und sie.",
		"nachdem er erfahren hatte",
		"Nett bin ich.",
		"Nett bist du.",
		"Nett ist er.",
		"Nett sind wir.",
		"Niemand ahnte, dass er gewinnen könne.",
		"Sie lebt und wir leben.",
		"Sie und er leben.",
		"Sind ich und Peter nicht nette Kinder?",
		"Sodass ich sagen möchte, dass unsere schönen Erinnerungen gut sind.",
		"Wann ich meinen letzten Film drehen werde, ist unbekannt.",
		"Was ich tun muss.",
		"Welche Aufgaben er dabei tatsächlich übernehmen könnte",
		"wie er beschaffen war",
		"Wir gelangen zu dir.",
		"Wir können heute leider nicht kommen.",
		"Wir leben noch.",
		"Wir sind nett.",
		"Wobei wir benutzt haben, dass der Satz gilt.",
		"Wünschst du dir mehr Zeit?",
		"Wyrjtjbst du?",
		"Wenn ich du wäre, würde ich das nicht machen.",
		"Er sagte: „Darf ich bitten, mir zu folgen?“",
		"Ja sind ab morgen dabei.",
		"Oh bin überfragt.",
		"Angenommen, du wärst ich.",
		"Ich denke, dass das Haus, in das er gehen will, heute Morgen gestrichen worden ist.",
		"Ich hab mein Leben, leb du deines!",
		"Da freut er sich, wenn er schlafen geht und was findet.",
		"John nimmt weiter an einem Abendkurs über Journalismus teil.",
		"Viele nahmen an der Aktion teil und am Ende des rAAd-Events war die Tafel zwar bunt, aber leider überwogen die roten Kärtchen sehr deutlich.",
		"Musst also nichts machen.",
		"Eine Situation, wo der Stadtrat gleich mal zum Du übergeht.",
		"Machen wir, sobald wir frische und neue Akkus haben.",
		"Darfst nicht so reden, Franz!",
		"Finde du den Jungen.",
		"Finde Du den Jungen.",
		"Kümmerst dich ja gar nicht um sie.",
		"Könntest was erfinden, wie dein Papa.",
		"Siehst aus wie ein Wachhund.",
		"Solltest es mal in seinem Büro versuchen.",
		"Stehst einfach nicht zu mir.",
		"Stellst für deinen Dad etwas zu Essen bereit.",
		"Springst weit, oder?",
		"Wirst groß, was?",
	}
	for _, s := range goods {
		vaAssertGood(t, lt, rule, s)
	}

	// incorrect sentences:
	vaAssertBad(t, lt, rule, "Auch morgen leben du.", 1)
	vaAssertBad(t, lt, rule, "Du weiß noch, dass du das gestern gesagt hast.", 1)
	vaAssertBad(t, lt, rule, "Auch morgen leben du", 1)
	vaAssertBad(t, lt, rule, "Auch morgen leben er.", 1)
	vaAssertBad(t, lt, rule, "Auch morgen leben ich.", 1)
	vaAssertBad(t, lt, rule, "Auch morgen lebte wir noch.", 1)
	vaAssertBad(t, lt, rule, "Du bin nett.", 2)
	vaAssertBad(t, lt, rule, "Du können heute leider nicht kommen.", 1)
	vaAssertBadSugs(t, lt, rule, "Du können heute leider nicht kommen.", 1,
		[]string{"Du könnest", "Du kannst", "Du könntest", "Wir können", "Sie können", "Du konntest"})
	vaAssertBad(t, lt, rule, "Du leben.", 1)
	vaAssertBad(t, lt, rule, "Du wünscht dir so viel.", 1)
	vaAssertBad(t, lt, rule, "Er bin nett.", 2)
	vaAssertBad(t, lt, rule, "Er gelangst zu ihr.", 2)
	vaAssertBadMsg(t, lt, rule, "Er können heute leider nicht kommen.", "Subjekt (Er) und Prädikat (können)")
	vaAssertBad(t, lt, rule, "Er lebst.", 2)
	vaAssertBad(t, lt, rule, "Ich bist nett.", 2)
	vaAssertBad(t, lt, rule, "Ich kannst heute leider nicht kommen.", 2)
	vaAssertBad(t, lt, rule, "Ich leben.", 1)
	vaAssertBadSugs(t, lt, rule, "Ich leben.", 1,
		[]string{"Ich lebe", "Ich lebte", "Wir leben", "Sie leben"})
	vaAssertBad(t, lt, rule, "Lebe du?", 1)
	vaAssertBadSugs(t, lt, rule, "Lebe du?", 1,
		[]string{"Lebest du", "Lebst du", "Lebe er", "Lebe es", "Lebtest du", "Lebe ich", "Lebe sie"})
	vaAssertBad(t, lt, rule, "Leben du?", 1)
	vaAssertBad(t, lt, rule, "Nett bist ich nicht.", 2)
	vaAssertBadSugs(t, lt, rule, "Nett bist ich nicht.", 2,
		[]string{"bin ich", "bist du", "sei ich", "wäre ich", "war ich"})
	vaAssertBad(t, lt, rule, "Nett sind du.", 1)
	vaAssertBad(t, lt, rule, "Nett sind er.", 1)
	vaAssertBadSugs(t, lt, rule, "Nett sind er.", 1,
		[]string{"sind wir", "sei er", "ist er", "sind sie", "wäre er", "war er"})
	vaAssertBad(t, lt, rule, "Nett warst wir.", 2)
	vaAssertBad(t, lt, rule, "Wir bin nett.", 2)
	vaAssertBad(t, lt, rule, "Wir gelangst zu ihr.", 2)
	vaAssertBad(t, lt, rule, "Wir könnt heute leider nicht kommen.", 1)
	vaAssertBadMsg(t, lt, rule, "Wünscht du dir mehr Zeit?", "Subjekt (du) und Prädikat (Wünscht)")
	vaAssertBad(t, lt, rule, "Wir lebst noch.", 2)
	vaAssertBadSugs(t, lt, rule, "Wir lebst noch.", 2,
		[]string{"Wir leben", "Wir lebten", "Du lebst"})
	vaAssertBadSugs(t, lt, rule, "Er sagte düster: „Ich brauchen mich nicht schuldig fühlen.“", 1,
		[]string{"Ich brauche", "Ich brauchte", "Ich bräuchte", "Wir brauchen", "Sie brauchen"})
	vaAssertBadSugs(t, lt, rule, "Er sagte: „Ich brauchen mich nicht schuldig fühlen.“", 1,
		[]string{"Ich brauche", "Ich brauchte", "Ich bräuchte", "Wir brauchen", "Sie brauchen"})
}

// TestVerbAgreementRule_BooleanAndFiniteVerb ports the inner helper control flow.
func TestVerbAgreementRule_BooleanAndFiniteVerb(t *testing.T) {
	// und between tokens → match true (coordinated subject)
	und := atrWithPOS("und", "KON:NEB", "und")
	bin := atrWithPOS("bin", "VER:1:SIN:PRÄ:NON", "sein")
	check := verbDoesMatchPersonAndNumber(und, bin, "1", "SIN", nil)
	require.True(t, check.verbDoesMatchPersonAndNumber)

	// finite verb mismatch → match false, finiteVerb set
	ist := atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein")
	check = verbDoesMatchPersonAndNumber(ist, atrWithPOS("müde", "ADJ:PRD:GRU", "müde"), "1", "SIN", nil)
	require.False(t, check.verbDoesMatchPersonAndNumber)
	require.NotNil(t, check.finiteVerb)
	require.Equal(t, "ist", check.finiteVerb.GetToken())

	// no finite verb → match true
	check = verbDoesMatchPersonAndNumber(
		atrWithPOS("sehr", "ADV", "sehr"),
		atrWithPOS("müde", "ADJ:PRD:GRU", "müde"),
		"1", "SIN", nil,
	)
	require.True(t, check.verbDoesMatchPersonAndNumber)
}
