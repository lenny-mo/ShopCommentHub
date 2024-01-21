package cron

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TaskElement struct {
	task     func(context.Context)
	pos      int64         // 定时任务挂载在环状数组中的索引位置，在删除任务的时候可以使用这个pos找到环形队列对应的list
	period   time.Duration // 任务规定的执行周期
	cycle    int64         // 定时任务的延迟轮次
	curcycle int64         // 当前经过的cycle
	key      string        // 定时任务的uuid，后续可以删除任务, 可以根据这个标志找到链表中的结点，并且删除
}

func NewTask(f func(context.Context), period time.Duration) *TaskElement {
	t := TaskElement{
		task:   f,
		key:    uuid.New().String(),
		period: period,
	}
	return &t
}

// 判断当前的任务是否应该执行，如果curcycle != 0 说明当前轮次还没到它执行
func (t *TaskElement) isReady() bool {
	return t.curcycle == 0
}
