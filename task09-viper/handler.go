package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler 持有一个 Store,负责处理 HTTP 请求。
type Handler struct {
	store *Store
}

// list 返回所有待办。
func (h *Handler) list(c *gin.Context) {
	c.JSON(http.StatusOK, h.store.List())
}

// create 新增一条待办。
func (h *Handler) create(c *gin.Context) {
	var input Todo
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.store.Create(&input)
	c.JSON(http.StatusCreated, input)
}

// delete 按 id 删除一条待办。
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
