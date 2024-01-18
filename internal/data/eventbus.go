package data

import (
	"comment/internal/biz"
	"context"

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

// Insert 向mongo数据库中的指定表插入一条数据
func (m *EventBusRepo) Insert(ctx context.Context, colle string, data interface{}) (interface{}, error) {
	m.data.mongo.Collection(colle).InsertOne(ctx, data)
	return data, nil
}
