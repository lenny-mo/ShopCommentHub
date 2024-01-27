package data

import (
	"comment/internal/conf"
	"comment/internal/data/query"
	"context"
	es8 "github.com/elastic/go-elasticsearch/v8"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewDB, NewRedis, NewMongo, NewKafka, NewEsWithIndex, NewCustomerRepo, NewMerchantRepo, NewEventBusRepo, NewReadModelRepo)

// Data .
type Data struct {
	//  wrapped database client
	q     *query.Query    // mysql
	mongo *mongo.Database // mongo
	kafka *kafka.Writer   // kafka
	es    *ESWithIndex    // elasticsearch
	redis *redis.Client   // redis
	log   *log.Helper
}

// NewData .
// 修改：添加gorm.DB
func NewData(db *gorm.DB, rdb *redis.Client, mongo *mongo.Database, kafka *kafka.Writer, es *ESWithIndex, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	query.SetDefault(db) // 选择数据库
	return &Data{q: query.Q, redis: rdb, mongo: mongo, kafka: kafka, es: es, log: log.NewHelper(logger)}, cleanup, nil
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

func NewRedis(cfg *conf.Data) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		Password:     "", // no password set
		DB:           0,  // use default DB
		ReadTimeout:  cfg.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: cfg.Redis.WriteTimeout.AsDuration(),
	})
	return rdb
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

func NewKafka(cfg *conf.Data) *kafka.Writer {
	// 创建一个writer
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Kafka.Source),
		Topic:        cfg.Kafka.Topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
	}

	// TODO 程序结束的时候关闭这个连接
	return writer
}

func ConnectMongo() (*mongo.Database, error) {

	source := "mongodb://localhost:27017"
	db := "test"

	clientOptions := options.Client().ApplyURI(source)

	// 连接到 MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}

	// 连接到指定数据库
	conn := client.Database(db)

	return conn, nil
}

func ConnectKafka() *kafka.Writer {
	source := "localhost:9092"
	topic := "eventbus"

	writer := &kafka.Writer{
		Addr:         kafka.TCP(source),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
	}

	return writer
}

type ESWithIndex struct {
	client *es8.TypedClient
	index  string
}

func NewEsWithIndex(cfg *conf.Data) *ESWithIndex {
	cli, err := es8.NewTypedClient(
		es8.Config{
			Addresses: cfg.Elasticsearch.Addrs,
		},
	)

	if err != nil {
		return nil
	}

	return &ESWithIndex{
		client: cli,
		index:  cfg.Elasticsearch.Index,
	}
}
