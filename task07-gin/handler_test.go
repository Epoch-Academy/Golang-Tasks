package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	mu.Lock()
	todos = nil
	nextID = 1
	mu.Unlock()

	r := gin.New()
	r.GET("/todos", listTodos)
	r.POST("/todos", createTodo)
	r.DELETE("/todos/:id", deleteTodo)
	return r
}

func TestCreateListAndDeleteTodo(t *testing.T) {
	r := setupTestRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(`{"title":"学习测试"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("创建待办状态码错误,实际 %d", w.Code)
	}

	var created Todo
	if err := json.NewDecoder(w.Body).Decode(&created); err != nil {
		t.Fatal("解析创建结果失败:", err)
	}
	if created.ID != 1 || created.Title != "学习测试" {
		t.Fatalf("创建结果不对:%+v", created)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/todos", nil)
	r.ServeHTTP(w, req)

	var list []Todo
	if err := json.NewDecoder(w.Body).Decode(&list); err != nil {
		t.Fatal("解析列表失败:", err)
	}
	if len(list) != 1 {
		t.Fatalf("期望 1 条待办,实际 %d 条", len(list))
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/todos/1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("删除待办状态码错误,实际 %d", w.Code)
	}
}
