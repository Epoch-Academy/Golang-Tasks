# 任务 4:结构体、方法与错误处理——做一个通讯录

> 本任务目标:学会把相关的数据**捆成一个整体**(结构体 struct),给它配上**动作**(方法),并正式掌握 Go 最有特色的**错误处理**方式。这是从"会写语法"迈向"会写程序"的关键一步。

---

## 一、先看看我们要做出什么

一个带菜单、能一直操作的通讯录:

```
== 简易通讯录 ==

请选择:[1]添加 [2]列出 [3]查找 [4]退出 > 1
姓名:张三
电话:13800138000
已添加。

请选择:[1]添加 [2]列出 [3]查找 [4]退出 > 2
1. 张三 - 13800138000

请选择:[1]添加 [2]列出 [3]查找 [4]退出 > 3
要查找的姓名:李四
查找失败: 没有找到联系人:李四

请选择:[1]添加 [2]列出 [3]查找 [4]退出 > 4
再见!
```

---

## 二、先把项目建好

```bash
mkdir task04-struct
cd task04-struct
go mod init gotask/task04-struct
```

---

## 三、结构体 `struct`:把数据捆成整体

一个联系人有姓名、有电话。如果用两个分开的变量,它们之间没有"关系",很容易乱。**结构体**能把它们捆成一个整体:

```go
type Contact struct {
	Name  string
	Phone string
}
```

- `type Contact struct {...}` 的意思是:定义一个**新类型**,叫 `Contact`,它里面有 `Name` 和 `Phone` 两个字段。
- 创建和使用:

```go
c := Contact{Name: "张三", Phone: "13800138000"}
fmt.Println(c.Name)  // 用"点"访问字段:张三
c.Phone = "111"      // 也可以修改
```

> 小细节:字段名 `Name`、`Phone` **首字母大写**。在 Go 里,首字母大写代表"对外公开"(别的文件/包能用到),小写代表"私有"。任务 5 把数据转成 JSON 时,这条规则会再次发挥作用。

---

## 四、方法:给结构体配上"动作"

**方法**就是"属于某个类型的函数"。比如让 `Contact` 自己知道怎么显示成一行字:

```go
func (c Contact) String() string {
	return fmt.Sprintf("%s - %s", c.Name, c.Phone)
}
```

- 和普通函数的区别就在开头多了 `(c Contact)`,叫**接收者**,表示"这个方法属于 `Contact`"。
- 调用时用"点":`c.String()`。
- `fmt.Sprintf` 和 `Printf` 很像,区别是它**不打印**,而是把格式化结果**作为字符串返回**。

> 顺带一提:方法名正好叫 `String()` 时,你直接 `fmt.Println(c)` 打印这个对象,Go 会**自动**调用它。算是个小福利。

---

## 五、指针:为什么有的方法带 `*`

看下面两个方法的接收者,一个带 `*`,一个不带,这不是手滑:

```go
func (c Contact) String() string         // 不带 *:只读取,不修改
func (b *AddressBook) Add(...)            // 带 *:要修改本体
```

用盒子打比方:
- **不带 `*`(值接收者)**:方法拿到的是本体的一份**复印件**。在复印件上改,不影响原件。适合"只读"的方法。
- **带 `*`(指针接收者)**:方法拿到的是本体的**地址**,能直接改到原件。**凡是要修改数据的方法,就用 `*`。**

我们的 `Add` 要往通讯录里加人(修改了数据),所以必须用 `*AddressBook`。

> 入门阶段记一条经验就够:**要改东西的方法用 `*`,只读的可以不用。** 拿不准时,统一用 `*` 也不会错。

---

## 六、Go 的错误处理:`(结果, error)` 和 `if err != nil`

这是 Go 最有特色、你**以后每天都要写**的部分。

很多语言用 `try / catch` 处理错误。**Go 不这样**:它让函数在返回正常结果的同时,**额外返回一个 `error`**,由你当场检查。

