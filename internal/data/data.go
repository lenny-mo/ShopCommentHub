package data

import (
	"comment/internal/conf"
	"comment/internal/data/query"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewDB, NewCustomerRepo, NewMerchantRepo)

// Data .
type Data struct {
	//  wrapped database client
	q   *query.Query
	log *log.Helper
}

// NewData .
// 修改：添加gorm.DB
func NewData(db *gorm.DB, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	query.SetDefault(db) // 选择数据库
	return &Data{q: query.Q, log: log.NewHelper(logger)}, cleanup, nil
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
