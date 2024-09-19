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
	return new(Runtime)
}

func GetRuntime(ctx context.Context) *Runtime {
	r, _ := ctx.Value(runtimeKey{}).(*Runtime)
	return r
}

func (r *Runtime) Context() context.Context {
	return context.WithValue(context.TODO(), runtimeKey{}, r)
}

func (r *Runtime) AddModule(name Atom) *Module {
	m := Module{name: name}
	_, ok := r.modules.LoadOrStore(name, &m)
	if ok {
		panic(fmt.Errorf("module %q already registered with runtime", name))
	}
	return &m
}

func (r *Runtime) GetModule(name Atom) (*Module, bool) {
	v, _ := r.modules.Load(name)
	m, ok := v.(*Module)
	return m, ok
}

type Module struct {
	name Atom
}

func (m *Module) Name() Atom {
	return m.name
}

func Eval(ctx context.Context, expr any, args *List) (any, context.Context, error) {
	switch expr := expr.(type) {
	case Evaluator:
		return expr.Eval(ctx, args)
	case Valuer:
		r, err := expr.Value(ctx)
		return r, ctx, err
	default:
		return expr, ctx, nil
	}
}

type Evaluator interface {
	Eval(ctx context.Context, args *List) (any, context.Context, error)
}

type Valuer interface {
	Value(ctx context.Context) (any, error)
}
