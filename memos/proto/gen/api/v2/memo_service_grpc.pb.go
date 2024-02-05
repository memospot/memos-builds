// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: api/v2/memo_service.proto

package apiv2

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	MemoService_CreateMemo_FullMethodName        = "/memos.api.v2.MemoService/CreateMemo"
	MemoService_ListMemos_FullMethodName         = "/memos.api.v2.MemoService/ListMemos"
	MemoService_GetMemo_FullMethodName           = "/memos.api.v2.MemoService/GetMemo"
	MemoService_GetMemoByName_FullMethodName     = "/memos.api.v2.MemoService/GetMemoByName"
	MemoService_UpdateMemo_FullMethodName        = "/memos.api.v2.MemoService/UpdateMemo"
	MemoService_DeleteMemo_FullMethodName        = "/memos.api.v2.MemoService/DeleteMemo"
	MemoService_SetMemoResources_FullMethodName  = "/memos.api.v2.MemoService/SetMemoResources"
	MemoService_ListMemoResources_FullMethodName = "/memos.api.v2.MemoService/ListMemoResources"
	MemoService_SetMemoRelations_FullMethodName  = "/memos.api.v2.MemoService/SetMemoRelations"
	MemoService_ListMemoRelations_FullMethodName = "/memos.api.v2.MemoService/ListMemoRelations"
	MemoService_CreateMemoComment_FullMethodName = "/memos.api.v2.MemoService/CreateMemoComment"
	MemoService_ListMemoComments_FullMethodName  = "/memos.api.v2.MemoService/ListMemoComments"
	MemoService_ExportMemos_FullMethodName       = "/memos.api.v2.MemoService/ExportMemos"
	MemoService_GetUserMemosStats_FullMethodName = "/memos.api.v2.MemoService/GetUserMemosStats"
)

// MemoServiceClient is the client API for MemoService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MemoServiceClient interface {
	// CreateMemo creates a memo.
	CreateMemo(ctx context.Context, in *CreateMemoRequest, opts ...grpc.CallOption) (*CreateMemoResponse, error)
	// ListMemos lists memos with pagination and filter.
	ListMemos(ctx context.Context, in *ListMemosRequest, opts ...grpc.CallOption) (*ListMemosResponse, error)
	// GetMemo gets a memo by id.
	GetMemo(ctx context.Context, in *GetMemoRequest, opts ...grpc.CallOption) (*GetMemoResponse, error)
	// GetMemoByName gets a memo by name.
	GetMemoByName(ctx context.Context, in *GetMemoByNameRequest, opts ...grpc.CallOption) (*GetMemoByNameResponse, error)
	// UpdateMemo updates a memo.
	UpdateMemo(ctx context.Context, in *UpdateMemoRequest, opts ...grpc.CallOption) (*UpdateMemoResponse, error)
	// DeleteMemo deletes a memo by id.
	DeleteMemo(ctx context.Context, in *DeleteMemoRequest, opts ...grpc.CallOption) (*DeleteMemoResponse, error)
	// SetMemoResources sets resources for a memo.
	SetMemoResources(ctx context.Context, in *SetMemoResourcesRequest, opts ...grpc.CallOption) (*SetMemoResourcesResponse, error)
	// ListMemoResources lists resources for a memo.
	ListMemoResources(ctx context.Context, in *ListMemoResourcesRequest, opts ...grpc.CallOption) (*ListMemoResourcesResponse, error)
	// SetMemoRelations sets relations for a memo.
	SetMemoRelations(ctx context.Context, in *SetMemoRelationsRequest, opts ...grpc.CallOption) (*SetMemoRelationsResponse, error)
	// ListMemoRelations lists relations for a memo.
	ListMemoRelations(ctx context.Context, in *ListMemoRelationsRequest, opts ...grpc.CallOption) (*ListMemoRelationsResponse, error)
	// CreateMemoComment creates a comment for a memo.
	CreateMemoComment(ctx context.Context, in *CreateMemoCommentRequest, opts ...grpc.CallOption) (*CreateMemoCommentResponse, error)
	// ListMemoComments lists comments for a memo.
	ListMemoComments(ctx context.Context, in *ListMemoCommentsRequest, opts ...grpc.CallOption) (*ListMemoCommentsResponse, error)
	// ExportMemos exports memos.
	ExportMemos(ctx context.Context, in *ExportMemosRequest, opts ...grpc.CallOption) (*ExportMemosResponse, error)
	// GetUserMemosStats gets stats of memos for a user.
	GetUserMemosStats(ctx context.Context, in *GetUserMemosStatsRequest, opts ...grpc.CallOption) (*GetUserMemosStatsResponse, error)
}

type memoServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewMemoServiceClient(cc grpc.ClientConnInterface) MemoServiceClient {
	return &memoServiceClient{cc}
}

