package data

import (
	"comment/internal/biz"
	"comment/internal/data/model"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"time"
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

// CreateComment cache aside: 使用事务写入mysql和mongo, 并且更新缓存
func (r *CustomerRepo) CreateComment(ctx context.Context, c *model.CustomerComment) (*model.CustomerComment, error) {
	tx := r.data.q.Begin()

	var err error
	// 1. 写入mysql
	if err = tx.CustomerComment.WithContext(ctx).Create(c); err != nil {
		return c, err
	}

	// 2. 写入mongo
	// mongo数据库要为这条数据添加一个额外字段：标记是否发送到kafka
	if _, err = r.data.mongo.Collection("AddCommentSuccess").InsertOne(ctx, c); err != nil {
		return c, err
	}

	// 把sendtomq 标记为未发送
	filter := bson.D{{"commentid", c.CommentID}}
	update := bson.D{{"$set", bson.D{{"sendtomq", false}}}}
	if _, err = r.data.mongo.Collection("AddCommentSuccess").UpdateOne(ctx, filter, update); err != nil {
		return c, err
	}

	// 在函数返回时检测 err，如果不为 nil，则执行回滚
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 提交事务
	if err = tx.Commit(); err != nil {
		fmt.Println("提交事务出错：", err)
		return c, err
	}

	// 更新缓存
	bytes, err := json.Marshal(c)
	if err != nil {
		r.log.Error(err)
		return c, err
	}
	if res := r.data.redis.Set(ctx, c.CommentID, bytes, 24*time.Hour); res.Err() != nil {
		return c, res.Err()
	}
	return c, err
}

func (r *CustomerRepo) ReplyComment(ctx context.Context, c *model.CustomerComment) (*model.CustomerComment, error) {
	// 遵循cache aside pattern
	// 根据评论id尝试从redis中获取，
	// 如果找到的话，比较两者的版本号
	// 如果新评论的版本号比旧评论的版本号低，不更新；否则用新评论覆盖旧评论
	// 如果没有找到，则尝试从mysql获取，
	// 如果找到的话，比较两者的版本号
	// 如果新评论的版本号比旧评论的版本号低，不更新；否则用新评论覆盖旧评论
	// 如果没有找到，直接插入

	// 1. 从存储中查找可能存在的评论，如果存在，尝试覆盖写
	res := r.data.redis.Get(ctx, c.CommentID)
	// 如果redis找不到，尝试从mysql中找
	if err := res.Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			r.log.Info(err)
			// 从mysql中找
			q := r.data.q
			cFromDB, err := q.CustomerComment.WithContext(ctx).Where(q.CustomerComment.CommentID.Eq(c.CommentID)).First()
			if err != nil {
				r.log.Error(err) // 如果mysql找不到 后续进行步骤2 写入mysql
				r.log.Info("mysql找不到数据，尝试写入新记录")
			} else {
				// 如果mysql找到的话，比较两者的版本号, 如果新评论的版本号比旧评论的版本号低，不更新直接返回
				// 如果新评论版本号高，则更新数据库
				if cFromDB.Version >= c.Version {
					return c, err
				} else {
					// cache aside: 更新数据库并删除缓存
					_, err := r.data.q.CustomerComment.WithContext(ctx).Updates(*c)
					if err != nil {
						r.log.Error(err)
						return c, err
					}
					return c, err
				}
			}
		}
	} else {
		// 如果redis找到的话，比较两者的版本号, 如果新评论的版本号比旧评论的版本号低，不更新直接返回
		// 否则，先把数据存到数据库中，成功后，再让缓存失效
		cFromRedis := new(model.CustomerComment)
		bytes, err := res.Bytes()
		if err != nil {
			r.log.Error()
			return c, err //获取redis结果失败返回
		} else {
			err := json.Unmarshal(bytes, cFromRedis)
			if err != nil {
				r.log.Error(err)
				return c, err // 反序列化失败返回
			}
			if cFromRedis.Version >= c.Version {
				return c, err
			} else {
				// cache aside 更新：先把数据存到数据库中，成功后，再让缓存失效
				_, err := r.data.q.CustomerComment.WithContext(ctx).Updates(*c)
				if err != nil {
					r.log.Error(err)
					return c, err // 更新数据库失败返回
				} else {
					if res := r.data.redis.Del(ctx, c.CommentID); res.Err() != nil {
						return c, res.Err() // 更新缓存失败返回
					}
				}
				return c, err
			}
		}
	}

	// 2. 如果存储找不到，则直接插入, 插入成功后更新缓存
	return r.CreateComment(ctx, c)
}
