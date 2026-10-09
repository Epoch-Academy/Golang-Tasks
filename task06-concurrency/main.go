package main

import (
	"context"
	"fmt"
	"time"
)

// Job 表示一个需要处理的任务。Cost 是它预计要花的时间。
type Job struct {
	Name string
	Cost time.Duration
}

func main() {
	jobs := []Job{
		{Name: "读取配置", Cost: 1 * time.Second},
		{Name: "调用外部服务", Cost: 2 * time.Second},
		{Name: "生成统计报告", Cost: 5 * time.Second},
	}

	// context 是一个"取消信号":3 秒后它会自动通知所有任务"时间到了"。
	// 真实项目里,常用它给一批操作设一个统一的超时上限。
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// defer 让这行在 main 结束前才执行,用来释放 context 占用的资源。
	// 即使是超时自动取消,也应当调用一次 cancel,这是固定搭配。
	defer cancel()

	// results 是一条传 string 的"管道"。它没有缓冲区,
	// 发送方会一直卡住,直到有接收方把数据取走——
	// 这个"必须碰头"的特性,下面正好被用来"等齐所有任务的结果"。
	results := make(chan string)

	fmt.Println("开始处理任务...")

	// 用 go 给每个 job 各启动一个 goroutine。
	// 这个循环瞬间就走完:它只负责"点火",不等任何一个任务做完。
	for _, job := range jobs {
		go work(ctx, job, results)
	}

	// 上面点了 len(jobs) 把火,这里就接收同样多的次数。
	// 我们不关心谁先完成(先后由各自的耗时决定),
	// 只要接收次数和任务数一致,就一定能等齐全部结果。
	for range jobs {
		fmt.Println(<-results)
	}

	fmt.Println("全部任务都有结果,程序结束。")
}

// work 处理单个任务。
// 参数 results 写成 chan<- string(箭头朝里),表示它在这里"只能发送"。
// 这是有意的限制:work 只负责报告结果,不该从这条管道读数据;
// 写成单向之后,一旦用错,编译器会直接报错。
func work(ctx context.Context, job Job, results chan<- string) {
	// select 同时盯住下面两个 channel,谁先来信号就走谁;两个都没来就一直等。
	// 这正是并发里典型的"谁先到就先处理谁"。
	select {
	case <-time.After(job.Cost):
		// 在超时之前做完了:time.After 到点会送来一个信号。
		results <- fmt.Sprintf("%s 完成,用时 %s", job.Name, job.Cost)
	case <-ctx.Done():
		// 还没做完,context 就先到了 3 秒上限:ctx.Done() 送来取消信号。
		results <- fmt.Sprintf("%s 被取消:%v", job.Name, ctx.Err())
	}
}
