package extract

import (
	"context"
	"iter"

	"deedles.dev/xsync"
)

var moduleIdent = MakeIdent("$module")

// Env is the language's state. It tracks global data that is
// necessary throughout an Extract program, such as declared modules.
// A runtime is necessary to properly evaluate Extract code. To do so,
// use the context returned by a runtime's [Context] method.
type Env struct {
	ctx           context.Context
	modules       *xsync.Map[Atom, *Module]
	currentModule *Module
	locals        *localList
}

// New returns a runtime that has been initialized with the standard
// global state.
func New(ctx context.Context) *Env {
	r := Env{
		ctx:     ctx,
		modules: new(xsync.Map[Atom, *Module]),
		locals:  kernel,
	}
	for name, m := range std {
		r.modules.Store(name, m)
	}
	return &r
}

func (env *Env) All() iter.Seq2[Ident, any] {
	return func(yield func(Ident, any) bool) {
		for ident, val := range env.locals.All() {
			switch ident {
			case moduleIdent:
				for ident, val := range env.currentModule.decls {
					if !yield(ident, val) {
						return
					}
				}
			default:
				if !yield(ident, val) {
					return
				}
			}
		}
	}
}

func (env Env) WithContext(ctx context.Context) *Env {
	env.ctx = ctx
	return &env
}

func (env Env) Context() context.Context {
	return env.ctx
}

func (env Env) Let(ident Ident, val any) *Env {
	env.locals = env.locals.Push(ident, val)
	return &env
}

func (env Env) Lookup(ident Ident) (any, bool) {
	for id, val := range env.All() {
		if id == ident {
			return val, true
		}
	}
	return nil, false
}

// AddModule declares a new module with the given name. If the module
// already exists, it returns nil.
func (env *Env) AddModule(name Atom) *Module {
	m := Module{name: name, decls: make(map[Ident]any)}
	_, ok := env.modules.LoadOrStore(name, &m)
	if ok {
		return nil
	}
	return &m
}

// GetModule finds a declared module with the given name. If no such
// module has been declared, it returns nil.
func (env *Env) GetModule(name Atom) *Module {
	v, _ := env.modules.Load(name)
	return v
}

func (env Env) withCurrentModule(m *Module) *Env {
	env.currentModule = m
	env.locals = env.locals.Push(moduleIdent, nil)
	return &env
}

// Module is a basic building block of an Extract program. All
// declared functions must be declared inside of a module. Modules are
// identified by an atom and are global to a [Env] once they are
// declared.
type Module struct {
	name  Atom
	decls map[Ident]any
}

// Name returns the name of the module.
func (m *Module) Name() Atom {
	return m.name
}

// Lookup returns the value associated with the given identifier
// inside of the module. If nothing with the given identifier has been
// declared in the module, it returns false as the second return
// value.
func (m *Module) Lookup(ident Ident) (any, bool) {
	v, ok := m.decls[ident]
	return v, ok
}

type localList struct {
	ident Ident
	val   any
	next  *localList
}

func (ll *localList) Push(ident Ident, val any) *localList {
	return &localList{
		ident: ident,
		val:   val,
		next:  ll,
	}
}

func (ll *localList) All() iter.Seq2[Ident, any] {
	return func(yield func(Ident, any) bool) {
		for ll != nil {
			if !yield(ll.ident, ll.val) {
				return
			}
			ll = ll.next
		}
	}
}
