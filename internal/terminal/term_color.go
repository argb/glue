package terminal

import (
	"fmt"
	"strings"
)

const START_SET = "\\x1b["

const (
FG_RED = "31"
FG_GREEN = "32"
FG_WHITE = "37"
FG_BLUE = "34"

)
const END_SET = "\x1b[0m"

func Set(color string) {
	ctlStrArr := []string{"\u001B[", color, "m"}
	ctrStr := strings.Join(ctlStrArr, "")
	fmt.Print(ctrStr)
}
func Unset() {
	fmt.Print(END_SET)
}

func TT() {
	fmt.Print("\x1b[32m")//设置颜色样式
	fmt.Print("Hello World")//打印文本内容
	fmt.Println("\x1b[0m")//样式结束符,清楚之前的显示属性
}