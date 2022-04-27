package main

import (
	"fmt"
)

type A struct {
	noCopy func()
	name   string
}

func main() {
	var a1, a2 A
	a1 = a2
	fmt.Println(a1)
}
