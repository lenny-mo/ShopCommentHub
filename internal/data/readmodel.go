package data

import (
	v1 "comment/api/comment/v1"
	"comment/internal/biz"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/go-kratos/kratos/v2/log"
	"strconv"
	"time"
)

// 实现 ReadModelRepo interfance
type ReadModelRepo struct {
	data *Data
	log  *log.Helper
}

// ReadModelRepo
func NewReadModelRepo(data *Data, logger log.Logger) biz.ReadModelRepo {
	return &ReadModelRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// Msg 从es中恢复出来的结构体
type Msg struct {
	ID            int64     `json:"id"`
	CreateAt      time.Time `json:"create_at"`
	UpdateAt      time.Time `json:"update_at"`
	Version       int64     `json:"version"`
	CustomerID    int64     `json:"customer_id"`
	CommentID     string    `json:"comment_id"`
	LastCommentID string    `json:"last_comment_id"`
	Content       string    `json:"content"`
	SKUID         string    `json:"sku_id"`
}

// GetAllCommentsBySKUID 从es 中根据skuid获取所有评论数据
//
// es 中获取到的数据内的Source_格式如下
//
//	{
//	   "id":50,
//	   "create_at":"2024-01-25T03:13:17.176Z",
//	   "update_at":"2024-01-25T03:13:17.176Z",
//	   "version":3,
//	   "customer_id":403,
//	   "comment_id":"7aae98f8-8155-44e0-9550-b4e4c8d61b1f",
//	   "last_comment_id":"",
//	   "content":"27 line: ES没问题",
//	   "sku_id":"1"
//	}
func (r *ReadModelRepo) GetAllCommentsBySKUID(ctx context.Context, skuid int32) ([]*v1.Comment, error) {

	// 创建查询
	query := &types.Query{
		Match: map[string]types.MatchQuery{
			"sku_id": {Query: strconv.FormatInt(int64(skuid), 10)},
		},
	}

	// 返回的结果
	res, err := r.data.es.client.Search().Index(r.data.es.index).
		Query(query).Do(ctx)

	if err != nil {
		r.log.Error(err)
		return nil, err
	}

	// 把结果拼接成我们想要的slice
	var comments []*v1.Comment

	// 遍历结果集合
	for _, v := range res.Hits.Hits {
		fmt.Printf("查询结果：%s\n", v.Source_)
		var c Msg
		// 反序列化
		if err := json.Unmarshal(v.Source_, &c); err != nil {
			r.data.log.Error(err)
		} else {
			comments = append(comments, &v1.Comment{
				CommentId:      c.CommentID,
				ConsumerId:     c.CustomerID,
				CommentContent: c.Content,
			})
		}
	}

	return comments, nil
}
