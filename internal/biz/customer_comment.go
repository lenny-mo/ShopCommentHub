package biz

import (
	"comment/internal/data/model"
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// 领域层代码
// CustomerUsecase is a Greeter usecase.
type CustomerUsecase struct {
	repo CustomerRepo
	log  *log.Helper
	c    *model.CustomerComment // 聚合根结构体
}

type CustomerRepo interface {
	// 存储评论
	CreateComment(context.Context, *model.CustomerComment) (*model.CustomerComment, error)
	// 回复已有的评论
	ReplyComment(context.Context, *model.CustomerComment) (*model.CustomerComment, error)
}

// CustomerUsecase 构造函数
// NewGreeterUsecase new a Greeter usecase.
func NewCustomerUsecase(repo CustomerRepo, logger log.Logger) *CustomerUsecase {
	return &CustomerUsecase{repo: repo, log: log.NewHelper(logger)}
}

// AddComment 添加用户对商品的新评价
func (uc *CustomerUsecase) AddComment(ctx context.Context, customer *model.CustomerComment) (*model.CustomerComment, error) {
	uc.log.WithContext(ctx).Debugf("AddComment: req: %v", customer)
	// 1. 数据校验

	// 2. 使用乐观锁判断数据是否被更改

// ReplyComment 用户对商品的回复
func (uc *CustomerUsecase) ReplyComment(ctx context.Context, customer *model.CustomerComment) (*model.CustomerComment, error) {
	uc.log.WithContext(ctx).Debugf("ReplyComment: req: %v", customer)
	uc.c = customer
	// 3. 新纪录插入数据库即可，缓存在后续查找的时候会添加
	return uc.repo.ReplyComment(ctx, uc.c)
}
