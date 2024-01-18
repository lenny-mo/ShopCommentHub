package service

import (
	"comment/internal/biz"
	"context"
	"fmt"
	"github.com/asaskevich/EventBus"
)

var (
	// 全局事件总线
	general_bus EventBus.Bus
)

// InitEventBus 启动事件总线的监听，并且开启一个goroutine 持续处理事件
func InitEventBus() {
	// 注册事件总线
	general_bus = EventBus.New()
	fmt.Println("事件总线启动！")

	// 监听事件
	general_bus.Subscribe("customer:AddComment:success", help)
	general_bus.Subscribe("customer:AddComment:fail", help)
}

// 接收事件参数 发现事件之后写入mongodb以及 kafka
func help(event string, data interface{}, b *biz.BusUseCase) {
	// TODO
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// 写入mongodb
	b.Insert(ctx, event, data)
	// 写入kafka

	fmt.Printf("eventbus: %s, %v, \n", event, data)
}
