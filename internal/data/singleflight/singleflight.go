package singleflight

import (
	"sync"
)

type Result struct {
	val interface{}
	err error
}

// 真正用于执行key 对应的func的结构体
//
// 参考实现
//
// type call struct {
//     wg sync.WaitGroup

//     // These fields are written once before the WaitGroup is done
//     // and are only read after the WaitGroup is done.
//     val interface{}
//     err error

//	    // These fields are read and written with the singleflight
//	    // mutex held before the WaitGroup is done, and are read but
//	    // not written after the WaitGroup is done.
//	    dups  int
//	    chans []chan<- Result
//	}
type call struct {
	done bool
	cond *sync.Cond // 守护临界区
	res  *Result    // 记录fn 返回的结果, 其他协程在等待的时候可以判断res是否为nil
}

func newCall() *call {
	return &call{
		cond: &sync.Cond{L: &sync.Mutex{}},
		res:  &Result{},
	}
}

// 对外暴露的结构体
//
// 参考实现
//
//	type Group struct {
//	    mu sync.Mutex       // protects m
//	    m  map[string]*call // lazily initialized
//	}
type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

func NewGroup() *Group {
	return &Group{}
}

// 对外暴露接口
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	// 先去拿锁
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}

	if c, ok := g.m[key]; ok {
		// 等待第一个gor 释放信号
		g.mu.Unlock()
		c.cond.L.Lock()
		for !c.done {
			c.cond.Wait()
		}
		c.cond.L.Unlock()
		return c.res.val, c.res.err
	}

	c := newCall()
	g.m[key] = c
	g.mu.Unlock()
	g.docall(c, fn)

	return c.res.val, c.res.err
}

// docall的调用本身就在锁，所以broadcast不用加锁
func (g *Group) docall(c *call, fn func() (interface{}, error)) {
	g.mu.Lock()
	defer g.mu.Unlock()
	c.res.val, c.res.err = fn()
	c.done = true
	c.cond.Broadcast()
}
