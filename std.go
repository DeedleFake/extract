package extract

import (
	"context"
	"reflect"
	"strings"
)

var std = map[Atom]*Module{
	NewAtom("String"): stdString(),
}

func stdString() *Module {
	m := Module{name: NewAtom("String")}
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

	return &m
}
