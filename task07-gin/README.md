# 任务 7:用 gin 写一个 Web 服务

> 本任务目标:换个身份——前面我们"调用别人的 API",这次让**你自己的程序变成一个 Web 后端**,对外提供接口给别人调用。我们用 Go 最流行的 Web 框架 **gin**,做一个待办事项(Todo)的增删查接口。

---

## 一、先看看我们要做出什么

一个一直运行的服务,别人(浏览器、手机 App、或命令行 `curl`)可以这样和它交互:

```
# 新增一条待办
$ curl -X POST http://localhost:8080/todos -H "Content-Type: application/json" -d '{"title":"学习Go","done":false}'
{"id":1,"title":"学习Go","done":false}

# 查看全部
$ curl http://localhost:8080/todos
[{"id":1,"title":"学习Go","done":false}]

# 删除第 1 条
$ curl -X DELETE http://localhost:8080/todos/1
{"message":"已删除"}
```

---

## 二、先把项目建好

```bash
mkdir task07-gin
cd task07-gin
go mod init gotask/task07-gin
```

---

## 三、先理解:什么是"Web 后端"

前面的程序都是"跑一下、出结果、结束"。**Web 服务不一样:它启动后会一直运行、一直等着**,谁来请求就回应谁。

- 它**监听**一个"端口"(本任务用 `8080`),好比一个一直开着的服务窗口。
- 别人通过 `http://地址:端口/路径` 来访问不同的功能。
- `/todos` 这种路径 + `GET`/`POST` 这种方法的组合,叫**路由(route)**。

---

## 四、安装第三方库:`go get`

gin 不是 Go 自带的,要从网上下载。这就是 `go.mod` 模块系统大显身手的地方:

```bash
go get github.com/gin-gonic/gin
```

它会下载 gin 及其依赖,并自动记录到 `go.mod` 里。

> **下载卡住怎么办?**(国内常见)
> - 确认用了国内代理:`go env -w GOPROXY=https://goproxy.cn,direct`
> - 如果卡在 `sum.golang.org`(校验库)连不上,可以关掉校验:`go env -w GOSUMDB=off`
> 设置一次即可,以后所有项目都生效。

---

## 五、gin 的最小服务

gin 的代码非常直观:

```go
r := gin.Default()                 // 创建一个 gin 引擎
r.GET("/hello", func(c *gin.Context) {
	c.JSON(200, gin.H{"msg": "你好"}) // 返回一段 JSON
})
r.Run(":8080")                     // 在 8080 端口启动,开始监听
```

- `r.GET("/hello", 处理函数)`:有人用 GET 访问 `/hello` 时,就执行这个处理函数。
- `c *gin.Context`:代表"这一次请求与响应"。读请求、写响应都靠它。
- `c.JSON(...)`:把数据作为 JSON 返回。`gin.H{}` 是 gin 提供的"随手造一个 JSON 对象"的快捷写法。

---

## 六、RESTful:用不同的 HTTP 方法操作资源

操作同一类资源(todos),用**不同的 HTTP 方法**表达不同意图,这套约定叫 REST:

| 方法 | 路径 | 含义 |
|---|---|---|
| `GET` | `/todos` | 查询全部 |
| `POST` | `/todos` | 新增一条 |
| `DELETE` | `/todos/:id` | 删除指定 id 的那条 |

`:id` 是**路径参数**——访问 `/todos/5` 时,`id` 就是 `5`。

---

## 七、读取请求里的数据

两种最常用的读取方式:

```go
// 1) 读路径参数 :id
id := c.Param("id")  // 字符串,需要的话用 strconv.Atoi 转成数字

// 2) 把请求体里的 JSON 解析进结构体
var input Todo
if err := c.ShouldBindJSON(&input); err != nil {
	// 客户端发来的 JSON 格式不对
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	return
}
```

> `http.StatusOK`、`http.StatusCreated`、`http.StatusBadRequest` 是标准库给 HTTP 状态码起的好记名字,分别是 200、201、400。用它们比直接写数字更清楚。

