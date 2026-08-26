package tools

// MostlySingularMultiMap ports org.languagetool.tools.MostlySingularMultiMap.
// Values are stored as T for single entries or []T for multiples.
type MostlySingularMultiMap[K comparable, V any] struct {
	m map[K]any
}

// NewMostlySingularMultiMap builds from a map of key → value lists.
func NewMostlySingularMultiMap[K comparable, V any](contents map[K][]V) *MostlySingularMultiMap[K, V] {
	m := make(map[K]any, len(contents))
	for k, vals := range contents {
		if len(vals) == 1 {
			m[k] = vals[0]
		} else if len(vals) > 1 {
			cp := append([]V(nil), vals...)
			m[k] = cp
		}
	}
	return &MostlySingularMultiMap[K, V]{m: m}
}

// GetList returns the values for key, or nil if absent.
func (m *MostlySingularMultiMap[K, V]) GetList(key K) []V {
	if m == nil || m.m == nil {
		return nil
	}
	o, ok := m.m[key]
	if !ok {
		return nil
	}
	switch v := o.(type) {
	case []V:
		return v
	case V:
		return []V{v}
	default:
		return nil
	}
}
