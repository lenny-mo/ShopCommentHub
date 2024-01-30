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
	"gorm.io/gorm"
	"time"
)

type MerchantRepo struct {
	data *Data
	log  *log.Helper
}

// NewMerchantRepo
func NewMerchantRepo(data *Data, logger log.Logger) biz.MerchantRepo {
	return &MerchantRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *MerchantRepo) AddProduct(ctx context.Context, p *model.Product) (*model.Product, error) {
	// 1. 检查skuid + merchant_id 是否已经被添加过，同一个商家不能重复添加商品
	// 在mysql表中创建联合索引
	_, err := r.data.q.Product.WithContext(ctx).Where(
		r.data.q.Product.MerchantID.Eq(p.MerchantID),
		r.data.q.Product.SkuID.Eq(p.SkuID)).First()

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		if err == nil {
			r.log.Error("this merchant cannnot add repeated skuid")
			return p, errors.New("this merchant cannnot add repeated skuid")
		}
		r.log.Error(err)
		return p, err
	}

	// 2. 执行插入商品操作
	if err := r.data.q.Product.WithContext(ctx).Create(p); err != nil {
		r.log.Error(err)
		return p, err
	}

	return p, nil
}

// CreateComment 使用事务保证写入mysql和mongo
func (r *MerchantRepo) CreateComment(ctx context.Context, mc *model.MerchantComment) (*model.MerchantComment, error) {
	tx := r.data.q.Begin()

	var err error
	// 1. 写入mysql
	if err = tx.MerchantComment.WithContext(ctx).Create(mc); err != nil {
		return mc, err
	}

	// 2. 写入mongo
	// mongo数据库要为这条数据添加一个额外字段：标记是否发送到kafka
	if _, err = r.data.mongo.Collection("AddCommentSuccess").InsertOne(ctx, mc); err != nil {
		return mc, err
	}

	// 把sendtomq 标记为未发送
	filter := bson.D{{"commentid", mc.CommentID}}
	update := bson.D{{"$set", bson.D{{"sendtomq", false}}}}
	if _, err = r.data.mongo.Collection("AddCommentSuccess").UpdateOne(ctx, filter, update); err != nil {
		return mc, err
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
		return mc, err
	}

	// 更新缓存
	bytes, err := json.Marshal(mc)
	if err != nil {
		r.log.Error(err)
		return mc, err
	}
	if res := r.data.redis.Set(ctx, mc.CommentID, bytes, 24*time.Hour); res.Err() != nil {
		return mc, res.Err()
	}
	return mc, err
}

func (r *MerchantRepo) ReplyComment(ctx context.Context, mc *model.MerchantComment) (*model.MerchantComment, error) {

	// 遵循cache aside pattern
	// 检查skuid 对应的merchant_id 是否属于自己，不属于商家直接退出，商家只能回复自己的商品评论（防止水平越权）
	// 根据评论id尝试从redis中获取
	// 如果找到的话, 比较两者的版本号
	// 		如果新评论的版本号比旧评论的版本号低，不更新；否则用新评论覆盖旧评论
	// 如果没有找到，则尝试从mysql获取，
	// 如果找到的话，重复上面步骤
	// 如果没有找到，直接插入

	// 检查skuid 对应的merchant_id 是否属于自己，不属于商家直接退出，商家只能回复自己的商品评论（防止水平越权）
	if p, err := r.data.q.Product.WithContext(ctx).Where(r.data.q.Product.SkuID.Eq(mc.SkuID)).First(); err != nil {
		return mc, err // 找不到这个商品，不能评论
	} else {
		if p.MerchantID != mc.MerchantID {
			return mc, errors.New("this product doesn't belong to the merchant")
		}
	}

	// 1. 从存储中查找可能存在的评论，如果存在，尝试覆盖写
	res := r.data.redis.Get(ctx, mc.CommentID)
	// 如果redis找不到，尝试从mysql中找
	if err := res.Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			r.log.Info(err)
			// 从mysql中找
			q := r.data.q
			cFromDB, err := q.MerchantComment.WithContext(ctx).Where(q.MerchantComment.CommentID.Eq(mc.CommentID)).First()
			if err != nil {
				r.log.Error(err) // 如果mysql找不到 后续进行步骤2 写入mysql
				r.log.Info("mysql找不到数据，尝试写入新记录")
			} else {
				// 如果mysql找到的话，比较两者的版本号, 如果新评论的版本号比旧评论的版本号低，不更新直接返回
				// 如果新评论版本号高，则更新数据库
				if cFromDB.Version >= mc.Version {
					return mc, err
				} else {
					// cache aside: 更新数据库并删除缓存
					_, err := r.data.q.MerchantComment.WithContext(ctx).Updates(*mc)
					if err != nil {
						r.log.Error(err)
						return mc, err
					}
					if res := r.data.redis.Del(ctx, mc.CommentID); res.Err() != nil {
						r.log.Error(res.Err())
						return mc, res.Err()
					}
					return mc, err
				}
			}
		}
	} else {
		// 如果redis找到的话，比较两者的版本号, 如果新评论的版本号比旧评论的版本号低，不更新直接返回
		// 否则，先把数据存到数据库中，成功后，再让缓存失效
		cFromRedis := new(model.MerchantComment)
		bytes, err := res.Bytes()
		if err != nil {
			r.log.Error()
			return mc, err //获取redis结果失败返回
		} else {
			err := json.Unmarshal(bytes, cFromRedis)
			if err != nil {
				r.log.Error(err)
				return mc, err // 反序列化失败返回
			}
			if cFromRedis.Version >= mc.Version {
				return mc, err
			} else {
				// cache aside 更新：先把数据存到数据库中，成功后，再让缓存失效
				_, err := r.data.q.MerchantComment.WithContext(ctx).Updates(*mc)
				if err != nil {
					r.log.Error(err)
					return mc, err // 更新数据库失败返回
				} else {
					if res := r.data.redis.Del(ctx, mc.CommentID); res.Err() != nil {
						return mc, res.Err() // 更新缓存失败返回
					}
				}
				return mc, err
			}
		}
	}

	// 2. 如果存储找不到，则直接插入, 插入成功后更新缓存
	return r.CreateComment(ctx, mc)
}
