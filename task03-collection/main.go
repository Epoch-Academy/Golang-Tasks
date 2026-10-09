package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// 统计一段英文里每个单词出现的次数。
func main() {
	fmt.Println("请输入一段英文(单词用空格隔开),然后回车:")

	// bufio.Scanner 可以读取一整行。
	// (fmt.Scanln 读到空格就停了,读不了一整句话,所以这里换它。)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()         // 读取一行
	line := scanner.Text() // 取出读到的那行文字

	// 转成小写,这样 "The" 和 "the" 会被当成同一个词。
	line = strings.ToLower(line)

	// strings.Fields 按空白把句子切成一个个单词,返回一个 slice(列表)。
	words := strings.Fields(line)

	// map 是"键值对"集合。这里用它记录:单词 -> 出现次数。
	counts := make(map[string]int)

	// for range 遍历 words 里的每个单词。
	// 下划线 _ 表示"我不关心下标,只要值"。
	for _, word := range words {
		counts[word]++ // 该单词计数加 1(从没出现过时,自动从 0 开始)
	}

	fmt.Printf("\n一共 %d 个单词,其中不同的单词 %d 种:\n", len(words), len(counts))

	// 遍历 map,打印每个单词和它出现的次数。
	for word, n := range counts {
		fmt.Printf("  %-12s %d 次\n", word, n) // %-12s:左对齐,占 12 个字符宽
	}
}
