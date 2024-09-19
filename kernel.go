package extract

import (
	"context"
	"reflect"
	"strings"
)

var kernel = func() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, Ident("to_upper"), EvalFunc(kernelToUpper))
	return ctx
}()

func kernelToUpper(ctx context.Context, args *List) (any, context.Context) {
	if args.Len() != 1 {
		return &ArgumentNumError{Num: args.Len(), Expected: 1}, ctx
	}

	head, _ := Eval(ctx, args.Head(), nil)
	str, ok := head.(string)
	if !ok {
		return NewTypeError(head, reflect.TypeFor[string]()), ctx
	}

	return strings.ToUpper(str), ctx
}
