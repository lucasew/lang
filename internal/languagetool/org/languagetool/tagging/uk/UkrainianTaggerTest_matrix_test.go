package uk

// Twin of UkrainianTaggerTest — full active myAssert / assertNotTagged matrix
// against official ukrainian.dict + UkrainianWordTokenizer (Java setup).

import (
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	tokuk "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/uk"
	"github.com/stretchr/testify/require"
)

// javaUkrainianTaggerMyAssertCases: every active TestTools.myAssert from UkrainianTaggerTest.
var javaUkrainianTaggerMyAssertCases = []struct{ in, want string }{
	{"києві", "києві/[кий]noun:inanim:m:v_dav|києві/[кий]noun:inanim:m:v_mis"},
	{"Києві", "Києві/[Кий]noun:anim:m:v_dav:prop:fname|Києві/[Кий]noun:anim:m:v_mis:prop:fname|Києві/[Київ]noun:inanim:m:v_mis:prop:geo|Києві/[кий]noun:inanim:m:v_dav|Києві/[кий]noun:inanim:m:v_mis"},
	{"Ізюму", "Ізюму/[Ізюм]noun:inanim:m:v_dav:prop:geo|Ізюму/[Ізюм]noun:inanim:m:v_mis:prop:geo|Ізюму/[Ізюм]noun:inanim:m:v_rod:prop:geo|Ізюму/[ізюм]noun:inanim:m:v_dav:bad|Ізюму/[ізюм]noun:inanim:m:v_mis:bad|Ізюму/[ізюм]noun:inanim:m:v_rod:bad"},
	{"віл", "віл/[віл]noun:anim:m:v_naz"},
	{"Віл", "Віл/[віл]noun:anim:m:v_naz"},
	{"ВІЛ", "ВІЛ/[ВІЛ]noun:inanim:m:v_dav:nv:abbr|ВІЛ/[ВІЛ]noun:inanim:m:v_mis:nv:abbr|ВІЛ/[ВІЛ]noun:inanim:m:v_naz:nv:abbr|ВІЛ/[ВІЛ]noun:inanim:m:v_oru:nv:abbr|ВІЛ/[ВІЛ]noun:inanim:m:v_rod:nv:abbr|ВІЛ/[ВІЛ]noun:inanim:m:v_zna:nv:abbr|ВІЛ/[віл]noun:anim:m:v_naz"},
	{"далі", "далі/[даль]noun:inanim:f:v_dav|далі/[даль]noun:inanim:f:v_mis|далі/[даль]noun:inanim:f:v_rod|далі/[даль]noun:inanim:p:v_kly|далі/[даль]noun:inanim:p:v_naz|далі/[даль]noun:inanim:p:v_zna|далі/[далі]adv:compc:predic"},
	{"Далі", "Далі/[Даль]noun:anim:m:v_mis:prop:lname|Далі/[Даля]noun:anim:f:v_dav:prop:fname|Далі/[Даля]noun:anim:f:v_mis:prop:fname|Далі/[Даля]noun:anim:f:v_rod:prop:fname|Далі/[Даля]noun:anim:p:v_kly:prop:fname|Далі/[Даля]noun:anim:p:v_naz:prop:fname|Далі/[Далі]noun:anim:m:v_dav:nv:prop:lname|Далі/[Далі]noun:anim:m:v_kly:nv:prop:lname|Далі/[Далі]noun:anim:m:v_mis:nv:prop:lname|Далі/[Далі]noun:anim:m:v_naz:nv:prop:lname|Далі/[Далі]noun:anim:m:v_oru:nv:prop:lname|Далі/[Далі]noun:anim:m:v_rod:nv:prop:lname|Далі/[Далі]noun:anim:m:v_zna:nv:prop:lname|Далі/[даль]noun:inanim:f:v_dav|Далі/[даль]noun:inanim:f:v_mis|Далі/[даль]noun:inanim:f:v_rod|Далі/[даль]noun:inanim:p:v_kly|Далі/[даль]noun:inanim:p:v_naz|Далі/[даль]noun:inanim:p:v_zna|Далі/[далі]adv:compc:predic"},
	{"Бен", "Бен/[Бен]noun:anim:m:v_naz:prop:fname|Бен/[бен]part:pers"},
	{"бен", "бен/[бен]part:pers"},
	{"френдити", "френдити/[френдити]verb:imperf:inf:slang"},
	{"Справу порушено судом", "Справу/[справа]noun:inanim:f:v_zna -- порушено/[порушити]verb:perf:impers -- судом/[суд]noun:inanim:m:v_oru|судом/[судома]noun:inanim:p:v_rod"},
	// testTagger multi-token myAssert (Java UkrainianTaggerTest.testTagger)
	{"Майже два роки тому Люба разом із чоловіком Степаном виїхали туди на проживання.", "Майже/[майже]adv -- два/[два]numr:p:v_naz|два/[два]numr:p:v_zna -- роки/[рік]noun:inanim:p:v_kly|роки/[рік]noun:inanim:p:v_naz|роки/[рік]noun:inanim:p:v_zna -- тому/[те]noun:inanim:n:v_dav:pron:dem|тому/[те]noun:inanim:n:v_mis:pron:dem|тому/[той]adj:m:v_dav:pron:dem|тому/[той]adj:m:v_mis:pron:dem|тому/[той]adj:n:v_dav:pron:dem|тому/[той]adj:n:v_mis:pron:dem|тому/[том]noun:inanim:m:v_dav|тому/[том]noun:inanim:m:v_mis|тому/[том]noun:inanim:m:v_rod|тому/[тому]adv|тому/[тому]conj:subord -- Люба/[Люба]noun:anim:f:v_naz:prop:fname|Люба/[любий]adj:f:v_kly:compb|Люба/[любий]adj:f:v_naz:compb -- разом/[раз]noun:inanim:m:v_oru|разом/[разом]adv -- із/[із]prep -- чоловіком/[чоловік]noun:anim:m:v_oru -- Степаном/[Степан]noun:anim:m:v_oru:prop:fname -- виїхали/[виїхати]verb:perf:past:p -- туди/[туди]adv:pron:dem -- на/[на]intj|на/[на]part|на/[на]prep -- проживання/[проживання]noun:inanim:n:v_kly|проживання/[проживання]noun:inanim:n:v_naz|проживання/[проживання]noun:inanim:n:v_rod|проживання/[проживання]noun:inanim:n:v_zna|проживання/[проживання]noun:inanim:p:v_kly|проживання/[проживання]noun:inanim:p:v_naz|проживання/[проживання]noun:inanim:p:v_zna"},
	{"аудіо-", "аудіо-/[аудіо-]noninfl"},
	{"вкраїна", "вкраїна/[вкраїна]noun:inanim:f:v_naz:prop:geo:bad:err"},
	{"101,234", "101,234/[101,234]number"},
	{"101 234", "101 234/[101 234]number"},
	{"3,5-5,6 7° 7,4°C", "3,5-5,6/[3,5-5,6]number -- 7/[7]number -- 7,4/[7,4]number -- C/[C]number:latin"},
	{"XIX", "XIX/[XIX]number:latin"},
	{"II", "II/[II]number:latin"},
	{"X", "X/[X]number:latin"},
	{"D", "D/[null]null"},
	{"ХІХ", "ХІХ/[ХІХ]number:latin:bad:err"},
	{"ІV", "ІV/[ІV]number:latin:bad"},
	{"ХІХ-го", "ХІХ-го/[ХІХ-го]number:latin:bad"},
	{"V-типу", "V-типу/[null]null"},
	{"10–15", "10–15/[10–15]number"},
	{"14.07.2001", "14.07.2001/[14.07.2001]date"},
	{"15:33", "15:33/[15:33]time"},
	{"15:33:00", "15:33:00/[15:33:00]time"},
	{"о 15.33", "о/[о]intj|о/[о]prep -- 15.33/[15.33]time"},
	{"О 1:05", "О/[о]intj|О/[о]prep -- 1:05/[1:05]time"},
	{"#янебоюсьсказати", "#янебоюсьсказати/[#янебоюсьсказати]hashtag"},
	{"300 р. до н. е.", "300/[300]number -- р./[р.]noun:inanim:f:v_dav:nv:abbr:xp2|р./[р.]noun:inanim:f:v_mis:nv:abbr:xp2|р./[р.]noun:inanim:f:v_naz:nv:abbr:xp2|р./[р.]noun:inanim:f:v_oru:nv:abbr:xp2|р./[р.]noun:inanim:f:v_rod:nv:abbr:xp2|р./[р.]noun:inanim:f:v_zna:nv:abbr:xp2|р./[р.]noun:inanim:m:v_dav:nv:abbr:xp1|р./[р.]noun:inanim:m:v_mis:nv:abbr:xp1|р./[р.]noun:inanim:m:v_naz:nv:abbr:xp1|р./[р.]noun:inanim:m:v_oru:nv:abbr:xp1|р./[р.]noun:inanim:m:v_rod:nv:abbr:xp1|р./[р.]noun:inanim:m:v_zna:nv:abbr:xp1 -- до/[до]noun:inanim:n:v_dav:nv|до/[до]noun:inanim:n:v_kly:nv|до/[до]noun:inanim:n:v_mis:nv|до/[до]noun:inanim:n:v_naz:nv|до/[до]noun:inanim:n:v_oru:nv|до/[до]noun:inanim:n:v_rod:nv|до/[до]noun:inanim:n:v_zna:nv|до/[до]noun:inanim:p:v_dav:nv|до/[до]noun:inanim:p:v_kly:nv|до/[до]noun:inanim:p:v_mis:nv|до/[до]noun:inanim:p:v_naz:nv|до/[до]noun:inanim:p:v_oru:nv|до/[до]noun:inanim:p:v_rod:nv|до/[до]noun:inanim:p:v_zna:nv|до/[до]prep -- н./[н.]adj:f:v_dav:nv:abbr|н./[н.]adj:f:v_mis:nv:abbr|н./[н.]adj:f:v_naz:nv:abbr|н./[н.]adj:f:v_oru:nv:abbr|н./[н.]adj:f:v_rod:nv:abbr|н./[н.]adj:f:v_zna:nv:abbr|н./[н.]noun:inanim:n:v_rod:nv:abbr -- е./[е.]noun:inanim:f:v_dav:nv:abbr|е./[е.]noun:inanim:f:v_mis:nv:abbr|е./[е.]noun:inanim:f:v_naz:nv:abbr|е./[е.]noun:inanim:f:v_oru:nv:abbr|е./[е.]noun:inanim:f:v_rod:nv:abbr|е./[е.]noun:inanim:f:v_zna:nv:abbr|е./[е.]noun:inanim:p:v_dav:nv:abbr|е./[е.]noun:inanim:p:v_mis:nv:abbr|е./[е.]noun:inanim:p:v_naz:nv:abbr|е./[е.]noun:inanim:p:v_oru:nv:abbr|е./[е.]noun:inanim:p:v_rod:nv:abbr|е./[е.]noun:inanim:p:v_zna:nv:abbr"},
	{"300 тис. гривень", "300/[300]number -- тис./[тис.]noun:inanim:f:v_dav:nv:numr:abbr|тис./[тис.]noun:inanim:f:v_mis:nv:numr:abbr|тис./[тис.]noun:inanim:f:v_naz:nv:numr:abbr|тис./[тис.]noun:inanim:f:v_oru:nv:numr:abbr|тис./[тис.]noun:inanim:f:v_rod:nv:numr:abbr|тис./[тис.]noun:inanim:f:v_zna:nv:numr:abbr|тис./[тис.]noun:inanim:n:v_dav:nv:abbr|тис./[тис.]noun:inanim:n:v_mis:nv:abbr|тис./[тис.]noun:inanim:n:v_naz:nv:abbr|тис./[тис.]noun:inanim:n:v_oru:nv:abbr|тис./[тис.]noun:inanim:n:v_rod:nv:abbr|тис./[тис.]noun:inanim:n:v_zna:nv:abbr|тис./[тис.]noun:inanim:p:v_dav:nv:abbr|тис./[тис.]noun:inanim:p:v_dav:nv:numr:abbr|тис./[тис.]noun:inanim:p:v_mis:nv:abbr|тис./[тис.]noun:inanim:p:v_mis:nv:numr:abbr|тис./[тис.]noun:inanim:p:v_naz:nv:abbr|тис./[тис.]noun:inanim:p:v_naz:nv:numr:abbr|тис./[тис.]noun:inanim:p:v_oru:nv:abbr|тис./[тис.]noun:inanim:p:v_oru:nv:numr:abbr|тис./[тис.]noun:inanim:p:v_rod:nv:abbr|тис./[тис.]noun:inanim:p:v_rod:nv:numr:abbr|тис./[тис.]noun:inanim:p:v_zna:nv:abbr|тис./[тис.]noun:inanim:p:v_zna:nv:numr:abbr -- гривень/[гривня]noun:inanim:p:v_rod"},
	{"Валерій (міліціонер-пародист. – Авт.) стане пародистом.", "Валерій/[Валерій]noun:anim:m:v_naz:prop:fname -- міліціонер-пародист/[міліціонер-пародист]noun:anim:m:v_naz -- Авт./[авт.]adj:f:v_dav:nv:abbr|Авт./[авт.]adj:f:v_mis:nv:abbr|Авт./[авт.]adj:f:v_naz:nv:abbr|Авт./[авт.]adj:f:v_oru:nv:abbr|Авт./[авт.]adj:f:v_rod:nv:abbr|Авт./[авт.]adj:f:v_zna:nv:abbr|Авт./[авт.]adj:m:v_dav:nv:abbr|Авт./[авт.]adj:m:v_mis:nv:abbr|Авт./[авт.]adj:m:v_naz:nv:abbr|Авт./[авт.]adj:m:v_oru:nv:abbr|Авт./[авт.]adj:m:v_rod:nv:abbr|Авт./[авт.]adj:m:v_zna:ranim:nv:abbr|Авт./[авт.]adj:m:v_zna:rinanim:nv:abbr|Авт./[авт.]adj:n:v_dav:nv:abbr|Авт./[авт.]adj:n:v_mis:nv:abbr|Авт./[авт.]adj:n:v_naz:nv:abbr|Авт./[авт.]adj:n:v_oru:nv:abbr|Авт./[авт.]adj:n:v_rod:nv:abbr|Авт./[авт.]adj:n:v_zna:nv:abbr|Авт./[авт.]adj:p:v_dav:nv:abbr|Авт./[авт.]adj:p:v_mis:nv:abbr|Авт./[авт.]adj:p:v_naz:nv:abbr|Авт./[авт.]adj:p:v_oru:nv:abbr|Авт./[авт.]adj:p:v_rod:nv:abbr|Авт./[авт.]adj:p:v_zna:ranim:nv:abbr|Авт./[авт.]adj:p:v_zna:rinanim:nv:abbr|Авт./[авт.]noun:anim:m:v_naz:nv:abbr -- стане/[стан]noun:inanim:m:v_kly:xp1|стане/[стан]noun:inanim:m:v_kly:xp2|стане/[станути]verb:perf:futr:s:3|стане/[стати]verb:perf:futr:s:3 -- пародистом/[пародист]noun:anim:m:v_oru"},
	{"Є. Бакуліна.", "Є./[null]null -- Бакуліна/[Бакулін]noun:anim:m:v_rod:prop:lname|Бакуліна/[Бакулін]noun:anim:m:v_zna:prop:lname|Бакуліна/[Бакуліна]noun:anim:f:v_kly:prop:lname|Бакуліна/[Бакуліна]noun:anim:f:v_naz:prop:lname"},
	{"і т. д.", "і/[і]conj:coord|і/[і]part -- т./[т.]adj:n:v_dav:nv:abbr|т./[т.]adj:n:v_mis:nv:abbr|т./[т.]adj:n:v_naz:nv:abbr|т./[т.]adj:n:v_oru:nv:abbr|т./[т.]adj:n:v_rod:nv:abbr|т./[т.]adj:n:v_zna:nv:abbr|т./[т.]adv:abbr|т./[т.]noun:inanim:m:v_dav:nv:abbr|т./[т.]noun:inanim:m:v_mis:nv:abbr|т./[т.]noun:inanim:m:v_naz:nv:abbr|т./[т.]noun:inanim:m:v_oru:nv:abbr|т./[т.]noun:inanim:m:v_rod:nv:abbr|т./[т.]noun:inanim:m:v_zna:nv:abbr|т./[т.]noun:inanim:p:v_dav:nv:abbr|т./[т.]noun:inanim:p:v_mis:nv:abbr|т./[т.]noun:inanim:p:v_naz:nv:abbr|т./[т.]noun:inanim:p:v_oru:nv:abbr|т./[т.]noun:inanim:p:v_rod:nv:abbr|т./[т.]noun:inanim:p:v_zna:nv:abbr -- д./[д.]adv:abbr"},
	{"УКРАЇНА", "УКРАЇНА/[Україна]noun:inanim:f:v_naz:prop:geo|УКРАЇНА/[україна]noun:inanim:f:v_naz:prop:geo:alt|УКРАЇНА/[україна]noun:inanim:f:v_naz:prop:geo:alt"},
	{"СИРІЮ", "СИРІЮ/[Сирія]noun:inanim:f:v_zna:prop:geo|СИРІЮ/[сиріти]verb:imperf:pres:s:1"},
	{"НЬЮ-ЙОРК", "НЬЮ-ЙОРК/[Нью-Йорк]noun:inanim:m:v_naz:prop:geo:xp1|НЬЮ-ЙОРК/[Нью-Йорк]noun:inanim:m:v_naz:prop:geo:xp2|НЬЮ-ЙОРК/[Нью-Йорк]noun:inanim:m:v_zna:prop:geo:xp1|НЬЮ-ЙОРК/[Нью-Йорк]noun:inanim:m:v_zna:prop:geo:xp2"},
	{"п'ятсотдвадцятип'ятиріччя", "п'ятсотдвадцятип'ятиріччя/[п'ятсотдвадцятип'ятиріччя]noun:inanim:n:v_naz|п'ятсотдвадцятип'ятиріччя/[п'ятсотдвадцятип'ятиріччя]noun:inanim:n:v_rod|п'ятсотдвадцятип'ятиріччя/[п'ятсотдвадцятип'ятиріччя]noun:inanim:n:v_zna"},
	{"п’ятсотдвадцятип'ятирі́ччям", "п'ятсотдвадцятип'ятиріччям/[п'ятсотдвадцятип'ятиріччя]noun:inanim:n:v_oru"},
	{"шістдесятип’ятимільйо́нний", "шістдесятип'ятимільйонний/[шістдесятип'ятимільйонний]adj:m:v_naz:numr|шістдесятип'ятимільйонний/[шістдесятип'ятимільйонний]adj:m:v_zna:rinanim:numr"},
	{"стодвадцятип’ятирі́чний", "стодвадцятип'ятирічний/[стодвадцятип'ятирічний]adj:m:v_naz|стодвадцятип'ятирічний/[стодвадцятип'ятирічний]adj:m:v_zna:rinanim"},
	{"кількадесятитисячної", "кількадесятитисячної/[кількадесятитисячний]adj:f:v_rod:numr"},
	{"Тридцятидвохрічний", "Тридцятидвохрічний/[Тридцятидвохрічний]adj:m:v_naz|Тридцятидвохрічний/[Тридцятидвохрічний]adj:m:v_zna:rinanim"},
	{"одним-трьома", "одним-трьома/[один-три]numr:p:v_oru|одним-трьома/[один-троє]numr:p:v_oru"},
	{"одного-другого", "одного-другого/[один-другий]adj:m:v_rod:numr|одного-другого/[один-другий]adj:m:v_zna:ranim:numr|одного-другого/[один-другий]adj:n:v_rod:numr"},
	{"п'яти-шести", "п'яти-шести/[п'ять-шість]numr:p:v_dav|п'яти-шести/[п'ять-шість]numr:p:v_mis|п'яти-шести/[п'ять-шість]numr:p:v_rod"},
	{"три-чотири", "три-чотири/[три-чотири]numr:p:v_naz|три-чотири/[три-чотири]numr:p:v_zna"},
	{"два-чотири", "два-чотири/[два-чотири]numr:p:v_naz|два-чотири/[два-чотири]numr:p:v_zna"},
	{"одному-двох", "одному-двох/[один-два]numr:p:v_mis"},
	{"три–чотири", "три-чотири/[три-чотири]numr:p:v_naz|три-чотири/[три-чотири]numr:p:v_zna|три–чотири/[null]null"},
	{"шість-сім", "шість-сім/[шість-сім]numr:p:v_naz|шість-сім/[шість-сім]numr:p:v_zna"},
	{"три-сім", "три-сім/[три-сім]numr:p:v_naz|три-сім/[три-сім]numr:p:v_zna"},
	{"абзац-два", "абзац-два/[абзац-два]noun:inanim:m:v_naz|абзац-два/[абзац-два]noun:inanim:m:v_zna|абзац-два/[абзац-два]noun:inanim:p:v_naz|абзац-два/[абзац-два]noun:inanim:p:v_zna"},
	{"сотні-дві", "сотні-дві/[сотня-два]noun:inanim:p:v_naz|сотні-дві/[сотня-два]noun:inanim:p:v_zna"},
	{"тисячею-трьома", "тисячею-трьома/[тисяча-три]noun:inanim:f:v_oru:numr|тисячею-трьома/[тисяча-три]noun:inanim:p:v_oru:numr|тисячею-трьома/[тисяча-троє]noun:inanim:f:v_oru:numr|тисячею-трьома/[тисяча-троє]noun:inanim:p:v_oru:numr"},
	{"друге-третє", "друге-третє/[другий-третій]adj:n:v_kly:numr|друге-третє/[другий-третій]adj:n:v_naz:numr|друге-третє/[другий-третій]adj:n:v_zna:numr"},
	{"100-річному", "100-річному/[100-річний]adj:m:v_dav|100-річному/[100-річний]adj:m:v_mis|100-річному/[100-річний]adj:n:v_dav|100-річному/[100-річний]adj:n:v_mis"},
	{"1-2-відсотковим", "1-2-відсотковим/[1-2-відсотковий]adj:m:v_oru|1-2-відсотковим/[1-2-відсотковий]adj:n:v_oru|1-2-відсотковим/[1-2-відсотковий]adj:p:v_dav"},
	{"10-класників", "10-класників/[10-класник]noun:anim:p:v_rod|10-класників/[10-класник]noun:anim:p:v_zna"},
	{"10-хвилинка", "10-хвилинка/[10-хвилинка]noun:inanim:f:v_naz"},
	{"11-12-річний", "11-12-річний/[11-12-річний]adj:m:v_kly|11-12-річний/[11-12-річний]adj:m:v_naz|11-12-річний/[11-12-річний]adj:m:v_zna:rinanim"},
	{"100-й", "100-й/[100-й]adj:f:v_dav:numr|100-й/[100-й]adj:f:v_mis:numr|100-й/[100-й]adj:m:v_naz:numr|100-й/[100-й]adj:m:v_zna:rinanim:numr"},
	{"50-х", "50-х/[50-й]adj:p:v_mis:numr|50-х/[50-й]adj:p:v_rod:numr|50-х/[50-й]adj:p:v_zna:ranim:numr"},
	{"11-ту", "11-ту/[11-й]adj:f:v_zna:numr"},
	{"3-ій", "3-ій/[3-й]adj:f:v_dav:numr|3-ій/[3-й]adj:f:v_mis:numr|3-ій/[3-й]adj:m:v_naz:numr|3-ій/[3-й]adj:m:v_zna:rinanim:numr"},
	{"5-ій", "5-ій/[5-й]adj:f:v_dav:numr|5-ій/[5-й]adj:f:v_mis:numr"},
	{"33-ми", "33-ми/[33-й]adj:p:v_oru:numr"},
	{"100%-й", "100%-й/[100%-й]adj:f:v_dav:numr|100%-й/[100%-й]adj:f:v_mis:numr|100%-й/[100%-й]adj:m:v_naz:numr|100%-й/[100%-й]adj:m:v_zna:rinanim:numr"},
	{"2000-ні", "2000-ні/[2000-й]adj:p:v_naz:numr|2000-ні/[2000-й]adj:p:v_zna:rinanim:numr"},
	{"58-мою", "58-мою/[58-й]adj:f:v_oru:numr:bad"},
	{"3-ми", "3-ми/[3-й]adj:p:v_oru:numr"},
	{"7-ма", "7-ма/[7-й]adj:f:v_naz:numr"},
	{"2-3-ма", "2-3-ма/[2-3]numr:p:v_oru:bad"},
	{"15-ти", "15-ти/[15]numr:p:v_dav:bad|15-ти/[15]numr:p:v_mis:bad|15-ти/[15]numr:p:v_rod:bad"},
	{"20-ці", "20-ці/[20]numr:f:v_dav:bad|20-ці/[20]numr:f:v_mis:bad"},
	{"42-ві", "42-ві/[42]numr:p:v_naz:bad|42-ві/[42]numr:p:v_zna:rinanim:bad"},
	{"148-ми", "148-ми/[148-й]adj:p:v_oru:numr"},
	{"54–річна", "54-річна/[54-річний]adj:f:v_kly|54-річна/[54-річний]adj:f:v_naz|54–річна/[null]null"},
	{"100-річчя", "100-річчя/[100-річчя]noun:inanim:n:v_kly|100-річчя/[100-річчя]noun:inanim:n:v_naz|100-річчя/[100-річчя]noun:inanim:n:v_rod|100-річчя/[100-річчя]noun:inanim:n:v_zna|100-річчя/[100-річчя]noun:inanim:p:v_kly|100-річчя/[100-річчя]noun:inanim:p:v_naz|100-річчя/[100-річчя]noun:inanim:p:v_zna"},
	{"100-РІЧЧЯМ", "100-РІЧЧЯМ/[100-річчя]noun:inanim:n:v_oru|100-РІЧЧЯМ/[100-річчя]noun:inanim:p:v_dav"},
	{"100-метрівка", "100-метрівка/[100-метрівка]noun:inanim:f:v_naz"},
	{"100-відсотково", "100-відсотково/[100-відсотково]adv"},
	{"120-мм", "120-мм/[120-мм]adj:f:v_dav|120-мм/[120-мм]adj:f:v_mis|120-мм/[120-мм]adj:f:v_naz|120-мм/[120-мм]adj:f:v_oru|120-мм/[120-мм]adj:f:v_rod|120-мм/[120-мм]adj:f:v_zna|120-мм/[120-мм]adj:m:v_dav|120-мм/[120-мм]adj:m:v_mis|120-мм/[120-мм]adj:m:v_naz|120-мм/[120-мм]adj:m:v_oru|120-мм/[120-мм]adj:m:v_rod|120-мм/[120-мм]adj:m:v_zna|120-мм/[120-мм]adj:n:v_dav|120-мм/[120-мм]adj:n:v_mis|120-мм/[120-мм]adj:n:v_naz|120-мм/[120-мм]adj:n:v_oru|120-мм/[120-мм]adj:n:v_rod|120-мм/[120-мм]adj:n:v_zna|120-мм/[120-мм]adj:p:v_dav|120-мм/[120-мм]adj:p:v_mis|120-мм/[120-мм]adj:p:v_naz|120-мм/[120-мм]adj:p:v_oru|120-мм/[120-мм]adj:p:v_rod|120-мм/[120-мм]adj:p:v_zna"},
	{"Євро-2014", "Євро-2014/[Євро-2014]noun:inanim:n:v_dav:nv:prop|Євро-2014/[Євро-2014]noun:inanim:n:v_mis:nv:prop|Євро-2014/[Євро-2014]noun:inanim:n:v_naz:nv:prop|Євро-2014/[Євро-2014]noun:inanim:n:v_oru:nv:prop|Євро-2014/[Євро-2014]noun:inanim:n:v_rod:nv:prop|Євро-2014/[Євро-2014]noun:inanim:n:v_zna:nv:prop"},
	{"Ігри-2014", "Ігри-2014/[Гра-2014]noun:inanim:p:v_naz:prop|Ігри-2014/[Гра-2014]noun:inanim:p:v_zna:prop"},
	{"ЧЄ-2014", "ЧЄ-2014/[ЧЄ-2014]noun:inanim:m:v_dav:nv:abbr:prop|ЧЄ-2014/[ЧЄ-2014]noun:inanim:m:v_mis:nv:abbr:prop|ЧЄ-2014/[ЧЄ-2014]noun:inanim:m:v_naz:nv:abbr:prop|ЧЄ-2014/[ЧЄ-2014]noun:inanim:m:v_oru:nv:abbr:prop|ЧЄ-2014/[ЧЄ-2014]noun:inanim:m:v_rod:nv:abbr:prop|ЧЄ-2014/[ЧЄ-2014]noun:inanim:m:v_zna:nv:abbr:prop"},
	{"Лондон-2014", "Лондон-2014/[Лондон-2014]noun:inanim:m:v_naz:prop|Лондон-2014/[Лондон-2014]noun:inanim:m:v_zna:prop"},
	{"Афіни-2014", "Афіни-2014/[Афіни-2014]noun:inanim:p:v_naz:ns:prop|Афіни-2014/[Афіни-2014]noun:inanim:p:v_zna:ns:prop"},
	{"Замок-2002", "Замок-2002/[Замок-2002]noun:inanim:m:v_naz:prop|Замок-2002/[Замок-2002]noun:inanim:m:v_zna:prop"},
	{"вибори-2002", "вибори-2002/[вибори-2002]noun:inanim:p:v_naz:ns|вибори-2002/[вибори-2002]noun:inanim:p:v_zna:ns"},
	{"Оскар-2021", "Оскар-2021/[Оскар-2021]noun:inanim:m:v_naz:prop|Оскар-2021/[Оскар-2021]noun:inanim:m:v_zna:prop"},
	{"Ан-140", "Ан-140/[Ан-140]noun:inanim:m:v_dav:nv|Ан-140/[Ан-140]noun:inanim:m:v_mis:nv|Ан-140/[Ан-140]noun:inanim:m:v_naz:nv|Ан-140/[Ан-140]noun:inanim:m:v_oru:nv|Ан-140/[Ан-140]noun:inanim:m:v_rod:nv|Ан-140/[Ан-140]noun:inanim:m:v_zna:nv|Ан-140/[Ан-140]noun:inanim:p:v_dav:nv|Ан-140/[Ан-140]noun:inanim:p:v_mis:nv|Ан-140/[Ан-140]noun:inanim:p:v_naz:nv|Ан-140/[Ан-140]noun:inanim:p:v_oru:nv|Ан-140/[Ан-140]noun:inanim:p:v_rod:nv|Ан-140/[Ан-140]noun:inanim:p:v_zna:nv"},
	{"ВАЗ-2104", "ВАЗ-2104/[ВАЗ-2104]noun:inanim:f:v_dav:nv|ВАЗ-2104/[ВАЗ-2104]noun:inanim:f:v_mis:nv|ВАЗ-2104/[ВАЗ-2104]noun:inanim:f:v_naz:nv|ВАЗ-2104/[ВАЗ-2104]noun:inanim:f:v_oru:nv|ВАЗ-2104/[ВАЗ-2104]noun:inanim:f:v_rod:nv|ВАЗ-2104/[ВАЗ-2104]noun:inanim:f:v_zna:nv|ВАЗ-2104/[ВАЗ-2104]noun:inanim:m:v_dav:nv|ВАЗ-2104/[ВАЗ-2104]noun:inanim:m:v_mis:nv|ВАЗ-2104/[ВАЗ-2104]noun:inanim:m:v_naz:nv|ВАЗ-2104/[ВАЗ-2104]noun:inanim:m:v_oru:nv|ВАЗ-2104/[ВАЗ-2104]noun:inanim:m:v_rod:nv|ВАЗ-2104/[ВАЗ-2104]noun:inanim:m:v_zna:nv|ВАЗ-2104/[ВАЗ-2104]noun:inanim:p:v_dav:nv|ВАЗ-2104/[ВАЗ-2104]noun:inanim:p:v_mis:nv|ВАЗ-2104/[ВАЗ-2104]noun:inanim:p:v_naz:nv|ВАЗ-2104/[ВАЗ-2104]noun:inanim:p:v_oru:nv|ВАЗ-2104/[ВАЗ-2104]noun:inanim:p:v_rod:nv|ВАЗ-2104/[ВАЗ-2104]noun:inanim:p:v_zna:nv"},
	{"Міг-21М", "Міг-21М/[Міг-21М]noun:inanim:m:v_dav:nv|Міг-21М/[Міг-21М]noun:inanim:m:v_mis:nv|Міг-21М/[Міг-21М]noun:inanim:m:v_naz:nv|Міг-21М/[Міг-21М]noun:inanim:m:v_oru:nv|Міг-21М/[Міг-21М]noun:inanim:m:v_rod:nv|Міг-21М/[Міг-21М]noun:inanim:m:v_zna:nv|Міг-21М/[Міг-21М]noun:inanim:p:v_dav:nv|Міг-21М/[Міг-21М]noun:inanim:p:v_mis:nv|Міг-21М/[Міг-21М]noun:inanim:p:v_naz:nv|Міг-21М/[Міг-21М]noun:inanim:p:v_oru:nv|Міг-21М/[Міг-21М]noun:inanim:p:v_rod:nv|Міг-21М/[Міг-21М]noun:inanim:p:v_zna:nv"},
	{"Ан-140-100", "Ан-140-100/[Ан-140-100]noun:inanim:m:v_dav:nv|Ан-140-100/[Ан-140-100]noun:inanim:m:v_mis:nv|Ан-140-100/[Ан-140-100]noun:inanim:m:v_naz:nv|Ан-140-100/[Ан-140-100]noun:inanim:m:v_oru:nv|Ан-140-100/[Ан-140-100]noun:inanim:m:v_rod:nv|Ан-140-100/[Ан-140-100]noun:inanim:m:v_zna:nv|Ан-140-100/[Ан-140-100]noun:inanim:p:v_dav:nv|Ан-140-100/[Ан-140-100]noun:inanim:p:v_mis:nv|Ан-140-100/[Ан-140-100]noun:inanim:p:v_naz:nv|Ан-140-100/[Ан-140-100]noun:inanim:p:v_oru:nv|Ан-140-100/[Ан-140-100]noun:inanim:p:v_rod:nv|Ан-140-100/[Ан-140-100]noun:inanim:p:v_zna:nv"},
	{"топ-10", "топ-10/[топ-10]noninfl:alt"},
	{"Формули-1", "Формули-1/[Формула-1]noun:inanim:f:v_rod:prop|Формули-1/[Формула-1]noun:inanim:p:v_naz:prop|Формули-1/[Формула-1]noun:inanim:p:v_zna:prop"},
	{"Карпати-2", "Карпати-2/[Карпати-2]noun:inanim:p:v_naz:ns:prop:geo|Карпати-2/[Карпати-2]noun:inanim:p:v_zna:ns:prop:geo"},
	{"ТЕЦ-1", "ТЕЦ-1/[ТЕЦ-1]noun:inanim:f:v_dav:nv|ТЕЦ-1/[ТЕЦ-1]noun:inanim:f:v_mis:nv|ТЕЦ-1/[ТЕЦ-1]noun:inanim:f:v_naz:nv|ТЕЦ-1/[ТЕЦ-1]noun:inanim:f:v_oru:nv|ТЕЦ-1/[ТЕЦ-1]noun:inanim:f:v_rod:nv|ТЕЦ-1/[ТЕЦ-1]noun:inanim:f:v_zna:nv"},
	{"омега-3", "омега-3/[омега-3]noninfl"},
	{"потік-2", "потік-2/[потік-2]noun:inanim:m:v_naz|потік-2/[потік-2]noun:inanim:m:v_zna"},
	{"СДПУ(о)", "СДПУ(о)/[СДПУ(о)]noun:inanim:f:v_dav:nv:abbr:prop|СДПУ(о)/[СДПУ(о)]noun:inanim:f:v_mis:nv:abbr:prop|СДПУ(о)/[СДПУ(о)]noun:inanim:f:v_naz:nv:abbr:prop|СДПУ(о)/[СДПУ(о)]noun:inanim:f:v_oru:nv:abbr:prop|СДПУ(о)/[СДПУ(о)]noun:inanim:f:v_rod:nv:abbr:prop|СДПУ(о)/[СДПУ(о)]noun:inanim:f:v_zna:nv:abbr:prop"},
	{"по-свинячому", "по-свинячому/[по-свинячому]adv"},
	{"по-сибірськи", "по-сибірськи/[по-сибірськи]adv"},
	{"по-абхазьки", "по-абхазьки/[по-абхазьки]adv"},
	{"по-авантурницьки", "по-авантурницьки/[по-авантурницьки]adv"},
	{"давай-но", "давай-но/[давати]verb:imperf:impr:s:2"},
	{"дивіться-но", "дивіться-но/[дивитися]verb:rev:imperf:impr:p:2"},
	{"гей-но", "гей-но/[гей]intj"},
	{"цить-но", "цить-но/[цить]intj"},
	{"хай-но", "хай-но/[хай]part"},
	{"бачиш-но", "бачиш-но/[бачити]verb:imperf:pres:s:2:insert"},
	{"той-таки", "той-таки/[той]adj:m:v_naz:pron:dem|той-таки/[той]adj:m:v_zna:rinanim:pron:dem|той-таки/[той]part"},
	{"буде-таки", "буде-таки/[бути]verb:imperf:futr:s:3"},
	{"там-таки", "там-таки/[там]adv:pron:dem|там-таки/[там]part"},
	{"зразу-таки", "зразу-таки/[зразу]adv"},
	{"першого-таки", "першого-таки/[перший]adj:m:v_rod:numr|першого-таки/[перший]adj:m:v_zna:ranim:numr|першого-таки/[перший]adj:n:v_rod:numr"},
	{"варто-таки", "варто-таки/[варто]noninfl:predic"},
	{"оцей-от", "оцей-от/[оцей]adj:m:v_naz:pron:dem|оцей-от/[оцей]adj:m:v_zna:rinanim:pron:dem"},
	{"ану-бо", "ану-бо/[ану]intj|ану-бо/[ану]part"},
	{"годі-бо", "годі-бо/[годі]adv:predic"},
	{"оттакий-то", "оттакий-то/[оттакий]adj:m:v_naz:pron:dem|оттакий-то/[оттакий]adj:m:v_zna:rinanim:pron:dem"},
	{"взагалі-то", "взагалі-то/[взагалі]adv:insert"},
	{"дуже-то", "дуже-то/[дуже]adv:compb|дуже-то/[дужий]adj:n:v_kly:compb|дуже-то/[дужий]adj:n:v_naz:compb|дуже-то/[дужий]adj:n:v_zna:compb"},
	{"десь-то", "десь-то/[десь]adv:pron:ind"},
	{"котрий-то", "котрий-то/[котрий]adj:m:v_naz:pron:int:rel|котрий-то/[котрий]adj:m:v_zna:rinanim:pron:int:rel"},
	{"ніби-то", "ніби-то/[ніби]conj:subord"},
	{"вони-то", "вони-то/[вони]noun:unanim:p:v_naz:pron:pers:3"},
	{"права-то", "права-то/[правий]adj:f:v_kly:compb|права-то/[правий]adj:f:v_naz:compb|права-то/[право]noun:inanim:n:v_rod|права-то/[право]noun:inanim:p:v_kly|права-то/[право]noun:inanim:p:v_naz|права-то/[право]noun:inanim:p:v_zna"},
	{"ти-ж", "ти-ж/[ти]noun:anim:s:v_kly:pron:pers:2:bad|ти-ж/[ти]noun:anim:s:v_naz:pron:pers:2:bad"},
	{"були-б", "були-б/[бути]verb:imperf:past:p:bad"},
	{"був-єм", "був-єм/[бути]verb:imperf:past:m:arch"},
	{"Са-вмісні", "Са-вмісні/[Са-вмісний]adj:p:v_naz|Са-вмісні/[Са-вмісний]adj:p:v_zna:rinanim"},
	{"карбонат-вмісні", "карбонат-вмісні/[карбонат-вмісний]adj:p:v_naz|карбонат-вмісні/[карбонат-вмісний]adj:p:v_zna:rinanim"},
	{"Ш-подібному", "Ш-подібному/[Ш-подібний]adj:m:v_dav|Ш-подібному/[Ш-подібний]adj:m:v_mis|Ш-подібному/[Ш-подібний]adj:n:v_dav|Ш-подібному/[Ш-подібний]adj:n:v_mis"},
	{"S-подібної", "S-подібної/[S-подібний]adj:f:v_rod"},
	{"PR-складову", "PR-складову/[PR-складова]noun:inanim:f:v_zna"},
	{"VIP-батьків", "VIP-батьків/[VIP-батьки]noun:anim:p:v_rod:ns|VIP-батьків/[VIP-батьки]noun:anim:p:v_zna:ns|VIP-батьків/[VIP-батько]noun:anim:p:v_rod|VIP-батьків/[VIP-батько]noun:anim:p:v_zna"},
	{"COVID-вакцинованих", "COVID-вакцинованих/[COVID-вакцинований]adj:p:v_mis:adjp:pasv:imperf|COVID-вакцинованих/[COVID-вакцинований]adj:p:v_mis:adjp:pasv:perf|COVID-вакцинованих/[COVID-вакцинований]adj:p:v_rod:adjp:pasv:imperf|COVID-вакцинованих/[COVID-вакцинований]adj:p:v_rod:adjp:pasv:perf|COVID-вакцинованих/[COVID-вакцинований]adj:p:v_zna:ranim:adjp:pasv:imperf|COVID-вакцинованих/[COVID-вакцинований]adj:p:v_zna:ranim:adjp:pasv:perf"},
	{"VIP–будинок", "VIP-будинок/[VIP-будинок]noun:inanim:m:v_naz|VIP-будинок/[VIP-будинок]noun:inanim:m:v_zna|VIP–будинок/[null]null"},
	{"PR-департаменту", "PR-департаменту/[PR-департамент]noun:inanim:m:v_dav|PR-департаменту/[PR-департамент]noun:inanim:m:v_mis|PR-департаменту/[PR-департамент]noun:inanim:m:v_rod"},
	{"RPM-пакунок", "RPM-пакунок/[RPM-пакунок]noun:inanim:m:v_naz|RPM-пакунок/[RPM-пакунок]noun:inanim:m:v_zna"},
	{"3D-друк", "3D-друк/[3D-друк]noun:inanim:m:v_naz|3D-друк/[3D-друк]noun:inanim:m:v_zna"},
	{"n-векторний", "n-векторний/[n-векторний]adj:m:v_naz|n-векторний/[n-векторний]adj:m:v_zna:rinanim"},
	{"α-векторний", "α-векторний/[α-векторний]adj:m:v_naz|α-векторний/[α-векторний]adj:m:v_zna:rinanim"},
	{"α-діапазон", "α-діапазон/[α-діапазон]noun:inanim:m:v_naz|α-діапазон/[α-діапазон]noun:inanim:m:v_zna"},
	{"топ-десять", "топ-десять/[топ-десять]numr:p:v_naz:bad|топ-десять/[топ-десять]numr:p:v_zna:bad"},
	{"Мустафа-ага", "Мустафа-ага/[Мустафа-ага]noun:anim:m:v_naz:prop:fname"},
	{"екс-«депутат»", "екс-«депутат»/[екс-депутат]noun:anim:m:v_naz:alt"},
	{"\"200\"-х", "\"200\"-х/[200-й]adj:p:v_mis:numr|\"200\"-х/[200-й]adj:p:v_rod:numr|\"200\"-х/[200-й]adj:p:v_zna:ranim:numr"},
	{"\"заступницю\"-колаборантку", "\"заступницю\"-колаборантку/[заступниця-колаборантка]noun:anim:f:v_zna"},
	{"псевдо-ФОП", "псевдо-ФОП/[псевдо-ФОП]noun:anim:f:v_dav:nv:abbr|псевдо-ФОП/[псевдо-ФОП]noun:anim:f:v_mis:nv:abbr|псевдо-ФОП/[псевдо-ФОП]noun:anim:f:v_naz:nv:abbr|псевдо-ФОП/[псевдо-ФОП]noun:anim:f:v_oru:nv:abbr|псевдо-ФОП/[псевдо-ФОП]noun:anim:f:v_rod:nv:abbr|псевдо-ФОП/[псевдо-ФОП]noun:anim:f:v_zna:nv:abbr|псевдо-ФОП/[псевдо-ФОП]noun:anim:m:v_naz:abbr|псевдо-ФОП/[псевдо-ФОП]noun:anim:p:v_dav:nv:abbr|псевдо-ФОП/[псевдо-ФОП]noun:anim:p:v_mis:nv:abbr|псевдо-ФОП/[псевдо-ФОП]noun:anim:p:v_naz:nv:abbr|псевдо-ФОП/[псевдо-ФОП]noun:anim:p:v_oru:nv:abbr|псевдо-ФОП/[псевдо-ФОП]noun:anim:p:v_rod:nv:abbr|псевдо-ФОП/[псевдо-ФОП]noun:anim:p:v_zna:nv:abbr"},
	{"пів-України", "пів-України/[пів-України]noun:inanim:p:v_dav:prop:geo:nv:alt|пів-України/[пів-України]noun:inanim:p:v_mis:prop:geo:nv:alt|пів-України/[пів-України]noun:inanim:p:v_naz:prop:geo:nv:alt|пів-України/[пів-України]noun:inanim:p:v_oru:prop:geo:nv:alt|пів-України/[пів-України]noun:inanim:p:v_rod:prop:geo:nv:alt|пів-України/[пів-України]noun:inanim:p:v_zna:prop:geo:nv:alt"},
	{"Пенсильванія-авеню", "Пенсильванія-авеню/[Пенсильванія-авеню]noun:inanim:f:v_dav:nv:prop|Пенсильванія-авеню/[Пенсильванія-авеню]noun:inanim:f:v_mis:nv:prop|Пенсильванія-авеню/[Пенсильванія-авеню]noun:inanim:f:v_naz:nv:prop|Пенсильванія-авеню/[Пенсильванія-авеню]noun:inanim:f:v_oru:nv:prop|Пенсильванія-авеню/[Пенсильванія-авеню]noun:inanim:f:v_rod:nv:prop|Пенсильванія-авеню/[Пенсильванія-авеню]noun:inanim:f:v_zna:nv:prop"},
	{"Уолл-стрит", "Уолл-стрит/[Уолл-стрит]noun:inanim:f:v_dav:nv:prop|Уолл-стрит/[Уолл-стрит]noun:inanim:f:v_mis:nv:prop|Уолл-стрит/[Уолл-стрит]noun:inanim:f:v_naz:nv:prop|Уолл-стрит/[Уолл-стрит]noun:inanim:f:v_oru:nv:prop|Уолл-стрит/[Уолл-стрит]noun:inanim:f:v_rod:nv:prop|Уолл-стрит/[Уолл-стрит]noun:inanim:f:v_zna:nv:prop"},
	{"пів-години", "пів-години/[пів-години]noun:inanim:p:v_dav:nv:bad|пів-години/[пів-години]noun:inanim:p:v_mis:nv:bad|пів-години/[пів-години]noun:inanim:p:v_naz:nv:bad|пів-години/[пів-години]noun:inanim:p:v_oru:nv:bad|пів-години/[пів-години]noun:inanim:p:v_rod:nv:bad|пів-години/[пів-години]noun:inanim:p:v_zna:nv:bad"},
	{"зїзду", "зїзду/[з'їзд]noun:inanim:m:v_dav:bad|зїзду/[з'їзд]noun:inanim:m:v_mis:bad|зїзду/[з'їзд]noun:inanim:m:v_rod:bad"},
	{"обвязаний", "обвязаний/[обв'язаний]adj:m:v_kly:adjp:pasv:perf:bad|обвязаний/[обв'язаний]adj:m:v_naz:adjp:pasv:perf:bad|обвязаний/[обв'язаний]adj:m:v_zna:rinanim:adjp:pasv:perf:bad"},
	{"время", "время/[null]null"},
	{"якогонебудь", "якогонебудь/[який-небудь]adj:m:v_rod:pron:int:rel:def:bad|якогонебудь/[який-небудь]adj:m:v_zna:ranim:pron:int:rel:def:bad|якогонебудь/[який-небудь]adj:n:v_rod:pron:int:rel:def:bad"},
	{"болнебудь", "болнебудь/[null]null"},
	{"півгодини-годину", "півгодини-годину/[півгодини-годину]noun:inanim:p:v_zna"},
	{"жило-було", "жило-було/[жити-бути]verb:imperf:past:n"},
	{"учиш-учиш", "учиш-учиш/[учити-учити]verb:imperf:pres:s:2"},
	{"а-а", "а-а/[а-а]intj"},
	{"Га-га", "Га-га/[га-га]intj"},
	{"кусь-кусь", "кусь-кусь/[кусь-кусь]noninfl:predic"},
	{"ось–ось", "ось-ось/[ось-ось]adv|ось–ось/[null]null"},
	{"ось‑ось", "ось-ось/[ось-ось]adv"},
	{"Івано–Франківськ", "Івано-Франківськ/[Івано-Франківськ]noun:inanim:m:v_naz:prop:geo|Івано-Франківськ/[Івано-Франківськ]noun:inanim:m:v_zna:prop:geo|Івано–Франківськ/[null]null"},
	{"вгору-вниз", "вгору-вниз/[вгору-вниз]adv"},
	{"Усе-усе", "Усе-усе/[увесь-увесь]adj:n:v_naz:pron:gen|Усе-усе/[увесь-увесь]adj:n:v_zna:pron:gen|Усе-усе/[усе-усе]adv|Усе-усе/[усе-усе]noun:inanim:n:v_naz:pron:gen|Усе-усе/[усе-усе]noun:inanim:n:v_zna:pron:gen"},
	{"всього-всього", "всього-всього/[ввесь-ввесь]adj:m:v_rod:pron:gen|всього-всього/[ввесь-ввесь]adj:m:v_zna:ranim:pron:gen|всього-всього/[ввесь-ввесь]adj:n:v_rod:pron:gen|всього-всього/[весь-весь]adj:m:v_rod:pron:gen|всього-всього/[весь-весь]adj:m:v_zna:ranim:pron:gen|всього-всього/[весь-весь]adj:n:v_rod:pron:gen|всього-всього/[все-все]noun:inanim:n:v_rod:pron:gen|всього-всього/[всього-всього]adv"},
	{"Ка-КА", "Ка-КА/[null]null"},
	{"годину-півтори", "годину-півтори/[година-півтори]noun:inanim:f:v_zna|годину-півтори/[година-півтори]noun:inanim:p:v_zna"},
	{"вівторок-середа", "вівторок-середа/[вівторок-середа]noun:inanim:m:v_naz|вівторок-середа/[вівторок-середа]noun:inanim:p:v_naz"},
	{"лютого-березня", "лютого-березня/[лютий-березень]noun:inanim:m:v_rod|лютого-березня/[лютий-березень]noun:inanim:p:v_rod"},
	{"низенько-низенько", "низенько-низенько/[низенько-низенько]adv"},
	{"неприховано-вороже", "неприховано-вороже/[неприховано-вороже]adv"},
	{"такого-сякого", "такого-сякого/[такий-сякий]adj:m:v_rod:pron:def|такого-сякого/[такий-сякий]adj:m:v_zna:ranim:pron:def|такого-сякого/[такий-сякий]adj:n:v_rod:pron:def"},
	{"чорній-чорній", "чорній-чорній/[чорна-чорна]noun:anim:f:v_dav|чорній-чорній/[чорна-чорна]noun:anim:f:v_mis|чорній-чорній/[чорний-чорний]adj:f:v_dav|чорній-чорній/[чорний-чорний]adj:f:v_mis|чорній-чорній/[чорніти-чорніти]verb:imperf:impr:s:2"},
	{"писаного-переписаного", "писаного-переписаного/[писаний-переписаний]adj:m:v_rod:adjp:pasv:imperf|писаного-переписаного/[писаний-переписаний]adj:m:v_zna:ranim:adjp:pasv:imperf|писаного-переписаного/[писаний-переписаний]adj:n:v_rod:adjp:pasv:imperf"},
	{"минулого–позаминулого", "минулого-позаминулого/[минулий-позаминулий]adj:m:v_rod|минулого-позаминулого/[минулий-позаминулий]adj:m:v_zna:ranim|минулого-позаминулого/[минулий-позаминулий]adj:n:v_rod|минулого–позаминулого/[null]null"},
	{"дівчинка-дівчинка", "дівчинка-дівчинка/[дівчинка-дівчинка]noun:anim:f:v_naz"},
	{"лікар-гомеопат", "лікар-гомеопат/[лікар-гомеопат]noun:anim:m:v_naz"},
	{"лікаря-гомеопата", "лікаря-гомеопата/[лікар-гомеопат]noun:anim:m:v_rod|лікаря-гомеопата/[лікар-гомеопат]noun:anim:m:v_zna"},
	{"шмкр-гомеопат", "шмкр-гомеопат/[null]null"},
	{"лікар-ткр", "лікар-ткр/[null]null"},
	{"Жінка-Актриса", "Жінка-Актриса/[жінка-актриса]noun:anim:f:v_naz"},
	{"пане-товаришу", "пане-товаришу/[пан-товариш]noun:anim:m:v_kly|пане-товаришу/[паня-товариш]noun:anim:f:v_kly"},
	{"вчинок-приклад", "вчинок-приклад/[вчинок-приклад]noun:inanim:m:v_naz|вчинок-приклад/[вчинок-приклад]noun:inanim:m:v_zna"},
	{"міста-фортеці", "міста-фортеці/[місто-фортеця]noun:inanim:n:v_rod|міста-фортеці/[місто-фортеця]noun:inanim:p:v_naz|міста-фортеці/[місто-фортеця]noun:inanim:p:v_zna"},
	{"Інтернет-пошуковик", "Інтернет-пошуковик/[інтернет-пошуковик]noun:inanim:m:v_naz|Інтернет-пошуковик/[інтернет-пошуковик]noun:inanim:m:v_zna"},
	{"ворог-стафілокок", "ворог-стафілокок/[null]null"},
	{"стафілокок-реагент", "стафілокок-реагент/[null]null"},
	{"Каладжі-бей", "Каладжі-бей/[Каладжі-бей]noun:anim:m:v_naz:prop:fname"},
	{"Гасан-огли́", "Гасан-огли/[Гасан-огли]noun:anim:m:v_naz:prop:fname"},
	{"Москви-ріки", "Москви-ріки/[Москва-ріка]noun:inanim:f:v_rod:prop:geo|Москви-ріки/[москва-ріка]noun:inanim:f:v_rod:prop:geo:alt"},
	{"Кочубієва-Одеси", "Кочубієва-Одеси/[Кочубіїв-Одеса]noun:inanim:m:v_rod:prop:geo"},
	{"Дівчинка-першокласниця", "Дівчинка-першокласниця/[дівчинка-першокласниця]noun:anim:f:v_naz"},
	{"ненависть-шоу", "ненависть-шоу/[ненависть-шоу]noun:inanim:f:v_naz|ненависть-шоу/[ненависть-шоу]noun:inanim:f:v_zna"},
	{"Київ-Прага", "Київ-Прага/[Київ-Прага]noninfl:prop:geo"},
	{"Україна-ЄС", "Україна-ЄС/[Україна-ЄС]noninfl:prop"},
	{"Хуана-Карлоса", "Хуана-Карлоса/[Хуан-Карлос]noun:anim:m:v_rod:prop:fname|Хуана-Карлоса/[Хуан-Карлос]noun:anim:m:v_zna:prop:fname"},
	{"Карпа-Хансен", "Карпа-Хансен/[Карпа-Хансен]noninfl:prop:lname"},
	{"Джолі-Пітт", "Джолі-Пітт/[Джолі-Пітт]noninfl:prop:lname"},
	{"Джеймса-Веніка", "Джеймса-Веніка/[Джеймса-Веніка]noninfl:prop:lname"},
	{"БАЛОГИ-КОЛЕСНІКОВА", "БАЛОГИ-КОЛЕСНІКОВА/[Балоги-Колеснікова]noninfl:prop:lname"},
	{"Івана-Павла", "Івана-Павла/[Іван-Павло]noun:anim:m:v_rod:prop:fname|Івана-Павла/[Іван-Павло]noun:anim:m:v_zna:prop:fname"},
	{"Гей-гей-гей", "Гей-гей-гей/[гей-гей-гей]intj"},
	{"Ого-го-го-го", "Ого-го-го-го/[ого-го-го-го]intj"},
	{"фу-фу", "фу-фу/[фу-фу]intj"},
	{"Соціал-Демократичний", "Соціал-Демократичний/[соціал-демократичний]adj:m:v_kly|Соціал-Демократичний/[соціал-демократичний]adj:m:v_naz|Соціал-Демократичний/[соціал-демократичний]adj:m:v_zna:rinanim"},
	{"Івано-Франківська", "Івано-Франківська/[Івано-Франківськ]noun:inanim:m:v_rod:prop:geo|Івано-Франківська/[івано-франківський]adj:f:v_kly|Івано-Франківська/[івано-франківський]adj:f:v_naz"},
	{"Івано-Франківської", "Івано-Франківської/[івано-франківський]adj:f:v_rod"},
	{"Переяслав-Хмельницького", "Переяслав-Хмельницького/[переяслав-хмельницький]adj:m:v_rod|Переяслав-Хмельницького/[переяслав-хмельницький]adj:m:v_zna:ranim|Переяслав-Хмельницького/[переяслав-хмельницький]adj:n:v_rod"},
	{"Південно-Західній", "Південно-Західній/[південно-західний]adj:f:v_dav|Південно-Західній/[південно-західний]adj:f:v_mis|Південно-Західній/[південно-західній]adj:f:v_dav:bad:err|Південно-Західній/[південно-західній]adj:f:v_mis:bad:err|Південно-Західній/[південно-західній]adj:m:v_kly:bad:err|Південно-Західній/[південно-західній]adj:m:v_naz:bad:err|Південно-Західній/[південно-західній]adj:m:v_zna:rinanim:bad:err"},
	{"Південно-Східної", "Південно-Східної/[південно-східний]adj:f:v_rod"},
	{"аль-Каїда", "Аль-Каїда/[Аль-Каїда]noun:inanim:f:v_naz:prop:bad"},
	{"екс-партнер", "екс-партнер/[екс-партнер]noun:anim:m:v_naz:alt"},
	{"еспресо-машина", "еспресо-машина/[еспресо-машина]noun:inanim:f:v_naz"},
	{"міні-БПЛА", "міні-БПЛА/[міні-БПЛА]noun:inanim:m:v_dav:nv:abbr|міні-БПЛА/[міні-БПЛА]noun:inanim:m:v_mis:nv:abbr|міні-БПЛА/[міні-БПЛА]noun:inanim:m:v_naz:nv:abbr|міні-БПЛА/[міні-БПЛА]noun:inanim:m:v_oru:nv:abbr|міні-БПЛА/[міні-БПЛА]noun:inanim:m:v_rod:nv:abbr|міні-БПЛА/[міні-БПЛА]noun:inanim:m:v_zna:nv:abbr|міні-БПЛА/[міні-БПЛА]noun:inanim:p:v_dav:nv:abbr|міні-БПЛА/[міні-БПЛА]noun:inanim:p:v_mis:nv:abbr|міні-БПЛА/[міні-БПЛА]noun:inanim:p:v_naz:nv:abbr|міні-БПЛА/[міні-БПЛА]noun:inanim:p:v_oru:nv:abbr|міні-БПЛА/[міні-БПЛА]noun:inanim:p:v_rod:nv:abbr|міні-БПЛА/[міні-БПЛА]noun:inanim:p:v_zna:nv:abbr"},
	{"супер-Маріо", "супер-Маріо/[супер-Маріо]noun:anim:m:v_dav:nv:prop:fname|супер-Маріо/[супер-Маріо]noun:anim:m:v_kly:nv:prop:fname|супер-Маріо/[супер-Маріо]noun:anim:m:v_mis:nv:prop:fname|супер-Маріо/[супер-Маріо]noun:anim:m:v_naz:nv:prop:fname|супер-Маріо/[супер-Маріо]noun:anim:m:v_oru:nv:prop:fname|супер-Маріо/[супер-Маріо]noun:anim:m:v_rod:nv:prop:fname|супер-Маріо/[супер-Маріо]noun:anim:m:v_zna:nv:prop:fname"},
	{"кава-еспресо", "кава-еспресо/[кава-еспресо]noun:inanim:f:v_naz"},
	{"кави-еспресо", "кави-еспресо/[кава-еспресо]noun:inanim:f:v_rod|кави-еспресо/[кава-еспресо]noun:inanim:p:v_naz|кави-еспресо/[кава-еспресо]noun:inanim:p:v_zna"},
	{"програмою-максимум", "програмою-максимум/[програма-максимум]noun:inanim:f:v_oru"},
	{"чарку-другу", "чарку-другу/[чарка-друга]noun:inanim:f:v_zna"},
	{"Алієва-старшого", "Алієва-старшого/[Алієв-старший]noun:anim:m:v_rod:prop:lname|Алієва-старшого/[Алієв-старший]noun:anim:m:v_zna:prop:lname"},
	// testDynamicTaggingInvalidLeft assertTagged (Java UkrainianTaggerTest; not-tagged twin: javaUkrainianTaggerNotTagged)
	{"авіа-переліт", "авіа-переліт/[авіа-переліт]noun:inanim:m:v_naz:bad|авіа-переліт/[авіа-переліт]noun:inanim:m:v_zna:bad"},
	{"авто-салон", "авто-салон/[авто-салон]noun:inanim:m:v_naz:bad|авто-салон/[авто-салон]noun:inanim:m:v_zna:bad"},
	{"квазі-держави", "квазі-держави/[квазі-держава]noun:inanim:f:v_rod:bad|квазі-держави/[квазі-держава]noun:inanim:p:v_kly:bad|квазі-держави/[квазі-держава]noun:inanim:p:v_naz:bad|квазі-держави/[квазі-держава]noun:inanim:p:v_zna:bad"},
	{"мульти-візу", "мульти-візу/[мульти-віза]noun:inanim:f:v_zna:bad"},
	{"контр-міри", "контр-міри/[контр-міра]noun:inanim:f:v_rod:bad|контр-міри/[контр-міра]noun:inanim:p:v_kly:bad|контр-міри/[контр-міра]noun:inanim:p:v_naz:bad|контр-міри/[контр-міра]noun:inanim:p:v_zna:bad"},
	{"кіно-критика", "кіно-критика/[кіно-критик]noun:anim:m:v_rod:bad|кіно-критика/[кіно-критик]noun:anim:m:v_zna:bad|кіно-критика/[кіно-критика]noun:inanim:f:v_naz:bad"},
	{"древньо-римський", "древньо-римський/[древньо-римський]adj:m:v_kly:bad|древньо-римський/[древньо-римський]adj:m:v_naz:bad|древньо-римський/[древньо-римський]adj:m:v_zna:rinanim:bad"},
	{"давньо-римський", "давньо-римський/[давньо-римський]adj:m:v_kly:bad|давньо-римський/[давньо-римський]adj:m:v_naz:bad|давньо-римський/[давньо-римський]adj:m:v_zna:rinanim:bad"},
	{"максимально-можливу", "максимально-можливу/[максимально-можливий]adj:f:v_zna:bad"},
	{"загально-правовий", "загально-правовий/[загально-правовий]adj:m:v_kly:bad|загально-правовий/[загально-правовий]adj:m:v_naz:bad|загально-правовий/[загально-правовий]adj:m:v_zna:rinanim:bad"},
	{"відео-навчання", "відео-навчання/[відео-навчання]noun:inanim:n:v_kly:bad|відео-навчання/[відео-навчання]noun:inanim:n:v_naz:bad|відео-навчання/[відео-навчання]noun:inanim:n:v_rod:bad|відео-навчання/[відео-навчання]noun:inanim:n:v_zna:bad|відео-навчання/[відео-навчання]noun:inanim:p:v_kly:bad|відео-навчання/[відео-навчання]noun:inanim:p:v_naz:bad|відео-навчання/[відео-навчання]noun:inanim:p:v_zna:bad"},
	{"В2В-рішень", "В2В-рішень/[В2В-рішення]noun:inanim:p:v_rod:bad"},
	{"кіно-Європа", "кіно-Європа/[кіно-Європа]noun:inanim:f:v_naz:prop:geo"},
	{"напів-японка", "напів-японка/[напів-японка]noun:anim:f:v_naz:bad"},
	{"напів-Європа", "напів-Європа/[напів-Європа]noun:inanim:f:v_naz:prop:geo"},
	{"напів'японка", "напів'японка/[напів'японка]noun:anim:f:v_naz"},
	{"напів–фантастичних", "напів-фантастичних/[напів-фантастичний]adj:p:v_mis:bad|напів-фантастичних/[напів-фантастичний]adj:p:v_rod:bad|напів-фантастичних/[напів-фантастичний]adj:p:v_zna:ranim:bad|напів–фантастичних/[null]null"},
	{"напівпольської-напіванглійської", "напівпольської-напіванглійської/[напівпольська-напіванглійська]noun:inanim:f:v_rod|напівпольської-напіванглійської/[напівпольський-напіванглійський]adj:f:v_rod"},
	{"напівбілоруски-напівукраїнки", "напівбілоруски-напівукраїнки/[напівбілоруска-напівукраїнка]noun:anim:f:v_rod|напівбілоруски-напівукраїнки/[напівбілоруска-напівукраїнка]noun:anim:p:v_kly|напівбілоруски-напівукраїнки/[напівбілоруска-напівукраїнка]noun:anim:p:v_naz|напівбілоруски-напівукраїнки/[напівбілоруска-напівукраїнка]noun:anim:p:v_zna:rare"},
	{"Лангштрассе", "Лангштрассе/[Лангштрассе]noun:inanim:f:v_dav:nv:prop:alt|Лангштрассе/[Лангштрассе]noun:inanim:f:v_mis:nv:prop:alt|Лангштрассе/[Лангштрассе]noun:inanim:f:v_naz:nv:prop:alt|Лангштрассе/[Лангштрассе]noun:inanim:f:v_oru:nv:prop:alt|Лангштрассе/[Лангштрассе]noun:inanim:f:v_rod:nv:prop:alt|Лангштрассе/[Лангштрассе]noun:inanim:f:v_zna:nv:prop:alt"},
	{"Єнукідзе", "Єнукідзе/[Єнукідзе]noun:inanim:f:v_dav:nv:prop:lname|Єнукідзе/[Єнукідзе]noun:inanim:f:v_mis:nv:prop:lname|Єнукідзе/[Єнукідзе]noun:inanim:f:v_naz:nv:prop:lname|Єнукідзе/[Єнукідзе]noun:inanim:f:v_oru:nv:prop:lname|Єнукідзе/[Єнукідзе]noun:inanim:f:v_rod:nv:prop:lname|Єнукідзе/[Єнукідзе]noun:inanim:f:v_zna:nv:prop:lname|Єнукідзе/[Єнукідзе]noun:inanim:m:v_dav:nv:prop:lname|Єнукідзе/[Єнукідзе]noun:inanim:m:v_mis:nv:prop:lname|Єнукідзе/[Єнукідзе]noun:inanim:m:v_naz:nv:prop:lname|Єнукідзе/[Єнукідзе]noun:inanim:m:v_oru:nv:prop:lname|Єнукідзе/[Єнукідзе]noun:inanim:m:v_rod:nv:prop:lname|Єнукідзе/[Єнукідзе]noun:inanim:m:v_zna:nv:prop:lname"},
	{"антиаденовірусну", "антиаденовірусну/[антиаденовірусний]adj:f:v_zna"},
	{"транс'ядерний", "транс'ядерний/[транс'ядерний]adj:m:v_kly|транс'ядерний/[транс'ядерний]adj:m:v_naz|транс'ядерний/[транс'ядерний]adj:m:v_zna:rinanim"},
	{"напівядерний", "напівядерний/[напівядерний]adj:m:v_kly:bad|напівядерний/[напівядерний]adj:m:v_naz:bad|напівядерний/[напівядерний]adj:m:v_zna:rinanim:bad"},
	{"ексюрист", "ексюрист/[ексюрист]noun:anim:m:v_naz:bad"},
	{"екс'прес", "екс'прес/[null]null"},
	{"напівяпонка", "напівяпонка/[напів'японка]noun:anim:f:v_naz:bad"},
	{"Напівсправедливий", "Напівсправедливий/[напівсправедливий]adj:m:v_kly|Напівсправедливий/[напівсправедливий]adj:m:v_naz|Напівсправедливий/[напівсправедливий]adj:m:v_zna:rinanim"},
	{"напіврозслабленого", "напіврозслабленого/[напіврозслаблений]adj:m:v_rod|напіврозслабленого/[напіврозслаблений]adj:m:v_zna:ranim|напіврозслабленого/[напіврозслаблений]adj:n:v_rod"},
	{"північноазіатський", "північноазіатський/[північноазіатський]adj:m:v_kly|північноазіатський/[північноазіатський]adj:m:v_naz|північноазіатський/[північноазіатський]adj:m:v_zna:rinanim"},
	{"гірко-прегірко", "гірко-прегірко/[гірко-прегірко]adv"},
	{"великий-превеликий", "великий-превеликий/[великий-превеликий]adj:m:v_kly|великий-превеликий/[великий-превеликий]adj:m:v_naz|великий-превеликий/[великий-превеликий]adj:m:v_zna:rinanim"},
	{"круглого-прекруглого", "круглого-прекруглого/[круглий-прекруглий]adj:m:v_rod|круглого-прекруглого/[круглий-прекруглий]adj:m:v_zna:ranim|круглого-прекруглого/[круглий-прекруглий]adj:n:v_rod"},
	{"патолого-анатомічний", "патолого-анатомічний/[патолого-анатомічний]adj:m:v_kly|патолого-анатомічний/[патолого-анатомічний]adj:m:v_naz|патолого-анатомічний/[патолого-анатомічний]adj:m:v_zna:rinanim"},
	{"дво-триметровий", "дво-триметровий/[дво-триметровий]adj:m:v_kly|дво-триметровий/[дво-триметровий]adj:m:v_naz|дво-триметровий/[дво-триметровий]adj:m:v_zna:rinanim"},
	{"три-чотириметровий", "три-чотириметровий/[три-чотириметровий]adj:m:v_kly|три-чотириметровий/[три-чотириметровий]adj:m:v_naz|три-чотириметровий/[три-чотириметровий]adj:m:v_zna:rinanim"},
	{"три-річний", "три-річний/[три-річний]adj:m:v_kly:bad|три-річний/[три-річний]adj:m:v_naz:bad|три-річний/[три-річний]adj:m:v_zna:rinanim:bad"},
	{"п'яти-зірковий", "п'яти-зірковий/[п'яти-зірковий]adj:m:v_kly:bad|п'яти-зірковий/[п'яти-зірковий]adj:m:v_naz:bad|п'яти-зірковий/[п'яти-зірковий]adj:m:v_zna:rinanim:bad"},
	{"Художньо-культурний", "Художньо-культурний/[художньо-культурний]adj:m:v_kly|Художньо-культурний/[художньо-культурний]adj:m:v_naz|Художньо-культурний/[художньо-культурний]adj:m:v_zna:rinanim"},
	{"сліпуче-яскравого", "сліпуче-яскравого/[сліпуче-яскравий]adj:m:v_rod|сліпуче-яскравого/[сліпуче-яскравий]adj:m:v_zna:ranim|сліпуче-яскравого/[сліпуче-яскравий]adj:n:v_rod"},
	{"україно-болгарський", "україно-болгарський/[україно-болгарський]adj:m:v_kly|україно-болгарський/[україно-болгарський]adj:m:v_naz|україно-болгарський/[україно-болгарський]adj:m:v_zna:rinanim"},
	{"учбово-спортивної", "учбово-спортивної/[учбово-спортивний]adj:f:v_rod:bad"},
	{"етико-філологічний", "етико-філологічний/[етико-філологічний]adj:m:v_kly|етико-філологічний/[етико-філологічний]adj:m:v_naz|етико-філологічний/[етико-філологічний]adj:m:v_zna:rinanim"},
	{"Альпійсько-Карпатського", "Альпійсько-Карпатського/[альпійсько-карпатський]adj:m:v_rod|Альпійсько-Карпатського/[альпійсько-карпатський]adj:m:v_zna:ranim|Альпійсько-Карпатського/[альпійсько-карпатський]adj:n:v_rod"},
	{"Азово-Чорноморського", "Азово-Чорноморського/[азово-чорноморський]adj:m:v_rod|Азово-Чорноморського/[азово-чорноморський]adj:m:v_zna:ranim|Азово-Чорноморського/[азово-чорноморський]adj:n:v_rod"},
	{"Західно-Удмуртської", "Західно-Удмуртської/[західно-удмуртський]adj:f:v_rod:bad"},
	{"Кримсько-Татарський", "Кримсько-Татарський/[кримсько-татарський]adj:m:v_kly|Кримсько-Татарський/[кримсько-татарський]adj:m:v_kly:bad:err|Кримсько-Татарський/[кримсько-татарський]adj:m:v_naz|Кримсько-Татарський/[кримсько-татарський]adj:m:v_naz:bad:err|Кримсько-Татарський/[кримсько-татарський]adj:m:v_zna:rinanim|Кримсько-Татарський/[кримсько-татарський]adj:m:v_zna:rinanim:bad:err"},
	{"внутрішньо-економічної", "внутрішньо-економічної/[внутрішньо-економічний]adj:f:v_rod:bad:err"},
	{"всесвітньо-відомої", "всесвітньо-відомої/[всесвітньо-відомий]adj:f:v_rod:bad"},
	{"Червоно-градського", "Червоно-градського/[червоно-градський]adj:m:v_rod:bad|Червоно-градського/[червоно-градський]adj:m:v_zna:ranim:bad|Червоно-градського/[червоно-градський]adj:n:v_rod:bad"},
	{"важко-напрацьований", "важко-напрацьований/[важко-напрацьований]adj:m:v_kly:adjp:pasv:perf:bad|важко-напрацьований/[важко-напрацьований]adj:m:v_naz:adjp:pasv:perf:bad|важко-напрацьований/[важко-напрацьований]adj:m:v_zna:rinanim:adjp:pasv:perf:bad"},
	{"вчених-новаторів", "вчених-новаторів/[вчений-новатор]noun:anim:p:v_rod|вчених-новаторів/[вчений-новатор]noun:anim:p:v_zna"},
	{"країна-виробник", "країна-виробник/[країна-виробник]noun:inanim:f:v_naz"},
	{"банк-виробник", "банк-виробник/[банк-виробник]noun:inanim:m:v_naz|банк-виробник/[банк-виробник]noun:inanim:m:v_zna"},
	{"банки-агенти", "банки-агенти/[банк-агент]noun:inanim:p:v_naz|банки-агенти/[банк-агент]noun:inanim:p:v_zna|банки-агенти/[банка-агент]noun:inanim:p:v_naz|банки-агенти/[банка-агент]noun:inanim:p:v_zna"},
	{"місто-гігант", "місто-гігант/[місто-гігант]noun:inanim:n:v_naz|місто-гігант/[місто-гігант]noun:inanim:n:v_zna"},
	{"країни-агресори", "країни-агресори/[країна-агресор]noun:inanim:p:v_naz|країни-агресори/[країна-агресор]noun:inanim:p:v_zna"},
	{"поселення-гігант", "поселення-гігант/[поселення-гігант]noun:inanim:n:v_naz|поселення-гігант/[поселення-гігант]noun:inanim:n:v_zna"},
	{"сонях-красень", "сонях-красень/[сонях-красень]noun:inanim:m:v_naz|сонях-красень/[сонях-красень]noun:inanim:m:v_zna"},
	{"красень-сонях", "красень-сонях/[красень-сонях]noun:inanim:m:v_naz|красень-сонях/[красень-сонях]noun:inanim:m:v_zna"},
	{"депутатів-привидів", "депутатів-привидів/[депутат-привид]noun:anim:p:v_rod|депутатів-привидів/[депутат-привид]noun:anim:p:v_zna"},
	{"україно-румуно-болгарський", "україно-румуно-болгарський/[україно-румуно-болгарський]adj:m:v_kly|україно-румуно-болгарський/[україно-румуно-болгарський]adj:m:v_naz|україно-румуно-болгарський/[україно-румуно-болгарський]adj:m:v_zna:rinanim"},
	{"синьо-біло-чорному", "синьо-біло-чорному/[синьо-біло-чорний]adj:m:v_dav|синьо-біло-чорному/[синьо-біло-чорний]adj:m:v_mis|синьо-біло-чорному/[синьо-біло-чорний]adj:n:v_dav|синьо-біло-чорному/[синьо-біло-чорний]adj:n:v_mis"},
	{"синьо-біло-жовтий", "синьо-біло-жовтий/[синьо-біло-жовтий]adj:m:v_kly|синьо-біло-жовтий/[синьо-біло-жовтий]adj:m:v_naz|синьо-біло-жовтий/[синьо-біло-жовтий]adj:m:v_zna:rinanim"},
	{"українсько-англійсько-французьким", "українсько-англійсько-французьким/[українсько-англійсько-французький]adj:m:v_oru|українсько-англійсько-французьким/[українсько-англійсько-французький]adj:n:v_oru|українсько-англійсько-французьким/[українсько-англійсько-французький]adj:p:v_dav"},
	{"седативного-снодійного-антигістамінного", "седативного-снодійного-антигістамінного/[снодійний-антигістамінний]adj:m:v_rod|седативного-снодійного-антигістамінного/[снодійний-антигістамінний]adj:m:v_zna:ranim|седативного-снодійного-антигістамінного/[снодійний-антигістамінний]adj:n:v_rod"},
	{"хіп-хоп-гуртом", "хіп-хоп-гуртом/[хіп-хоп-гурт]noun:inanim:m:v_oru:xp1|хіп-хоп-гуртом/[хіп-хоп-гурт]noun:inanim:m:v_oru:xp2"},
	{"он-лайн-навчання", "он-лайн-навчання/[он-лайн-навчання]noun:inanim:n:v_naz:bad|он-лайн-навчання/[он-лайн-навчання]noun:inanim:n:v_rod:bad|он-лайн-навчання/[он-лайн-навчання]noun:inanim:n:v_zna:bad|он-лайн-навчання/[он-лайн-навчання]noun:inanim:p:v_naz:bad|он-лайн-навчання/[он-лайн-навчання]noun:inanim:p:v_zna:bad"},
	{"Та-а-ак", "Та-а-ак/[так]adv:pron:dem:alt|Та-а-ак/[так]conj:coord:alt|Та-а-ак/[так]conj:subord:alt|Та-а-ак/[так]part:alt"},
	{"Му-у-у", "Му-у-у/[му]noninfl:onomat:predic:alt"},
	{"С-с-с-лава", "С-с-с-лава/[Слава]noun:anim:f:v_naz:prop:fname:alt|С-с-с-лава/[Слава]noun:anim:m:v_naz:prop:fname:alt|С-с-с-лава/[Слава]noun:inanim:f:v_naz:prop:geo:alt|С-с-с-лава/[слава]noun:inanim:f:v_naz:alt"},
	{"ва-ре-ни-ки", "ва-ре-ни-ки/[вареник]noun:inanim:p:v_kly:alt|ва-ре-ни-ки/[вареник]noun:inanim:p:v_naz:alt|ва-ре-ни-ки/[вареник]noun:inanim:p:v_zna:alt"},
	{"ч-чо-л-ло-в-вік", "ч-чо-л-ло-в-вік/[чоловік]noun:anim:m:v_naz:alt|ч-чо-л-ло-в-вік/[чоловік]noun:anim:p:v_rod:subst:alt|ч-чо-л-ло-в-вік/[чоловік]noun:anim:p:v_zna:subst:alt"},
	{"Я-вор-рова", "Я-вор-рова/[Яворів]noun:inanim:m:v_rod:prop:geo:alt|Я-вор-рова/[яворовий]adj:f:v_kly:alt|Я-вор-рова/[яворовий]adj:f:v_naz:alt"},
	{"мілі-ціо-нерів", "мілі-ціо-нерів/[null]null"},
	{"Тааак", "Тааак/[так]adv:pron:dem:alt|Тааак/[так]conj:coord:alt|Тааак/[так]conj:subord:alt|Тааак/[так]part:alt"},
	{"загубииилася", "загубииилася/[загубитися]verb:rev:perf:past:f:alt"},
	{"гаааа", "гаааа/[га]intj:alt"},
	{"з-зателефоную", "з-зателефоную/[зателефонувати]verb:perf:futr:s:1:alt"},
	{"т-то", "т-то/[null]null"},
	{"г-г-г", "г-г-г/[null]null"},
	{"йо-га", "йо-га/[null]null"},
	{"с-г", "с-г/[null]null"},
	{"де-куди", "де-куди/[null]null"},
	{"все-транс", "все-транс/[null]null"},
	{"транс-все", "транс-все/[null]null"},
	{"тренінґ", "тренінґ/[тренінґ]noun:inanim:m:v_naz:alt|тренінґ/[тренінґ]noun:inanim:m:v_zna:alt"},
	{"антирадіяційно", "антирадіяційно/[антирадіяційно]adv:alt"},
	{"фотометер", "фотометер/[фотометер]noun:inanim:m:v_naz:alt|фотометер/[фотометер]noun:inanim:m:v_zna:alt"},
	{"ідеольогічними", "ідеольогічними/[ідеольогічний]adj:p:v_oru:alt"},
	{"Мелітопольский", "Мелітопольский/[мелітопольский]adj:m:v_kly:bad|Мелітопольский/[мелітопольский]adj:m:v_naz:bad|Мелітопольский/[мелітопольский]adj:m:v_zna:rinanim:bad"},
	{"сьвідомими", "сьвідомими/[сьвідомий]adj:p:v_oru:compb:arch"},
	{"сьвятинь", "сьвятинь/[сьвятиня]noun:inanim:p:v_rod:arch"},
	{"вДягнутися", "вДягнутися/[вдягнутися]verb:rev:perf:inf:alt"},
	{"зфотографував", "зфотографував/[зфотографувати]verb:perf:past:m:alt"},
	{"Зформувати", "Зформувати/[зформувати]verb:perf:inf:alt"},
	{"км2", "км/[км]noun:inanim:m:v_dav:nv:abbr|км/[км]noun:inanim:m:v_mis:nv:abbr|км/[км]noun:inanim:m:v_naz:nv:abbr|км/[км]noun:inanim:m:v_oru:nv:abbr|км/[км]noun:inanim:m:v_rod:nv:abbr|км/[км]noun:inanim:m:v_zna:nv:abbr|км/[км]noun:inanim:p:v_dav:nv:abbr|км/[км]noun:inanim:p:v_mis:nv:abbr|км/[км]noun:inanim:p:v_naz:nv:abbr|км/[км]noun:inanim:p:v_oru:nv:abbr|км/[км]noun:inanim:p:v_rod:nv:abbr|км/[км]noun:inanim:p:v_zna:nv:abbr -- 2/[2]number"},
	{"м3", "м/[м]noun:inanim:m:v_dav:nv:abbr|м/[м]noun:inanim:m:v_mis:nv:abbr|м/[м]noun:inanim:m:v_naz:nv:abbr|м/[м]noun:inanim:m:v_oru:nv:abbr|м/[м]noun:inanim:m:v_rod:nv:abbr|м/[м]noun:inanim:m:v_zna:nv:abbr|м/[м]noun:inanim:p:v_dav:nv:abbr|м/[м]noun:inanim:p:v_mis:nv:abbr|м/[м]noun:inanim:p:v_naz:nv:abbr|м/[м]noun:inanim:p:v_oru:nv:abbr|м/[м]noun:inanim:p:v_rod:nv:abbr|м/[м]noun:inanim:p:v_zna:nv:abbr -- 3/[3]number"},
	{"д[окто]р[ом]", "д[окто]р[ом]/[доктор]noun:anim:m:v_oru:alt"},
	{"Акад[емія]", "Акад[емія]/[академія]noun:inanim:f:v_naz:alt"},
	{"дївчини", "дївчини/[дівчина]noun:anim:f:v_rod:alt|дївчини/[дівчина]noun:anim:p:v_kly:alt|дївчини/[дівчина]noun:anim:p:v_naz:alt|дївчини/[дівчина]noun:anim:p:v_zna:alt"},
	{"П`єр", "П'єр/[П'єр]noun:anim:m:v_naz:prop:fname"},
}

