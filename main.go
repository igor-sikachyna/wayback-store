package main

import (
	"fmt"
	"wayback-store/chain"
	"wayback-store/store"
)

func main() {
	var c = chain.Chain{}
	var s, err = store.MakeStore(&c)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	s.Write("a", "hello")
	s.Write("b", "world")
	c.ProduceBlock()
	s.Write("c", "!")
	s.Write("z", "abc")
	s.Write("y", "def")
	c.ProduceBlock()
	s.Write("x", "ghi")
	fmt.Println(s.Get("a"))

	c.ProduceBlock()
	s2, err := store.MakeStore(&c)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(s2.Get("x"))
}
