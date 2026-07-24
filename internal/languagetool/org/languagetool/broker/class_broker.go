package broker

import "fmt"

// ClassBroker ports org.languagetool.broker.ClassBroker as a name→factory registry
// (Go has no Class.forName; callers register constructors by qualified name).
type ClassBroker interface {
	// ForName returns a zero-value or prototype for the registered name.
	ForName(qualifiedName string) (any, error)
}

// MapClassBroker is a simple registry.
type MapClassBroker struct {
	factories map[string]func() any
}

func NewMapClassBroker() *MapClassBroker {
	return &MapClassBroker{factories: map[string]func() any{}}
}

// Register associates qualifiedName with a factory.
func (b *MapClassBroker) Register(qualifiedName string, factory func() any) {
	b.factories[qualifiedName] = factory
}

func (b *MapClassBroker) ForName(qualifiedName string) (any, error) {
	f, ok := b.factories[qualifiedName]
	if !ok {
		return nil, fmt.Errorf("class not found: %s", qualifiedName)
	}
	return f(), nil
}

// DefaultClassBroker is an empty registry (Java default uses Class.forName).
func NewDefaultClassBroker() *MapClassBroker {
	return NewMapClassBroker()
}
