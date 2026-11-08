package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"helpdesk/backend/internal/auth"
	"helpdesk/backend/internal/domain"
)

type AuthService struct {
	users      domain.UserRepository
	jwt        *auth.Manager
	bcryptCost int
}

func NewAuthService(users domain.UserRepository, jwt *auth.Manager, bcryptCost int) *AuthService {
	return &AuthService{users: users, jwt: jwt, bcryptCost: bcryptCost}
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginResult struct {
	Token     string
	ExpiresAt time.Time
	User      domain.User
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (domain.User, error) {
	details := map[string]string{}
	name, msg := validateName(in.Name)
	if msg != "" {
		details["name"] = msg
	}
	email, msg := validateEmail(in.Email)
	if msg != "" {
		details["email"] = msg
	}
	if msg := validatePassword(in.Password); msg != "" {
		details["password"] = msg
	}
	if len(details) > 0 {
		return domain.User{}, domain.Validation(details)
	}

	hash, err := auth.HashPassword(in.Password, s.bcryptCost)
	if err != nil {
		return domain.User{}, domain.Internal(fmt.Errorf("hash password: %w", err))
	}

	user := domain.User{
		Name:         name,
		Email:        email,
		PasswordHash: hash,
		Role:         domain.RoleRequester,
	}
	created, err := s.users.Create(ctx, user)
	if err != nil {
		return domain.User{}, err
	}
	return created, nil
}

func (s *AuthService) Login(ctx context.Context, in LoginInput) (LoginResult, error) {
	email, msg := validateEmail(in.Email)
	if msg != "" {
		// Do not reveal whether the email exists.
		return LoginResult{}, domain.InvalidCredentials()
	}
	if in.Password == "" {
		return LoginResult{}, domain.InvalidCredentials()
	}

	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		var appErr *domain.AppError
		if errors.As(err, &appErr) && appErr.Code == "NOT_FOUND" {
			return LoginResult{}, domain.InvalidCredentials()
		}
		return LoginResult{}, err
	}
	if !auth.VerifyPassword(user.PasswordHash, in.Password) {
		return LoginResult{}, domain.InvalidCredentials()
	}

	token, expiresAt, err := s.jwt.Generate(user.ID, user.Role)
	if err != nil {
		return LoginResult{}, domain.Internal(fmt.Errorf("buat token: %w", err))
	}
	return LoginResult{Token: token, ExpiresAt: expiresAt, User: user}, nil
}

func (s *AuthService) Me(ctx context.Context, userID int64) (domain.User, error) {
	return s.users.GetByID(ctx, userID)
}
