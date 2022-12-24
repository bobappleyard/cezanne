Cezanne
=======

Cezanne is a simple programming language to experiment with algebraic
effects.

It currently has a very simple compiler in Go, which generates Scheme
code. An interpreter in Go is under construction, which will
hopefully provide a platform to experiment with optimising an
explicit-stack implementation of effect handling.

The ultimate aim is to generate e.g. WASM and have a runtime in Rust
or something like that. That's probably a bit of a way off though.
