package monitor

import "fmt"

type VMMonitor struct {
	instructions [] string

}

var monitor *VMMonitor

func (mn *VMMonitor) ShowInstructions() {
	for index, instruction := range mn.instructions {
		fmt.Printf("%d, %s",index, instruction)
	}
}

func (mn *VMMonitor) AddInstruction(instruction string) *VMMonitor {
	mn.instructions = append(mn.instructions, instruction)

	return mn
}

// SingletonNew /**
/*
只需要一个监视器的实例
 */
func SingletonNew() *VMMonitor {
	if monitor == nil {
		monitor = new(VMMonitor)
	}
	return monitor
}
