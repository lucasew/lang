package commandline

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEN_HASNT_and_waiting_rules_fire(t *testing.T) {
	opts := &CommandLineOptions{Language: "en"}
	lt, err := configureCoreLT("en", opts)
	require.NoError(t, err)
	ch := &CoreRulesChecker{Lang: "en", lt: lt}
	cases := []struct {
		text, wantID string
	}{
		{"They haven't bit their tongue.", "HASNT_IRREGULAR_VERB"},
		{"Tom hasn't send any message yet.", "HASNT_IRREGULAR_VERB"},
		{"These fly by night companies are not reliable.", "CA_FLY_BY_NIGHT"},
		{"I am waiting my patient finish the sample collection.", "WAITING_MY_PATIENT_FINISH"},
	}
	for _, tc := range cases {
		t.Run(tc.wantID+"/"+tc.text[:20], func(t *testing.T) {
			ms, err := ch.Check(tc.text)
			require.NoError(t, err)
			found := false
			var ids []string
			for _, m := range ms {
				if m == nil {
					continue
				}
				id := ruleIDOfMatch(m)
				ids = append(ids, id)
				if id == tc.wantID {
					found = true
				}
			}
			require.True(t, found, "want %s in %v for %q", tc.wantID, ids, tc.text)
		})
	}
}
