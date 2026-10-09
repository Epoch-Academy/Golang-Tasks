# 任务 6:goroutine、channel 与 context——让程序同时做事

> 本任务目标:理解 Go 最有特色的并发写法。你会让几个任务同时运行,用 channel 收集结果,再用 context 控制超时取消。

---

## 一、先看看我们要做出什么

运行后,程序会同时处理 3 个任务:

```text
开始处理任务...
读取配置 完成,用时 1s
调用外部服务 完成,用时 2s
生成统计报告 被取消:context deadline exceeded
全部任务都有结果,程序结束。
```

前两个任务来得及完成,第三个任务要 5 秒,但我们设置了 3 秒超时,所以它会被取消。

> 这三行的先后基本是固定的:三个任务分别在第 1、2、3 秒给出结果,界限分明。
> 之所以这么稳,是因为它们耗时差得明显。如果几个任务耗时很接近、几乎同时完成,那谁的结果先打印就说不准了——**输出顺序本来就是不确定的**,只是本例的耗时差距大,才让它看起来很稳定。

---

## 二、先把项目建好

```bash
mkdir task06-concurrency
cd task06-concurrency
go mod init gotask/task06-concurrency
```

---

## 三、为什么需要并发

前面的程序基本都是"一件事做完,再做下一件事"。

但真实后端经常不是这样:

- 一个请求进来,可能要同时查数据库、调外部 API、写日志
- 一个服务会同时处理很多用户请求
- 有些任务太慢,不能让它一直卡着不结束

Go 的并发模型很轻量,核心先记住三个词:

| 词 | 直觉理解 |
|---|---|
| `goroutine` | 让一段代码"同时跑起来"。 |
| `channel` | goroutine 之间传消息的管道。 |
| `context` | 控制超时、取消和请求生命周期。 |

---

## 四、goroutine:让函数同时跑

普通函数调用是这样:

```go
work()
```

这表示:等 `work` 跑完,程序才继续往下走。

如果前面加一个 `go`:

```go
go work()
```

意思就变成:启动一个新的 goroutine 去跑 `work`,当前函数继续往下走。

可以先把 goroutine 理解成一个很轻量的"小任务"。它和操作系统的"线程"并不完全是一回事,但在入门阶段,把它当成"一段能同时跑起来的代码"就够用了。

---

## 五、channel:把结果传回来

goroutine 跑到后台之后,主程序怎么知道它做完了?

一种常见做法是用 channel:

```go
results := make(chan string)
```

这行创建了一个传 `string` 的通道。

发送数据:

```go
results <- "任务完成"
```

接收数据:

```go
msg := <-results
```

箭头 `<-` 指向哪里,数据就往哪里走。

还有一个关键细节,正是它让本任务能"等齐所有结果":

我们用 `make(chan string)` 创建的是**无缓冲 channel**——它自己不存数据。发送方执行 `results <- ...` 时会**卡在那里**,直到另一边有人来接收;反过来,接收方 `<-results` 也会卡住,直到有人发送。

也就是说,一次发送和一次接收必须"碰头"才算完成,像两个人交接东西、必须同时在场。正因如此,主程序只要 `<-results` 接收和任务数一样多的次数,就一定能等到所有 goroutine 都交出结果,不需要额外的"等待"机制。

---

## 六、context:让任务知道该停了

真实项目不能只会"启动任务",还要会"及时停下"。

比如:

- 用户取消了请求
- 接口超过 3 秒还没返回
- 服务准备关闭,不想再继续做无意义的工作

这时就会用 `context`。

```go
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()
```

这段代码创建了一个 3 秒后自动取消的 `ctx`。

`defer cancel()` 的作用是:函数结束前释放 context 相关资源。即使是超时自动取消,手动 `cancel` 也是好习惯。

---

## 七、select:等多个结果里先来的那个

`select` 很像专门给 channel 用的 `switch`。

```go
select {
case <-time.After(job.Cost):
	results <- "任务完成"
case <-ctx.Done():
	results <- "任务被取消"
}
```

它会等几个 `case` 中**最先发生**的那个:

