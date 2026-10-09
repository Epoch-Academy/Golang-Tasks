package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Contact 是一个"结构体":把一个联系人的相关信息(姓名、电话)捆成一个整体。
type Contact struct {
	Name  string
	Phone string
}

// String 是 Contact 的"方法":定义这个联系人怎么显示成一行字。
// 开头的 (c Contact) 表示:这个方法属于 Contact。
func (c Contact) String() string {
	return fmt.Sprintf("%s - %s", c.Name, c.Phone)
}

// AddressBook 通讯录:内部用一个 Contact 的 slice 存放所有联系人。
type AddressBook struct {
	contacts []Contact
}

// Add 添加一个联系人。
// 接收者写成 *AddressBook(带星号的指针),方法里的修改才能真正改到本体。
func (b *AddressBook) Add(name, phone string) {
	b.contacts = append(b.contacts, Contact{Name: name, Phone: phone})
}

// Find 按姓名查找。找到就返回联系人;找不到就返回一个 error。
// 返回 (结果, error) 是 Go 里最典型的写法。
func (b *AddressBook) Find(name string) (Contact, error) {
	for _, c := range b.contacts {
		if c.Name == name {
			return c, nil // nil 表示"没有错误"
		}
	}
	return Contact{}, fmt.Errorf("没有找到联系人:%s", name)
}

// List 列出所有联系人。
func (b *AddressBook) List() {
	if len(b.contacts) == 0 {
		fmt.Println("(通讯录是空的)")
		return
	}
	for i, c := range b.contacts {
		fmt.Printf("%d. %s\n", i+1, c.String())
	}
}

func main() {
	book := &AddressBook{} // 创建一个空通讯录(注意前面的 &)
	reader := bufio.NewScanner(os.Stdin)

	fmt.Println("== 简易通讯录 ==")
	for {
		fmt.Print("\n请选择:[1]添加 [2]列出 [3]查找 [4]退出 > ")
		if !reader.Scan() {
			break // 没有更多输入了,退出
		}
		choice := strings.TrimSpace(reader.Text())

		switch choice {
		case "1":
			fmt.Print("姓名:")
			reader.Scan()
			name := strings.TrimSpace(reader.Text())
			fmt.Print("电话:")
			reader.Scan()
			phone := strings.TrimSpace(reader.Text())
			book.Add(name, phone)
			fmt.Println("已添加。")
		case "2":
			book.List()
		case "3":
			fmt.Print("要查找的姓名:")
			reader.Scan()
			name := strings.TrimSpace(reader.Text())
			c, err := book.Find(name) // 经典的错误处理
			if err != nil {
				fmt.Println("查找失败:", err)
			} else {
				fmt.Println("找到了:", c.String())
			}
		case "4":
			fmt.Println("再见!")
			return
		default:
			fmt.Println("无效选择,请输入 1-4。")
		}
	}
}
