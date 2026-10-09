# 任务 8:用 gorm + SQLite 让数据存下来

> 本任务目标:解决上一个任务的致命缺陷——**程序一关,数据就没了**。我们用 **gorm**(Go 最流行的数据库工具)配合 **SQLite**(一个零配置、免安装的数据库),把待办真正存进硬盘,关掉重启也不丢。

---

## 一、先看看我们要做出什么

接口和任务 7 一模一样,但数据会**存进一个数据库文件**(`todos.db`)。哪怕你把程序关掉再重新启动:

```
# 重启服务后再查询,之前的数据还在!
$ curl http://localhost:8080/todos
[{"ID":1,"title":"learn go","done":false,...}]
```

---

## 二、先把项目建好

```bash
mkdir task08-gorm
cd task08-gorm
go mod init gotask/task08-gorm
```

---

## 三、为什么需要数据库

任务 7 把待办存在内存(一个 slice)里。内存的特点是**断电/关程序就清空**。要让数据"长久保存",得写进硬盘。

直接自己读写文件很麻烦(还记得任务 4 读文件吗)。**数据库**就是专门干这个的:帮你高效、安全地存取大量结构化数据。

我们选 **SQLite**:它不需要单独安装任何"数据库服务器",整个数据库就是**一个文件**,最适合入门和小项目。

---

## 四、安装 gorm 和 SQLite 驱动(避坑重点!)

```bash
go get gorm.io/gorm
go get github.com/glebarez/sqlite
```

> ⚠️ **为什么用 `github.com/glebarez/sqlite`,而不是 gorm 官网示例里的 `gorm.io/driver/sqlite`?**
> 官方那个驱动底层是 C 语言写的,需要你电脑装好 **gcc 编译器**,Windows 新手在这一步**极易卡死**(各种编译报错)。
> 而 `glebarez/sqlite` 是**纯 Go** 实现的,**不需要 gcc**,下载完直接能用。这是给初学者最省心的选择。

- `gorm` 是 **ORM**(对象关系映射):让你用"操作结构体"的方式来操作数据库,不用手写 SQL 语句。
- `glebarez/sqlite` 是让 gorm 能连上 SQLite 的"驱动"。

---

## 五、定义模型:`gorm.Model`

把任务 7 的 `Todo` 改造一下,**嵌入** `gorm.Model`:

```go
type Todo struct {
	gorm.Model        // 自动带上 ID、CreatedAt、UpdatedAt、DeletedAt 四个字段
	Title string `json:"title"`
	Done  bool   `json:"done"`
}
```

`gorm.Model` 是 gorm 提供的"标配字段包",嵌入它,你的表就自动有了:
- `ID`:主键(自动递增,不用自己维护 `nextID` 了)
- `CreatedAt` / `UpdatedAt`:创建、更新时间(gorm 自动填)
- `DeletedAt`:用于"软删除"(见第十节)

---

## 六、连接数据库 + 自动建表

```go
db, err := gorm.Open(sqlite.Open("todos.db"), &gorm.Config{})
if err != nil {
	panic("连接数据库失败:" + err.Error())
}

db.AutoMigrate(&Todo{}) // 自动迁移
```

- `gorm.Open(sqlite.Open("todos.db"), ...)`:打开(或创建)名为 `todos.db` 的数据库文件。
- `db.AutoMigrate(&Todo{})`:**自动迁移**——gorm 会照着 `Todo` 结构体,帮你把数据库里的表建好/改好。以后你给结构体加字段,它也能自动加上对应的列。新手不用手写一句建表 SQL。

---

## 七、增删查:不用写 SQL

gorm 把常见操作都包好了:

```go
db.Create(&todo)       // 新增一条(todo.ID 会被自动填上)
db.Find(&todos)        // 查询全部,结果填进 todos 这个 slice
db.First(&todo, id)    // 按主键查一条
db.Delete(&Todo{}, id) // 按主键删除
```

是不是比手写 SQL 直观多了?这就是 ORM 的价值。

---

## 八、写代码

完整 `main.go`:

```go
package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type Todo struct {
	gorm.Model
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

var db *gorm.DB

func main() {
	var err error
	db, err = gorm.Open(sqlite.Open("todos.db"), &gorm.Config{})
	if err != nil {
		panic("连接数据库失败:" + err.Error())
	}
	db.AutoMigrate(&Todo{})

	r := gin.Default()
	r.GET("/todos", listTodos)
	r.POST("/todos", createTodo)
	r.DELETE("/todos/:id", deleteTodo)
	r.Run(":8080")
}

func listTodos(c *gin.Context) {
	var todos []Todo
	db.Find(&todos)
	c.JSON(http.StatusOK, todos)
}

func createTodo(c *gin.Context) {
	var input Todo
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.Create(&input)
	c.JSON(http.StatusCreated, input)
}

func deleteTodo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 必须是数字"})
		return
	}
	result := db.Delete(&Todo{}, id)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "没找到这条待办"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}
```

对比任务 7,变化很小:去掉了内存 slice 和那把锁(数据库会处理并发),把存取换成了 gorm 的几行调用。

---

## 九、运行,并亲自验证"持久化"

```bash
go mod tidy
go run .
```

测试步骤(另开终端):

```bash
# 1. 创建两条
curl -X POST http://localhost:8080/todos -H "Content-Type: application/json" -d "{\"title\":\"learn go\"}"
curl -X POST http://localhost:8080/todos -H "Content-Type: application/json" -d "{\"title\":\"do homework\"}"

# 2. 此时目录里出现了 todos.db 文件

# 3. 回到服务窗口按 Ctrl+C 停掉,再 go run . 重新启动

# 4. 再查一次,数据还在!这就是"持久化"
curl http://localhost:8080/todos
```

**关掉重启数据还在**——这是数据库带来的、内存版做不到的能力。

---

## 十、一个小知识:软删除

你删除一条后,会发现 `todos.db` 里其实没真删,只是把那条的 `DeletedAt` 填上了时间。这叫**软删除**:数据还在,只是查询时默认不显示了(像回收站)。这是 `gorm.Model` 自带的保护机制,防止误删丢数据。想要彻底删可以用 `db.Unscoped().Delete(...)`。

---

## 十一、动手练习

1. 加一个 `GET /todos/:id` 接口,用 `db.First(&todo, id)` 查单条,找不到返回 404。
2. 加一个 `PUT /todos/:id` 接口,用 `db.Save(&todo)` 更新待办的 `done` 状态。
3. 让 `GET /todos` 支持只看未完成的:加个查询参数 `?done=false`,用 `db.Where("done = ?", false).Find(&todos)`。

---

## 十二、小结

- ✅ 理解了内存 vs 数据库(持久化)的区别
- ✅ 会用 gorm + SQLite,且知道**用 `glebarez/sqlite` 避开 gcc 的坑**
- ✅ 用 `gorm.Model` 自动获得 ID 和时间戳
- ✅ 用 `AutoMigrate` 自动建表,用 `Create/Find/Delete` 增删查,不用写 SQL
- ✅ 知道了 gorm 的"软删除"

现在你有一个数据能存下来的后端了!但还有些"业余":端口、数据库名都**写死**在代码里,改一下就得重新编译;代码也全挤在一个文件里。**最后一个任务**,我们用 **viper** 把配置抽出来,把项目**分层整理**,再写个**自动化测试**,做成一个像模像样的真实项目。

➡️ 下一站:`task09-viper`
