package main

import (
	"fmt"
	"math/rand"
)

// 猜数字游戏:程序随机想一个 1~100 的数,你来猜,它提示大了还是小了。
func main() {
	// rand.Intn(100) 返回 0~99 的随机整数,+1 就变成 1~100。
	// Go 1.20 起随机数会自动初始化,不用再手动设置"种子"。
	target := rand.Intn(100) + 1

	count := 0 // 记录猜了几次

	fmt.Println("我想好了一个 1 到 100 之间的数,你来猜!")

	// for 后面什么条件都不写,就是"无限循环",直到 break 才停下来。
	for {
		fmt.Print("请输入你猜的数:")
		guess := readInt() // 调用我们自己写的函数,读取一个整数

		count++ // 次数加 1,等价于 count = count + 1

		// if / else if / else:从上往下依次判断,命中哪个就执行哪个。
		if guess < target {
			fmt.Println("小了,再大一点!")
		} else if guess > target {
			fmt.Println("大了,再小一点!")
		} else {
			fmt.Printf("恭喜你猜对了!答案就是 %d,你一共猜了 %d 次。\n", target, count)
			break // 跳出循环,游戏结束
		}
	}
}

// readInt 是我们自己定义的函数:读取用户输入的一个整数并返回。
// 函数名后面 () 里是"参数"(这里不需要,所以是空的),
// () 后面的 int 表示这个函数会"返回"一个整数。
func readInt() int {
	var n int
	fmt.Scanln(&n)
	return n // 把读到的数,交回给调用它的地方
}