- 如果任务先完成,就发送"任务完成"
- 如果 context 先取消,就发送"任务被取消"

这正是并发程序里很常见的"谁先到就处理谁"。

---

## 八、写代码

新建 `main.go`:

```go
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
```

---

## 九、运行

```bash
go run .
```

你会看到前两个任务完成,第三个任务被取消。

重点不是去记输出长什么样,而是看懂这条链路:

1. `go work(...)` 同时启动多个任务
2. 每个任务通过 `results` 把结果传回来
3. `context.WithTimeout` 控制总超时时间
4. `select` 决定是任务完成,还是被取消

---

## 多想一层:为什么这个例子里没用到"锁"

你可能听说并发很容易写出 bug。可上面的程序却很干净,为什么?

关键在于:**每个 goroutine 只是把结果"发"进 channel,谁也没有去改同一个变量。**

设想另一种写法:让三个 goroutine 同时给一个共享的计数器加一。

```go
count := 0
// 三个 goroutine 同时执行 count++
```

`count++` 看着是一步,其实是三步:读出 count、加一、写回去。多个 goroutine 同时做这三步,就可能互相覆盖,最后结果比预期小。这种"多个 goroutine 同时读写同一个变量"造成的混乱,叫**数据竞争(data race)**。

Go 给了两条思路来避免它:

1. **不共享,用 channel 传**:像本任务这样,数据只在 channel 里流动,从不被多人同时修改。Go 有句名言——**"不要通过共享内存来通信,而要通过通信来共享内存。"**
2. **要共享,就加锁**:如果确实需要多个 goroutine 改同一份数据,就用一把"锁"(`sync.Mutex`),保证同一时刻只有一个在改。

本任务用的是第一条。**下一个任务(task07)写 Web 服务时,你会见到第二条**——很多请求同时往一个待办列表里增删,那里就需要锁了。现在不必深究锁怎么用,只要记住一句话:**一旦多个 goroutine 碰同一个变量,就要警惕数据竞争。**

> 想亲眼看看数据竞争?Go 自带检测器,用 `go run -race .` 运行,它能在程序跑动时当场抓出竞争。(这个检测器在 Windows 上需要额外的 C 编译器,装起来略麻烦;现在只要知道有它就行,真正用到时再查也不迟。)

---

## 十、为什么后端里经常见到 context

后面写 Web 服务时,每个 HTTP 请求都有自己的生命周期:

- 请求来了,开始处理
- 用户断开连接,请求就该取消
- 超过服务器设定的时间,也该取消

所以真实项目里的函数经常会长这样:

```go
func FindUser(ctx context.Context, id int) {
	// 用 ctx 控制数据库查询、外部请求等操作
}
```

`context.Context` 通常放在第一个参数,这是 Go 项目里很常见的约定。

---

## 十一、动手练习

1. 把超时时间从 3 秒改成 6 秒,观察第三个任务会不会完成。
2. 再加一个任务,比如 `{Name: "发送邮件", Cost: 4 * time.Second}`。
3. 把 `go work(...)` 前面的 `go` 去掉,观察程序变成"一个接一个执行"后有什么不同。
4. 把 `results := make(chan string)` 改成 `results := make(chan string, len(jobs))`,查一下这叫"带缓冲的 channel"。
5. 把三个任务的 `Cost` 都改成一样(比如都 1 秒),多跑几次,看看输出顺序是不是开始变得不固定了——这能帮你体会"顺序不确定"到底是怎么来的。

---

## 十二、小结

完成这个任务,你已经知道:

- `go 函数调用` 可以启动 goroutine
- channel 用来在 goroutine 之间传数据;无缓冲 channel 的发送和接收必须"碰头",这天然成了一种同步手段
- `select` 可以等待多个 channel 操作中先发生的一个
- `context` 常用来控制超时和取消
- 多个 goroutine 共享同一个变量会有数据竞争;Go 更推荐"用 channel 传数据"来避开它

下一步,我们换个身份:用 **gin 框架**,让你自己的程序变成一个**对外提供服务的 Web 后端**,别人也能来调用你。

➡️ 下一站:`task07-gin`

⬅️ 返回 [总览](../README.md)
