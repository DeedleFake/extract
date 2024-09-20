package extract

import (
	"context"
	"sync"
)

type (
	runtimeKey struct{}
	moduleKey  struct{}
)

// Runtime is the language's state. It tracks global data that is
// necessary throughout an Extract program, such as declared modules.
// A runtime is necessary to properly evaluate Extract code. To do so,
// use the context returned by a runtime's [Context] method.
type Runtime struct {
	modules sync.Map // map[Atom]*Module
}

// NewRuntime returns a runtime that has been initialized with the
// standard global state.
func NewRuntime() *Runtime {
	var r Runtime
	for name, m := range std {
		r.modules.Store(name, m)
	}
	return &r
}

// GetRuntime returns the runtime associated with given context, or
// nil if none is found. Many pieces of the Extract system assume that
// a runtime will always be available in their provided contexts
// during evaluation.
func GetRuntime(ctx context.Context) *Runtime {
	r, _ := ctx.Value(runtimeKey{}).(*Runtime)
	return r
}

// Context returns the base context for the Runtime. This is usually
// what should be passed to top-level Extract code during evaluation.
func (r *Runtime) Context() context.Context {
	return context.WithValue(kernel, runtimeKey{}, r)
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
	v, ok := r.modules.Load(name)
	if !ok {
		return nil
	}
	return v.(*Module)
}

// Module is a basic building block of an Extract program. All
// declared functions must be declared inside of a module. Modules are
// identified by an atom and are global to a [Runtime] once they are
// declared.
type Module struct {
	name  Atom
	decls sync.Map // map[Ident]any
}

// GetModule gets the current module from the context. If the context
// does not have a current module, likely because the code being
// evaluated is outside of a module declaration, it returns nil.
func GetModule(ctx context.Context) *Module {
	m, _ := ctx.Value(moduleKey{}).(*Module)
	return m
}

// Context returns a context suitable for executing code inside of a
// module declaration using the provided context as a base.
func (m *Module) Context(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, moduleKey{}, m)
	for name, f := range m.decls.Range {
		ctx = context.WithValue(ctx, name, f)
	}
	return ctx
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
