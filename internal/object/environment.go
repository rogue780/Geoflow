package object

import "fmt"

// Binding represents a variable binding with mutability info.
type Binding struct {
	Value   Object
	Mutable bool
}

// Environment represents a scope with variable bindings.
type Environment struct {
	store map[string]*Binding
	outer *Environment
}

// NewEnvironment creates a new empty environment.
func NewEnvironment() *Environment {
	return &Environment{
		store: make(map[string]*Binding),
		outer: nil,
	}
}

// NewEnclosedEnvironment creates a new environment enclosed by an outer one.
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

// Get retrieves a binding by name, searching up the scope chain.
func (e *Environment) Get(name string) (Object, bool) {
	binding, ok := e.store[name]
	if ok {
		return binding.Value, true
	}
	if e.outer != nil {
		return e.outer.Get(name)
	}
	return nil, false
}

// Set creates a new binding in the current scope.
func (e *Environment) Set(name string, val Object, mutable bool) Object {
	e.store[name] = &Binding{Value: val, Mutable: mutable}
	return val
}

// Update updates an existing mutable binding, searching up the scope chain.
func (e *Environment) Update(name string, val Object) error {
	binding, ok := e.store[name]
	if ok {
		if !binding.Mutable {
			return fmt.Errorf("cannot reassign immutable binding '%s'", name)
		}
		binding.Value = val
		return nil
	}
	if e.outer != nil {
		return e.outer.Update(name, val)
	}
	return fmt.Errorf("undefined variable '%s'", name)
}
