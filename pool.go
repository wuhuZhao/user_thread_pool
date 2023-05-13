package user_thread_pool

import (
	"log"
	"sync"
)

var exe Executor = initExecutor()

func Submit(task Task) {
	// 对外暴露的Submit接口
	exe.Submit(task)
}

type Task func()

type Executor struct {
	tasks chan Task
	once  *sync.Once
	queue Queue
}

func initExecutor() Executor {
	executor := newExecutor()
	// 控制初始化执行一次
	executor.once.Do(func() {
		executor.execute()
		// 启动主协程，负责调度其他协程的任务
		go executor.Run()
	})
	return executor
}

func newExecutor() Executor {
	return Executor{
		tasks: make(chan Task, 100),
		once:  &sync.Once{},
		queue: NewDefaultTaskQueue(),
	}
}

// Submit Submit: 提交任务，如果有空闲协程则添加到空闲协程执行，否则添加到任务队列中
func (e *Executor) Submit(task Task) {
	select {
	case e.tasks <- task:
		return
	default:
		e.queue.Offer(task)
	}
}

// Run Run: 主协程负责调度任务，如果有空闲协程则添加到协程中执行，否则就自己执行
func (e *Executor) Run() {
	for {
		task := e.queue.Poll()
		if task == nil {
			continue
		}
		select {
		case e.tasks <- task.(Task):
			continue
		default:
			log.Println("thread_pool_master run task!")
			task.(Task)()
		}
	}
}

// execute execute: 创建协程池的初始协程，不断地轮训任务，不存在任务的时候会阻塞该协程
func (e *Executor) execute() {
	for i := 0; i < 8; i++ {
		go func(threadId int) {
			for {
				t := <-e.tasks
				log.Printf("thread_pool_%d run task!\n", threadId)
				t()
			}
		}(i)
	}
}