var javaUkrainianTaggerNotTagged = []string{
	"5610100000:01:041:0487",
	"H1",
	"УКРАЇ",
	"агатомільйонних",
	"54–бкх",
	"Берлін-7",
	"м-б",
	"хто-то",
	"що-то",
	"как-то",
	"кто-то",
	"до-пари",
	"Малишко-це",
	"відносини-коли",
	"лісо-та",
	"Ліс-наш",
	"Нью-Париж",
	"кохання-найщиріше",
	"Донець-кий",
	"мас-штаби",
	"підо-пічні",
	"рибо-полювання",
	"вовіки-вічні",
	"Гірник-спорт",
	"Квітку-Основ'яненко",
	"йо-йо-тусовки",
	"дер-жав",
	"теле-та",
	"квазі-я",
	"макро-і",
	"авіамені",
	"антиЄвропа",
	"напіврозслабеному",
	"напіви",
	"паталого-анатомічний",
	"патолога-анатомічний",
	"патолого-гмкнх",
	"патолого-голова",
	"бірмюково-блакитний",
	"во-політичний",
	"о-політичний",
	"рово-часового",
	"військо-во-політичний",
	"ііі",
	"відео-конференц-зв'язком",
	"Шаабан",
	"спа-салоне",
	"піт-стопою",
	"па-холка",
	"АвСтріях",
	"аВс",
	"вДягнУтися",
	"австріях",
	"польская",
}


