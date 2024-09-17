// Package extract implements the core of the Extract language.
package extract

import "iter"

// List is a singly-linked list. It is the core building block of the
// language. Both a zero-value List and a nil *List are valid lists of
// length 0.
type List struct {
	head any
	tail *List
	len  int
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
