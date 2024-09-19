package extract

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
)

var kernel = func() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, Ident("defmodule"), EvalFunc(kernelDefModule))
	ctx = context.WithValue(ctx, Ident("def"), EvalFunc(kernelDef))
	ctx = context.WithValue(ctx, Ident("add"), EvalFunc(kernelAdd))
	ctx = context.WithValue(ctx, Ident("sub"), EvalFunc(kernelSub))
	return ctx
}()

func kernelDefModule(ctx context.Context, args *List) (any, context.Context) {
	if args.Len() == 0 {
		return &ArgumentNumError{Num: args.Len(), Expected: -1}, ctx
	}

	name, ok := args.Head().(Atom)
	if !ok {
		return NewTypeError(name, reflect.TypeFor[Atom]()), ctx
	}

	runtime := GetRuntime(ctx)
	if runtime == nil {
		panic(errors.New("no runtime in context"))
	}

	m := runtime.AddModule(name)
	r := args.Tail().Run(m.Context(ctx))
	if err, ok := r.(error); ok {
		return err, ctx
	}
	return name, ctx
}

func kernelDef(ctx context.Context, args *List) (any, context.Context) {
	if args.Len() < 2 {
		return &ArgumentNumError{Num: args.Len(), Expected: -1}, ctx
	}

	m := GetModule(ctx)
	if m == nil {
		panic(errors.New("def used outside of module"))
	}

	var name Ident
	var f any
	switch pattern := args.Head().(type) {
	case Ident:
		name = pattern
		f = EvalFunc(func(fctx context.Context, args *List) (any, context.Context) {
			if args.Len() != 0 {
				return &ArgumentNumError{Num: args.Len(), Expected: 0}, fctx
			}

			return args.Tail().Run(m.Context(fctx)), fctx
		})

	case *List:
		if pattern.Len() == 0 {
			return errors.New("function pattern list must contain at least one element"), ctx
		}

		name, _ = pattern.Head().(Ident)
		if name == "" {
			return NewTypeError(name, reflect.TypeFor[Ident]()), ctx
		}

		tail := pattern.Tail()
		for arg := range tail.All() {
			if _, ok := arg.(Ident); !ok {
				return NewTypeError(arg, reflect.TypeFor[Ident]()), ctx
			}
		}

		f = EvalFunc(func(fctx context.Context, fargs *List) (any, context.Context) {
			if fargs.Len() != tail.Len() {
				return &ArgumentNumError{Num: fargs.Len(), Expected: tail.Len()}, fctx
			}

			evalargs := slices.Collect(EvalAll(ctx, fargs.All()))
			ectx := m.Context(fctx)
			var i int
			for name := range tail.All() {
				ectx = context.WithValue(ectx, name, evalargs[i])
				i++
			}

			return args.Tail().Run(ectx), fctx
		})

	default:
		return NewTypeError(pattern, reflect.TypeFor[*List](), reflect.TypeFor[Ident]()), ctx
	}

	_, ok := m.decls.LoadOrStore(name, f)
	if ok {
		return fmt.Errorf("attempted to redeclare function %q", string(name)), ctx
	}
	return f, ctx
}

func kernelAdd(ctx context.Context, args *List) (any, context.Context) {
	if args.Len() < 2 {
		return &ArgumentNumError{Num: args.Len(), Expected: -1}, ctx
	}

	var total int64
	var totalf float64
	for arg := range EvalAll(ctx, args.All()) {
		switch arg := arg.(type) {
		case int64:
			total += arg
		case float64:
			totalf += arg
		case error:
			return arg, ctx
		default:
			return NewTypeError(arg, reflect.TypeFor[int64](), reflect.TypeFor[float64]()), ctx
		}
	}

	if totalf != 0 {
		return float64(total) + totalf, ctx
	}
	return total, ctx
}

func kernelSub(ctx context.Context, args *List) (any, context.Context) {
	if args.Len() != 2 {
		return &ArgumentNumError{Num: args.Len(), Expected: 2}, ctx
	}

	first, _ := Eval(ctx, args.Head(), nil)
	second, _ := Eval(ctx, args.Tail().Head(), nil)

	var i int64
	var f float64
	switch a := first.(type) {
	case int64:
		i = a
	case float64:
		f = a
	default:
		return NewTypeError(a, reflect.TypeFor[int64](), reflect.TypeFor[float64]()), ctx
	}

	switch b := second.(type) {
	case int64:
		if f != 0 {
			return f - float64(b), ctx
		}
		return i - b, ctx
	case float64:
		if i != 0 {
			return float64(i) - b, ctx
		}
		return f - b, ctx
	default:
		return NewTypeError(b, reflect.TypeFor[int64](), reflect.TypeFor[float64]()), ctx
	}
}