---

## 八、并发安全:为什么要加一把"锁"

Web 服务会**同时**处理很多请求。如果两个请求**同时**往同一个 slice 里写数据,可能会冲突、出乱子。

我们用 `sync.Mutex`(互斥锁)解决:

```go
var mu sync.Mutex

mu.Lock()         // 上锁:我要开始改数据了,别人等一下
defer mu.Unlock() // 函数结束自动解锁
// ……这中间的代码,同一时刻只有一个请求能进入
```

> 入门只需记住:**多个请求会共享、争抢同一份数据时,改它之前先 `Lock`,用 `defer Unlock` 善后。** 到任务 8 用上数据库后,这种并发问题数据库会帮我们处理,就不用自己加锁了。

---

## 九、写代码

完整 `main.go`:

```go
package main

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

var (
	todos  []Todo
	nextID = 1
	mu     sync.Mutex
)

func main() {
	r := gin.Default()

	r.GET("/todos", listTodos)
	r.POST("/todos", createTodo)
	r.DELETE("/todos/:id", deleteTodo)

	r.Run(":8080") // 端口先写死,任务 9 会用 viper 把它变成可配置
}

func listTodos(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()
	c.JSON(http.StatusOK, todos)
}

func createTodo(c *gin.Context) {
	var input Todo
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mu.Lock()
	defer mu.Unlock()
	input.ID = nextID
	nextID++
	todos = append(todos, input)
	c.JSON(http.StatusCreated, input)
}

func deleteTodo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 必须是数字"})
		return
	}

	mu.Lock()
	defer mu.Unlock()
	for i, t := range todos {
		if t.ID == id {
			todos = append(todos[:i], todos[i+1:]...)
			c.JSON(http.StatusOK, gin.H{"message": "已删除"})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "没找到这条待办"})
}
```

---

## 十、运行并测试

先把依赖补齐再启动:

```bash
go mod tidy   # 下载/整理依赖
go run .
```

看到类似 `Listening and serving HTTP on :8080` 就说明启动成功了,**这个窗口会一直占用,别关**。

另开**一个新终端**来测试(用 `curl`,Windows 10 以上自带):

```bash
curl -X POST http://localhost:8080/todos -H "Content-Type: application/json" -d "{\"title\":\"学习Go\",\"done\":false}"
curl http://localhost:8080/todos
curl -X DELETE http://localhost:8080/todos/1
```

也可以直接用浏览器打开 http://localhost:8080/todos 看 `GET` 的结果。测完回到服务窗口按 `Ctrl + C` 停止。

> 💡 两个小提示:
> - **Windows 终端里 curl 返回的中文可能是乱码**,这只是终端显示编码的问题,数据本身是对的——用浏览器或 Postman 看就正常。
> - 启动时 gin 会提示 "running in debug mode"。想让日志清爽些,可以设环境变量 `GIN_MODE=release` 再运行。

---

## 十一、动手练习

1. 加一个 `GET /todos/:id` 接口,只返回指定 id 的那一条(找不到返回 404)。
2. 加一个 `PUT /todos/:id` 接口,用于修改某条待办的 `title` 或 `done`。
3. 给 `POST` 加校验:`title` 为空时返回 400 错误,提示"标题不能为空"。

---

## 十二、小结

- ✅ 理解了 Web 后端"启动后一直监听、等人请求"的模型
- ✅ 用 `go get` 安装第三方库,会处理国内的下载/校验问题
- ✅ 用 gin 写路由:`GET`/`POST`/`DELETE` + 处理函数
- ✅ 用 `c.Param`、`c.ShouldBindJSON` 读请求,用 `c.JSON` 返回响应
- ✅ 知道用 `sync.Mutex` 保护被并发访问的共享数据

但现在有个大问题:**程序一关,数据就全没了**(都存在内存里)。**下一个任务**,我们用 **gorm + SQLite**,把待办真正**存进数据库**,关掉重启也不丢。

➡️ 下一站:`task08-gorm`
