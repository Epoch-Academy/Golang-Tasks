# 任务 9:用 viper 配置化,整合成一个真实项目

> 本任务目标:这是**收官之作**。我们把前面学的 gin + gorm 整合起来,用 **viper** 把配置从代码里抽离,把代码**分层整理**成多个文件,再写一个**自动化测试**,最后**打包发布**。做完,你就拥有一个结构清晰、像模像样的真实后端项目。

---

## 一、先看看我们要做出什么

功能还是待办增删查,但工程质量上了一个台阶:

```
task09-viper/
├── go.mod          # 依赖清单
├── config.yaml     # 配置文件(端口、数据库路径……改它不用改代码)
├── main.go         # 入口:把各部分组装起来并启动
├── config.go       # 配置:用 viper 读取 config.yaml
├── model.go        # 数据:Todo 模型 + Store(封装数据库操作)
├── handler.go      # 接口:处理 HTTP 请求
├── model_test.go   # 数据层自动化测试
└── handler_test.go # 接口层自动化测试
```

---

## 二、为什么要"分层"

前面所有代码都挤在一个 `main.go` 里。文件一长就难维护、难找东西。真实项目会**按职责把代码拆开**:

- `config.go` 只管"读配置"
- `model.go` 只管"和数据库打交道"
- `handler.go` 只管"处理网络请求"
- `main.go` 只管"把它们组装起来、启动"

这样改其中一块,基本不影响其它块,协作和排错都方便。

> 说明:本任务这些文件都还是 `package main`(同一个包,放在同一文件夹,可以互相直接调用)。更大的项目会进一步拆成**不同的包**(放进 `internal/` 等子目录、互相 `import`),那是你之后进阶的内容。现在先体会"按职责拆文件"这一步。

---

## 三、viper:把配置从代码里抽出来

任务 7、8 里,端口 `:8080`、数据库名 `todos.db` 都**写死在代码**中。想换个端口?得改代码、重新编译。这很不专业。

**正确做法:把这些可变的东西放进配置文件**,程序启动时读进来。好处:
- 改配置不用改代码、不用重新编译
- 同一份代码,开发用一套配置、上线用另一套
- 密钥之类的敏感信息可以放配置/环境变量,不进代码

**viper** 就是 Go 里最流行的配置读取库,支持 yaml、json、环境变量等多种来源。

```bash
go get github.com/spf13/viper
```

---

## 四、config.yaml:配置文件

`yaml` 是一种很易读的配置格式,用缩进表示层级:

```yaml
# 服务相关配置
server:
  port: ":8080"      # 监听的端口

# 数据库相关配置
database:
  path: "todos.db"   # SQLite 数据库文件名

# 是否开启调试模式
debug: true
```

---

## 五、读配置:`config.go`

```go
package main

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config 对应 config.yaml 的结构。
type Config struct {
	Server struct {
		Port string
	}
	Database struct {
		Path string
	}
	Debug bool
}

func loadConfig() (*Config, error) {
	viper.SetConfigName("config") // 文件名 config(不带扩展名)
	viper.SetConfigType("yaml")   // 格式 yaml
	viper.AddConfigPath(".")      // 在当前目录找
	viper.AutomaticEnv()          // 允许环境变量覆盖

	// 默认值:即使没有 config.yaml 也能启动
	viper.SetDefault("server.port", ":8080")
	viper.SetDefault("database.path", "todos.db")
	viper.SetDefault("debug", true)

	if err := viper.ReadInConfig(); err != nil {
		// 只有"文件不存在"可以容忍(用默认值);其它错误要上报
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("读取配置文件出错:%w", err)
		}
		fmt.Println("提示:未找到 config.yaml,使用默认配置。")
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil { // 把配置塞进 Config 结构体
		return nil, fmt.Errorf("解析配置出错:%w", err)
	}
	return &cfg, nil
}
```

要点:`viper.Unmarshal(&cfg)` 把读到的配置一次性装进结构体,之后用 `cfg.Server.Port` 这样访问,很清爽。

---

## 六、数据层:`model.go`

这次我们把数据库操作封装进一个 `Store` 结构体(比任务 8 用全局变量更整洁):

```go
package main

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type Todo struct {
	gorm.Model
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type Store struct {
	db *gorm.DB
}

func NewStore(dbPath string) (*Store, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&Todo{}); err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() {
	if sqlDB, err := s.db.DB(); err == nil {
		sqlDB.Close()
	}
}

func (s *Store) List() []Todo {
	var todos []Todo
	s.db.Find(&todos)
	return todos
}

func (s *Store) Create(t *Todo) {
	s.db.Create(t)
}

func (s *Store) Delete(id int) bool {
	result := s.db.Delete(&Todo{}, id)
	return result.RowsAffected > 0
}
```

---

## 七、接口层:`handler.go`

处理函数挂在 `Handler` 上,`Handler` 持有 `Store`——请求来了就调用 `store` 的方法:

```go
package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	store *Store
}

func (h *Handler) list(c *gin.Context) {
	c.JSON(http.StatusOK, h.store.List())
}

func (h *Handler) create(c *gin.Context) {
	var input Todo
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.store.Create(&input)
	c.JSON(http.StatusCreated, input)
}

func (h *Handler) delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 必须是数字"})
		return
	}
	if !h.store.Delete(id) {
		c.JSON(http.StatusNotFound, gin.H{"error": "没找到这条待办"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}
```

