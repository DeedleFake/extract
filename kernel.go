package extract

import (
	"context"
	"errors"
	"reflect"
)

var kernel = func() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, Ident("defmodule"), EvalFunc(kernelDefModule))
	ctx = context.WithValue(ctx, Ident("def"), EvalFunc(kernelDef))
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

	pattern, ok := args.Head().(*List)
	if !ok {
		return NewTypeError(pattern, reflect.TypeFor[*List]()), ctx
	}

	m := GetModule(ctx)
	if m == nil {
		panic(errors.New("def used outside of module"))
	}

	panic("Not implemented.")
}
