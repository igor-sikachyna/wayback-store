package main

import (
	"fmt"
	"wayback-store/chain"
	"wayback-store/store"
)

func main() {
	var c = chain.Chain{}
	var s = store.MakeStore(&c)

	s.Write("a", "hello")
	s.Write("b", "world")
	s.Write("c", "!")
	s.Write("z", "abc")
	s.Write("y", "def")
	s.Write("x", "ghi")
	fmt.Println(s.Get("a"))

	c.ProduceBlock()
	var s2 = store.MakeStore(&c)
	fmt.Println(s2.Get("hello"))
}