---

## 八、入口:`main.go`

`main` 现在很短,只负责"组装"——这正是分层的好处:

```go
package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := loadConfig()           // 1. 读配置
	if err != nil {
		log.Fatal(err)
	}

	store, err := NewStore(cfg.Database.Path) // 2. 连数据库
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	if !cfg.Debug {                    // 3. 非调试模式让日志更干净
		gin.SetMode(gin.ReleaseMode)
	}

	h := &Handler{store: store}        // 4. 注册路由
	r := gin.Default()
	r.GET("/todos", h.list)
	r.POST("/todos", h.create)
	r.DELETE("/todos/:id", h.delete)

	log.Println("服务启动,监听", cfg.Server.Port) // 5. 启动
	if err := r.Run(cfg.Server.Port); err != nil {
		log.Fatal(err)
	}
}
```

---

## 九、自动化测试:`go test`

到现在我们都是手动 `curl` 测试。**自动化测试**让你写一段代码去检验另一段代码,以后改了东西跑一下就知道有没有改坏。

Go 内置测试,规矩很简单:
- 文件名以 `_test.go` 结尾
- 测试函数以 `Test` 开头、参数是 `t *testing.T`

```go
package main

import (
	"path/filepath"
	"testing"
)

func TestStoreCreateAndList(t *testing.T) {
	dbFile := filepath.Join(t.TempDir(), "test.db") // 临时数据库,测完自动清理

	store, err := NewStore(dbFile)
	if err != nil {
		t.Fatal("创建 Store 失败:", err)
	}
	defer store.Close()

	store.Create(&Todo{Title: "测试任务"})

	todos := store.List()
	if len(todos) != 1 {
		t.Fatalf("期望 1 条待办,实际 %d 条", len(todos))
	}
	if todos[0].Title != "测试任务" {
		t.Errorf("标题不对,实际「%s」", todos[0].Title)
	}
}
```

运行所有测试:

```bash
go test -v .
```

看到 `PASS` 就说明通过了。

---

## 十、运行与打包发布

**运行**(开发时):

```bash
go mod tidy
go run .
```

**打包成一个可执行文件**(发给别人 / 部署上线):

```bash
go build -o todo-server.exe .
```

得到 `todo-server.exe`,把它和 `config.yaml` 一起拷给别人,对方**不装 Go 也能直接运行**。这就是任务 1 说过的——Go 程序好分发。

> 进阶小知识:Go 还能"交叉编译"。比如在 Windows 上打包出 Linux 服务器能跑的版本:
> ```bash
> # Git Bash 里:
> GOOS=linux GOARCH=amd64 go build -o todo-server .
> ```

---

## 十一、(选学)给它加一个"AI / 外部接口"

还记得任务 5 调用外部 API 吗?把那套搬进来,就能给后端加一个新接口。比如加一个 `/quote`,返回一句随机的话:

```go
// 在 handler.go 里加(需要 import "io"、"net/http"、"encoding/json")
func (h *Handler) quote(c *gin.Context) {
	resp, err := http.Get("https://v1.hitokoto.cn/")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "获取失败"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var data map[string]any
	json.Unmarshal(body, &data)
	c.JSON(http.StatusOK, gin.H{"quote": data["hitokoto"]})
}

// 在 main.go 注册:
// r.GET("/quote", h.quote)
```

把这里的 `http.Get` 换成**调用大模型(LLM)的 POST 请求**(带上从环境变量读的 API Key,见任务 5 选学部分),你的后端就有了一个"AI 接口"——比如 `/ask?q=...` 让用户提问、由大模型回答。这就是把 Go 后端接入 AI 应用的最基本姿势。

---

## 十二、动手练习

1. 在 `config.yaml` 里加一项 `app_name`,在 `Config` 里加对应字段,启动时打印出来。
2. 给 `handler.go` 的列表接口和错误输入各写一个测试(提示:用 `net/http/httptest` 配合 gin)。
3. 完成第十一节的 `/quote` 接口,亲手跑通。

---

## 十三、全教程总结:你已经会了什么

走完九个任务,回头看看这条路:

| 阶段 | 任务 | 关键能力 |
|---|---|---|
| 语言基础 | 1–4 | 变量、流程控制、函数、slice/map、结构体、方法、**错误处理** |
| 联网 | 5 | JSON、调用 HTTP API |
| 并发 | 6 | goroutine、channel、context |
| 后端框架 | 7–9 | **gin** 写接口、**gorm** 存数据库、**viper** 管配置、分层、测试、发布 |

你不只是"学了语法",而是**做出了一个真实可用、能放进 GitHub 的后端项目**。

**接下来可以往哪走:**
- **中间件与认证**:给接口加日志、加 JWT 登录鉴权
- **更复杂的数据**:gorm 的表关联(一对多、多对多)
- **部署上线**:用 Docker 把项目打包,部署到云服务器
- **并发进阶**:深入 `goroutine` / `channel` / `context`
- **和 AI 协作**:用 AI 帮你读源码、写测试、做 code review——但始终保持"看懂它为什么对"的能力

恭喜你从零走到这里。剩下的,就是不断做项目、不断踩坑、不断查 AI 又不依赖 AI。祝你写代码愉快!

⬅️ 返回 [总览](../README.md)
