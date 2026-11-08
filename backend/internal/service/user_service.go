package service

import (
	"context"
	"strings"

	"helpdesk/backend/internal/domain"
)

type UserService struct {
	users domain.UserRepository
}

func NewUserService(users domain.UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) List(ctx context.Context, actor domain.User, filter domain.UserFilter) ([]domain.User, int, error) {
	if actor.Role != domain.RoleAdmin {
		return nil, 0, domain.Forbidden("Hanya admin yang dapat melihat daftar pengguna.")
	}
	if filter.Page == 0 {
		filter.Page = 1
	}
	if filter.Page < 1 {
		return nil, 0, domain.Validation(map[string]string{"page": "Parameter page tidak valid."})
	}
	if filter.PerPage == 0 {
		filter.PerPage = 20
	}
	if filter.PerPage < 1 || filter.PerPage > 100 {
		return nil, 0, domain.Validation(map[string]string{"per_page": "per_page harus di antara 1 dan 100."})
	}
	filter.Search = strings.TrimSpace(filter.Search)
	if runeLen(filter.Search) > 100 {
		return nil, 0, domain.Validation(map[string]string{"search": "Pencarian maksimal 100 karakter."})
	}
	if filter.Role != nil && !filter.Role.Valid() {
		return nil, 0, domain.Validation(map[string]string{"role": "Peran tidak valid."})
	}
	return s.users.List(ctx, filter)
}

func (s *UserService) UpdateRole(ctx context.Context, actor domain.User, targetID int64, rawRole string) (domain.User, error) {
	if actor.Role != domain.RoleAdmin {
		return domain.User{}, domain.Forbidden("Hanya admin yang dapat mengubah peran pengguna.")
	}
	role, ok := domain.ParseRole(rawRole)
	if !ok {
		return domain.User{}, domain.Validation(map[string]string{"role": "Peran tidak valid."})
	}

	target, err := s.users.GetByID(ctx, targetID)
	if err != nil {
		return domain.User{}, err
	}
	if target.ID == actor.ID {
		return domain.User{}, domain.SelfRoleChange()
	}
	if target.Role == domain.RoleAdmin && role != domain.RoleAdmin {
		count, err := s.users.CountByRole(ctx, domain.RoleAdmin)
		if err != nil {
			return domain.User{}, err
		}
		if count <= 1 {
			return domain.User{}, domain.LastAdmin()
		}
	}
	return s.users.UpdateRole(ctx, targetID, role)
}
