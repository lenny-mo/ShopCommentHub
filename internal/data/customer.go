package data

import (
	"comment/internal/biz"
	"comment/internal/data/model"
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// 实现CustomerRepo interfance
type CustomerRepo struct {
	data *Data
	log  *log.Helper
}

// NewCustomerRepo
func NewCustomerRepo(data *Data, logger log.Logger) biz.CustomerRepo {
	return &CustomerRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *CustomerRepo) CreateComment(ctx context.Context, c *model.CustomerComment) (*model.CustomerComment, error) {
	err := r.data.q.CustomerComment.WithContext(ctx).Create(c)
	if err != nil {
		return c, err
	}
	return c, nil
}
