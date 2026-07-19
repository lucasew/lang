package ngrams

import "fmt"

// GetContextForTerm ports LanguageModelUtils.getContext(token, tokens, term, toLeft, toRight).
func GetContextForTerm(pos int, tokens []GoogleToken, term string, toLeft, toRight int) []string {
	return GetContextGoogleTokens(pos, tokens, term, toLeft, toRight)
}

// Get3gramProbabilityFor ports LanguageModelUtils.get3gramProbabilityFor for a single-token candidate.
// Multi-token candidates (Google split >2) return 0 (Java warns and returns 0).
func Get3gramProbabilityFor(lm LanguageModel, pos int, tokens []GoogleToken, term string, tokenize TokenizerFunc) float64 {
	if lm == nil || pos < 0 || pos >= len(tokens) {
		return 0
	}
	if tokenize == nil {
		// treat term as one token
		return get3gramProbabilityForTokens(lm, pos, tokens, []GoogleToken{NewGoogleToken(term, 0, len(term))})
	}
	newTokens := GetGoogleTokens(term, false, tokenize)
	return get3gramProbabilityForTokens(lm, pos, tokens, newTokens)
}

func get3gramProbabilityForTokens(lm LanguageModel, pos int, tokens, newTokens []GoogleToken) float64 {
	var ngram3Left, ngram3Middle, ngram3Right Probability
	switch len(newTokens) {
	case 1:
		term := newTokens[0].Token
		ngram3Left = lm.GetPseudoProbability(GetContextGoogleTokens(pos, tokens, term, 0, 2))
		ngram3Middle = lm.GetPseudoProbability(GetContextGoogleTokens(pos, tokens, term, 1, 1))
		ngram3Right = lm.GetPseudoProbability(GetContextGoogleTokens(pos, tokens, term, 2, 0))
	case 2:
		// e.g. you're -> you 're
		ctxL := getContextMulti(pos, tokens, newTokens, 0, 1)
		ctxR := getContextMulti(pos, tokens, newTokens, 1, 0)
		ngram3Left = lm.GetPseudoProbability(ctxL)
		ngram3Right = lm.GetPseudoProbability(ctxR)
		ngram3Middle = NewProbabilitySimple((ngram3Left.GetProb()+ngram3Right.GetProb())/2, 1.0)
	default:
		// Java: not supported yet → 0
		return 0
	}
	if ngram3Left.GetCoverage() < MinCoverage && ngram3Middle.GetCoverage() < MinCoverage && ngram3Right.GetCoverage() < MinCoverage {
		return 0
	}
	return ngram3Left.GetProb() * ngram3Middle.GetProb() * ngram3Right.GetProb()
}

// Get4gramProbabilityFor ports LanguageModelUtils.get4gramProbabilityFor for 1–2 token candidates.
func Get4gramProbabilityFor(lm LanguageModel, pos int, tokens []GoogleToken, term string, tokenize TokenizerFunc) float64 {
	if lm == nil || pos < 0 || pos >= len(tokens) {
		return 0
	}
	var newTokens []GoogleToken
	if tokenize == nil {
		newTokens = []GoogleToken{NewGoogleToken(term, 0, len(term))}
	} else {
		newTokens = GetGoogleTokens(term, false, tokenize)
	}
	return get4gramProbabilityForTokens(lm, pos, tokens, newTokens)
}

