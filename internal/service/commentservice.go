package service

import (
	"context"
	"time"

	pb "comment/api/comment/v1"
	v1 "comment/api/comment/v1"
	"comment/internal/biz"
	"comment/internal/data/model"

	"github.com/google/uuid"
)

type CommentServiceService struct {
	pb.UnimplementedCommentServiceServer

	// 聚合根
	customer  *biz.CustomerUsecase
	merchant  *biz.MerchantUsecase
	bus       *biz.BusUseCase
	readmodel *biz.ReadModelUsecase
}

// service 层
// 传递两个聚合根实体
func NewCommentServiceService(c *biz.CustomerUsecase, m *biz.MerchantUsecase, b *biz.BusUseCase, r *biz.ReadModelUsecase) *CommentServiceService {
	return &CommentServiceService{
		customer:  c,
		merchant:  m,
		bus:       b,
		readmodel: r,
	}
}

// GetComments 读模型
//
// 直接获取ES的数据
func (s *CommentServiceService) GetComments(ctx context.Context, req *pb.GetCommentsRequest) (*pb.GetCommentsResponse, error) {
	// TODO 查看商品SKUID=10的所有评价
	data, err := s.readmodel.GetAllCommentsBySKUID(ctx, req.SkuId)
	if err != nil {
		return &pb.GetCommentsResponse{}, err
	}

	return &pb.GetCommentsResponse{
		CommentList: data,
	}, nil

}

// ----------------------- customer --------------------------------//
// AddComment customer add comment
func (s *CommentServiceService) AddComment(ctx context.Context, req *pb.CommentRequest) (*pb.CommentResponse, error) {
	// 1. 参数格式转换 DTO转VO，从外部pb传递进来的参数
	// 需要构造成聚合根传递给领域业务，领域内的数据修改，只能依靠聚合根的数据
	comment, err := s.customer.AddComment(ctx, &model.CustomerComment{
		CustomerID: req.ConsumerId,
		CommentID:  uuid.New().String(),
		Version:    req.LastVersion + 1,
		Content:    req.CommentContent,
		SkuID:      req.SkuId,
	})

	// 2. 写入事件总线: 如果领域层执行成功，发送“消费者发表评价成功” 事件， 否则发送“消费者发表评价失败”
	if err != nil {
		bus.Publish("customer:AddComment:fail", "fail", comment)
	}

	bus.Publish("customer:AddComment:success", "success", comment)

	return &pb.CommentResponse{
		Success: true,
		Message: "success",
	}, nil
}

// AddReply customer reply to an existing comment
func (s *CommentServiceService) AddReply(ctx context.Context, req *pb.ReplyRequest) (*pb.ReplyResponse, error) {
	// 1. 参数格式转换 DTO转VO，从外部pb传递进来的参数
	// 需要构造成聚合根传递给领域业务，领域内的数据修改，只能依靠聚合根的数据
	comment, err := s.customer.ReplyComment(ctx, &model.CustomerComment{
		CustomerID:    req.ConsumerId,
		CommentID:     uuid.New().String(),
		Version:       req.LastVersion + 1,
		Content:       req.ReplyContent,
		SkuID:         req.SkuId,
		LastCommentID: req.LastCommentId,
		CreateAt:      time.Now(),
		UpdateAt:      time.Now(),
	})

	// 2. 写入事件总线: 如果领域层执行成功，发送“消费者回复评价成功” 事件， 否则发送“消费者回复评价失败”
	if err != nil {
		nerr := v1.ErrorUserIDError("写入数据库失败") // 使用自定义错误

		general_bus.Publish("customer:AddReply:fail", "AddReplyFail", comment, s.bus)
		return &pb.ReplyResponse{
			Success: false,
			Message: nerr.String(),
		}, err
	}
	general_bus.Publish("customer:AddReply:success", "AddReplySuccess", comment, s.bus)

	return &pb.ReplyResponse{
		Success: true,
		Message: "success",
	}, err
}

// ----------------------- merchant --------------------------------//
// AddProduct merchant add a product
func (s *CommentServiceService) AddProduct(ctx context.Context, req *pb.MerchantAddProductRequest) (*pb.MerchantAddProductResponse, error) {
	//
	return &pb.MerchantAddProductResponse{}, nil
}

// AddProductReply merchant reply to an existing comment
func (s *CommentServiceService) AddProductReply(ctx context.Context, req *pb.MerchantReplyRequest) (*pb.MerchantReplyResponse, error) {
	return &pb.MerchantReplyResponse{}, nil
}
