package ca

import (
	"regexp"
	"strings"
)

// ApostophationHelper ports org.languagetool.rules.ca.ApostophationHelper.
var prepDet = map[string]string{
	"MS": "el ", "FS": "la ", "MP": "els ", "FP": "les ",
	"MSapos": "l'", "FSapos": "l'",
	"aMS": "al ", "aFS": "a la ", "aMP": "als ", "aFP": "a les ",
	"aMSapos": "a l'", "aFSapos": "a l'",
	"dMS": "del ", "dFS": "de la ", "dMP": "dels ", "dFP": "de les ",
	"dMSapos": "de l'", "dFSapos": "de l'",
	"pMS": "pel ", "pFS": "per la ", "pMP": "pels ", "pFP": "per les ",
	"pMSapos": "per l'", "pFSapos": "per l'",
}

var (
	pMascYes     = regexp.MustCompile(`(?i)^h?[aeiouàèéíòóú].*`)
	pMascNo      = regexp.MustCompile(`(?i)^h?[ui][aeioàèéóò].+`)
	pFemYes      = regexp.MustCompile(`(?i)^h?[aeoàèéíòóú].*|^h?[ui][^aeiouàèéíòóúüï]+[aeiou][ns]?$|^urbs$`)
	pFemNo       = regexp.MustCompile(`(?i)^host$|^ira$|^inxa$`)
	pHacAspirada = regexp.MustCompile(`(?i)Higgs|high|Hildesheim|Hill|hijabs?|Hillary|Himmler|hip-hop|hippies|hippy|hipsters?|Hirado|His|hits?|Hubei|Hudson|Hunter|Husserl|Huygens|husky|Utah|hides?|honey.*|Hartle.*|happy|happi.*|Hulk|Heart.*|Haakon|Halberstadt|Harley|Huck.*|Hanna|haka|hakes|Hama|Hornbostel|Heidi|Hayao|Hansi|Haas|Hindemith|user|users|one|head|history|Human|Hampshire|Hovedstaden|Handmade|Helm|Hahnem.*|hikimor.*|Houdini|Hugging|Heritage|hardcore|hancock|hender.*|h[ei][zs]b[ou]l.*|harira|Hawth.*|Henk.*|Humphry|Hohle|Höhle|Hooke|hajj.*|Hochschule|Hoch.*|Hutt|Hansel|Henley|hook|Handstand|Hull|Hatshepsut|Hatchepsut|Hana|Hamri|Hanley|Halis|Huxley|Hess|Hatteras|Herzberg|Hanlon|Harriet|hawl.*|hard|hip|herderi.*|Hangouts|Hayes|hostings?|Hal|hajj|Hermann|Hannah|Hertzsprung|Hotmail|Homrani|Harris|Harvey|Hunspell|Hassan|Haddock|Haarle[mn].*|Hainan|haendel.*|händel.*|habermas.*|hadits?|Hanuk?kà|hack.*|Harlem|Harper|Hartford|Haifa|haikus?|haima|haimes|Haikou|halal|halar|Halifax|Halmstad|halls?|Halle|Halley|Hallstatt|Hallstein|Halloweens?|Hals|herr|Herut|Hamadan|Hamas|Hamàs|hamilton.*|Hamlet.*|hammams?|Hammond|Hampton|hàmsters?|h[aà]ndicaps?|Hangzhou|Hannover|Hanoi|Hans|Hansa|hanseàti[cq].*|happenings?|Harbin|hardware|Haneke|harolds?|Hatay|Hamleigh |Harrisburg|Harrison|harrods?|harry|Hartley|Hartmann?|Hartree|Haruki|Har[td]?vard|Harz|hash.*|Hastings|Havel|Havilland|hawai.*|hawk.*|Hayek|Haydn.*|Hayworth|Heard|hearst|Heathrow|heav.*|hegel.*|Hebei|Hedmark|Heerenveen|Hedw.*|Heerlen|Hefei|Heidelberg|Heide[gn].*|Heilbronn|Heilongjiang|Heilig.*|hei[nk].*|Heisen.*|Heitz|Helmand|Helmholtz|Helen|Helsingborg|Hèlsinki|Heming.*|Henan|henna|hennes|Henry|Hepburn|herbert.*|Herder|Hereford|Herford|Herning|Hertfordshire|Herzog|Hesse|Hessen.*|Hewlett.*|H[ie]zbol·?l.+|high.*|hilbert.*|Hilda|Hillingdon|hinden.*|Hilton|hinterlands?|Hirsch.*|Hitch.*|hitler.*|Hilversum|Hobart|Hockenheim|Hodeida|Hohhot|Hokkaido|hobbes.*|hobby|Hogw.*|hobbies|Hodgkin|Hohen.*|Hölderlin|h[òo]ldings?|holy.*|hollywood.*|Holmes.*|Holstein|Hong|Hong-Kong|hongk.+|Honolu.+|Honsh[uū]|h[òo]bbits?|hooligan.*|hoover.*|hopkins|Hork.*|Horowitz|horst|H[ou]f.*|Houla|house|Houston|Howard|Hoyerswerda|Hunan|Huddersfield|Hunedoara|huskys?|huskies|hubs?|Hubble|humbold.*|Hume|hunting.*|Hussein|husseinit.+|Unity|university|united.*|European|OneDrive`)
)

// GetPrepositionAndDeterminer returns the determiner/preposition phrase for a form.
// genderNumber is MS/FS/MP/FP (or C*/ *N normalized); preposition is "", "a", "de", or "per".
func GetPrepositionAndDeterminer(newForm, genderNumber, preposition string) string {
	if strings.HasPrefix(genderNumber, "C") && len(genderNumber) >= 2 {
		genderNumber = "M" + string(genderNumber[1])
	}
	if strings.HasSuffix(genderNumber, "N") && len(genderNumber) >= 1 {
		genderNumber = string(genderNumber[0]) + "S"
	}
	prep := ""
	if preposition != "" {
		prep = strings.ToLower(string([]rune(preposition)[0]))
	}
	apos := ""
	if !pHacAspirada.MatchString(newForm) {
		switch genderNumber {
		case "MS":
			if pMascYes.MatchString(newForm) && !pMascNo.MatchString(newForm) {
				apos = "apos"
			}
		case "FS":
			if pFemYes.MatchString(newForm) && !pFemNo.MatchString(newForm) {
				apos = "apos"
			}
		}
	}
	return prepDet[prep+genderNumber+apos]
}
