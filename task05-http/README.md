# 任务 5:JSON 与调用网络 API——让程序联网

> 本任务目标:让程序走出"本机",去**调用别人提供的网络服务**。你会学到两样真实开发离不开的东西:**JSON**(数据交换格式)和 **HTTP 请求**。这也是后面 gin、gorm 的基础。

---

## 一、先看看我们要做出什么

程序联网,从一个免费接口拉一句随机的话打印出来:

```
「岁月不饶人,我亦未曾饶过岁月。」
    —— 《岁月》
```

(每次运行内容都不同。)

---

## 二、先把项目建好

```bash
mkdir task05-http
cd task05-http
go mod init gotask/task05-http
```

---

## 三、先搞懂两个词:API 和 JSON

**API**(应用程序接口)你可以理解成"**别人程序对外开的一个服务窗口**"。你按它规定的方式发一个请求(比如访问一个网址),它就返回数据给你。天气、汇率、地图、大模型……几乎都是通过 API 提供的。

**JSON** 是这些数据最常用的**格式**,长这样:

```json
{
	"hitokoto": "岁月不饶人,我亦未曾饶过岁月。",
	"from": "《岁月》"
}
```

它就是一组"**键: 值**"——是不是和任务 4 的 `map`、任务 5 的 `struct` 很像?没错,我们的工作就是把这段 JSON 文本,**装进一个 Go 结构体**里来使用。

---

## 四、JSON 与结构体:`struct` 标签

我们定义一个结构体来"接住"返回的数据:

```go
type Hitokoto struct {
	Text string `json:"hitokoto"`
	From string `json:"from"`
}
```

注意字段后面反引号 `` ` `` 里的 `json:"hitokoto"`,这叫**标签(tag)**:

> JSON 里的字段名是 `hitokoto`(小写),但 Go 结构体字段必须**大写**才能被 JSON 工具读写(还记得任务 5 说的"大写=公开"吗)。
> 于是用标签做个"对应关系":告诉 Go——"JSON 的 `hitokoto`,请放进我的 `Text` 字段"。

---

## 五、发起请求:`net/http`

标准库 `net/http` 直接能发请求,最简单的是 `http.Get`:

```go
resp, err := http.Get("https://v1.hitokoto.cn/")
if err != nil {
	// 网络不通、网址错误等,都会走到这里
}
```

返回两个值:`resp`(响应)和 `err`(错误)。**联网随时可能失败**(断网、超时、对方挂了),所以这里的 `if err != nil` 不是走形式,是真的需要。

---

## 六、`defer` 和读取响应内容

拿到响应后要做两件事:

```go
defer resp.Body.Close()        // 用完后关闭,释放资源
body, err := io.ReadAll(resp.Body) // 把响应内容全部读成一段文本(字节)
```

**`defer` 是什么?**

> `defer` 后面跟的那行代码,会"**推迟到当前函数结束时才执行**"。
> 这里用它来保证 `resp.Body.Close()` 一定会被调用——不管函数从哪儿返回,资源都能被正确关闭。这是 Go 里非常常见的"善后"写法。

---

## 七、解析 JSON:`json.Unmarshal`

把读到的 JSON 文本(`body`)装进结构体:

```go
var h Hitokoto
err := json.Unmarshal(body, &h) // 注意 &h:把数据填进 h 这个盒子
```

`Unmarshal` 是"解组/反序列化"的意思——把 JSON 文本变成 Go 能直接用的结构体。成功后,`h.Text`、`h.From` 就有值了。(注意又出现了 `&`,道理和任务 2 的 `Scanln` 一样:告诉它"填到哪个盒子"。)

---

## 八、写代码

完整 `main.go`:

```go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Hitokoto struct {
	Text string `json:"hitokoto"`
	From string `json:"from"`
}

func main() {
	url := "https://v1.hitokoto.cn/"

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("请求失败(检查一下网络):", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取响应失败:", err)
		return
	}

	var h Hitokoto
	if err := json.Unmarshal(body, &h); err != nil {
		fmt.Println("解析 JSON 失败:", err)
		return
	}

	fmt.Printf("「%s」\n    —— %s\n", h.Text, h.From)
}
```

---

## 九、运行

```bash
go run .
```

如果联网正常,你会看到一句随机的话。多跑几次,每次都不同。

> 如果报"请求失败",先检查网络;公司/学校网络偶尔会拦截,可以换个网络再试。**这正好说明:写联网程序时,错误处理是必需的。**

---

## 十、(选学)进阶:调用 AI 大模型

调用大模型(LLM)和上面调一言**本质是同一回事**:发 HTTP 请求、收 JSON、解析。区别是:大模型要用 `POST`、要带上你的 **API Key**、请求体也是一段 JSON。

最关键的一条安全规矩:**API Key 绝对不能写死在代码里,更不能提交到 GitHub**。正确做法是放进**环境变量**,代码用 `os.Getenv` 读:

```go
apiKey := os.Getenv("AI_API_KEY") // 从环境变量读取,代码里看不到真实 key
if apiKey == "" {
	fmt.Println("请先设置环境变量 AI_API_KEY")
	return
}
```

运行前先设置环境变量(以一个 OpenAI 兼容接口为例):

```bash
# Windows PowerShell:
$env:AI_API_KEY="你的key"
# Git Bash / Linux / Mac:
export AI_API_KEY="你的key"
```

发 POST 请求的核心写法(片段,供你拼成完整程序去试):

```go
reqBody := `{"model":"gpt-4o-mini","messages":[{"role":"user","content":"你好"}]}`
req, _ := http.NewRequest("POST", "https://你的接口地址/v1/chat/completions",
	strings.NewReader(reqBody))
req.Header.Set("Content-Type", "application/json")
req.Header.Set("Authorization", "Bearer "+apiKey) // 把 key 放在请求头里
resp, err := http.DefaultClient.Do(req)
// ……后面读取 resp.Body、json.Unmarshal,和本任务一模一样
```

这是**选学**,不影响主线。等你学完后面的内容,可以回来把它拼成一个完整的命令行 AI 问答工具。

---

## 十一、动手练习

1. 把接口换成 `https://v1.hitokoto.cn/?c=a`(`c=a` 表示只要"动画"分类),观察结果变化。
2. 给 `Hitokoto` 结构体再加一个字段去接收 JSON 里的 `creator`(作者),并打印出来。
3. 连续调用 3 次,打印 3 句话。提示:用任务 3 学的 `for` 循环。

---

## 十二、小结

- ✅ 理解了 **API**(别人的服务窗口)和 **JSON**(数据格式)
- ✅ 用 `struct` + `json:"..."` 标签接住返回数据
- ✅ 用 `http.Get` 发请求,并认真处理 `err`
- ✅ 用 `defer` 善后、`io.ReadAll` 读内容、`json.Unmarshal` 解析
- ✅ (选学)知道调用大模型的套路,以及 **API Key 要走环境变量**

你已经能让程序"消费"别人的 API 了。真实后端还经常要**同时处理很多事**,并且在超时或取消时及时停下来。**下一个任务**,我们学习 goroutine、channel 和 context。

➡️ 下一站:`task06-concurrency`
