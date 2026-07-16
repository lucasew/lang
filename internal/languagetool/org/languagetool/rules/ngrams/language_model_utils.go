package ngrams

import "fmt"

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
