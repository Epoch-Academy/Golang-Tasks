package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Hitokoto 用来"接住"API 返回的 JSON 数据。
// 字段后面反引号里的 `json:"hitokoto"` 叫标签(tag),作用是告诉 Go:
// JSON 里那个叫 "hitokoto" 的字段,请装进结构体的 Text 字段。
type Hitokoto struct {
	Text string `json:"hitokoto"`
	From string `json:"from"`
}

func main() {
	url := "https://v1.hitokoto.cn/" // 一个免费的"一言"接口,会随机返回一句话

	// 第 1 步:发起 GET 请求。
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("请求失败(检查一下网络):", err)
		return
	}
	// defer 表示"等本函数结束时再执行这行",用来确保连接被正确关闭。
	defer resp.Body.Close()

	// 第 2 步:把返回的内容(一段 JSON 文本)全部读出来。
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取响应失败:", err)
		return
	}

	// 第 3 步:把 JSON 文本"解析"进我们定义的结构体。
	var h Hitokoto
	if err := json.Unmarshal(body, &h); err != nil {
		fmt.Println("解析 JSON 失败:", err)
		return
	}

	// 第 4 步:使用解析出来的数据。
	fmt.Printf("「%s」\n    —— %s\n", h.Text, h.From)
}
