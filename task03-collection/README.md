# 任务 3:slice、map 与字符串——做一个词频统计

> 本任务目标:学会管理"一堆"数据。`slice` 是有序的列表,`map` 是键值对字典——这是 Go 里最常用的两种数据结构。我们用它们做一个统计英文单词出现次数的小工具。

---

## 一、先看看我们要做出什么

```
请输入一段英文(单词用空格隔开),然后回车:
the cat sat on the mat the cat

一共 8 个单词,其中不同的单词 5 种:
  the          3 次
  cat          2 次
  sat          1 次
  on           1 次
  mat          1 次
```

---

## 二、先把项目建好

```bash
mkdir task03-collection
cd task03-collection
go mod init gotask/task03-collection
```

---

## 三、slice:有序的列表

前面我们的变量一次只装一个值。但现实中常要装"一串"东西——比如一串单词、一串分数。这时用 **slice**。

```go
fruits := []string{"苹果", "香蕉"}  // 创建一个装 string 的 slice,里面有两个

fruits = append(fruits, "橘子")     // append 往后面追加,现在有 3 个
fmt.Println(fruits[0])             // 用下标取值:苹果(下标从 0 开始!)
fmt.Println(len(fruits))           // len 取长度:3

// for range 遍历:i 是下标,f 是对应的值
for i, f := range fruits {
	fmt.Println(i, f)
}
```

要点:
- `[]string` 表示"一串字符串";`[]int` 就是"一串整数"。
- **下标从 0 开始**:第一个是 `fruits[0]`,第二个是 `fruits[1]`。
- `append` 用来加元素,`len` 取长度。

---

## 四、map:键值对字典

**map** 像一本字典:通过"键(key)"快速找到"值(value)"。比如用单词当键、用次数当值。

```go
age := make(map[string]int)  // 创建一个"键是 string、值是 int"的 map

age["小明"] = 19             // 存:键"小明"对应值 19
fmt.Println(age["小明"])     // 取:19
age["小明"]++                // 让"小明"对应的值 +1,变成 20

// 判断某个键存不存在:第二个返回值 ok 是 true/false
score, ok := age["小红"]
if !ok {
	fmt.Println("没有小红的记录")
}
_ = score

// 遍历 map:每次拿到一个键和它的值
for name, a := range age {
	fmt.Println(name, a)
}
```

要点:
- 用 `make(map[键类型]值类型)` 创建。
- `m[键] = 值` 存,`m[键]` 取。
- 一个很方便的特性:对**不存在的键**做 `m[key]++`,它会**自动从 0 开始**再加 1。我们的词频统计就靠这个。

---

## 五、读取一整行:`bufio.Scanner`

之前用的 `fmt.Scanln` 读到**空格就停**,没法读一整句话。要读一整行(含空格),用 `bufio.Scanner`:

```go
scanner := bufio.NewScanner(os.Stdin) // os.Stdin 代表"键盘输入"
scanner.Scan()                        // 读一行
line := scanner.Text()                // 取出读到的内容
```

这几行现在照抄即可,意思是"从键盘读一整行存进 `line`"。

---

## 六、字符串处理:`strings` 包

`strings` 包提供了一堆处理文字的工具,常用的:

| 函数 | 作用 | 例子 |
|---|---|---|
| `strings.ToLower(s)` | 转小写 | `"ABC"` → `"abc"` |
| `strings.ToUpper(s)` | 转大写 | `"abc"` → `"ABC"` |
| `strings.Fields(s)` | 按空白切成单词 slice | `"a b c"` → `["a","b","c"]` |
| `strings.Split(s, ",")` | 按指定符号切分 | `"a,b"` → `["a","b"]` |
| `strings.Contains(s, "x")` | 是否包含某段文字 | 返回 `true`/`false` |
| `strings.TrimSpace(s)` | 去掉首尾空白 | `" a "` → `"a"` |

我们的词频统计会用到 `ToLower`(统一大小写)和 `Fields`(切单词)。

---

## 七、写代码

新建 `main.go`:

```go
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// 统计一段英文里每个单词出现的次数。
func main() {
	fmt.Println("请输入一段英文(单词用空格隔开),然后回车:")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	line := scanner.Text()

	line = strings.ToLower(line)   // 统一转小写
	words := strings.Fields(line)  // 切成单词 slice

	counts := make(map[string]int) // 单词 -> 次数

	for _, word := range words {
		counts[word]++
	}

	fmt.Printf("\n一共 %d 个单词,其中不同的单词 %d 种:\n", len(words), len(counts))

	for word, n := range counts {
		fmt.Printf("  %-12s %d 次\n", word, n)
	}
}
```

---

## 八、运行

```bash
go run .
```

输入一句英文(比如 `the cat sat on the mat the cat`),回车,看统计结果。

---

## 九、关于 map 的顺序(重要)

多运行几次你会发现:**每次打印出来的单词顺序都不一样**。

这不是 bug——**Go 的 map 是无序的**,故意设计成这样,所以你不能依赖它的遍历顺序。如果想按固定顺序(比如按次数从高到低)输出,需要额外用 `sort` 包排序。这是**选学**练习,刚入门先知道"map 无序"这个事实就够了。

---

## 十、动手练习

1. 把切分用的 `strings.Fields` 换成你自己造的句子,试试不同输入。
2. 统计完后,额外打印"出现次数最多的单词是哪个"。提示:遍历 `counts`,用一个变量记住目前见过的最大次数。
3. (选学)用 `sort` 包,让输出按次数从多到少排列,做成一个真正的"热词榜"。

---

## 十一、小结

- ✅ `slice`:有序列表,会用 `append`、下标、`len`、`for range`
- ✅ `map`:键值字典,会用 `make` 创建、存取、`m[key]++` 计数、判断键是否存在
- ✅ 用 `bufio.Scanner` 读取一整行输入
- ✅ 用 `strings` 包处理文字(`ToLower`、`Fields` 等)
- ✅ 知道 **map 是无序的**

到这里,Go 的"基本功"——变量、判断、循环、函数、列表、字典——你都有了。**下一个任务**,我们学习怎么把相关的数据**捆成一个整体**(结构体),给它配上**方法**,并正式学习 Go 最有特色的**错误处理**。我们会把一个**通讯录**做得像模像样。

➡️ 下一站:`task04-struct`
