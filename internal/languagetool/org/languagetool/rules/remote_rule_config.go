package rules

import (
	"encoding/json"
	"io"
	"regexp"
)

const (
	RemoteDefaultPort                       = 443
	RemoteDefaultBaseTimeoutMS              = 1000
	RemoteDefaultTimeoutPerCharMS           = 0
	RemoteDefaultDownMS               int64 = 5000
	RemoteDefaultFailureRate                = 50.0
	RemoteDefaultSlidingWindowType          = "TIME_BASED"
	RemoteDefaultSlidingWindowSize          = 60
	RemoteDefaultMinimumNumberOfCalls       = 10

	RemoteOptionThirdPartyAI = "thirdPartyAI"
	RemoteOptionFallbackRule = "fallbackRuleId"
	RemoteOptionPremium      = "premium"
)

// RemoteRuleConfig ports org.languagetool.rules.RemoteRuleConfig.
type RemoteRuleConfig struct {
	RuleID                          string            `json:"ruleId"`
	URL                             string            `json:"url"`
	Port                            *int              `json:"port"`
	BaseTimeoutMilliseconds         int64             `json:"baseTimeoutMilliseconds"`
	TimeoutPerCharacterMilliseconds float64           `json:"timeoutPerCharacterMilliseconds"`
	DownMilliseconds                int64             `json:"downMilliseconds"`
	FailureRateThreshold            float64           `json:"failureRateThreshold"`
	SlidingWindowType               string            `json:"slidingWindowType"`
	SlidingWindowSize               int               `json:"slidingWindowSize"`
	MinimumNumberOfCalls            int               `json:"minimumNumberOfCalls"`
	Options                         map[string]string `json:"options"`
	Language                        string            `json:"language"`
	Type                            string            `json:"type"`
}

func NewRemoteRuleConfig() *RemoteRuleConfig {
	p := RemoteDefaultPort
	return &RemoteRuleConfig{
		Port:                            &p,
		BaseTimeoutMilliseconds:         RemoteDefaultBaseTimeoutMS,
		TimeoutPerCharacterMilliseconds: RemoteDefaultTimeoutPerCharMS,
		DownMilliseconds:                RemoteDefaultDownMS,
		FailureRateThreshold:            RemoteDefaultFailureRate,
		SlidingWindowType:               RemoteDefaultSlidingWindowType,
		SlidingWindowSize:               RemoteDefaultSlidingWindowSize,
		MinimumNumberOfCalls:            RemoteDefaultMinimumNumberOfCalls,
		Options:                         map[string]string{},
	}
}

func CopyRemoteRuleConfig(c *RemoteRuleConfig) *RemoteRuleConfig {
	if c == nil {
		return nil
	}
	out := *c
	if c.Port != nil {
		p := *c.Port
		out.Port = &p
	}
	if c.Options != nil {
		out.Options = make(map[string]string, len(c.Options))
		for k, v := range c.Options {
			out.Options[k] = v
		}
	}
	return &out
}

func (c *RemoteRuleConfig) GetRuleID() string { return c.RuleID }
func (c *RemoteRuleConfig) GetURL() string    { return c.URL }

// GetFallbackRuleId returns options["fallbackRuleId"] if set.
func (c *RemoteRuleConfig) GetFallbackRuleId() string {
	if c == nil || c.Options == nil {
		return ""
	}
	return c.Options[RemoteOptionFallbackRule]
}

// IsUsingThirdPartyAI ports isUsingThirdPartyAI (options thirdPartyAI == "true").
func (c *RemoteRuleConfig) IsUsingThirdPartyAI() bool {
	if c == nil || c.Options == nil {
		return false
	}
	return c.Options[RemoteOptionThirdPartyAI] == "true"
}

func (c *RemoteRuleConfig) GetPort() int {
	if c.Port != nil {
		return *c.Port
	}
	return RemoteDefaultPort
}
func (c *RemoteRuleConfig) GetFailureRateThreshold() float64 { return c.FailureRateThreshold }
func (c *RemoteRuleConfig) GetSlidingWindowType() string     { return c.SlidingWindowType }
func (c *RemoteRuleConfig) GetSlidingWindowSize() int        { return c.SlidingWindowSize }
func (c *RemoteRuleConfig) GetDownMilliseconds() int64       { return c.DownMilliseconds }
func (c *RemoteRuleConfig) GetBaseTimeoutMilliseconds() int64 {
	return c.BaseTimeoutMilliseconds
}

// GetRelevantConfig returns the first config with matching rule id.
func GetRelevantRemoteRuleConfig(rule string, configs []*RemoteRuleConfig) *RemoteRuleConfig {
	for _, c := range configs {
		if c != nil && c.GetRuleID() == rule {
			return c
		}
	}
	return nil
}

// IsRelevantRemoteRuleConfig matches type and optional language regex (full short code).
func IsRelevantRemoteRuleConfig(typ, langShortCodeWithVariant string) func(*RemoteRuleConfig) bool {
	return func(r *RemoteRuleConfig) bool {
		if r == nil || r.Type != typ {
			return false
		}
		if r.Language == "" {
			return true
		}
		ok, err := regexp.MatchString(r.Language, langShortCodeWithVariant)
		return err == nil && ok
	}
}

// ParseRemoteRuleConfigs parses a JSON array of RemoteRuleConfig (// comments not supported).
func ParseRemoteRuleConfigs(r io.Reader) ([]*RemoteRuleConfig, error) {
	var list []*RemoteRuleConfig
	dec := json.NewDecoder(r)
	if err := dec.Decode(&list); err != nil {
		return nil, err
	}
	for _, c := range list {
		if c == nil {
			continue
		}
		if c.Options == nil {
			c.Options = map[string]string{}
		}
		if c.Port == nil {
			p := RemoteDefaultPort
			c.Port = &p
		}
		if c.BaseTimeoutMilliseconds == 0 {
			c.BaseTimeoutMilliseconds = RemoteDefaultBaseTimeoutMS
		}
		if c.DownMilliseconds == 0 {
			c.DownMilliseconds = RemoteDefaultDownMS
		}
		if c.FailureRateThreshold == 0 {
			c.FailureRateThreshold = RemoteDefaultFailureRate
		}
		if c.SlidingWindowType == "" {
			c.SlidingWindowType = RemoteDefaultSlidingWindowType
		}
		if c.SlidingWindowSize == 0 {
			c.SlidingWindowSize = RemoteDefaultSlidingWindowSize
		}
		if c.MinimumNumberOfCalls == 0 {
			c.MinimumNumberOfCalls = RemoteDefaultMinimumNumberOfCalls
		}
	}
	return list, nil
}
