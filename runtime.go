package extract

import (
	"context"
	"fmt"
	"sync"
)

type (
	runtimeKey struct{}
	moduleKey  struct{}
)

type Runtime struct {
	modules sync.Map // map[Atom]*Module
}

func NewRuntime() *Runtime {
	var r Runtime
	for name, m := range std {
		r.modules.Store(name, m)
	}
	return &r
}

func GetRuntime(ctx context.Context) *Runtime {
	r, _ := ctx.Value(runtimeKey{}).(*Runtime)
	return r
}

func (r *Runtime) Context() context.Context {
	return context.WithValue(kernel, runtimeKey{}, r)
}

func (r *Runtime) AddModule(name Atom) *Module {
	m := Module{name: name}
	_, ok := r.modules.LoadOrStore(name, &m)
	if ok {
		panic(fmt.Errorf("module %q already registered with runtime", name))
	}
	return &m
}

func (r *Runtime) GetModule(name Atom) *Module {
	v, ok := r.modules.Load(name)
	if !ok {
		return nil
	}
	return v.(*Module)
}

type Module struct {
	name  Atom
	decls sync.Map // map[Ident]any
}

func GetModule(ctx context.Context) *Module {
	m, _ := ctx.Value(moduleKey{}).(*Module)
	return m
}

func (m *Module) Context(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, moduleKey{}, m)
	for name, f := range m.decls.Range {
		ctx = context.WithValue(ctx, name, f)
	}
	return ctx
}

func (m *Module) Name() Atom {
	return m.name
}

func (m *Module) Lookup(ident Ident) (any, bool) {
	return m.decls.Load(ident)
}
