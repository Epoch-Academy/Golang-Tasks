package main

import (
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) {
	t.Helper()

	var err error
	db, err = gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.db")), &gorm.Config{})
	if err != nil {
		t.Fatal("连接测试数据库失败:", err)
	}
	if err := db.AutoMigrate(&Todo{}); err != nil {
		t.Fatal("创建测试表失败:", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	})
}

func TestCreateTodoPersistsInDatabase(t *testing.T) {
	setupTestDB(t)

	todo := Todo{Title: "测试持久化"}
	if err := db.Create(&todo).Error; err != nil {
		t.Fatal("写入待办失败:", err)
	}

	var got Todo
	if err := db.First(&got, todo.ID).Error; err != nil {
		t.Fatal("读取待办失败:", err)
	}
	if got.Title != "测试持久化" {
		t.Fatalf("标题不对,实际 %q", got.Title)
	}
}
