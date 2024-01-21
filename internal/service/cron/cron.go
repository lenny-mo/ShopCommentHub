package cron

import (
	"container/list"
	"context"
	"sync"
	"time"
)

type Cron struct {
	//  执行关闭函数
	sync.Once
	// ticker 每走一步要用的时间
	step time.Duration
	// 一轮时间的总长度 = step * len(slots)
	span time.Duration
	// 时间轮定时器
	ticker *time.Ticker
	// 停止时间轮的 channel，1个缓冲
	done chan struct{}
	// 新增定时任务的入口 channel
	addTask chan TaskElement
	// 删除定时任务的入口 channel
	removeTask chan string
	// 通过 list 组成的环状数组. 通过遍历环状数组的方式实现时间轮
	slots []*list.List
	// 当前遍历到的环状数组的索引
	curSlot int
	// map[task.UUID] = *list.Element
	tasks map[string]*list.Element
}

// NewCron 创建一个时间轮
func NewCron(step time.Duration, size int) *Cron {
	cron := Cron{
		step:       step,
		span:       time.Duration(int64(size) * int64(step)),
		ticker:     time.NewTicker(step),
		done:       make(chan struct{}, 1),
		addTask:    make(chan TaskElement, 1),
		removeTask: make(chan string, 1),
		slots:      make([]*list.List, size),
		curSlot:    0,
		tasks:      make(map[string]*list.Element),
	}
	return &cron
}

// AddTask 把任务添加到队列
func (c *Cron) AddTask(t TaskElement) {
	c.addTask <- t
}

// RemoveTask 添加删除指令到队列
func (c *Cron) RemoveTask(key string) error {
	c.removeTask <- key
	return nil
}

// Stop 关闭Cron
func (c *Cron) Stop() {
	c.done <- struct{}{}
}

// Run 启动一个不会停止的Cron
func (c *Cron) Run() {
	go c.run()
}

func (c *Cron) run() {
	for {
		select {
		case <-c.done:
			c.stop()
			return
		case <-c.ticker.C:
			c.tick()
		case element := <-c.addTask:
			// 根据传入的task, 执行c.AddTask
			c.addToTask(&element)
		case key := <-c.removeTask:
			// 根据key 从map中找到
			c.removeFromTask(key)
		}
	}
}

// addToTask 实际添加任务到环形队列中的某个双向链表
// 同时添加到c.map中
func (c *Cron) addToTask(t *TaskElement) {
	// 1. 计算task在环形队列中的position以及需要经过的时间轮次
	t.cycle = int64(t.period) / int64(c.span)
	t.curcycle = t.cycle
	t.pos = (int64(c.curSlot) + ceilDiv(int64(t.period), int64(c.step))) % int64(len(c.slots))

	// 2. 插入到双向链表
	var e *list.Element
	if l := c.slots[t.pos]; l != nil { // 判断该位置的list是否为nil
		e = l.PushBack(t)
	} else {
		c.slots[t.pos] = list.New()
		e = c.slots[t.pos].PushBack(t)
	}

	// 3. 添加到c.tasks
	c.tasks[t.key] = e
}

func (c *Cron) removeFromTask(key string) error {
	// 1. 从链表中删除任务
	e := c.tasks[key]
	if task, ok := e.Value.(*TaskElement); ok {
		c.slots[task.pos].Remove(e) // 获取到pos对应list执行对应的删除方法
	}

	// 2. 从c.tasks中删除任务
	delete(c.tasks, key)
	return nil
}

// 判断当前的slot的list是否存在，如果存在则遍历list的所有任务
// 调用任务的isReady函数，如果返回true, 开启一个go 执行任务, 并且从当前链表中删除该任务，重新添加到新位置的链表
// 移动cron内的curslot指针到下一个位置
func (c *Cron) tick() {
	// 1. 判断当前的slot的list是否存在，如果存在则遍历list的所有任务
	// 调用任务的isReady函数
	if l := c.slots[c.curSlot]; l != nil {
		for e := l.Front(); e != nil; e = e.Next() {
			task := e.Value.(*TaskElement)
			if task.isReady() {
				ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
				go task.task(ctx)
				task.curcycle = task.cycle
				c.removeFromTask(task.key)
				c.addToTask(task)
			} else {
				task.curcycle--
			}
		}
	}

	// 2. 更新cur的位置
	c.curSlot = (c.curSlot + 1) % len(c.slots)
}

func (c *Cron) stop() {
	close(c.done)
	close(c.addTask)
	close(c.removeTask)
	c.ticker.Stop()
}

// ceilDiv 向上取整
func ceilDiv(a, b int64) int64 {
	result := a / b
	remainder := a % b
	if remainder > 0 {
		result++
	}
	return result
}
