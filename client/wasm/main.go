//go:build js

package main

import (
	"syscall/js"

	"github.com/charlieplate/TinyHash/client/wasm/wasm"
)

func main() {
	c := make(chan struct{})

	hi := wasm.NewHashInstance()

	js.Global().Set("hashLoop", js.FuncOf(hi.HashLoop))
	js.Global().Set("cancelLoop", js.FuncOf(hi.CancelLoop))
	js.Global().Set("setCurrentMin", js.FuncOf(wasm.SetCurrentMin))

	<-c
}
