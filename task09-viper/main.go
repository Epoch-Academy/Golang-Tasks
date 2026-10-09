package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

// main 是程序入口:读配置 → 连数据库 → 注册路由 → 启动服务。
// 它只负责"把各部分组装起来",具体逻辑分散在 config.go / model.go / handler.go 里。
func main() {
	// 1. 读取配置
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err) // log.Fatal 会打印错误并直接退出程序
	}

	// 2. 连接数据库
	store, err := NewStore(cfg.Database.Path)
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	// 3. 根据配置决定 gin 的运行模式
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// 4. 注册路由
	h := &Handler{store: store}
	r := gin.Default()
	r.GET("/todos", h.list)
	r.POST("/todos", h.create)
	r.DELETE("/todos/:id", h.delete)

	// 5. 启动服务,端口来自配置文件
	log.Println("服务启动,监听", cfg.Server.Port)
	if err := r.Run(cfg.Server.Port); err != nil {
		log.Fatal(err)
	}
}
