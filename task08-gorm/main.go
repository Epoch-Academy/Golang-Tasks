package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite" // 纯 Go 实现的 SQLite 驱动,不需要装 gcc
	"gorm.io/gorm"
)

// Todo 一条待办。
// 嵌入 gorm.Model 后,会自动带上 ID、CreatedAt、UpdatedAt、DeletedAt 几个字段。
type Todo struct {
	gorm.Model
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// db 是全局的数据库连接,各个处理函数都用它。
var db *gorm.DB

func main() {
	var err error
	// 连接 SQLite。数据存进 todos.db 这个文件,文件不存在会自动创建。
	db, err = gorm.Open(sqlite.Open("todos.db"), &gorm.Config{})
	if err != nil {
		panic("连接数据库失败:" + err.Error())
	}

	// 自动迁移:让数据库里的表结构,和 Todo 结构体保持一致(没有表就建表)。
	db.AutoMigrate(&Todo{})

	r := gin.Default()
	r.GET("/todos", listTodos)
	r.POST("/todos", createTodo)
	r.DELETE("/todos/:id", deleteTodo)
	r.Run(":8080")
}

// listTodos 查询所有待办。
func listTodos(c *gin.Context) {
	var todos []Todo
	db.Find(&todos) // 查出所有记录,填进 todos
	c.JSON(http.StatusOK, todos)
}

// createTodo 新增一条待办。
func createTodo(c *gin.Context) {
	var input Todo
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.Create(&input) // 插入数据库,input.ID 会被自动填上
	c.JSON(http.StatusCreated, input)
}

// deleteTodo 按 id 删除一条待办。
func deleteTodo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 必须是数字"})
		return
	}
	result := db.Delete(&Todo{}, id) // 按主键 id 删除
	if result.RowsAffected == 0 {    // 没有任何行被删除,说明这条不存在
		c.JSON(http.StatusNotFound, gin.H{"error": "没找到这条待办"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}
