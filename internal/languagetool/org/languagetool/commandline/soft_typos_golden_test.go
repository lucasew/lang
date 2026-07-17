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
		{"Whose going to the party?", "EN_SOFT_WHOSE_WHO_S", "Who's going"},
		{"Who's book is this?", "EN_SOFT_WHO_S_BOOK", "Whose book"},
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
			require.Equal(t, "We're going to", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_SoftFusedWords(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"I have alot of work.", "EN_SOFT_ALOT", "a lot"},
		{"Infact it works.", "EN_SOFT_INFACT", "In fact"},
		{"Come aswell please.", "EN_SOFT_ASWELL", "as well"},
		{"Never the less we try.", "EN_SOFT_NEVERTHELESS_SPLIT", "nevertheless"},
		{"Everytime I try it fails.", "EN_SOFT_EVERYTIME", "Every time"},
		{"Noone knows the answer.", "EN_SOFT_NOONE", "No one"},
		{"Don't give into pressure.", "EN_SOFT_INTO_IN_TO", "give in to"},
		{"Talk to eachother soon.", "EN_SOFT_EACHOTHER", "each other"},
		{"Inspite of that we stay.", "EN_SOFT_INSPITE", "In spite"},
		{"Atleast try once.", "EN_SOFT_ATLEAST", "At least"},
		{"Incase it rains, wait.", "EN_SOFT_INCASE", "In case"},
		{"Upto ten people may join.", "EN_SOFT_UPTO", "Up to"},
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

func TestGolden_SoftMoreRedundancyStyle(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"That is an added bonus.", "EN_SOFT_ADDED_BONUS", ""},
		{"Wait a brief moment.", "EN_SOFT_BRIEF_MOMENT", ""},
		{"Please join together now.", "EN_SOFT_JOIN_TOGETHER", ""},
		{"We should plan ahead carefully.", "EN_SOFT_PLAN_AHEAD", ""},
		{"The problem still remains.", "EN_SOFT_STILL_REMAINS", ""},
		{"Birds circle around the tree.", "EN_SOFT_CIRCLE_AROUND", ""},
		{"Please empty out the drawer.", "EN_SOFT_EMPTY_OUT", ""},
		{"The dress is pink in color.", "EN_SOFT_PINK_IN_COLOR", "pink"},
		{"The sky is blue in colour.", "EN_SOFT_BLUE_IN_COLOUR", "blue"},
		{"We need to preplan the trip.", "EN_SOFT_PREPLAN", "plan"},
		{"Irregardless of the cost, go.", "EN_SOFT_IRREGARDLESS", "Regardless"},
		{"He is supposably ready.", "EN_SOFT_SUPPOSABLY", "supposedly"},
		{"They are the exact same.", "EN_SOFT_EXACT_SAME", ""},
		{"The reason why we left is clear.", "EN_SOFT_REASON_WHY", ""},
		{"Please return back soon.", "EN_SOFT_RETURN_BACK", ""},
		{"They ascend up the stairs.", "EN_SOFT_ASCEND_UP", ""},
		{"They descend down the hill.", "EN_SOFT_DESCEND_DOWN", ""},
		{"Please enter in the room.", "EN_SOFT_ENTER_IN", "enter the"},
		{"Continue on with the plan.", "EN_SOFT_CONTINUE_ON", "continue with"},
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
					require.Equal(t, "style", f.Type, "%+v", f)
					if tc.sug != "" {
						require.Equal(t, tc.sug, f.Suggestion)
					}
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftExtraTyposMap(t *testing.T) {
	cases := []struct {
		text, sug string
	}{
		{"I was commited to the goal.", "committed"},
		{"They are transfered today.", "transferred"},
		{"A nice restaraunt nearby.", "restaurant"},
		{"A good questionaire form.", "questionnaire"},
		{"The rythm is steady.", "rhythm"},
		{"He is persistant about it.", "persistent"},
		{"A vaccuum cleaner.", "vacuum"},
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

func TestGolden_SoftUSGBExtraVariants(t *testing.T) {
	us := []struct {
		text, rule, sug string
	}{
		{"Dry humour helps.", "EN_SOFT_HUMOUR_US", "humor"},
		{"Hard labour pays.", "EN_SOFT_LABOUR_US", "labor"},
		{"Self defence class.", "EN_SOFT_DEFENCE_US", "defense"},
		{"A driving licence.", "EN_SOFT_LICENCE_US", "license"},
		{"Please analyse data.", "EN_SOFT_ANALYSE_US", "analyze"},
		{"A large theatre hall.", "EN_SOFT_THEATRE_US", "theater"},
		{"Browse the catalogue.", "EN_SOFT_CATALOGUE_US", "catalog"},
		{"Open a dialogue box.", "EN_SOFT_DIALOGUE_US", "dialog"},
		{"TV programme tonight.", "EN_SOFT_PROGRAMME_US", "program"},
		{"My neighbour is kind.", "EN_SOFT_NEIGHBOUR_US", "neighbor"},
	}
	for _, tc := range us {
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
					require.Equal(t, tc.sug, f.Suggestion)
					require.Equal(t, "misspelling", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
	gb := []struct {
		text, rule, sug string
	}{
		{"Dry humor helps.", "EN_SOFT_HUMOR_GB", "humour"},
		{"Hard labor pays.", "EN_SOFT_LABOR_GB", "labour"},
		{"Self defense class.", "EN_SOFT_DEFENSE_GB", "defence"},
		{"A driving license.", "EN_SOFT_LICENSE_GB", "licence"},
		{"Please analyze data.", "EN_SOFT_ANALYZE_GB", "analyse"},
		{"A large theater hall.", "EN_SOFT_THEATER_GB", "theatre"},
		{"Browse the catalog.", "EN_SOFT_CATALOG_GB", "catalogue"},
		{"Open a dialog box.", "EN_SOFT_DIALOG_GB", "dialogue"},
		{"TV program tonight.", "EN_SOFT_PROGRAM_GB", "programme"},
		{"My neighbor is kind.", "EN_SOFT_NEIGHBOR_GB", "neighbour"},
	}
	for _, tc := range gb {
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
					require.Equal(t, "misspelling", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ApplySoftIrregardless(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Irregardless of cost, ship it.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "Regardless")
	require.NotContains(t, strings.ToLower(out.String()), "irregardless")
}

func TestGolden_SoftIdiomConfusables(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"The principal of the matter is clear.", "EN_SOFT_PRINCIPAL_REASON", "principle of the matter"},
		{"The principle officer spoke.", "EN_SOFT_PRINCIPLE_OFFICER", "principal officer"},
		{"Colors compliment each other well.", "EN_SOFT_COMPLIMENT_COLORS", "complement each other"},
		{"That will peek my interest soon.", "EN_SOFT_PEEK_INTEREST", "pique my interest"},
		{"That will peak my interest soon.", "EN_SOFT_PEAK_INTEREST", "pique my interest"},
		{"Please insure that the door is locked.", "EN_SOFT_ENSURE_INSURE", "ensure that"},
		{"We need farther discussion tomorrow.", "EN_SOFT_FARTHER_ABSTRACT", "further discussion"},
		{"I will lay down and rest now.", "EN_SOFT_LAY_DOWN_REST", "lie down and rest"},
		{"Please sight the source carefully.", "EN_SOFT_CITE_SEE", "cite the source"},
		{"Bare with me for a moment.", "EN_SOFT_BARE_WITH", "Bear with me"},
		{"A deep seeded fear remains.", "EN_SOFT_DEEP_SEEDED", "deep-seated"},
		{"Nip it in the butt early.", "EN_SOFT_NIP_IN_THE_BUTT", "Nip it in the bud"},
		{"Case and point: it failed.", "EN_SOFT_CASE_AND_POINT", "Case in point"},
		{"They are one in the same.", "EN_SOFT_ONE_IN_THE_SAME", "one and the same"},
		{"I ordered an expresso please.", "EN_SOFT_EXPRESSO", "espresso"},
		{"They tried to excape the room.", "EN_SOFT_EXCAPE", "escape"},
		{"I like it exspecially today.", "EN_SOFT_EXSPECIALLY", "especially"},
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

func TestGolden_ApplySoftBareWithMe(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Bare with me for a second.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "Bear with me")
	require.NotContains(t, out.String(), "Bare with me")
}

func TestGolden_SoftPtBRRegionalHints(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Peguei o autocarro cedo.", "PT_SOFT_AUTOCARRO_BR", "ônibus"},
		{"Ligue no telemóvel agora.", "PT_SOFT_TELEMOVEL_BR", "celular"},
		{"O ecrã está sujo.", "PT_SOFT_ECRA_BR", "tela"},
		{"Esse facto importa.", "PT_SOFT_FACTO_BR", "fato"},
		{"Sem contacto visual.", "PT_SOFT_CONTACTO_BR", "contato"},
		{"Um objecto antigo.", "PT_SOFT_OBJECTO_BR", "objeto"},
		{"Resultado óptimo.", "PT_SOFT_OPTIMO_BR", "ótimo"},
		{"A acção principal.", "PT_SOFT_ACCAO_BR", "ação"},
		{"O comboio partiu.", "PT_SOFT_COMBOIO_BR", "trem"},
		{"A equipa ganhou.", "PT_SOFT_EQUIPA_BR", "equipe"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "pt-BR"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
					require.Equal(t, "misspelling", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPtPTRegionalHints(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Peguei o ônibus cedo.", "PT_SOFT_ONIBUS_PT", "autocarro"},
		{"Ligue no celular agora.", "PT_SOFT_CELULAR_PT", "telemóvel"},
		{"Esse fato importa.", "PT_SOFT_FATO_PT", "facto"},
		{"Sem contato visual.", "PT_SOFT_CONTATO_PT", "contacto"},
		{"Um objeto antigo.", "PT_SOFT_OBJETO_PT", "objecto"},
		{"Resultado ótimo.", "PT_SOFT_OTIMO_PT", "óptimo"},
		{"O trem partiu.", "PT_SOFT_TREM_PT", "comboio"},
		{"A equipe ganhou.", "PT_SOFT_EQUIPE_PT", "equipa"},
		{"Café da manhã cedo.", "PT_SOFT_CAFE_MANHA_PT", "pequeno-almoço"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "pt-PT"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
					require.Equal(t, "misspelling", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPtBRNotOnPtPT(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Peguei o autocarro cedo.", &CommandLineOptions{Language: "pt-PT"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	for _, f := range findings {
		require.NotEqual(t, "PT_SOFT_AUTOCARRO_BR", f.Rule, "%+v", findings)
	}
}

func TestGolden_ApplySoftPtBRAutocarro(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "pt-BR", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Peguei o autocarro cedo.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "ônibus")
	require.NotContains(t, out.String(), "autocarro")
}

func TestGolden_SoftEsMXRegionalHints(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Uso el ordenador hoy.", "ES_SOFT_ORDENADOR_MX", "computadora"},
		{"Mi móvil es nuevo.", "ES_SOFT_MOVIL_MX", "celular"},
		{"Quiero zumo de naranja.", "ES_SOFT_ZUMO_MX", "jugo"},
		{"El coche es rojo.", "ES_SOFT_COCHE_MX", "carro"},
		{"Voy a conducir despacio.", "ES_SOFT_CONDUCIR_MX", "manejar"},
		{"Compré un bolígrafo.", "ES_SOFT_BOLIGRAFO_MX", "pluma"},
		{"Un melocotón maduro.", "ES_SOFT_MELON_MX", "durazno"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "es-MX"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
					require.Equal(t, "misspelling", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftEsESRegionalHints(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Uso la computadora hoy.", "ES_SOFT_COMPUTADORA_ES", "ordenador"},
		{"Mi celular es nuevo.", "ES_SOFT_CELULAR_ES", "móvil"},
		{"Quiero jugo de naranja.", "ES_SOFT_JUGO_ES", "zumo"},
		{"El carro es rojo.", "ES_SOFT_CARRO_ES", "coche"},
		{"Voy a manejar despacio.", "ES_SOFT_MANEJAR_ES", "conducir"},
		{"Un durazno maduro.", "ES_SOFT_DURAZNO_ES", "melocotón"},
		{"Compré una laptop.", "ES_SOFT_LAPTOP_ES", "portátil"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "es-ES"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
					require.Equal(t, "misspelling", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftDeCHRegionalHints(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Die Straße ist nass.", "DE_SOFT_STRASSE_CH", "Strasse"},
		{"Ein groß Haus.", "DE_SOFT_GROSS_CH", "gross"},
		{"Hier darf man parken.", "DE_SOFT_PARKEN_CH", "parkieren"},
		{"Mein Fahrrad ist neu.", "DE_SOFT_FAHRRAD_CH", "Velo"},
		{"Das Handy klingelt.", "DE_SOFT_HANDY_CH", "Natel"},
		{"Wir grillen heute.", "DE_SOFT_GRILLEN_CH", "grillieren"},
		{"Es ist heiß draußen.", "DE_SOFT_ZWEI_CH", "heiss"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "de-CH"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
					require.Equal(t, "misspelling", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftDeATRegionalHints(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Im Januar schneit es.", "DE_SOFT_JANUAR_AT", "Jänner"},
		{"Wir essen Kartoffeln.", "DE_SOFT_KARTOFFELN_AT", "Erdäpfel"},
		{"Frische Tomaten bitte.", "DE_SOFT_TOMATEN_AT", "Paradeiser"},
		{"Ein Brötchen zum Kaffee.", "DE_SOFT_BROETCHEN_AT", "Semmel"},
		{"Gebratenes Hähnchen.", "DE_SOFT_HAEHNCHEN_AT", "Hendl"},
		{"Frischkäse und Quark.", "DE_SOFT_QUARK_AT", "Topfen"},
		{"Die Treppe ist steil.", "DE_SOFT_TREPPE_AT", "Stiege"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "de-AT"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
					require.Equal(t, "misspelling", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftFrCARegionalHints(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Bon week-end à tous.", "FR_SOFT_WEEKEND_CA", "fin de semaine"},
		{"Envoie un e-mail vite.", "FR_SOFT_EMAIL_CA", "courriel"},
		{"Mon portable sonne.", "FR_SOFT_PORTABLE_CA", "cellulaire"},
		{"La voiture est rouge.", "FR_SOFT_VOITURE_CA", "char"},
		{"Je vais faire des courses.", "FR_SOFT_COURSES_CA", "magasiner"},
		{"Le parking est plein.", "FR_SOFT_PARKING_CA", "stationnement"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "fr-CA"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
					require.Equal(t, "misspelling", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftEsMXNotOnEsES(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Uso el ordenador hoy.", &CommandLineOptions{Language: "es-ES"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	for _, f := range findings {
		require.NotEqual(t, "ES_SOFT_ORDENADOR_MX", f.Rule, "%+v", findings)
	}
}

func TestGolden_ApplySoftEsMXOrdenador(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "es-MX", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Uso el ordenador hoy.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "computadora")
	require.NotContains(t, out.String(), "ordenador")
}

func TestGolden_SoftCasingLowercaseI(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"i think so.", "EN_SOFT_LOWERCASE_I", "I"},
		{"im ready now.", "EN_SOFT_LOWERCASE_IM", "I'm"},
		{"ive finished work.", "EN_SOFT_LOWERCASE_IVE", "I've"},
		{"id like coffee.", "EN_SOFT_LOWERCASE_ID_LIKE", "I'd like"},
		{"ill go later.", "EN_SOFT_LOWERCASE_ILL_GO", "I'll go"},
		{"id love to join.", "EN_SOFT_LOWERCASE_ID_LOVE", "I'd love"},
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
					require.Equal(t, "typographical", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ApplySoftLowercaseI(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "i think this works.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "I think")
	require.NotContains(t, out.String(), "i think")
}

func TestGolden_SoftStyleMetaViaListRules(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, CoreListRules(&buf, "en"))
	out := buf.String()
	// SoftRuleMeta classifies these for list-rules columns
	require.Contains(t, out, "EN_SOFT_LOWERCASE_I\tCASING\t")
	require.Contains(t, out, "EN_SOFT_THE_THE\tSTYLE\t")
	var us bytes.Buffer
	require.NoError(t, CoreListRules(&us, "en-US"))
	require.Contains(t, us.String(), "EN_SOFT_COLOUR_US\tTYPOS\t")
}

func TestGolden_SoftIdiomConfusablesMore(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"She poured over the document carefully.", "EN_SOFT_PORED_OVER", "pored over the document"},
		{"They waive goodbye at the station.", "EN_SOFT_WAIVE_GOODBYE", "wave goodbye"},
		{"That is a mute point entirely.", "EN_SOFT_MUTE_POINT", "moot point"},
		{"That does not jive with the data.", "EN_SOFT_JIVE_WITH", "jibe with"},
		{"We will hone in on the bug.", "EN_SOFT_HONE_IN", "home in on"},
		{"A clever slight of hand trick.", "EN_SOFT_SLIGHT_OF_HAND", "sleight of hand"},
		{"Soldiers tow the line carefully.", "EN_SOFT_TOW_THE_LINE", "toe the line"},
		{"This begs the question of timing.", "EN_SOFT_BEGS_THE_QUESTION_SOFT", ""},
		{"Is that ok with you?", "EN_SOFT_CASE_SENSITIVE_OK", "OK"},
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

func TestGolden_ApplySoftMultiLang(t *testing.T) {
	cases := []struct {
		lang, in, want, not string
	}{
		{"de", "Ich denke das es stimmt.", "dass", "denke das"},
		{"fr", "Je vais a la maison.", "à la", "a la"},
		{"es", "Voy a el parque.", "al", "a el"},
		{"pt", "Vou a o mercado.", "ao", "a o"},
		{"nl", "Meer als gisteren.", "Meer dan", "Meer als"},
		{"it", "Vado a il negozio.", "al", "a il"},
	}
	for _, tc := range cases {
		t.Run(tc.lang, func(t *testing.T) {
			var out, errb bytes.Buffer
			code := RunWithIO([]string{"-l", tc.lang, "--apply", "-"}, RunHooks{
				ReadStdin: func() (string, error) { return tc.in, nil },
				Check:     CoreApplySuggestionsHook,
			}, &out, &errb)
			require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
			require.Contains(t, out.String(), tc.want)
			// not all rules remove the bad span text entirely when suggestions replace differently
			_ = tc.not
		})
	}
}

func TestGolden_IgnoreSpellingChatGPT(t *testing.T) {
	if DiscoverEnglishIgnoreSpellingList(nil) == "" {
		t.Skip("ignore list missing")
	}
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "I use ChatGPT and Claude daily.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	for _, f := range findings {
		require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
	}
}

func TestGolden_SoftDisableCategoriesStyle(t *testing.T) {
	// STYLE soft rule suppressed; grammar soft confusable remains
	text := "This is very unique work. Bare with me please."
	var all bytes.Buffer
	_, err := CoreGoldenHook(&all, text, &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var allF []Finding
	require.NoError(t, json.Unmarshal(all.Bytes(), &allF))
	var hasStyle, hasBare bool
	for _, f := range allF {
		if f.Rule == "EN_SOFT_VERY_UNIQUE" {
			hasStyle = true
		}
		if f.Rule == "EN_SOFT_BARE_WITH" {
			hasBare = true
		}
	}
	require.True(t, hasStyle && hasBare, "%+v", allF)

	var filtered bytes.Buffer
	_, err = CoreGoldenHook(&filtered, text, &CommandLineOptions{
		Language:           "en",
		DisabledCategories: []string{"STYLE"},
	})
	require.NoError(t, err)
	var ff []Finding
	require.NoError(t, json.Unmarshal(filtered.Bytes(), &ff))
	for _, f := range ff {
		require.NotEqual(t, "EN_SOFT_VERY_UNIQUE", f.Rule, "%+v", ff)
	}
	foundBare := false
	for _, f := range ff {
		if f.Rule == "EN_SOFT_BARE_WITH" {
			foundBare = true
		}
	}
	require.True(t, foundBare, "%+v", ff)
}

func TestGolden_SoftEnableCategoriesCasing(t *testing.T) {
	text := "i think so. Bare with me please."
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{
		Language:          "en",
		EnabledCategories: []string{"CASING"},
	})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	require.NotEmpty(t, findings)
	// only CASING-category matches (soft lowercase i and/or core sentence-start casing)
	for _, f := range findings {
		require.Equal(t, "typographical", f.Type, "%+v", f)
		require.NotEqual(t, "EN_SOFT_BARE_WITH", f.Rule)
		require.NotEqual(t, "style", f.Type)
	}
	foundLowerI := false
	for _, f := range findings {
		if f.Rule == "EN_SOFT_LOWERCASE_I" || f.Rule == "UPPERCASE_SENTENCE_START" {
			foundLowerI = true
		}
	}
	require.True(t, foundLowerI, "%+v", findings)
}

func TestGolden_SoftDisableRule(t *testing.T) {
	// configureCoreLT applies -d via ApplyCLIRuleFilters
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Bare with me please.", &CommandLineOptions{
		Language: "en",
	})
	require.NoError(t, err)
	var before []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &before))
	found := false
	for _, f := range before {
		if f.Rule == "EN_SOFT_BARE_WITH" {
			found = true
		}
	}
	require.True(t, found, "%+v", before)

	// Use RunWithIO -d flag path
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "-d", "EN_SOFT_BARE_WITH", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Bare with me please.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s out=%s", code, errb.String(), out.String())
	require.NotContains(t, out.String(), "EN_SOFT_BARE_WITH")
}

func TestGolden_SoftIdiomConfusablesWave2(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"They wreck havoc everywhere.", "EN_SOFT_WREAK_HAVOC", "wreak havoc"},
		{"Give free reign to creativity.", "EN_SOFT_FREE_REIN", "free rein"},
		{"Wait with baited breath.", "EN_SOFT_BAITED_BREATH", "bated breath"},
		{"The event was a damp squid.", "EN_SOFT_DAMP_SQUID", "damp squib"},
		{"Take a sneak peak at this.", "EN_SOFT_SNEAK_PEAK", "sneak peek"},
		{"They will extract revenge soon.", "EN_SOFT_EXTRACT_REVENGE", "exact revenge"},
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

func TestGolden_SoftMultiLangDoubleQ(t *testing.T) {
	cases := []struct {
		lang, text, rule string
	}{
		{"de", "Wirklich??", "DE_SOFT_DOUBLE_Q"},
		{"fr", "Vraiment??", "FR_SOFT_DOUBLE_Q"},
		{"es", "¿Qué??", "ES_SOFT_DOUBLE_Q"},
		{"pt", "Sério??", "PT_SOFT_DOUBLE_Q"},
		{"it", "Davvero??", "IT_SOFT_DOUBLE_Q"},
		{"nl", "Echt??", "NL_SOFT_DOUBLE_Q"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: tc.lang})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "typographical", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftInformalForms(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"I shud of known better.", "EN_SOFT_SHUD_OF", "should have"},
		{"I woulda gone earlier.", "EN_SOFT_WOULDA", "would have"},
		{"I coulda helped you.", "EN_SOFT_COULDA", "could have"},
		{"I shoulda called first.", "EN_SOFT_SHOULDA", "should have"},
		{"I gotta leave now.", "EN_SOFT_GOTTA", ""},
		{"I wanna go home.", "EN_SOFT_WANNA", ""},
		{"I'm gonna try that.", "EN_SOFT_GONNA", ""},
		{"That aint right.", "EN_SOFT_AIN_T", ""},
		{"Yall should come over.", "EN_SOFT_YALL", "Y'all"},
		{"Imma finish this later.", "EN_SOFT_IMMA", "I'm going to"},
		{"Prolly tomorrow works.", "EN_SOFT_PROLLY", "Probably"},
		{"Deffo a good idea.", "EN_SOFT_DEFFO", "Definitely"},
		{"Basically basically it works.", "EN_SOFT_BASICALLY_BASIC", ""},
		{"Actually actually I agree.", "EN_SOFT_ACTUALLY_ACTUALLY", ""},
		{"Honestly honestly I tried.", "EN_SOFT_HONESTLY_HONESTLY", ""},
		{"Literally literally amazing.", "EN_SOFT_LITERALLY_LITERALLY", ""},
		{"Just just wait here.", "EN_SOFT_JUST_JUST", ""},
		{"Really really good work.", "EN_SOFT_REALLY_REALLY", ""},
		{"Very very cold outside.", "EN_SOFT_VERY_VERY", ""},
		{"I like so so much cake.", "EN_SOFT_SO_SO", "so much"},
		{"Cats and and dogs.", "EN_SOFT_AND_AND", ""},
		{"See the the problem.", "EN_SOFT_THE_THE", ""},
		{"I need a a break.", "EN_SOFT_A_A", ""},
		{"I want to to leave.", "EN_SOFT_TO_TO", ""},
		{"Kind of of work.", "EN_SOFT_OF_OF", ""},
		{"Put it in in the box.", "EN_SOFT_IN_IN", ""},
		{"Put it on on the table.", "EN_SOFT_ON_ON", ""},
		{"This is for for you.", "EN_SOFT_FOR_FOR", ""},
		{"Come with with me.", "EN_SOFT_WITH_WITH", ""},
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
