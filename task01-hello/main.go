package main

import "fmt"

func main() {
	fmt.Println("你好,Go!")
	fmt.Println("这是我的第一个 Go 程序,它会先打印文字,再和你简单聊两句。")

	const thisYear = 2026

	var name string
	fmt.Print("请输入你的名字:")
	fmt.Scanln(&name)

	var birthYear int
	fmt.Print("请输入你的出生年份(比如 2007):")
	fmt.Scanln(&birthYear)

	age := thisYear - birthYear

	fmt.Printf("你好,%s!\n", name)
	fmt.Printf("%d 年出生的你,今年大约 %d 岁。\n", birthYear, age)
}
