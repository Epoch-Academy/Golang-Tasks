package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func newTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	store, err := NewStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal("创建 Store 失败:", err)
	}
	t.Cleanup(store.Close)

	h := &Handler{store: store}
	r := gin.New()
	r.GET("/todos", h.list)
	r.POST("/todos", h.create)
	r.DELETE("/todos/:id", h.delete)
	return r
}

func TestHandlerCreateAndDelete(t *testing.T) {
	r := newTestRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(`{"title":"接口测试"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("创建待办状态码错误,实际 %d", w.Code)
	}

	var created Todo
	if err := json.NewDecoder(w.Body).Decode(&created); err != nil {
		t.Fatal("解析创建结果失败:", err)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/todos/"+strconv.Itoa(int(created.ID)), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("删除待办状态码错误,实际 %d", w.Code)
	}
}
