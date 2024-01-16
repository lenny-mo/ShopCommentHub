package data

import (
	"comment/internal/conf"
	"comment/internal/data/query"
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewDB, NewMongo, NewCustomerRepo, NewMerchantRepo, NewMongoRepo)

// Data .
type Data struct {
	//  wrapped database client
	q     *query.Query    // mysql
	mongo *mongo.Database // mongo
	log   *log.Helper
}

// NewData .
// 修改：添加gorm.DB
func NewData(db *gorm.DB, mongo *mongo.Database, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	query.SetDefault(db) // 选择数据库
	return &Data{q: query.Q, mongo: mongo, log: log.NewHelper(logger)}, cleanup, nil
}

func NewDB(cfg *conf.Data) *gorm.DB {
	switch cfg.Database.Driver {
	case "mysql":
		db, err := gorm.Open(mysql.Open(cfg.Database.Source))
		if err != nil {
			panic(err)
		}
		return db
	case "tidb":
	}

	return nil
}

func NewMongo(cfg *conf.Data) *mongo.Database {
	// 设置连接选项
	clientOptions := options.Client().ApplyURI(cfg.Mongo.Source)

	// 连接到 MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil
	}

	// 连接到指定数据库
	db := client.Database(cfg.Mongo.Db)

	return db
}
