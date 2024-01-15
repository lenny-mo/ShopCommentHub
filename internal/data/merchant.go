package data

import (
	"comment/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type MerchantRepo struct {
	data *Data
	log  *log.Helper
}

// NewCustomerRepo
func NewMerchantRepo(data *Data, logger log.Logger) biz.MerchantRepo {
	return &CustomerRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}
