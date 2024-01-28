package biz

import (
	v1 "comment/api/comment/v1"
	"context"
	"github.com/go-kratos/kratos/v2/log"
)

// 读模型：提供ES检索功能给C端和O端，目前支持根据skuid获取到所有对应的评论
type ReadModelUsecase struct {
	repo  ReadModelRepo
	log   *log.Helper
	skuId int32 // 聚合根值对象
}

type ReadModelRepo interface {
	// 根据skuid查询评论
	GetAllCommentsBySKUID(context.Context, int32) ([]*v1.Comment, error)
}

func NewReadModelUsecase(repo ReadModelRepo, logger log.Logger) *ReadModelUsecase {
	return &ReadModelUsecase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *ReadModelUsecase) GetAllCommentsBySKUID(ctx context.Context, skuid int32) ([]*v1.Comment, error) {

	uc.skuId = skuid
	data, err := uc.repo.GetAllCommentsBySKUID(ctx, uc.skuId)
	if err != nil {
		uc.log.Error(err)
		return nil, err
	}

	return data, nil
}
