package extract

import (
	"context"
	"iter"

	"deedles.dev/xsync"
)

// Runtime is the language's state. It tracks global data that is
// necessary throughout an Extract program, such as declared modules.
// A runtime is necessary to properly evaluate Extract code. To do so,
// use the context returned by a runtime's [Context] method.
type Runtime struct {
	ctx           context.Context
	modules       *xsync.Map[Atom, *Module]
	currentModule *Module
	locals        *localList
}

// New returns a runtime that has been initialized with the standard
// global state.
func New(ctx context.Context) *Runtime {
	r := Runtime{
		ctx:     ctx,
		modules: new(xsync.Map[Atom, *Module]),
		locals:  kernel,
	}
	for name, m := range std {
		r.modules.Store(name, m)
	}
	return &r
}

func (r *Runtime) All() iter.Seq2[Ident, any] {
	// TODO: Also provide module-level declarations.
	return r.locals.All()
}

func (r Runtime) WithContext(ctx context.Context) *Runtime {
	r.ctx = ctx
	return &r
}

func (r Runtime) Context() context.Context {
	return r.ctx
}

func (r Runtime) Let(ident Ident, val any) *Runtime {
	r.locals = r.locals.Push(ident, val)
	return &r
}

func (r Runtime) Lookup(ident Ident) (any, bool) {
	for id, val := range r.All() {
		if id == ident {
			return val, true
		}
	}
	return nil, false
}

// AddModule declares a new module with the given name. If the module
// already exists, it returns nil.
func (r *Runtime) AddModule(name Atom) *Module {
	m := Module{name: name}
	_, ok := r.modules.LoadOrStore(name, &m)
	if ok {
		return nil
	}
	return &m
}

// GetModule finds a declared module with the given name. If no such
// module has been declared, it returns nil.
func (r *Runtime) GetModule(name Atom) *Module {
	v, _ := r.modules.Load(name)
	return v
}

// Module is a basic building block of an Extract program. All
// declared functions must be declared inside of a module. Modules are
// identified by an atom and are global to a [Runtime] once they are
// declared.
type Module struct {
	name  Atom
	decls xsync.Map[Ident, any]
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
	return m.decls.Load(ident)
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
