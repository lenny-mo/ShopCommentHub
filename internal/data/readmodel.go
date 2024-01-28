package data

import (
	v1 "comment/api/comment/v1"
	"comment/internal/biz"
	"comment/internal/data/singleflight"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

var single *singleflight.Group

func init() {
	single = singleflight.NewGroup()
}

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

	// 阶段1：获取到skuid 对应的字节流数据
	// 1. 查询缓存，只有返回找不到err 再查es
	// 2. 使用singleflight 查询es
	// 3. 获取结果并且更新缓存
	bytes, err := r.getAllCommentsByRedis(ctx, int(skuid))
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// 使用自定义的singleflight 查询ES
			bytes, err = r.getAllCommentsByES(ctx, int(skuid))
			if err != nil {
				return nil, err
			}
			// 写回redis
			go r.setAllCommentsByRedis(ctx, int(skuid), bytes)
		} else {
			return nil, err // 如果redis出现其他错误，直接直接返回空值
		}
	}

	// 阶段2: 对数据进行unmarshal
	var res types.HitsMetadata
	if err := json.Unmarshal(bytes, &res); err != nil {
		return nil, err
	}
	var comments []*v1.Comment

	// 遍历结果集合
	for _, v := range res.Hits {
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

// 根据skuid 查询redis
func (r *ReadModelRepo) getAllCommentsByRedis(ctx context.Context, skuid int) ([]byte, error) {
	res := r.data.redis.Get(ctx, strconv.Itoa(skuid))
	return res.Bytes()
}

// 根据skuid 设置redis
func (r *ReadModelRepo) setAllCommentsByRedis(ctx context.Context, skuid int, val []byte) error {
	r.data.redis.Set(ctx, strconv.Itoa(skuid), val, 24*time.Hour)
	return nil
}

// 根据skuid 查询es, 使用singleflight
func (r *ReadModelRepo) getAllCommentsByES(ctx context.Context, skuid int) ([]byte, error) {

	res, err := single.Do(strconv.Itoa(skuid), func() (interface{}, error) {
		// 创建es查询 使用filter
		query := &types.Query{
			Bool: &types.BoolQuery{
				Filter: []types.Query{
					{
						Term: map[string]types.TermQuery{
							"sku_id": {Value: strconv.FormatInt(int64(skuid), 10)},
						},
					},
				},
			},
		}

		// 返回的结果
		res, err := r.data.es.client.Search().Index(r.data.es.index).
			Query(query).Do(ctx)

		if err != nil {
			r.log.Error(err)
			return nil, err
		}

		return json.Marshal(res.Hits) // 把结果转化成bytes存储到redis
	})

	return res.([]byte), err
}
