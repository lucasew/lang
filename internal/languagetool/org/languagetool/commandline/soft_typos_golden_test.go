package commandline

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiscoverEnglishTyposFile(t *testing.T) {
	p := DiscoverEnglishTyposFile(nil)
	if p == "" {
		t.Skip("en-typos.tsv not found")
	}
	require.FileExists(t, p)
}

func TestGolden_SoftTyposSuggestions(t *testing.T) {
	if DiscoverEnglishTyposFile(nil) == "" && DiscoverEnglishUSDict(nil) == "" {
		t.Skip("need typos file or binary dict")
	}
	cases := []struct{ text, sug string }{
		{"I will go tommorow.", "tomorrow"},
		{"That is wierd.", "weird"},
		{"Please recieve this.", "receive"},
	}
	for _, tc := range cases {
		t.Run(tc.sug, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == "MORFOLOGIK_RULE_EN_US" && f.Suggestion == tc.sug {
					found = true
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftCanCan(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "They can can fish.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_SOFT_CAN_CAN" {
			found = true
			require.Equal(t, "can", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_SoftThatThat(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "I know that that is true.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_SOFT_THAT_THAT" {
			found = true
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_SoftHadOf(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "I had of known better.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_SOFT_HAD_OF" {
			found = true
			require.Equal(t, "had", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_SoftAgreement(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"They is ready.", "EN_SOFT_THEY_IS", "are"},
		{"I are happy.", "EN_SOFT_I_ARE", "am"},
		{"He are late.", "EN_SOFT_HE_ARE", "is"},
		{"This are wrong.", "EN_SOFT_THIS_ARE", "is"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					if tc.sug != "" {
						require.Equal(t, tc.sug, f.Suggestion)
					}
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftAPlural(t *testing.T) {
	// needs POS tagger for NNS on "books"
	if DiscoverEnglishPOSDict(nil) == "" {
		t.Skip("english.dict not found")
	}
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "A books are here.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_SOFT_A_PLURAL" {
			found = true
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_Soft3sgBareVerb(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"He go home.", "EN_SOFT_HE_GO", "goes"},
		{"She walk fast.", "EN_SOFT_HE_WALK", "walks"},
		{"It like rain.", "EN_SOFT_HE_LIKE", "likes"},
		{"He want more.", "EN_SOFT_HE_WANT", "wants"},
		{"She need help.", "EN_SOFT_HE_NEED", "needs"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftLooseLose(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "I want to loose weight.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_SOFT_LOOSE_LOSE" {
			found = true
			require.Equal(t, "lose weight", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_SoftConfusablesExtra(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"The rules will take affect soon.", "EN_SOFT_TAKE_AFFECT", "take effect"},
		{"Please be quite in the library.", "EN_SOFT_QUITE_QUIET", "be quiet"},
		{"That peaked my interest a lot.", "EN_SOFT_PEAKED_INTEREST", "piqued my interest"},
		{"I could care less about that.", "EN_SOFT_COULD_CARE_LESS", "couldn't care less"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ApplySoft3sg(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "He go home every day.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "goes")
}

func TestGolden_ApplySoftPhrase(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "I did it on accident yesterday.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "by accident")
}

func TestGolden_SoftConfusablesMore(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Whose going to the party?", "EN_SOFT_WHOSE_WHO_S", "who's going"},
		{"Who's book is this?", "EN_SOFT_WHO_S_BOOK", "whose book"},
		{"Please breath deeply now.", "EN_SOFT_BREATH_BREATHE", "breathe deeply"},
		{"I want to advice you.", "EN_SOFT_ADVICE_ADVISE", "to advise"},
		{"He do the work.", "EN_SOFT_HE_DO", "does"},
		{"She have a car.", "EN_SOFT_HE_HAVE", "has"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftContractionForms(t *testing.T) {
	// cant/wont are valid dict words; soft grammar forces apostrophe suggestions.
	cases := []struct {
		text, rule, sug string
	}{
		{"I dont know.", "EN_SOFT_DONT", "don't"},
		{"She cant come.", "EN_SOFT_CANT", "can't"},
		{"They wont mind.", "EN_SOFT_WONT", "won't"},
		{"He didnt call.", "EN_SOFT_DIDNT", "didn't"},
		{"It isnt ready.", "EN_SOFT_ISNT", "isn't"},
		{"They arent here.", "EN_SOFT_ARENT", "aren't"},
		{"He wasnt late.", "EN_SOFT_WASNT", "wasn't"},
		{"You shouldnt go.", "EN_SOFT_SHOULDNT", "shouldn't"},
		{"I wouldnt care.", "EN_SOFT_WOULDNT", "wouldn't"},
		{"Youre welcome.", "EN_SOFT_YOURE", "You're"},
		{"Theyre leaving.", "EN_SOFT_THEYRE", "They're"},
		{"I wouldve gone.", "EN_SOFT_WOULDVE", "would've"},
		{"She couldve won.", "EN_SOFT_COULDVE", "could've"},
		{"You shouldve called.", "EN_SOFT_SHOULDVE", "should've"},
		{"He mustve left.", "EN_SOFT_MUSTVE", "must've"},
		{"It mightve worked.", "EN_SOFT_MIGHTVE", "might've"},
		{"Thatll do.", "EN_SOFT_THATLL", "That'll"},
		{"Itll rain soon.", "EN_SOFT_ITLL", "It'll"},
		{"I know thatll work.", "EN_SOFT_THATLL", "that'll"},
		{"Say youre ready.", "EN_SOFT_YOURE", "you're"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			// fallback: typos/speller path also acceptable for some forms
			if !found {
				for _, f := range findings {
					if f.Rule == "MORFOLOGIK_RULE_EN_US" && f.Suggestion == tc.sug {
						found = true
					}
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ApplySoftContraction(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "I dont know if she cant come.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	s := out.String()
	require.Contains(t, s, "don't")
	require.Contains(t, s, "can't")
}

func TestGolden_ApplySoftPhraseCase(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "On Accident I slipped.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "By accident")
}

func TestGolden_SoftTypography(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Hello!!", "EN_SOFT_DOUBLE_BANG"},
		{"What??", "EN_SOFT_DOUBLE_Q"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "typographical", f.Type, "%+v", f)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftWereGoingTo(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Were going to leave soon.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_SOFT_WERE_WE_RE" {
			found = true
			require.Equal(t, "we're going to", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_SoftFusedWords(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"I have alot of work.", "EN_SOFT_ALOT", "a lot"},
		{"Infact it works.", "EN_SOFT_INFACT", "in fact"},
		{"Come aswell please.", "EN_SOFT_ASWELL", "as well"},
		{"Never the less we try.", "EN_SOFT_NEVERTHELESS_SPLIT", "nevertheless"},
		{"Everytime I try it fails.", "EN_SOFT_EVERYTIME", "every time"},
		{"Noone knows the answer.", "EN_SOFT_NOONE", "no one"},
		{"Don't give into pressure.", "EN_SOFT_INTO_IN_TO", "give in to"},
		{"Talk to eachother soon.", "EN_SOFT_EACHOTHER", "each other"},
		{"Inspite of that we stay.", "EN_SOFT_INSPITE", "in spite"},
		{"Atleast try once.", "EN_SOFT_ATLEAST", "at least"},
		{"Incase it rains, wait.", "EN_SOFT_INCASE", "in case"},
		{"Upto ten people may join.", "EN_SOFT_UPTO", "up to"},
		{"I need to workout daily.", "EN_SOFT_WORKOUT_VERB", "to work out"},
		{"I need to setup the tool.", "EN_SOFT_SETUP_VERB", "to set up"},
		{"You need to login first.", "EN_SOFT_LOGIN_VERB", "to log in"},
		{"Please to checkout the code.", "EN_SOFT_CHECKOUT_VERB", "to check out"},
		{"Remember to backup data.", "EN_SOFT_BACKUP_VERB", "to back up"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ApplySoftDoubleBang(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Hello!!", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	// first suggestion "!" replaces the "!!" span
	require.Equal(t, "Hello!", strings.TrimSpace(out.String()))
}

func TestGolden_ApplySoftAlot(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "I have alot of work.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "a lot")
	require.NotContains(t, out.String(), "alot")
}

func TestGolden_SoftStyleCategory(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"This is very unique work.", "EN_SOFT_VERY_UNIQUE"},
		{"He literally died laughing.", "EN_SOFT_LITERALLY_FIG"},
		{"She is kind of a genius.", "EN_SOFT_KIND_OF_A"},
		{"We left due to the fact that it rained.", "EN_SOFT_DUE_TO_THE_FACT"},
		{"At this point in time we wait.", "EN_SOFT_AT_THIS_POINT_IN_TIME"},
		{"In the event that it fails, retry.", "EN_SOFT_IN_THE_EVENT_THAT"},
		{"The end result is clear.", "EN_SOFT_END_RESULT"},
		{"His past history is known.", "EN_SOFT_PAST_HISTORY"},
		{"Get a free gift today.", "EN_SOFT_FREE_GIFT"},
		{"We must completely eliminate waste.", "EN_SOFT_COMPLETELY_ELIMINATE"},
		{"This is different than that.", "EN_SOFT_DIFFERENT_THAN"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					// STYLE category maps to style issue type via soft loader
					require.Equal(t, "style", f.Type, "%+v", f)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ApplySoftWouldve(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "I wouldve gone earlier.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "would've")
	require.NotContains(t, out.String(), "wouldve")
}

func TestGolden_SoftCaseAndCount(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"This is between you and I.", "EN_SOFT_BETWEEN_YOU_AND_I", "between you and me"},
		{"It is for you and I.", "EN_SOFT_FOR_YOU_AND_I", "for you and me"},
		{"Come with you and I.", "EN_SOFT_WITH_YOU_AND_I", "with you and me"},
		{"Give it to who asks.", "EN_SOFT_TO_WHO", "to whom"},
		{"Less people came today.", "EN_SOFT_LESS_PEOPLE", "Fewer people"},
		{"The amount of people grew.", "EN_SOFT_AMOUNT_OF_PEOPLE", "number of people"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ApplySoftBetweenYouAndI(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Keep this between you and I.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "between you and me")
}

func TestGolden_SoftMoreStyleAndTryAnd(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Please try and see the difference.", "EN_SOFT_TRY_AND", "try to see"},
		{"Each and every student passed.", "EN_SOFT_EACH_AND_EVERY", ""},
		{"First and foremost, plan carefully.", "EN_SOFT_FIRST_AND_FOREMOST", ""},
		{"Learn the basic fundamentals first.", "EN_SOFT_BASIC_FUNDAMENTALS", ""},
		{"The reason is because it rained.", "EN_SOFT_REASON_IS_BECAUSE", ""},
		{"Decide whether or not to go.", "EN_SOFT_WHETHER_OR_NOT", ""},
		{"That is an actual fact.", "EN_SOFT_ACTUAL_FACT", ""},
		{"A true fact remains.", "EN_SOFT_TRUE_FACT", ""},
		{"Do advance planning early.", "EN_SOFT_ADVANCE_PLANNING", ""},
		{"Stay in close proximity.", "EN_SOFT_CLOSE_PROXIMITY", ""},
		{"Share your future plans.", "EN_SOFT_FUTURE_PLANS", ""},
		{"What an unexpected surprise.", "EN_SOFT_UNEXPECTED_SURPRISE", ""},
		{"Please revert back soon.", "EN_SOFT_REVERT_BACK", ""},
		{"Do not repeat again.", "EN_SOFT_REPEAT_AGAIN", ""},
		{"The final outcome is known.", "EN_SOFT_FINAL_OUTCOME", ""},
		{"There is a general consensus.", "EN_SOFT_GENERAL_CONSENSUS", ""},
		{"In my personal opinion, wait.", "EN_SOFT_PERSONAL_OPINION", ""},
		{"The train came to a complete stop.", "EN_SOFT_COMPLETE_STOP", ""},
		{"This is absolutely essential.", "EN_SOFT_ABSOLUTELY_ESSENTIAL", ""},
		{"They are exactly the same.", "EN_SOFT_EXACTLY_THE_SAME", ""},
		{"Work is currently in progress.", "EN_SOFT_CURRENTLY_IN_PROGRESS", ""},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					if tc.sug != "" {
						require.Equal(t, tc.sug, f.Suggestion)
					}
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ApplySoftLessPeople(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Less people attended.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "Fewer people")
}

func TestGolden_ApplySoftAmountOfPeople(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "The amount of people grew.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "number of people")
}

func TestGolden_ApplySoftDontCase(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Dont forget keys.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "Don't")
}

func TestGolden_ApplySoftAnyways(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Anyways, we left early.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	// SoftPreserveCase capitalizes suggestion when match is sentence-initial.
	require.Contains(t, out.String(), "Anyway")
	require.NotContains(t, strings.ToLower(out.String()), "anyways")
}

func TestGolden_SoftUSVariantHints(t *testing.T) {
	// Loaded from testdata/grammar/en-US-soft.xml when language is en-US.
	cases := []struct {
		text, rule, sug string
	}{
		{"Walk towards the door.", "EN_SOFT_TOWARDS_US", "toward"},
		{"Sit amongst friends.", "EN_SOFT_AMONGST_US", "among"},
		{"Wait whilst I check.", "EN_SOFT_WHILST_US", "while"},
		{"A grey sky.", "EN_SOFT_GREY_US", "gray"},
		{"Pick a colour.", "EN_SOFT_COLOUR_US", "color"},
		{"My favourite book.", "EN_SOFT_FAVOURITE_US", "favorite"},
		{"City centre is busy.", "EN_SOFT_CENTRE_US", "center"},
		{"Please organise files.", "EN_SOFT_ORGANISE_US", "organize"},
		{"I realise now.", "EN_SOFT_REALISE_US", "realize"},
		{"Good behaviour matters.", "EN_SOFT_BEHAVIOUR_US", "behavior"},
		{"We travelled far.", "EN_SOFT_TRAVELLED_US", "traveled"},
		{"The flight was cancelled.", "EN_SOFT_CANCELLED_US", "canceled"},
		{"The data was modelled.", "EN_SOFT_MODELLED_US", "modeled"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en-US"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					if tc.sug != "" {
						require.Equal(t, tc.sug, f.Suggestion)
					}
					// TYPOS category → misspelling issue type
					require.Equal(t, "misspelling", f.Type, "%+v", f)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftUSVariantsNotOnPlainEN(t *testing.T) {
	// en-US-soft.xml must not load for language "en"
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Pick a colour.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	for _, f := range findings {
		require.NotEqual(t, "EN_SOFT_COLOUR_US", f.Rule, "%+v", findings)
	}
}

func TestGolden_SoftGBVariantHints(t *testing.T) {
	// Loaded from testdata/grammar/en-GB-soft.xml when language is en-GB.
	cases := []struct {
		text, rule, sug string
	}{
		{"Walk toward the door.", "EN_SOFT_TOWARD_GB", "towards"},
		{"A gray sky.", "EN_SOFT_GRAY_GB", "grey"},
		{"Pick a color.", "EN_SOFT_COLOR_GB", "colour"},
		{"My favorite book.", "EN_SOFT_FAVORITE_GB", "favourite"},
		{"City center is busy.", "EN_SOFT_CENTER_GB", "centre"},
		{"Please organize files.", "EN_SOFT_ORGANIZE_GB", "organise"},
		{"I realize now.", "EN_SOFT_REALIZE_GB", "realise"},
		{"Good behavior matters.", "EN_SOFT_BEHAVIOR_GB", "behaviour"},
		{"We traveled far.", "EN_SOFT_TRAVELED_GB", "travelled"},
		{"The flight was canceled.", "EN_SOFT_CANCELED_GB", "cancelled"},
		{"The data was modeled.", "EN_SOFT_MODELED_GB", "modelled"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en-GB"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
					require.Equal(t, "misspelling", f.Type, "%+v", f)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftGBVariantsNotOnUS(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Pick a color.", &CommandLineOptions{Language: "en-US"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	for _, f := range findings {
		require.NotEqual(t, "EN_SOFT_COLOR_GB", f.Rule, "%+v", findings)
	}
}

func TestGolden_ApplySoftGBColour(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en-GB", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Pick a color please.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "colour")
	require.NotContains(t, out.String(), "color")
}

func TestGolden_SoftAnywaysOnPlainEN(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Anyways, we left.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_SOFT_ANYWAYS" {
			found = true
			require.Equal(t, "Anyway", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_SoftSupposeTo(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "You suppose to leave now.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_SUPPOSE_TO" || f.Rule == "EN_SOFT_SUPPOSE_TO" {
			found = true
			require.Equal(t, "supposed to", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_SoftTokenSequenceExtras(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"I use to go there often.", "EN_USED_TO_GO", "used to go"},
		{"For all intensive purposes it works.", "EN_FOR_ALL_INTENSIVE", "for all intents and purposes"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftDialectForms(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"I seen him yesterday.", "EN_SOFT_I_SEEN", "I saw"},
		{"I done the work already.", "EN_SOFT_I_DONE", "I did"},
		{"They was late again.", "EN_SOFT_WE_WAS", "were"},
		{"I is ready now.", "EN_SOFT_I_IS", "am"},
		{"Me and John went home.", "EN_SOFT_ME_AND", ""},
		{"She should of been here.", "EN_SOFT_SHOULD_OF_SPACED", "should have been"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					if tc.sug != "" {
						require.Equal(t, tc.sug, f.Suggestion)
					}
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}
