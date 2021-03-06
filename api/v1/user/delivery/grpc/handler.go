//
//  Practicing gRPC
//
//  Copyright © 2020. All rights reserved.
//

package grpc

import (
	"github.com/moemoe89/go-grpc-server-tisa/api/v1/api_struct/form"
	"github.com/moemoe89/go-grpc-server-tisa/api/v1/api_struct/model"
	usr "github.com/moemoe89/go-grpc-server-tisa/api/v1/user"
	usrProto "github.com/moemoe89/go-grpc-server-tisa/api/v1/user/delivery/grpc/proto"

	"context"
	"errors"
	"math"
	"strings"

	ts "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/moemoe89/go-helpers"
	"github.com/rs/xid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	svc usr.Service
}

func NewAUserServerGrpc(s *grpc.Server, svc usr.Service) *server {
	usrCtrl := &server{svc}
	usrProto.RegisterUserServiceServer(s, usrCtrl)
	reflection.Register(s)
	return usrCtrl
}

func (s *server) Create(c context.Context, r *usrProto.UserCreateReq) (*usrProto.User, error) {
	req := &form.UserForm{}
	req.ID = xid.New().String()
	req.Name = r.GetName()
	req.Phone = r.GetPhone()
	req.Email = r.GetEmail()
	req.Address = r.GetAddress()

	errs := req.Validate()
	if len(errs) > 0 {
		return nil, errors.New(errs[0])
	}

	user, err := s.svc.Create(req)
	if err != nil {
		return nil, err
	}

	createdAt := &ts.Timestamp{
		Seconds: user.CreatedAt.Unix(),
	}

	updatedAt := &ts.Timestamp{
		Seconds: user.UpdatedAt.Unix(),
	}

	resp := &usrProto.User{
		Id:        user.ID,
		Name:      user.Name,
		Phone:     user.Phone,
		Email:     user.Email,
		Address:   user.Address,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	return resp, nil
}

func (s *server) List(c context.Context, r *usrProto.UsersReq) (*usrProto.Users, error) {
	userModel := model.UserModel{}

	offset, perPage, showPage, err := helpers.PaginationSetter(r.GetPerPage(), r.GetPage())
	if err != nil {
		return nil, err
	}

	orderBy := r.GetOrderBy()
	orderBy = helpers.OrderByHandler(orderBy, "db", userModel)
	if len(orderBy) < 1 {
		orderBy = "created_at DESC"
	}

	where := "WHERE deleted_at IS NULL"
	filter := map[string]interface{}{}

	name := r.GetName()
	if len(name) > 0 {
		where += " AND name LIKE :name"
		filter["name"] = "%" + name + "%"
	}

	email := r.GetEmail()
	if len(email) > 0 {
		where += " AND email LIKE :email"
		filter["email"] = "%" + email + "%"
	}

	phone := r.GetPhone()
	if len(phone) > 0 {
		where += " AND phone LIKE :phone"
		filter["phone"] = "%" + phone + "%"
	}

	createdAtStart := r.GetCreatedAtStart()
	if len(createdAtStart) > 0 {
		where += " AND created_at >= :created_at_start"
		filter["created_at_start"] = createdAtStart
	}

	createdAtEnd := r.GetCreatedAtEnd()
	if len(createdAtEnd) > 0 {
		where += " AND created_at <= :created_at_end"
		filter["created_at_end"] = createdAtEnd
	}

	filterCount := filter
	filter["limit"] = perPage
	filter["offset"] = offset

	selectField := model.UserSelectField
	filterField := r.GetSelectField()
	if len(filterField) > 0 {
		res := helpers.CheckInTag(userModel, filterField, "db")
		if len(res) > 0 {
			selectField = strings.Join(res, ",")
		}
	}

	usersRaw, count, err := s.svc.List(filter, filterCount, where, orderBy, selectField)
	if err != nil {
		return nil, err
	}

	totalPage := int(math.Ceil(float64(count) / float64(perPage)))

	var users []*usrProto.User
	for _, user := range usersRaw {
		createdAt := &ts.Timestamp{
			Seconds: user.CreatedAt.Unix(),
		}

		updatedAt := &ts.Timestamp{
			Seconds: user.UpdatedAt.Unix(),
		}

		user := &usrProto.User{
			Id:        user.ID,
			Name:      user.Name,
			Phone:     user.Phone,
			Email:     user.Email,
			Address:   user.Address,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}

		users = append(users, user)
	}

	resp := &usrProto.Users{
		Users:     users,
		Page:      int64(showPage),
		PerPage:   int64(perPage),
		TotalPage: int64(totalPage),
		TotalData: int64(count),
	}

	return resp, nil
}

func (s *server) Detail(c context.Context, r *usrProto.UserIDReq) (*usrProto.User, error) {
	id := r.GetId()
	user, err := s.svc.Detail(id, model.UserSelectField)
	if err != nil {
		return nil, err
	}

	createdAt := &ts.Timestamp{
		Seconds: user.CreatedAt.Unix(),
	}

	updatedAt := &ts.Timestamp{
		Seconds: user.UpdatedAt.Unix(),
	}

	resp := &usrProto.User{
		Id:        user.ID,
		Name:      user.Name,
		Phone:     user.Phone,
		Email:     user.Email,
		Address:   user.Address,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	return resp, nil
}

func (s *server) Update(c context.Context, r *usrProto.UserUpdateReq) (*usrProto.User, error) {
	id := r.GetId()
	user, err := s.svc.Detail(id, "id")
	if err != nil {
		return nil, err
	}

	req := &form.UserForm{}
	req.Name = r.GetName()
	req.Phone = r.GetPhone()
	req.Email = r.GetEmail()
	req.Address = r.GetAddress()

	errs := req.Validate()
	if len(errs) > 0 {
		return nil, errors.New(errs[0])
	}

	user, err = s.svc.Update(req, id)
	if err != nil {
		return nil, err
	}

	createdAt := &ts.Timestamp{
		Seconds: user.CreatedAt.Unix(),
	}

	updatedAt := &ts.Timestamp{
		Seconds: user.UpdatedAt.Unix(),
	}

	resp := &usrProto.User{
		Id:        user.ID,
		Name:      user.Name,
		Phone:     user.Phone,
		Email:     user.Email,
		Address:   user.Address,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	return resp, nil
}

func (s *server) Delete(c context.Context, r *usrProto.UserIDReq) (*usrProto.UserIDReq, error) {
	id := r.GetId()
	err := s.svc.Delete(id)
	if err != nil {
		return nil, err
	}

	resp := &usrProto.UserIDReq{
		Id: id,
	}

	return resp, nil
}
