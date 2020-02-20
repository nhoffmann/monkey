package object

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func NewEnvironment() *Environment {
	store := make(map[string]Object)
	return &Environment{store: store, outer: nil}
}

type Environment struct {
	store map[string]Object
	outer *Environment
}

func (e *Environment) Get(name string) (Object, bool) {
	object, ok := e.store[name]
	if !ok && e.outer != nil {
		object, ok = e.outer.Get(name)
	}
	return object, ok
}

func (e *Environment) Set(name string, value Object) Object {
	e.store[name] = value
	return value
}
