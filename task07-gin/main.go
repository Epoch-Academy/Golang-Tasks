package main

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

// Todo 一条待办事项。json 标签决定它转成 JSON 时用什么字段名。
type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// 用内存(一个 slice)存放所有待办。
// Web 服务会同时处理很多请求,多个请求一起改这个 slice 会出乱子,
// 所以配一把"锁"(sync.Mutex),保证同一时刻只有一个请求在改它。
var (
	todos  []Todo
	nextID = 1
	mu     sync.Mutex
)

func main() {
	r := gin.Default() // 创建一个带默认配置(自带日志和错误恢复)的 gin 引擎

	r.GET("/todos", listTodos)         // 查全部
	r.POST("/todos", createTodo)       // 新增一条
	r.DELETE("/todos/:id", deleteTodo) // 按 id 删除

	// 在 8080 端口启动服务。端口先写死,任务 9 会用 viper 把它变成可配置。
	r.Run(":8080")
}

// listTodos 返回所有待办。
func listTodos(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()
	c.JSON(http.StatusOK, todos)
}

// createTodo 新增一条待办。
func createTodo(c *gin.Context) {
	var input Todo
	// 把请求体里的 JSON 解析进 input。解析失败说明客户端发来的数据有问题。
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

// deleteTodo 按 id 删除一条待办。
func deleteTodo(c *gin.Context) {
	// 路径里的 :id 拿出来是字符串,用 strconv.Atoi 转成整数。
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 必须是数字"})
		return
	}

	mu.Lock()
	defer mu.Unlock()
	for i, t := range todos {
		if t.ID == id {
			todos = append(todos[:i], todos[i+1:]...) // 从 slice 中删掉第 i 个
			c.JSON(http.StatusOK, gin.H{"message": "已删除"})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "没找到这条待办"})
}
