package grpc

// NullAsEmpty ports ProtoHelper.nullAsEmpty for optional strings.
func NullAsEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// EmptyAsNull ports ProtoHelper.emptyAsNull.
func EmptyAsNull(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// CoalesceURL prefers matchURL over ruleURL (ProtoHelper.getUrl without RuleMatch type).
func CoalesceURL(matchURL, ruleURL string) string {
	if matchURL != "" {
		return matchURL
	}
	return ruleURL
}