```go
func (b *AddressBook) Find(name string) (Contact, error) {
	for _, c := range b.contacts {
		if c.Name == name {
			return c, nil // 找到了:返回联系人 + nil(nil 表示"无错误")
		}
	}
	// 没找到:返回一个空联系人 + 一个错误
	return Contact{}, fmt.Errorf("没有找到联系人:%s", name)
}
```

调用方**必须当场检查**这个错误:

```go
c, err := book.Find("李四")
if err != nil {
	fmt.Println("查找失败:", err) // 出错了,处理它
} else {
	fmt.Println("找到了:", c)     // 没出错,正常用结果
}
```

记住这个万能句式:

```go
结果, err := 某个可能出错的操作()
if err != nil {
	// 出错时怎么办
}
// 没出错,继续往下
```

- `nil` 表示"空 / 没有",这里代表"没有错误"。
- `fmt.Errorf(...)` 用来造一个带说明文字的错误,`%s` 之类的占位符可以把变量拼进去。

> 这套 `if err != nil` 写起来略啰嗦,但好处是:**每一个可能失败的地方,你都被强制正视**,程序不会"偷偷出错"。这是 Go 工程质量的基石。

---

## 七、写代码

完整 `main.go`(综合了结构体、方法、指针、错误处理和前几个任务的菜单循环):

```go
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Contact struct {
	Name  string
	Phone string
}

func (c Contact) String() string {
	return fmt.Sprintf("%s - %s", c.Name, c.Phone)
}

type AddressBook struct {
	contacts []Contact
}

func (b *AddressBook) Add(name, phone string) {
	b.contacts = append(b.contacts, Contact{Name: name, Phone: phone})
}

func (b *AddressBook) Find(name string) (Contact, error) {
	for _, c := range b.contacts {
		if c.Name == name {
			return c, nil
		}
	}
	return Contact{}, fmt.Errorf("没有找到联系人:%s", name)
}

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
	book := &AddressBook{}
	reader := bufio.NewScanner(os.Stdin)

	fmt.Println("== 简易通讯录 ==")
	for {
		fmt.Print("\n请选择:[1]添加 [2]列出 [3]查找 [4]退出 > ")
		if !reader.Scan() {
			break
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
			c, err := book.Find(name)
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
```

---

## 八、运行

```bash
go run .
```

按菜单提示操作:添加几个人、列出、查找一个不存在的名字看看报错、最后退出。

---

## 九、(选学)关于拆分文件和"包"

现在所有代码挤在一个 `main.go` 里,文件越来越长。真实项目会**拆开**:比如把 `Contact`、`AddressBook` 放进单独的文件,甚至单独的"包(package)"。

入门阶段你只要知道:
- 同一个文件夹里可以有多个 `.go` 文件,它们同属一个包,能互相直接调用。
- 真正把代码拆成不同的包、互相 `import`,我们留到**任务 9** 做综合项目时再正式实践。

---

## 十、动手练习

1. 加一个 `Delete(name string)` 方法,按姓名删除联系人。提示:遍历找到下标后,用 `append(b.contacts[:i], b.contacts[i+1:]...)` 删除。
2. 给 `Add` 加个检查:如果姓名已存在,返回一个 `error`(让 `Add` 也返回 `error`),并在 `main` 里 `if err != nil` 处理。
3. 在 `Contact` 里再加一个字段(比如 `Email string`),让添加、显示都带上它。

---

## 十一、小结

- ✅ `struct`:把相关数据捆成一个新类型,用"点"访问字段
- ✅ 方法:属于某类型的函数,靠接收者 `(x 类型)` 声明
- ✅ 指针 `*`:要**修改**本体的方法就用它
- ✅ **错误处理**:函数返回 `(结果, error)`,调用方 `if err != nil` 检查——这是 Go 的灵魂
- ✅ 知道首字母大小写决定"公开/私有"

恭喜!到这里,Go 的**语言基础全部打通了**。从下一个任务开始,我们走进**真实世界**:让程序**联网**。**下一个任务**学习处理 JSON、用代码去**调用一个网络 API**,把别人的服务为你所用。

➡️ 下一站:`task05-http`
