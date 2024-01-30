package biz

import (
	"comment/internal/data/model"
	"context"
	"github.com/go-kratos/kratos/v2/log"
)

type MerchantUsecase struct {
	repo MerchantRepo
	log  *log.Helper
	mc   *model.MerchantComment // 聚合根结构体
}

type MerchantRepo interface {
	// 回复已有的评论
	ReplyComment(context.Context, *model.MerchantComment) (*model.MerchantComment, error)
	// 添加商品
	AddProduct(context.Context, *model.Product) (*model.Product, error)
}

// MerchantUsecase 构造函数
// MerchantUsecase new a MerchantUsecase usecase.
func NewMerchantUsecase(repo MerchantRepo, logger log.Logger) *MerchantUsecase {
	return &MerchantUsecase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *MerchantUsecase) AddProduct(ctx context.Context, p *model.Product) (*model.Product, error) {
	uc.log.WithContext(ctx).Debugf("AddProduct: req: %v", p)
	return uc.repo.AddProduct(ctx, p)
}

func (uc *MerchantUsecase) ReplyComment(ctx context.Context, mc *model.MerchantComment) (*model.MerchantComment, error) {
	uc.log.WithContext(ctx).Debugf("ReplyComment: req: %v", mc)
	uc.mc = mc
	return uc.repo.ReplyComment(ctx, uc.mc)
}
