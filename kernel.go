package extract

import (
	"errors"
	"fmt"
	"reflect"
)

// kernel is the base scope containing the built-in, top-level
// functions.
var kernel = func() (ll *localList) {
	ll = ll.Push(MakeIdent("defmodule"), EvalFunc(kernelDefModule))
	ll = ll.Push(MakeIdent("def"), EvalFunc(kernelDef))
	ll = ll.Push(MakeIdent("add"), EvalFunc(kernelAdd))
	ll = ll.Push(MakeIdent("sub"), EvalFunc(kernelSub))
	return ll
}()

func kernelDefModule(env *Env, args *List) (*Env, any) {
	if args.Len() == 0 {
		return env, &ArgumentNumError{Num: args.Len(), Expected: -1}
	}

	name, ok := args.Head().(Atom)
	if !ok {
		return env, NewTypeError(name, reflect.TypeFor[Atom]())
	}

	m := env.AddModule(name)
	if m == nil {
		return env, fmt.Errorf("attempted to redeclare module %q", name)
	}
	mr := *env
	mr.currentModule = m
	body := args.Tail().Run(&mr)
	if err, ok := body.(error); ok {
		return env, err
	}
	return env, name
}

func kernelDef(env *Env, args *List) (*Env, any) {
	if args.Len() < 2 {
		return env, &ArgumentNumError{Num: args.Len(), Expected: -1}
	}

	m := env.currentModule
	if m == nil {
		return env, errors.New("def used outside of module")
	}

	var name Ident
	var f any
	switch pattern := args.Head().(type) {
	case Ident:
		name = pattern
		f = EvalFunc(func(fenv *Env, args *List) (*Env, any) {
			if args.Len() != 0 {
				return fenv, &ArgumentNumError{Num: args.Len(), Expected: 0}
			}

			return fenv, args.Tail().Run(fenv)
		})

	case Call:
		if pattern.Len() == 0 {
			return env, errors.New("function pattern list must contain at least one element")
		}

		n, ok := pattern.Head().(Ident)
		if !ok {
			return env, NewTypeError(name, reflect.TypeFor[Ident]())
		}
		name = n

		tail := pattern.Tail()
		params := make([]Ident, 0, tail.Len())
		for arg := range tail.All() {
			name, ok := arg.(Ident)
			if !ok {
				return env, NewTypeError(arg, reflect.TypeFor[Ident]())
			}
			params = append(params, name)
		}

		f = EvalFunc(func(fenv *Env, fargs *List) (*Env, any) {
			if fargs.Len() != len(params) {
				return fenv, &ArgumentNumError{Num: fargs.Len(), Expected: tail.Len()}
			}

			var i int
			for arg := range EvalAll(env, fargs.All()) {
				fenv = fenv.Let(params[i], arg)
				i++
			}

			return fenv, args.Tail().Run(fenv)
		})

	default:
		return env, NewTypeError(pattern, reflect.TypeFor[*List](), reflect.TypeFor[Ident]())
	}

	_, ok := m.decls.LoadOrStore(name, f)
	if ok {
		return env, fmt.Errorf("attempted to redeclare function %q", name)
	}
	return env, f
}

func kernelAdd(env *Env, args *List) (*Env, any) {
	if args.Len() < 2 {
		return env, &ArgumentNumError{Num: args.Len(), Expected: -1}
	}

	var total int64
	var totalf float64
	for arg := range EvalAll(env, args.All()) {
		switch arg := arg.(type) {
		case int64:
			total += arg
		case float64:
			totalf += arg
		case error:
			// TODO: Don't handle errors like this?
			return env, arg
		default:
			return env, NewTypeError(arg, reflect.TypeFor[int64](), reflect.TypeFor[float64]())
		}
	}

	if totalf != 0 {
		return env, float64(total) + totalf
	}
	return env, total
}

func kernelSub(env *Env, args *List) (*Env, any) {
	if args.Len() != 2 {
		return env, &ArgumentNumError{Num: args.Len(), Expected: 2}
	}

	_, first := Eval(env, args.Head(), nil)
	_, second := Eval(env, args.Tail().Head(), nil)

	var i int64
	var f float64
	switch a := first.(type) {
	case int64:
		i = a
	case float64:
		f = a
	default:
		return env, NewTypeError(a, reflect.TypeFor[int64](), reflect.TypeFor[float64]())
	}

	switch b := second.(type) {
	case int64:
		if f != 0 {
			return env, f - float64(b)
		}
		return env, i - b
	case float64:
		if i != 0 {
			return env, float64(i) - b
		}
		return env, f - b
	default:
		return env, NewTypeError(b, reflect.TypeFor[int64](), reflect.TypeFor[float64]())
	}
}
