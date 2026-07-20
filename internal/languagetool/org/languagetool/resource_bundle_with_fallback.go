package languagetool

import (
	"hash/fnv"
	"sort"
)

// MessageBundle is a minimal string-key message lookup (Java ResourceBundle surface).
type MessageBundle map[string]string

func (b MessageBundle) GetString(key string) string {
	if b == nil {
		return ""
	}
	return b[key]
}

// Keys returns keys present in the bundle (unordered).
func (b MessageBundle) Keys() []string {
	if b == nil {
		return nil
	}
	out := make([]string, 0, len(b))
	for k := range b {
		out = append(out, k)
	}
	return out
}

// ResourceBundleWithFallback ports org.languagetool.ResourceBundleWithFallback.
type ResourceBundleWithFallback struct {
	Bundle   MessageBundle
	Fallback MessageBundle
}

func NewResourceBundleWithFallback(bundle, fallback MessageBundle) *ResourceBundleWithFallback {
	return &ResourceBundleWithFallback{Bundle: bundle, Fallback: fallback}
}

// GetString ports handleGetObject: use bundle value unless trim-empty, then fallback.
// Note: Java ResourceBundle.getString throws MissingResourceException if absent;
// Go map returns "" which is treated as empty → fallback (practical twin).
func (r *ResourceBundleWithFallback) GetString(key string) string {
	if r == nil {
		return ""
	}
	s := r.Bundle.GetString(key)
	if stringsTrimSpaceEmpty(s) {
		return r.Fallback.GetString(key)
	}
	return s
}

// GetKeys ports getKeys — keys of the primary bundle only.
func (r *ResourceBundleWithFallback) GetKeys() []string {
	if r == nil {
		return nil
	}
	return r.Bundle.Keys()
}

// Equal ports equals (bundle + fallbackBundle identity by content map equality).
func (r *ResourceBundleWithFallback) Equal(o *ResourceBundleWithFallback) bool {
	if r == o {
		return true
	}
	if r == nil || o == nil {
		return false
	}
	return messageBundleEqual(r.Bundle, o.Bundle) && messageBundleEqual(r.Fallback, o.Fallback)
}

// Hash ports hashCode: 31 * hash(bundle) + hash(fallback).
func (r *ResourceBundleWithFallback) Hash() uint64 {
	if r == nil {
		return 0
	}
	h := messageBundleHash(r.Bundle)
	return 31*h + messageBundleHash(r.Fallback)
}

func messageBundleEqual(a, b MessageBundle) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func messageBundleHash(b MessageBundle) uint64 {
	h := fnv.New64a()
	keys := make([]string, 0, len(b))
	for k := range b {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		_, _ = h.Write([]byte(k))
		_, _ = h.Write([]byte{0})
		_, _ = h.Write([]byte(b[k]))
		_, _ = h.Write([]byte{0})
	}
	return h.Sum64()
}

func stringsTrimSpaceEmpty(s string) bool {
	for _, r := range s {
		if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
			return false
		}
	}
	return true
}