func get4gramProbabilityForTokens(lm LanguageModel, pos int, tokens, newTokens []GoogleToken) float64 {
	var n4L, n4ML, n4MR, n4R Probability
	switch len(newTokens) {
	case 1:
		n4L = lm.GetPseudoProbability(getContextMulti(pos, tokens, newTokens, 0, 3))
		n4ML = lm.GetPseudoProbability(getContextMulti(pos, tokens, newTokens, 2, 1))
		n4MR = lm.GetPseudoProbability(getContextMulti(pos, tokens, newTokens, 1, 2))
		n4R = lm.GetPseudoProbability(getContextMulti(pos, tokens, newTokens, 3, 0))
	case 2:
		n4L = lm.GetPseudoProbability(getContextMulti(pos, tokens, newTokens, 0, 2))
		n4ML = lm.GetPseudoProbability(getContextMulti(pos, tokens, newTokens, 1, 1))
		n4MR = n4ML // Java TODO: is this okay?
		n4R = lm.GetPseudoProbability(getContextMulti(pos, tokens, newTokens, 2, 0))
	default:
		return 0
	}
	if n4L.GetCoverage() < MinCoverage && n4ML.GetCoverage() < MinCoverage &&
		n4MR.GetCoverage() < MinCoverage && n4R.GetCoverage() < MinCoverage {
		return 0
	}
	// Java: Math.exp(sum of log probs)
	return n4L.GetProb() * n4ML.GetProb() * n4MR.GetProb() * n4R.GetProb()
}

// getContextMulti builds context around tokens[pos] substituting newTokens (multi-token center).
func getContextMulti(pos int, tokens, newTokens []GoogleToken, toLeft, toRight int) []string {
	if pos < 0 || pos >= len(tokens) {
		return nil
	}
	ctx := GetContextAtIndex(pos, tokens, newTokens, toLeft, toRight, GoogleToken.IsWhitespace, NewGoogleToken(".", 0, 0))
	out := make([]string, len(ctx))
	for i, t := range ctx {
		out[i] = t.Token
	}
	return out
}

// GetContextStrings ports LanguageModelUtils.getContext for string token lists:
// around token in tokens, substitute newToken, take toLeft/toRight neighbors.
func GetContextStrings(token string, tokens []string, newToken string, toLeft, toRight int) []string {
	pos := -1
	for i, t := range tokens {
		if t == token {
			pos = i
			break
		}
	}
	if pos < 0 {
		panic(fmt.Sprintf("Token not found: '%s' in tokens %v", token, tokens))
	}
	var left []string
	for i := pos - 1; i >= 0 && len(left) < toLeft; i-- {
		left = append([]string{tokens[i]}, left...)
	}
	var right []string
	for i := pos + 1; i < len(tokens) && len(right) < toRight; i++ {
		right = append(right, tokens[i])
	}
	out := append(append([]string{}, left...), newToken)
	return append(out, right...)
}

// GetContextAtIndex builds context around tokens[pos], substituting newTokens in the center.
func GetContextAtIndex[T any](pos int, tokens []T, newTokens []T, toLeft, toRight int, isWhitespace func(T) bool, endToken T) []T {
	if pos < 0 || pos >= len(tokens) {
		panic(fmt.Sprintf("Token index out of range: %d (len %d)", pos, len(tokens)))
	}
	var result []T
	for i, added := 1, 0; added < toLeft; i++ {
		if pos-i < 0 {
			var early []T
			for j := 0; j < pos; j++ {
				if !isWhitespace(tokens[j]) {
					early = append(early, tokens[j])
				}
			}
			return append(early, newTokens...)
		}
		cand := tokens[pos-i]
		if isWhitespace(cand) {
			continue
		}
		result = append([]T{cand}, result...)
		added++
	}
	result = append(result, newTokens...)
	for i, added := 1, 0; added < toRight; i++ {
		if pos+i >= len(tokens) {
			result = append(result, endToken)
			break
		}
		cand := tokens[pos+i]
		if isWhitespace(cand) {
			continue
		}
		result = append(result, cand)
		added++
	}
	return result
}

// GetContextGoogleTokens ports getContext for GoogleToken by index of the focus token.
func GetContextGoogleTokens(pos int, tokens []GoogleToken, newToken string, toLeft, toRight int) []string {
	newToks := []GoogleToken{NewGoogleToken(newToken, 0, len(newToken))}
	ctx := GetContextAtIndex(pos, tokens, newToks, toLeft, toRight, GoogleToken.IsWhitespace, NewGoogleToken(".", 0, 0))
	out := make([]string, len(ctx))
	for i, t := range ctx {
		out[i] = t.Token
	}
	return out
}
