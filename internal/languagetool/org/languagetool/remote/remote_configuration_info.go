package remote

import (
	"encoding/json"
	"fmt"
	"io"
)

// RemoteConfigurationInfo ports org.languagetool.remote.RemoteConfigurationInfo.
type RemoteConfigurationInfo struct {
	MaxTextLength int
	SoftwareInfo  map[string]any
	Rules         []map[string]string
}

// ParseRemoteConfigurationInfo reads a /v2/configinfo JSON body.
func ParseRemoteConfigurationInfo(r io.Reader) (*RemoteConfigurationInfo, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return ParseRemoteConfigurationInfoJSON(data)
}

// ParseRemoteConfigurationInfoJSON parses configinfo JSON bytes.
func ParseRemoteConfigurationInfoJSON(data []byte) (*RemoteConfigurationInfo, error) {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	info := &RemoteConfigurationInfo{
		SoftwareInfo: map[string]any{},
	}
	if sw, ok := raw["software"].(map[string]any); ok {
		info.SoftwareInfo = sw
	}
	if param, ok := raw["parameter"].(map[string]any); ok {
		switch v := param["maxTextLength"].(type) {
		case float64:
			info.MaxTextLength = int(v)
		case int:
			info.MaxTextLength = v
		case json.Number:
			n, _ := v.Int64()
			info.MaxTextLength = int(n)
		}
	}
	if rules, ok := raw["rules"].([]any); ok {
		for _, item := range rules {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			row := map[string]string{}
			for k, v := range m {
				row[k] = fmt.Sprint(v)
			}
			info.Rules = append(info.Rules, row)
		}
	}
	return info, nil
}

func (i *RemoteConfigurationInfo) GetSoftwareInfo() map[string]any {
	if i == nil {
		return nil
	}
	return i.SoftwareInfo
}

func (i *RemoteConfigurationInfo) GetMaxTextLength() int {
	if i == nil {
		return 0
	}
	return i.MaxTextLength
}

func (i *RemoteConfigurationInfo) GetRemoteRules() []map[string]string {
	if i == nil {
		return nil
	}
	return append([]map[string]string(nil), i.Rules...)
}
