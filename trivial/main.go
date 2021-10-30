package main

import (
	"fmt"
	"strings"
)

func main() {
	//log.WriteLog()
	tt()
}

func t1() {
	var b strings.Builder
	b.Grow(100)
	fmt.Println(b.Cap())
}

func transform(node Node) {
	n := node.(*BaseNode)
	n.Str()
	node.Str()
}

type Node interface {
	Str()
}

type BaseNode struct {
	Name string
	Uid string
}

func (bn *BaseNode) Str() {
	fmt.Println("BaseNode")
}

type RootNode struct {
	Name string
	Children [] Node
}

func (rn *RootNode) Str() {
	fmt.Println("RootNode", len(rn.Children))
}

func tt() {
	//n := &RootNode{Name: "root"}
	n1 := &BaseNode{Name: "base"}

	//transform(n)
	transform(n1)

}

