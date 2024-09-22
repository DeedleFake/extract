package extract

import (
	"errors"
	"fmt"
	"reflect"
)

// kernel is the base scope containing the built-in, top-level
// functions.
var kernel = func() (ll *localList) {
	ll = ll.Push("defmodule", EvalFunc(kernelDefModule))
	ll = ll.Push("def", EvalFunc(kernelDef))
	ll = ll.Push("add", EvalFunc(kernelAdd))
	ll = ll.Push("sub", EvalFunc(kernelSub))
	return ll
}()

func kernelDefModule(r *Runtime, args *List) (*Runtime, any) {
	if args.Len() == 0 {
		return r, &ArgumentNumError{Num: args.Len(), Expected: -1}
	}

	name, ok := args.Head().(Atom)
	if !ok {
		return r, NewTypeError(name, reflect.TypeFor[Atom]())
	}

	m := r.AddModule(name)
	if m == nil {
		return r, fmt.Errorf("attempted to redeclare module %q", name)
	}
	mr := *r
	mr.currentModule = m
	body := args.Tail().Run(&mr)
	if err, ok := body.(error); ok {
		return r, err
	}
	return r, name
}

func kernelDef(r *Runtime, args *List) (*Runtime, any) {
	if args.Len() < 2 {
		return r, &ArgumentNumError{Num: args.Len(), Expected: -1}
	}

	m := r.currentModule
	if m == nil {
		return r, errors.New("def used outside of module")
	}

	var name Ident
	var f any
	switch pattern := args.Head().(type) {
	case Ident:
		name = pattern
		f = EvalFunc(func(fr *Runtime, args *List) (*Runtime, any) {
			if args.Len() != 0 {
				return fr, &ArgumentNumError{Num: args.Len(), Expected: 0}
			}

			return fr, args.Tail().Run(fr)
		})

	case *List:
		if pattern.Len() == 0 {
			return r, errors.New("function pattern list must contain at least one element")
		}

		name, _ = pattern.Head().(Ident)
		if name == "" {
			return r, NewTypeError(name, reflect.TypeFor[Ident]())
		}

		tail := pattern.Tail()
		params := make([]Ident, 0, tail.Len())
		for arg := range tail.All() {
			name, ok := arg.(Ident)
			if !ok {
				return r, NewTypeError(arg, reflect.TypeFor[Ident]())
			}
			params = append(params, name)
		}

		f = EvalFunc(func(fr *Runtime, fargs *List) (*Runtime, any) {
			if fargs.Len() != len(params) {
				return fr, &ArgumentNumError{Num: fargs.Len(), Expected: tail.Len()}
			}

			var i int
			for arg := range EvalAll(r, fargs.All()) {
				fr = fr.Let(params[i], arg)
				i++
			}

			return fr, args.Tail().Run(fr)
		})

	default:
		return r, NewTypeError(pattern, reflect.TypeFor[*List](), reflect.TypeFor[Ident]())
	}

	_, ok := m.decls.LoadOrStore(name, f)
	if ok {
		return r, fmt.Errorf("attempted to redeclare function %q", string(name))
	}
	return r, f
}

func kernelAdd(r *Runtime, args *List) (*Runtime, any) {
	if args.Len() < 2 {
		return r, &ArgumentNumError{Num: args.Len(), Expected: -1}
	}

	var total int64
	var totalf float64
	for arg := range EvalAll(r, args.All()) {
		switch arg := arg.(type) {
		case int64:
			total += arg
		case float64:
			totalf += arg
		case error:
			// TODO: Don't handle errors like this?
			return r, arg
		default:
			return r, NewTypeError(arg, reflect.TypeFor[int64](), reflect.TypeFor[float64]())
		}
	}

	if totalf != 0 {
		return r, float64(total) + totalf
	}
	return r, total
}

func kernelSub(r *Runtime, args *List) (*Runtime, any) {
	if args.Len() != 2 {
		return r, &ArgumentNumError{Num: args.Len(), Expected: 2}
	}

	_, first := Eval(r, args.Head(), nil)
	_, second := Eval(r, args.Tail().Head(), nil)

	var i int64
	var f float64
	switch a := first.(type) {
	case int64:
		i = a
	case float64:
		f = a
	default:
		return r, NewTypeError(a, reflect.TypeFor[int64](), reflect.TypeFor[float64]())
	}

	switch b := second.(type) {
	case int64:
		if f != 0 {
			return r, f - float64(b)
		}
		return r, i - b
	case float64:
		if i != 0 {
			return r, float64(i) - b
		}
		return r, f - b
	default:
		return r, NewTypeError(b, reflect.TypeFor[int64](), reflect.TypeFor[float64]())
	}
}