func (c *memoServiceClient) CreateMemo(ctx context.Context, in *CreateMemoRequest, opts ...grpc.CallOption) (*CreateMemoResponse, error) {
	out := new(CreateMemoResponse)
	err := c.cc.Invoke(ctx, MemoService_CreateMemo_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *memoServiceClient) ListMemos(ctx context.Context, in *ListMemosRequest, opts ...grpc.CallOption) (*ListMemosResponse, error) {
	out := new(ListMemosResponse)
	err := c.cc.Invoke(ctx, MemoService_ListMemos_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *memoServiceClient) GetMemo(ctx context.Context, in *GetMemoRequest, opts ...grpc.CallOption) (*GetMemoResponse, error) {
	out := new(GetMemoResponse)
	err := c.cc.Invoke(ctx, MemoService_GetMemo_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *memoServiceClient) GetMemoByName(ctx context.Context, in *GetMemoByNameRequest, opts ...grpc.CallOption) (*GetMemoByNameResponse, error) {
	out := new(GetMemoByNameResponse)
	err := c.cc.Invoke(ctx, MemoService_GetMemoByName_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *memoServiceClient) UpdateMemo(ctx context.Context, in *UpdateMemoRequest, opts ...grpc.CallOption) (*UpdateMemoResponse, error) {
	out := new(UpdateMemoResponse)
	err := c.cc.Invoke(ctx, MemoService_UpdateMemo_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *memoServiceClient) DeleteMemo(ctx context.Context, in *DeleteMemoRequest, opts ...grpc.CallOption) (*DeleteMemoResponse, error) {
	out := new(DeleteMemoResponse)
	err := c.cc.Invoke(ctx, MemoService_DeleteMemo_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *memoServiceClient) SetMemoResources(ctx context.Context, in *SetMemoResourcesRequest, opts ...grpc.CallOption) (*SetMemoResourcesResponse, error) {
	out := new(SetMemoResourcesResponse)
	err := c.cc.Invoke(ctx, MemoService_SetMemoResources_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *memoServiceClient) ListMemoResources(ctx context.Context, in *ListMemoResourcesRequest, opts ...grpc.CallOption) (*ListMemoResourcesResponse, error) {
	out := new(ListMemoResourcesResponse)
	err := c.cc.Invoke(ctx, MemoService_ListMemoResources_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *memoServiceClient) SetMemoRelations(ctx context.Context, in *SetMemoRelationsRequest, opts ...grpc.CallOption) (*SetMemoRelationsResponse, error) {
	out := new(SetMemoRelationsResponse)
	err := c.cc.Invoke(ctx, MemoService_SetMemoRelations_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *memoServiceClient) ListMemoRelations(ctx context.Context, in *ListMemoRelationsRequest, opts ...grpc.CallOption) (*ListMemoRelationsResponse, error) {
	out := new(ListMemoRelationsResponse)
	err := c.cc.Invoke(ctx, MemoService_ListMemoRelations_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *memoServiceClient) CreateMemoComment(ctx context.Context, in *CreateMemoCommentRequest, opts ...grpc.CallOption) (*CreateMemoCommentResponse, error) {
	out := new(CreateMemoCommentResponse)
	err := c.cc.Invoke(ctx, MemoService_CreateMemoComment_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *memoServiceClient) ListMemoComments(ctx context.Context, in *ListMemoCommentsRequest, opts ...grpc.CallOption) (*ListMemoCommentsResponse, error) {
	out := new(ListMemoCommentsResponse)
	err := c.cc.Invoke(ctx, MemoService_ListMemoComments_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *memoServiceClient) ExportMemos(ctx context.Context, in *ExportMemosRequest, opts ...grpc.CallOption) (*ExportMemosResponse, error) {
	out := new(ExportMemosResponse)
	err := c.cc.Invoke(ctx, MemoService_ExportMemos_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *memoServiceClient) GetUserMemosStats(ctx context.Context, in *GetUserMemosStatsRequest, opts ...grpc.CallOption) (*GetUserMemosStatsResponse, error) {
	out := new(GetUserMemosStatsResponse)
	err := c.cc.Invoke(ctx, MemoService_GetUserMemosStats_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MemoServiceServer is the server API for MemoService service.
// All implementations must embed UnimplementedMemoServiceServer
// for forward compatibility
type MemoServiceServer interface {
	// CreateMemo creates a memo.
	CreateMemo(context.Context, *CreateMemoRequest) (*CreateMemoResponse, error)
	// ListMemos lists memos with pagination and filter.
	ListMemos(context.Context, *ListMemosRequest) (*ListMemosResponse, error)
	// GetMemo gets a memo by id.
	GetMemo(context.Context, *GetMemoRequest) (*GetMemoResponse, error)
	// GetMemoByName gets a memo by name.
	GetMemoByName(context.Context, *GetMemoByNameRequest) (*GetMemoByNameResponse, error)
	// UpdateMemo updates a memo.
	UpdateMemo(context.Context, *UpdateMemoRequest) (*UpdateMemoResponse, error)
	// DeleteMemo deletes a memo by id.
	DeleteMemo(context.Context, *DeleteMemoRequest) (*DeleteMemoResponse, error)
	// SetMemoResources sets resources for a memo.
	SetMemoResources(context.Context, *SetMemoResourcesRequest) (*SetMemoResourcesResponse, error)
	// ListMemoResources lists resources for a memo.
	ListMemoResources(context.Context, *ListMemoResourcesRequest) (*ListMemoResourcesResponse, error)
	// SetMemoRelations sets relations for a memo.
	SetMemoRelations(context.Context, *SetMemoRelationsRequest) (*SetMemoRelationsResponse, error)
	// ListMemoRelations lists relations for a memo.
	ListMemoRelations(context.Context, *ListMemoRelationsRequest) (*ListMemoRelationsResponse, error)
	// CreateMemoComment creates a comment for a memo.
	CreateMemoComment(context.Context, *CreateMemoCommentRequest) (*CreateMemoCommentResponse, error)
	// ListMemoComments lists comments for a memo.
	ListMemoComments(context.Context, *ListMemoCommentsRequest) (*ListMemoCommentsResponse, error)
	// ExportMemos exports memos.
	ExportMemos(context.Context, *ExportMemosRequest) (*ExportMemosResponse, error)
	// GetUserMemosStats gets stats of memos for a user.
	GetUserMemosStats(context.Context, *GetUserMemosStatsRequest) (*GetUserMemosStatsResponse, error)
	mustEmbedUnimplementedMemoServiceServer()
}

// UnimplementedMemoServiceServer must be embedded to have forward compatible implementations.
type UnimplementedMemoServiceServer struct {
}

func (UnimplementedMemoServiceServer) CreateMemo(context.Context, *CreateMemoRequest) (*CreateMemoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateMemo not implemented")
}
func (UnimplementedMemoServiceServer) ListMemos(context.Context, *ListMemosRequest) (*ListMemosResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListMemos not implemented")
}
func (UnimplementedMemoServiceServer) GetMemo(context.Context, *GetMemoRequest) (*GetMemoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMemo not implemented")
}
func (UnimplementedMemoServiceServer) GetMemoByName(context.Context, *GetMemoByNameRequest) (*GetMemoByNameResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMemoByName not implemented")
}
func (UnimplementedMemoServiceServer) UpdateMemo(context.Context, *UpdateMemoRequest) (*UpdateMemoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateMemo not implemented")
}
func (UnimplementedMemoServiceServer) DeleteMemo(context.Context, *DeleteMemoRequest) (*DeleteMemoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteMemo not implemented")
}
func (UnimplementedMemoServiceServer) SetMemoResources(context.Context, *SetMemoResourcesRequest) (*SetMemoResourcesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetMemoResources not implemented")
}
func (UnimplementedMemoServiceServer) ListMemoResources(context.Context, *ListMemoResourcesRequest) (*ListMemoResourcesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListMemoResources not implemented")
}
func (UnimplementedMemoServiceServer) SetMemoRelations(context.Context, *SetMemoRelationsRequest) (*SetMemoRelationsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetMemoRelations not implemented")
}
func (UnimplementedMemoServiceServer) ListMemoRelations(context.Context, *ListMemoRelationsRequest) (*ListMemoRelationsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListMemoRelations not implemented")
}
func (UnimplementedMemoServiceServer) CreateMemoComment(context.Context, *CreateMemoCommentRequest) (*CreateMemoCommentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateMemoComment not implemented")
}
func (UnimplementedMemoServiceServer) ListMemoComments(context.Context, *ListMemoCommentsRequest) (*ListMemoCommentsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListMemoComments not implemented")
}
func (UnimplementedMemoServiceServer) ExportMemos(context.Context, *ExportMemosRequest) (*ExportMemosResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExportMemos not implemented")
}
func (UnimplementedMemoServiceServer) GetUserMemosStats(context.Context, *GetUserMemosStatsRequest) (*GetUserMemosStatsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserMemosStats not implemented")
}
func (UnimplementedMemoServiceServer) mustEmbedUnimplementedMemoServiceServer() {}

// UnsafeMemoServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MemoServiceServer will
// result in compilation errors.
type UnsafeMemoServiceServer interface {
	mustEmbedUnimplementedMemoServiceServer()
}

func RegisterMemoServiceServer(s grpc.ServiceRegistrar, srv MemoServiceServer) {
	s.RegisterService(&MemoService_ServiceDesc, srv)
}

func _MemoService_CreateMemo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateMemoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MemoServiceServer).CreateMemo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MemoService_CreateMemo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MemoServiceServer).CreateMemo(ctx, req.(*CreateMemoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MemoService_ListMemos_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListMemosRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MemoServiceServer).ListMemos(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MemoService_ListMemos_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MemoServiceServer).ListMemos(ctx, req.(*ListMemosRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MemoService_GetMemo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetMemoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MemoServiceServer).GetMemo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MemoService_GetMemo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MemoServiceServer).GetMemo(ctx, req.(*GetMemoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MemoService_GetMemoByName_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetMemoByNameRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MemoServiceServer).GetMemoByName(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MemoService_GetMemoByName_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MemoServiceServer).GetMemoByName(ctx, req.(*GetMemoByNameRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MemoService_UpdateMemo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateMemoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MemoServiceServer).UpdateMemo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MemoService_UpdateMemo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MemoServiceServer).UpdateMemo(ctx, req.(*UpdateMemoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MemoService_DeleteMemo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteMemoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MemoServiceServer).DeleteMemo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MemoService_DeleteMemo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MemoServiceServer).DeleteMemo(ctx, req.(*DeleteMemoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MemoService_SetMemoResources_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetMemoResourcesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MemoServiceServer).SetMemoResources(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MemoService_SetMemoResources_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MemoServiceServer).SetMemoResources(ctx, req.(*SetMemoResourcesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MemoService_ListMemoResources_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListMemoResourcesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MemoServiceServer).ListMemoResources(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MemoService_ListMemoResources_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MemoServiceServer).ListMemoResources(ctx, req.(*ListMemoResourcesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MemoService_SetMemoRelations_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetMemoRelationsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MemoServiceServer).SetMemoRelations(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MemoService_SetMemoRelations_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MemoServiceServer).SetMemoRelations(ctx, req.(*SetMemoRelationsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MemoService_ListMemoRelations_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListMemoRelationsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MemoServiceServer).ListMemoRelations(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MemoService_ListMemoRelations_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MemoServiceServer).ListMemoRelations(ctx, req.(*ListMemoRelationsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MemoService_CreateMemoComment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateMemoCommentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MemoServiceServer).CreateMemoComment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MemoService_CreateMemoComment_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MemoServiceServer).CreateMemoComment(ctx, req.(*CreateMemoCommentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MemoService_ListMemoComments_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListMemoCommentsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MemoServiceServer).ListMemoComments(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MemoService_ListMemoComments_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MemoServiceServer).ListMemoComments(ctx, req.(*ListMemoCommentsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MemoService_ExportMemos_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExportMemosRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MemoServiceServer).ExportMemos(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MemoService_ExportMemos_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MemoServiceServer).ExportMemos(ctx, req.(*ExportMemosRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MemoService_GetUserMemosStats_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUserMemosStatsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MemoServiceServer).GetUserMemosStats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MemoService_GetUserMemosStats_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MemoServiceServer).GetUserMemosStats(ctx, req.(*GetUserMemosStatsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// MemoService_ServiceDesc is the grpc.ServiceDesc for MemoService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MemoService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "memos.api.v2.MemoService",
	HandlerType: (*MemoServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateMemo",
			Handler:    _MemoService_CreateMemo_Handler,
		},
		{
			MethodName: "ListMemos",
			Handler:    _MemoService_ListMemos_Handler,
		},
		{
			MethodName: "GetMemo",
			Handler:    _MemoService_GetMemo_Handler,
		},
		{
			MethodName: "GetMemoByName",
			Handler:    _MemoService_GetMemoByName_Handler,
		},
		{
			MethodName: "UpdateMemo",
			Handler:    _MemoService_UpdateMemo_Handler,
		},
		{
			MethodName: "DeleteMemo",
			Handler:    _MemoService_DeleteMemo_Handler,
		},
		{
			MethodName: "SetMemoResources",
			Handler:    _MemoService_SetMemoResources_Handler,
		},
		{
			MethodName: "ListMemoResources",
			Handler:    _MemoService_ListMemoResources_Handler,
		},
		{
			MethodName: "SetMemoRelations",
			Handler:    _MemoService_SetMemoRelations_Handler,
		},
		{
			MethodName: "ListMemoRelations",
			Handler:    _MemoService_ListMemoRelations_Handler,
		},
		{
			MethodName: "CreateMemoComment",
			Handler:    _MemoService_CreateMemoComment_Handler,
		},
		{
			MethodName: "ListMemoComments",
			Handler:    _MemoService_ListMemoComments_Handler,
		},
		{
			MethodName: "ExportMemos",
			Handler:    _MemoService_ExportMemos_Handler,
		},
		{
			MethodName: "GetUserMemosStats",
			Handler:    _MemoService_GetUserMemosStats_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/v2/memo_service.proto",
}