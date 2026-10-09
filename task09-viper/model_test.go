package main

import (
	"path/filepath"
	"testing"
)

// TestStoreCreateAndList 测试:新建一条待办后,能在列表里查到它。
// 运行方式:go test
func TestStoreCreateAndList(t *testing.T) {
	// t.TempDir() 提供一个临时目录,测试结束后会自动清理,不污染项目。
	dbFile := filepath.Join(t.TempDir(), "test.db")

	store, err := NewStore(dbFile)
	if err != nil {
		t.Fatal("创建 Store 失败:", err)
	}
	defer store.Close() // 关闭连接,临时目录才能被顺利清理

	store.Create(&Todo{Title: "测试任务"})

	todos := store.List()
	if len(todos) != 1 {
		t.Fatalf("期望 1 条待办,实际 %d 条", len(todos))
	}
	if todos[0].Title != "测试任务" {
		t.Errorf("标题不对,期望「测试任务」,实际「%s」", todos[0].Title)
	}
}
