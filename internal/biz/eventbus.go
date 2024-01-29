package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// 作为支撑领域服务
type BusUseCase struct {
	repo EventBusRepo
	log  *log.Helper
}

func NewBusUseCase(repo EventBusRepo, logger log.Logger) *BusUseCase {
	return &BusUseCase{repo: repo, log: log.NewHelper(logger)}
}

type EventBusRepo interface {
	Insert(context.Context, interface{}) (interface{}, error)
}

func (b *BusUseCase) Insert(ctx context.Context, data interface{}) (interface{}, error) {
	return b.repo.Insert(ctx, data)
}
