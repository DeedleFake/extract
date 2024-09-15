Extract
=======

_Note: This project is in heavy early development and many, if not all, features described below do not actually exist yet._

The Extract programming language is a procedural, statically-typed language that runs on the Go runtime. It includes features from Erlang, in particular pattern matching and process mailboxes. File structure follows Go conventions with projects being organized into modules containing packages, with, generally, one package per directory. The language is implemented via a transpiler that produces Go code.

Example
-------

As the language is still in early planning stages, this example is subject to change in backwards-incompatible ways.

```extract
package main

import (
  "fmt"
)

func add(parent pid) {
  select {
  case {:add, $a, $b}:
    parent <- {:result, a + b}
  }
}

func main() {
  $child = go add(self())
  child <- {:add, 1, 2}
  select {
  case {:result, $v}:
    fmt.Printf("Result: %v\n", v)
  }
}
```
