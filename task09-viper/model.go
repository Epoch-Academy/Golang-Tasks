package main

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// Todo 一条待办(和任务 8 一样)。
type Todo struct {
	gorm.Model
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// Store 把所有对数据库的操作封装在一起。
// 好处:其它代码只跟 Store 打交道,不用关心底层用的是哪种数据库。
type Store struct {
	db *gorm.DB
}

// NewStore 连接数据库并自动建表。
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

// Close 关闭数据库连接(测试结束时会用到,释放对数据库文件的占用)。
func (s *Store) Close() {
	if sqlDB, err := s.db.DB(); err == nil {
		sqlDB.Close()
	}
}

// List 返回所有待办。
func (s *Store) List() []Todo {
	var todos []Todo
	s.db.Find(&todos)
	return todos
}

// Create 新增一条待办。
func (s *Store) Create(t *Todo) {
	s.db.Create(t)
}

// Delete 按 id 删除,返回是否真的删掉了(false 表示这条不存在)。
func (s *Store) Delete(id int) bool {
	result := s.db.Delete(&Todo{}, id)
	return result.RowsAffected > 0
}
