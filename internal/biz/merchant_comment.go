package biz

import "github.com/go-kratos/kratos/v2/log"

type MerchantUsecase struct {
	repo MerchantRepo
	log  *log.Helper
}

type MerchantRepo interface {
}

// MerchantUsecase 构造函数
// MerchantUsecase new a MerchantUsecase usecase.
func NewMerchantUsecase(repo MerchantRepo, logger log.Logger) *MerchantUsecase {
	return &MerchantUsecase{repo: repo, log: log.NewHelper(logger)}
}
