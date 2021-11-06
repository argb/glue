这是关于这个项目的说明。
这是一句废话。

Glue is a toy language for learning.

麻雀虽小，五脏俱全，嗯。。。。。。实际上并不全，不过一些基础语法还是具备的。


用法：
./glue ./examples/t2.gl

可执行文件glue是glue语言的解释器。

举个栗子：
fn getAdder(seed){
    let add = fn(n){
        return seed+n
    }
    return add;
}

let add = getAdder(10);
let n = add(10)

print(n)

结果：20

gendot 是一个生成可视化ast的工具
用法：
./glue ./examples/t2.gl
会生成 ast.dot文件，自己安装一下graphviz, 然后 dot -Tpng ast.dot -o ast.png 就可以得到ast.png文件了。

可以把glue的源文件翻译成dot文件，然后用graphviz生成图形。
示例：
https://github.com/argb/glue/blob/master/ast.png
https://github.com/argb/glue/blob/master/tools/ast.png
