package zh

import (
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// sentencesUtilToSentenceList ports HanLP
// com.hankcs.hanlp.utility.SentencesUtil.toSentenceList(String)
// which is toSentenceList(chars, true /* insert */).
//
// Used by ChineseSentenceTokenizer (not a public LT type). Algorithm transcribed
// from HanLP portable-1.8.2 bytecode (javap SentencesUtil.toSentenceList([CZ)).
//
// Java twin: com.hankcs.hanlp.utility.SentencesUtil (dependency of
// org.languagetool.tokenizers.zh.ChineseSentenceTokenizer).
func sentencesUtilToSentenceList(text string) []string {
	return sentencesUtilToSentenceListInsert(text, true)
}

// sentencesUtilToSentenceListInsert ports
// SentencesUtil.toSentenceList(char[] chars, boolean insert).
func sentencesUtilToSentenceListInsert(text string, insert bool) []string {
	chars := utf16.Encode([]rune(text))
	var sb []uint16
	var list []string
	for i := 0; i < len(chars); i++ {
		c := chars[i]
		// Leading skip only when buffer empty: Character.isWhitespace OR space (32).
		// Space is already isWhitespace; both branches exist in bytecode.
		if len(sb) == 0 {
			if tools.CharacterIsWhitespace(rune(c)) || c == 32 {
				continue
			}
		}
		sb = append(sb, c)
		switch c {
		case 9, 10, 13, 32, 33, 63, 160, 12290, 65281, 65311:
			// \t \n \r space ! ? NBSP 。！？ — always split
			list = sentencesUtilInsertIntoList(sb, list)
			sb = sb[:0]
		case 44, 59, 65292, 65307:
			// , ; ，； — split only when insert==true
			if !insert {
				continue
			}
			list = sentencesUtilInsertIntoList(sb, list)
			sb = sb[:0]
		case 46:
			// '.' — split only if not last char AND next char > 128
			if i >= len(chars)-1 {
				continue
			}
			if chars[i+1] <= 128 {
				continue
			}
			list = sentencesUtilInsertIntoList(sb, list)
			sb = sb[:0]
		case 8230:
			// '…' (U+2026) — if next is also 8230, append second, advance, split
			if i >= len(chars)-1 {
				continue
			}
			if chars[i+1] != 8230 {
				continue
			}
			sb = append(sb, 8230)
			i++
			list = sentencesUtilInsertIntoList(sb, list)
			sb = sb[:0]
		default:
			// no split
		}
	}
	if len(sb) > 0 {
		list = sentencesUtilInsertIntoList(sb, list)
	}
	return list
}

// sentencesUtilInsertIntoList ports private SentencesUtil.insertIntoList:
// sb.toString().trim(); add only if length > 0.
// Java String.trim() = strip UTF-16 units <= U+0020 (tools.JavaStringTrim).
func sentencesUtilInsertIntoList(sb []uint16, list []string) []string {
	s := tools.JavaStringTrim(utf16UnitsToString(sb))
	if len(s) > 0 {
		list = append(list, s)
	}
	return list
}

func utf16UnitsToString(u []uint16) string {
	if len(u) == 0 {
		return ""
	}
	return string(utf16.Decode(u))
}
