package extract

import (
	"context"
	"iter"
	"slices"
	"sync"
)

// List is a singly-linked list. It is the core building block of the
// language. Both a zero-value List and a nil *List are valid lists of
// length 0.
type List struct {
	head any
	tail *List
	len  int
}

func ListOf(vals ...any) (list *List) {
	for _, v := range slices.Backward(vals) {
		list = list.Push(v)
	}
	return list
}

var listPool sync.Pool

func CollectList[T any](seq iter.Seq[T]) (list *List) {
	s, _ := listPool.Get().(*[]any)
	if s == nil {
		s = new([]any)
	}
	defer func() {
		clear(*s)
		*s = (*s)[:0]
		listPool.Put(&s)
	}()

	anys := func(yield func(any) bool) {
		for v := range seq {
			if !yield(v) {
				return
			}
		}
	}
	*s = slices.AppendSeq(*s, anys)
	return ListOf((*s)...)
}

// Head returns the value at the head of the list. In other words, the
// value of the this node in the linked list.
func (list *List) Head() any {
	if list == nil {
		return nil
	}
	return list.head
}

// Push pushes an element onto the list, effectively prepending it. It
// returns the node representing the new list that is formed.
//
// Note that the old list is still valid, but unmodified.
func (list *List) Push(val any) *List {
	return &List{
		head: val,
		tail: list,
		len:  list.Len() + 1,
	}
}

func PushAll[T any](list *List, seq iter.Seq[T]) *List {
	for v := range seq {
		list = list.Push(v)
	}
	return list
}

// Tail returns the tail of the list.
func (list *List) Tail() *List {
	if list == nil || list.tail == nil {
		return nil
	}
	if list.tail.len == list.len-1 {
		return list.tail
	}

	return &List{
		head: list.tail,
		tail: list.tail.tail,
		len:  list.len - 1,
	}
}

// Len returns the length of the list. Each node caches the length, so
// this operation is O(1) despite the linked list nature of the
// implementation.
func (list *List) Len() int {
	if list == nil {
		return 0
	}
	return list.len
}

// All returns an iterator over the values stored in the list.
func (list *List) All() iter.Seq[any] {
	return func(yield func(any) bool) {
		cur := list
		for cur.Len() > 0 {
			if !yield(cur.head) {
				return
			}
			cur = cur.Tail()
		}
	}
}

func (list *List) Eval(ctx context.Context, args *List) (any, context.Context) {
	if list.Len() == 0 {
		return list, ctx
	}

	return Eval(ctx, list.Head(), list.Tail())
}

// Run runs a list like it's the body of a function. If any elements
// of the list return an error when evaluated, this function returns
// early with that error. Otherwise, it returns the result of the
// evaluation of the last element of the list.
func (list *List) Run(ctx context.Context) (r any) {
	for v := range list.All() {
		r, ctx = Eval(ctx, v, nil)
		if err, ok := r.(error); ok {
			return err
		}
	}
	return r
}
