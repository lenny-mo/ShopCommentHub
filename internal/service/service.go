package service

import (
	"context"
	"fmt"

	"github.com/asaskevich/EventBus"
	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewCommentServiceService)

var (
	// 全局事件总线
	bus EventBus.Bus
)

type event struct {
	eventMsg string
	data     interface{}
}

func InitEventBus(ctx context.Context) {
	// 注册事件总线
	bus = EventBus.New()
	fmt.Println("事件总线启动！")

	// 事件总线监听事件，发现事件之后写入mongodb以及 kafka
	go func(ctx context.Context) {
		// 1. 订阅事件
		bus.Subscribe("customer:AddComment:success", help)
		bus.Subscribe("customer:AddComment:fail", help)

		// 2. 持续监听事件总线, 有事件发生则写入mongodb和 kafka
		for {
			select {
			case <-ctx.Done():
				fmt.Println("eventbus is out!")
				return
			default:
			}
		}
	}(ctx)
}

// 接收事件参数 发现事件之后写入mongodb以及 kafka
func help(event string, data interface{}) {

	fmt.Printf("eventbus: %s, %v, \n", event, data)
}
