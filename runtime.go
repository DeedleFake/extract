package extract

import (
	"context"
	"fmt"
	"sync"
)

type runtimeKey struct{}

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

func (m *Module) Name() Atom {
	return m.name
}

func (m *Module) Lookup(ident Ident) (any, bool) {
	return m.decls.Load(ident)
}

func Eval(ctx context.Context, expr any, args *List) (any, context.Context) {
	switch expr := expr.(type) {
	case Evaluator:
		return expr.Eval(ctx, args)
	default:
		if args.Len() > 0 {
			expr = args.Push(expr)
		}
		return expr, ctx
	}
}

type Evaluator interface {
	Eval(ctx context.Context, args *List) (any, context.Context)
}

type EvalFunc func(ctx context.Context, args *List) (any, context.Context)

func (f EvalFunc) Eval(ctx context.Context, args *List) (any, context.Context) {
	return f(ctx, args)
}