// TestUkrainianTagger_Dictionary twins TestTools.testDictionary (sample POS non-empty).
func TestUkrainianTagger_Dictionary_official(t *testing.T) {
	if DiscoverUkrainianPOSDict() == "" {
		t.Skip("ukrainian.dict not discoverable")
	}
	EnsureDefaultUkrainianTagger()
	tagger := DefaultUkrainianTagger
	require.NotNil(t, tagger)
	require.Equal(t, UkrainianDictPath, tagger.GetDictionaryPath())
	require.NotEmpty(t, UkrainianPOSDictPath())
	require.NotNil(t, tagger.GetWordTagger())
	samples := []string{"дім", "віл", "києві", "майже", "два", "рік", "бути", "україна"}
	for _, w := range samples {
		tw := tagger.TagWordExact(w)
		if len(tw) == 0 {
			// try common surfaces
			continue
		}
		for _, tword := range tw {
			require.NotEmpty(t, tword.PosTag, "POS empty for %q", w)
		}
	}
}

// TestUkrainianTagger_Tagger twins every active myAssert in UkrainianTaggerTest.
func TestUkrainianTagger_Tagger(t *testing.T) {
	if DiscoverUkrainianPOSDict() == "" {
		t.Skip("ukrainian.dict not discoverable")
	}
	EnsureDefaultUkrainianTagger()
	require.NotNil(t, DefaultUkrainianTagger)

	for _, tc := range javaUkrainianTaggerMyAssertCases {
		t.Run(tc.in, func(t *testing.T) {
			got := myAssertUkrainianTagger(tc.in)
			require.Equal(t, tc.want, got, "input=%q", tc.in)
		})
	}
}

