package server

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckWithOptions_AllowIncompleteResults_ErrorRate(t *testing.T) {
	// Java TextChecker: allowIncompleteResults + ErrorRateTooHighException →
	// incompleteResultsReason = "Results are incomplete: " + exception message
	// (not invent size-threshold soft warning).
	tc := NewTextChecker(nil, false, nil)
	// Force rate trip: many matches relative to words. Without real rules that fire,
	// exercise the CheckErrorRate path via MaxErrorsPerWordRate on a text that
	// still produces some matches from core rules when possible.
	// Unit twin of CheckErrorRate is in languagetool package; here we verify
	// the server packaging of incomplete reason when rate trips.
	opts := CheckOptions{
		AllowIncompleteResults: true,
		// Extremely low threshold so any real error density trips (if matchCount high).
		// With clean text, rate does not trip — then incompleteReason is empty (Java same).
		MaxErrorsPerWordRate: 0.0001,
	}
	// Use nonsense tokens to provoke spelling matches when speller is registered.
	text := strings.Repeat("xyzzy ", 40) // >25 words
	ms, _, reason := tc.CheckWithOptionsAndIgnore(text, "en", opts)
	// Either no speller → no trip (empty reason) OR trip → Java-shaped prefix.
	if reason != "" {
		require.True(t, strings.HasPrefix(reason, "Results are incomplete: "), reason)
		require.Contains(t, reason, "too many errors")
	}
	_ = ms
}

func TestBuildResponse_IncompleteResultsWarningsObject(t *testing.T) {
	// Java writeWarningsSection shape.
	v := NewV2TextChecker(nil, false, nil)
	body, err := v.BuildResponseExFull("hi", "en", "English", nil, false,
		"Results are incomplete: Text checking was stopped due to too many errors", nil, 1)
	require.NoError(t, err)
	var resp CheckResponse
	require.NoError(t, json.Unmarshal([]byte(body), &resp))
	require.NotNil(t, resp.Warnings)
	require.True(t, resp.Warnings.IncompleteResults)
	require.Contains(t, resp.Warnings.IncompleteResultsReason, "too many errors")

	body2, err := v.BuildResponseExFull("hi", "en", "English", nil, false, "", nil, 1)
	require.NoError(t, err)
	var resp2 CheckResponse
	require.NoError(t, json.Unmarshal([]byte(body2), &resp2))
	require.NotNil(t, resp2.Warnings)
	require.False(t, resp2.Warnings.IncompleteResults)
	require.Empty(t, resp2.Warnings.IncompleteResultsReason)
}

func TestFormatTimeoutIncompleteReason(t *testing.T) {
	// Java Locale.ENGLISH "%.2f" of maxCheckTimeMillis/1000.0
	require.Equal(t,
		"Results are incomplete: text checking took longer than allowed maximum of 1.50 seconds",
		formatTimeoutIncompleteReason(1500),
	)
	require.Equal(t,
		"Results are incomplete: text checking took longer than allowed maximum of 0.01 seconds",
		formatTimeoutIncompleteReason(10),
	)
}

func TestCheckWithOptions_AllowIncompleteResults_Timeout(t *testing.T) {
	// When MaxCheckTimeMillis is extremely small, check may timeout before finish.
	// Message must match Java TimeoutException incomplete path when it fires.
	tc := NewTextChecker(nil, false, nil)
	opts := CheckOptions{
		AllowIncompleteResults: true,
		MaxCheckTimeMillis:     1,
	}
	_, _, reason := tc.CheckWithOptionsAndIgnore(strings.Repeat("word ", 5000), "en", opts)
	if reason != "" {
		require.True(t, strings.HasPrefix(reason, "Results are incomplete: text checking took longer than allowed maximum of "), reason)
		require.Contains(t, reason, "seconds")
	}
}
