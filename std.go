package extract

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

// std is the Extract standard library in the form of a map of module
// names to modules.
var std = map[Atom]*Module{
	MakeAtom("String"): stdString(),
}

func stdString() *Module {
	m := Module{name: MakeAtom("String")}
	m.decls.Store(Ident("to_upper"), EvalFunc(func(ctx context.Context, args *List) (any, context.Context) {
		if args.Len() != 1 {
			return &ArgumentNumError{Num: args.Len(), Expected: 1}, ctx
		}

		head, _ := Eval(ctx, args.Head(), nil)
		str, ok := head.(string)
		if !ok {
			return NewTypeError(head, reflect.TypeFor[string]()), ctx
		}

		return strings.ToUpper(str), ctx
	}))
	m.decls.Store(Ident("to_lower"), EvalFunc(func(ctx context.Context, args *List) (any, context.Context) {
		if args.Len() != 1 {
			return &ArgumentNumError{Num: args.Len(), Expected: 1}, ctx
		}

		head, _ := Eval(ctx, args.Head(), nil)
		str, ok := head.(string)
		if !ok {
			return NewTypeError(head, reflect.TypeFor[string]()), ctx
		}

		return strings.ToLower(str), ctx
	}))
	m.decls.Store(Ident("format"), EvalFunc(func(ctx context.Context, args *List) (any, context.Context) {
		if args.Len() == 0 {
			return &ArgumentNumError{Num: args.Len(), Expected: -1}, ctx
		}

		head, _ := Eval(ctx, args.Head(), nil)
		str, ok := head.(string)
		if !ok {
			return NewTypeError(head, reflect.TypeFor[string]()), ctx
		}

		verbs := slices.Collect(EvalAll(ctx, args.Tail().All()))
		return fmt.Sprintf(str, verbs...), ctx
	}))

	return &m
}
