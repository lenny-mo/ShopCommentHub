package data

import (
	"comment/internal/biz"
	"comment/internal/data/model"
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/go-kratos/kratos/v2/log"
)

type EventBusRepo struct {
	data *Data
	log  *log.Helper
}

// NewCustomerRepo
func NewEventBusRepo(data *Data, logger log.Logger) biz.EventBusRepo {
	return &EventBusRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// Insert kafka 中添加事件消息
func (e *EventBusRepo) Insert(ctx context.Context, data interface{}) (interface{}, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return data, err
	}
	repeatTimes := 3

	for repeatTimes > 0 {
		// 1. 写入消息队列
		if err = e.data.kafka.WriteMessages(ctx, kafka.Message{Value: bytes}); err != nil {
			if repeatTimes > 0 {
				repeatTimes--
				continue
			}
			// 3. 执行失败，交给一个定时任务后期重新发送
			return data, err
		} else {
			// 2. 执行成功，把mongo中对应的字段设置为true, 并且退出循环
			filter := bson.D{{"commentid", data.(*model.CustomerComment).CommentID}}
			update := bson.D{{"$set", bson.D{{"sendtomq", true}}}}
			e.data.mongo.Collection("AddCommentSuccess").UpdateOne(ctx, filter, update)
			break
		}
	}

	defer e.data.kafka.Close() // 关闭writer
	return data, nil
}
