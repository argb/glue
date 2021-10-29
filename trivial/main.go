package main

import (
	"fmt"
	"strings"
)

func main() {
	//log.WriteLog()
	t1()
}

func t1() {
	var b strings.Builder
	b.Grow(100)
	fmt.Println(b.Cap())
}

