Extract
=======

_Note: This project is in heavy early development and many, if not all, features described below do not actually exist yet._

Extract is a functional, dynamically-typed scripting language inspired by Lisp and Elixir and running on top of the Go runtime. It has Erlang-like concurrency features and good interaction with Go.

Example
-------

As the language is still in early planning stages, this example is subject to change in backwards-incompatible ways.

```extract
(defmodule Example
    (defwhen (fib n) (lte? n 1) n)

    (def (fib n) (+
        (fib (- n 1))
        (fib (- n 2))
    ))
)

(IO.println (fib 5))
```
