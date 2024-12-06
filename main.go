package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

//type Task func()

type TaskQueue struct {
	taskQueue chan func()
	wg        sync.WaitGroup
	maxTasks  int
	maxQueue  int
	mu        sync.Mutex
	stopped   bool
}

func NewTaskQueue(maxTasks, maxQueue int) *TaskQueue {
	return &TaskQueue{
		taskQueue: make(chan func(), maxQueue),
		maxTasks:  maxTasks,
		maxQueue:  maxQueue,
		stopped:   false,
	}
}

func (tq *TaskQueue) Start() {
	for i := 0; i < tq.maxTasks; i++ {
		go func() {
			for task := range tq.taskQueue {
				task()
				tq.wg.Done()
			}
		}()
	}
}

func (tq *TaskQueue) EnqueueTask(task func()) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	if tq.stopped {
		return errors.New("TaskQueue has been stopped, cannot enqueue new tasks.")
	}

	if len(tq.taskQueue) >= tq.maxQueue {
		return errors.New("TaskQueue is full, cannot enqueue new task.")
	}

	tq.taskQueue <- task
	tq.wg.Add(1)

	return nil
}

func (tq *TaskQueue) Stop() {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	tq.stopped = true
	close(tq.taskQueue)
	tq.wg.Wait()
}

func main() {
	taskQueue := NewTaskQueue(3, 5)
	taskQueue.Start()

	for i := 0; i < 7; i++ {
		err := taskQueue.EnqueueTask(func() {
			fmt.Printf("Task processed %d\n", i)
			time.Sleep(1 * time.Second)
		})
		if err != nil {
			fmt.Println(err)
		}
	}

	taskQueue.Stop()
}
