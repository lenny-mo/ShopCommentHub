package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

type MongoRepo struct {
	data *Data
	log  *log.Helper
}

// NewCustomerRepo
func NewMongoRepo(data *Data, logger log.Logger) *MongoRepo {
	return &MongoRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// Insert 向mongo数据库中的指定表插入一条数据
func (m *MongoRepo) Insert(ctx context.Context, colle string, data interface{}) (interface{}, error) {
	m.data.mongo.Collection(colle).InsertOne(ctx, data)
	return data, nil
}
