package service

import (
	"comment/internal/biz"
	"comment/internal/data"
	"comment/internal/data/model"
	"comment/internal/service/cron"
	"context"
	"encoding/json"
	"fmt"
	"github.com/asaskevich/EventBus"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"
	"time"
)

var (
	// 全局事件总线
	general_bus EventBus.Bus
	c           *cron.Cron
)

// InitEventBus 启动事件总线的监听，并且开启一个goroutine 持续处理事件
func InitEventBus() {
	// 注册事件总线
	general_bus = EventBus.New()
	fmt.Println("事件总线启动！")

	// 监听事件
	general_bus.Subscribe("customer:AddComment:success", sendToMQ)
	general_bus.Subscribe("customer:AddComment:fail", printFail)
	general_bus.Subscribe("customer:AddReply:success", sendToMQ)
	general_bus.Subscribe("customer:AddReply:fail", printFail)
	general_bus.Subscribe("merchant:AddReply:success", sendToMQ)
	general_bus.Subscribe("merchant:AddReply:fail", printFail)

	general_bus.Subscribe("scanMongoSendUnsuccessMsg", scanUserCommentMongo) // 定时任务

	// 创建 cron, 开启一个永不停止的时间轮, 注意defer stop
	// 定时任务：定时查询mongo中是否有sentMQ = false 的字段，挑选出来重新发送, 时间轮长度100
	c = cron.NewCron(time.Second, 100)
	c.Run()
	// 创建task
	var f = func(ctx context.Context) {
		once := new(sync.Once)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// 向事件总线发送事件,扫描mongo中的sentMQ字段, 如果为false 则拿出来重新发送
				once.Do(func() {
					general_bus.Publish("scanMongoSendUnsuccessMsg")
				})
			}
		}
	}
	// 添加定时任务
	c.AddTask(*cron.NewTask(f, 60*time.Second))
}

func CloseBus() {
	c.Stop()
}

// 接收事件参数 发现事件之后写入 kafka
func sendToMQ(event string, data interface{}, b *biz.BusUseCase) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// 写入kafka，写入操作最大重试3次，后续即使失败，会有另一个程序执行这个没有发布成功的事件
	if _, err := b.Insert(ctx, data); err != nil {
		fmt.Println("写入 kafka 失败", event)
		return
	}
	fmt.Println("写入 kafka 成功", event)
}

func printFail(event string, data interface{}, b *biz.BusUseCase) {
	fmt.Println(event, data)
}

// scanUserCommentMongo 扫描mongo中的sentMQ字段, 如果为false 则拿出来重新发送
func scanUserCommentMongo() {
	conn, err := data.ConnectMongo()
	if err != nil {
		fmt.Println("出错了", conn)
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second) // 限定每个goroutine执行5秒
	defer cancel()

	filter := bson.M{"sendtomq": false}    // 筛选条件
	option := options.Find().SetLimit(100) // 设置最多返回的结果集
	cur, err := conn.Collection("AddCommentSuccess").Find(ctx, filter, option)
	if err != nil {
		fmt.Println("查询出错")
	}
	defer cur.Close(ctx)

	var wg sync.WaitGroup
	for cur.Next(ctx) { // 异步把消息发送到kafka
		var comment model.CustomerComment
		if err := cur.Decode(&comment); err != nil {
			fmt.Println("解码出错了")
		} else {
			fmt.Println("scanUserCommentMongo任务正在发送到kafka: ", comment)
			wg.Add(1)
			go sentToKafka(&wg, ctx, comment)
		}
	}
	wg.Wait() // 等待所有kafka写入结束
}

// sentToKafka 把msg发送到kafka
func sentToKafka(wg *sync.WaitGroup, ctx context.Context, msg interface{}) error {
	defer wg.Done()

	select {
	case <-ctx.Done():
		return nil
	default:
		// 1. 开启kafka连接
		// 把数据写入kafka，只执行一次，失败了无所谓
		writer := data.ConnectKafka()
		defer writer.Close() // 一定要关闭
		comment := msg.(model.CustomerComment)
		bytes, err := json.Marshal(comment)
		if err != nil {
			return err
		}
		if err := writer.WriteMessages(ctx, kafka.Message{
			Value: bytes,
		}); err != nil {
			return err
		}

		// 2. 发送成功，根据commentid 找到mongo中的数据，修改sendtomq字段为true
		conn, err := data.ConnectMongo()
		if err != nil {
			return err
		}

		filter := bson.D{{"commentid", comment.CommentID}}
		update := bson.D{{"$set", bson.D{{"sendtomq", true}}}}
		if _, err = conn.Collection("AddCommentSuccess").UpdateOne(ctx, filter, update); err != nil {
			return err
		}
	}
	return nil
}
