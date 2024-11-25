package main

import (
	"fmt"
	"wayback-store/chain"
	"wayback-store/store"
)

func main() {
	var c = chain.Chain{}
	var s = store.MakeStore(&c)

	s.Write("hello", "world")
	fmt.Println(s.Get("hello"))

	c.ProduceBlock()
	var s2 = store.MakeStore(&c)
	fmt.Println(s2.Get("hello"))
}