// TestUkrainianTagger_NotTagged twins assertNotTagged helpers.
func TestUkrainianTagger_NotTagged(t *testing.T) {
	if DiscoverUkrainianPOSDict() == "" {
		t.Skip("ukrainian.dict not discoverable")
	}
	EnsureDefaultUkrainianTagger()
	for _, w := range javaUkrainianTaggerNotTagged {
		t.Run(w, func(t *testing.T) {
			got := myAssertUkrainianTagger(w)
			require.Equal(t, w+"/[null]null", got, "assertNotTagged %q", w)
		})
	}
}

// myAssertUkrainianTagger ports TestTools.myAssert(input, expected, UkrainianWordTokenizer, UkrainianTagger).
func myAssertUkrainianTagger(input string) string {
	EnsureDefaultUkrainianTagger()
	tagger := DefaultUkrainianTagger
	wt := tokuk.NewUkrainianWordTokenizer()
	tokens := wt.Tokenize(input)
	var noWS []string
	for _, tok := range tokens {
		if testToolsIsWord(tok) {
			noWS = append(noWS, tok)
		}
	}
	output := tagger.Tag(noWS)
	var parts []string
	for _, atr := range output {
		parts = append(parts, strings.Join(testToolsGetAsStrings(atr), "|"))
	}
	return strings.Join(parts, " -- ")
}

func testToolsIsWord(token string) bool {
	for _, r := range token {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func testToolsGetAsStrings(atr *languagetool.AnalyzedTokenReadings) []string {
	if atr == nil {
		return nil
	}
	var readings []string
	for _, r := range atr.GetReadings() {
		if r != nil {
			readings = append(readings, testToolsGetAsString(r))
		}
	}
	sort.Strings(readings)
	return readings
}

func testToolsGetAsString(tok *languagetool.AnalyzedToken) string {
	lemma, pos := "null", "null"
	if tok.GetLemma() != nil {
		lemma = *tok.GetLemma()
	}
	if tok.GetPOSTag() != nil {
		pos = *tok.GetPOSTag()
	}
	return tok.GetToken() + "/[" + lemma + "]" + pos
}
