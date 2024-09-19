package extract_test

import (
	"slices"
	"testing"

	"deedles.dev/extract"
)

func TestList(t *testing.T) {
	var list *extract.List
	list = list.Push(5)
	list = list.Push(2)
	list = list.Push(3)
	if list.Len() != 3 {
		t.Fatal(list.Len())
	}
	if s := slices.Collect(list.All()); !slices.Equal(s, []any{3, 2, 5}) {
		t.Fatal(s)
	}
}

func TestCollectList(t *testing.T) {
	list := extract.CollectList(slices.Values([]int{3, 2, 5}))
	if list.Len() != 3 {
		t.Fatal(list.Len())
	}
	if s := slices.Collect(list.All()); !slices.Equal(s, []any{3, 2, 5}) {
		t.Fatal(s)
	}
}
