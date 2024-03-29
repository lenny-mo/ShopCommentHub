// Code generated by protoc-gen-go-http. DO NOT EDIT.
// versions:
// - protoc-gen-go-http v2.7.2
// - protoc             v4.23.3
// source: api/comment/v1/comment.proto

package v1

import (
	context "context"
	http "github.com/go-kratos/kratos/v2/transport/http"
	binding "github.com/go-kratos/kratos/v2/transport/http/binding"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the kratos package it is being compiled against.
var _ = new(context.Context)
var _ = binding.EncodeURL

const _ = http.SupportPackageIsVersion1

const OperationCommentServiceAddComment = "/api.comment.v1.CommentService/AddComment"
const OperationCommentServiceAddProduct = "/api.comment.v1.CommentService/AddProduct"
const OperationCommentServiceAddProductReply = "/api.comment.v1.CommentService/AddProductReply"
const OperationCommentServiceAddReply = "/api.comment.v1.CommentService/AddReply"
const OperationCommentServiceGetComments = "/api.comment.v1.CommentService/GetComments"

type CommentServiceHTTPServer interface {
	// AddComment 消费者流程
	// 对商品SKUID=10发表评价
	AddComment(context.Context, *CommentRequest) (*CommentResponse, error)
	// AddProduct 商家流程
	// 发布商品SKUID=10
	AddProduct(context.Context, *MerchantAddProductRequest) (*MerchantAddProductResponse, error)
	// AddProductReply 对商品SKUID=10的其中一条评价发表回复
	AddProductReply(context.Context, *MerchantReplyRequest) (*MerchantReplyResponse, error)
	// AddReply 对商品SKUID=10的其中的一条评价发表回复
	AddReply(context.Context, *ReplyRequest) (*ReplyResponse, error)
	// GetComments common 读模型
	// 查看商品SKUID=10的所有评价
	GetComments(context.Context, *GetCommentsRequest) (*GetCommentsResponse, error)
}

func RegisterCommentServiceHTTPServer(s *http.Server, srv CommentServiceHTTPServer) {
	r := s.Route("/")
	r.GET("/product-review/v1/get-comments/{sku_id}", _CommentService_GetComments0_HTTP_Handler(srv))
	r.POST("/product-review/v1/add-comment", _CommentService_AddComment0_HTTP_Handler(srv))
	r.POST("/product-review/v1/add-reply", _CommentService_AddReply0_HTTP_Handler(srv))
	r.POST("/product-review/v1/add-product", _CommentService_AddProduct0_HTTP_Handler(srv))
	r.POST("/product-review/v1/add-product-reply", _CommentService_AddProductReply0_HTTP_Handler(srv))
}

func _CommentService_GetComments0_HTTP_Handler(srv CommentServiceHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in GetCommentsRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		if err := ctx.BindVars(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationCommentServiceGetComments)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.GetComments(ctx, req.(*GetCommentsRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*GetCommentsResponse)
		return ctx.Result(200, reply)
	}
}

func _CommentService_AddComment0_HTTP_Handler(srv CommentServiceHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in CommentRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationCommentServiceAddComment)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.AddComment(ctx, req.(*CommentRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*CommentResponse)
		return ctx.Result(200, reply)
	}
}

func _CommentService_AddReply0_HTTP_Handler(srv CommentServiceHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in ReplyRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationCommentServiceAddReply)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.AddReply(ctx, req.(*ReplyRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*ReplyResponse)
		return ctx.Result(200, reply)
	}
}

func _CommentService_AddProduct0_HTTP_Handler(srv CommentServiceHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in MerchantAddProductRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationCommentServiceAddProduct)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.AddProduct(ctx, req.(*MerchantAddProductRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*MerchantAddProductResponse)
		return ctx.Result(200, reply)
	}
}

func _CommentService_AddProductReply0_HTTP_Handler(srv CommentServiceHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in MerchantReplyRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationCommentServiceAddProductReply)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.AddProductReply(ctx, req.(*MerchantReplyRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*MerchantReplyResponse)
		return ctx.Result(200, reply)
	}
}

type CommentServiceHTTPClient interface {
	AddComment(ctx context.Context, req *CommentRequest, opts ...http.CallOption) (rsp *CommentResponse, err error)
	AddProduct(ctx context.Context, req *MerchantAddProductRequest, opts ...http.CallOption) (rsp *MerchantAddProductResponse, err error)
	AddProductReply(ctx context.Context, req *MerchantReplyRequest, opts ...http.CallOption) (rsp *MerchantReplyResponse, err error)
	AddReply(ctx context.Context, req *ReplyRequest, opts ...http.CallOption) (rsp *ReplyResponse, err error)
	GetComments(ctx context.Context, req *GetCommentsRequest, opts ...http.CallOption) (rsp *GetCommentsResponse, err error)
}

type CommentServiceHTTPClientImpl struct {
	cc *http.Client
}

func NewCommentServiceHTTPClient(client *http.Client) CommentServiceHTTPClient {
	return &CommentServiceHTTPClientImpl{client}
}

func (c *CommentServiceHTTPClientImpl) AddComment(ctx context.Context, in *CommentRequest, opts ...http.CallOption) (*CommentResponse, error) {
	var out CommentResponse
	pattern := "/product-review/v1/add-comment"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationCommentServiceAddComment))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, err
}

func (c *CommentServiceHTTPClientImpl) AddProduct(ctx context.Context, in *MerchantAddProductRequest, opts ...http.CallOption) (*MerchantAddProductResponse, error) {
	var out MerchantAddProductResponse
	pattern := "/product-review/v1/add-product"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationCommentServiceAddProduct))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, err
}

func (c *CommentServiceHTTPClientImpl) AddProductReply(ctx context.Context, in *MerchantReplyRequest, opts ...http.CallOption) (*MerchantReplyResponse, error) {
	var out MerchantReplyResponse
	pattern := "/product-review/v1/add-product-reply"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationCommentServiceAddProductReply))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, err
}

func (c *CommentServiceHTTPClientImpl) AddReply(ctx context.Context, in *ReplyRequest, opts ...http.CallOption) (*ReplyResponse, error) {
	var out ReplyResponse
	pattern := "/product-review/v1/add-reply"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationCommentServiceAddReply))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, err
}

func (c *CommentServiceHTTPClientImpl) GetComments(ctx context.Context, in *GetCommentsRequest, opts ...http.CallOption) (*GetCommentsResponse, error) {
	var out GetCommentsResponse
	pattern := "/product-review/v1/get-comments/{sku_id}"
	path := binding.EncodeURL(pattern, in, true)
	opts = append(opts, http.Operation(OperationCommentServiceGetComments))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "GET", path, nil, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, err
}
