package en

import "regexp"

// enVariantBlogPatterns ports AbstractEnglishSpellerRule.wordPatterns/blogLinks (9/9).
// Patterns use Java CASE_INSENSITIVE | UNICODE_CASE → Go (?i).
var enVariantBlogPatterns = []struct {
	pat *regexp.Regexp
	url string
}{
	{regexp.MustCompile(`(?i).*[yi][zs]e([sd])?|.*[yi][zs]ings?|.*i[zs]ations?`), "https://quillbot.com/blog/category/uk-vs-us/"},
	{regexp.MustCompile(`(?i).*(defen[cs]e|offen[sc]e|preten[sc]e).*`), "https://quillbot.com/blog/category/uk-vs-us/"},
	{regexp.MustCompile(`(?i).*og|.*ogue`), "https://quillbot.com/blog/category/uk-vs-us/"},
	{regexp.MustCompile(`(?i).*(or|our).*`), "https://languagetool.org/insights/post/our-or/#colour-or-color%E2%80%94colourise-or-colorize"},
	{regexp.MustCompile(`(?i).*e?able|.*dge?ments?|aging|ageing|ax|axe|.*grame?s?|neuron|neurone|neurons|neurones`), "https://languagetool.org/insights/post/our-or/#likeable-vs-likable-judgement-vs-judgment-oestrogen-vs-estrogen"},
	{regexp.MustCompile(`(?i).*(centre|center).*|.*(re|er)`), "https://languagetool.org/insights/post/re-vs-er/#the-difference-of-%E2%80%9C-reer%E2%80%9D-at-the-center-of-attention"},
	{regexp.MustCompile(`(?i)canceled|cancelled|canceling|cancelling|chili|chilli|chilies|chillies|chilis|chillis|counselor|counsellor|counselors|counsellors|defueled|defuelled|defueling|defuelling|defuelings|defuellings|dialed|dialled|dialer|dialler|dialers|diallers|dialing|dialling|dialog|dialogue|dialogize|dialogise|dialogized|dialogised|dialogizes|dialogises|dialogizing|dialogising|dialogs|dialogues|dialyzable|dialysable|dialyze|dialyse|dialyzed|dialysed|dialyzes|dialyses|dialyzing|dialysing|enroll|enrol|enrolled|enroled|enrolling|enroling|enrollment|enrolment|enrollments|enrolments|enrolls|enrols|fueled|fuelled|fueling|fuelling|fulfill|fulfil|fulfillment|fulfilment|fulfills|fulfils|installment|instalment|installments|instalments|jewelry|jewellery|labeled|labelled|labeling|labelling|marvelous|marvellous|medalist|medallist|medalists|medallists|modeled|modelled|modeling|modelling|noise-canceling|noise-cancelling|refueled|refuelled|refueling|refuelling|relabeled|relabelled|relabeling|relabelling|remodeled|remodelled|remodeling|remodelling|signalization|signalisation|signalize|signalise|signalized|signalised|signalizes|signalises|signalizing|signalising|skillful|skilful|skillfully|skilfully|tranquilize|tranquillize|tranquilized|tranquillized|tranquilizes|tranquillizes|traveled|travelled|traveler|traveller|travelers|travellers|traveling|travelling|uncanceled|uncancelled|uncanceling|uncancelling|unlabeled|unlabelled|wooly|woolly`), "https://languagetool.org/insights/post/re-vs-er/#british-english-prefers-doubling-consonants-doesn%E2%80%99t-it"},
	{regexp.MustCompile(`(?i)airfoil|aerofoil|airfoils|aerofoils|airplane|aeroplane|airplanes|aeroplanes|aluminum|aluminium|artifact|artefact|artifacts|artefacts|backdraft|backdraught|cozy|cosy|`), "https://languagetool.org/insights/post/re-vs-er/#more-radical-differences-between-british-and-american-english-spellings"},
	{regexp.MustCompile(`(?i)amenorrhea|amenorrhoea|anesthesia|anaesthesia|anesthesias|anaesthesias|anesthetic|anaesthetic|anesthetically|anaesthetically|anesthetics|anaesthetics|anesthetist|anaesthetist|anesthetists|anaesthetists|anesthetization|anaesthetisation|anesthetizations|anaesthetisations|anesthetize|anaesthetise|anesthetized|anaesthetised|anesthetizes|anaesthetises|anesthetizing|anaesthetising|archeological|archaeological|archeologically|archaeologically|archeologies|archaeologies|archeology|archaeology|cesium|caesium|diarrhea|diarrhoea|diarrheal|diarrhoeal|dyslipidemia|dyslipidaemia|dyslipidemias|dyslipidaemias|edematous|oedematous|encyclopedia|encyclopaedia|encyclopedias|encyclopaedias|eon|aeon|eons|aeons|esophagi|oesophagi|esophagus|oesophagus|esophaguses|oesophaguses|esthetic|aesthetic|esthetical|aesthetical|esthetically|aesthetically|esthetician|aesthetician|estheticians|aestheticians|estrogen|oestrogen|estrus|oestrus|etiologies|aetiologies|etiology|aetiology|feces|faeces|fetal|foetal|fetus|foetus|fetuses|foetuses|gastroesophageal|gastro-oesophageal|glycemic|glycaemic|gynecomastia|gynaecomastia|hematemesis|haematemesis|hematoma|haematoma|hematomas|haematomas|hematopoietic|haematopoietic|hematuria|haematuria|hematurias|haematurias|hemolytic|haemolytic|hemophilia|haemophilia|hemorrhage|haemorrhage|hemorrhages|haemorrhages|hemostasis|haemostasis|homeopathies|homoeopathies|homeopathy|homoeopathy|hyperemia|hyperaemia|hyperemic|hyperaemic|hypnopedia|hypnopaedia|hypnopedic|hypnopaedic|hypocalcaemia|hypocalcaemia|hypokalaemic|hypokalemic|kinesthesia|kinaesthesia|kinesthesis|kinaesthesis|kinesthetic|kinaesthetic|kinesthetically|kinaesthetically|maneuver|manoeuvre|maneuvers|manoeuvres|orthopedic|orthopaedic|orthopedics|orthopaedics|paleoecology|palaeoecology|paleogeographical|palaeogeographical|paleogeographically|palaeogeographically|paleogeography|palaeogeography|paresthesia|paraesthesia|pediatric|paediatric|pediatrically|paediatrically|pediatrician|paediatrician|pediatricians|paediatricians|pedomorphic|paedomorphic|pedophile|paedophile|pedophiles|paedophiles|polycythemia|polycythaemia|pretorium|praetorium|pyorrhea|pyorrhoea|septicemia|septicaemia|synesthesia|synaesthesia|synesthete|synaesthete|synesthetes|synaesthetes|tracheoesophageal|tracheo-oesophageal`), "https://languagetool.org/insights/post/our-or/#likeable-vs-likable-judgement-vs-judgment-oestrogen-vs-estrogen"},
}

// enVariantBlogURL ports Match loop that sets blog URL when isValidInOtherVariant.
func enVariantBlogURL(word string) string {
	for _, e := range enVariantBlogPatterns {
		if e.pat.MatchString(word) {
			return e.url
		}
	}
	return ""
}
